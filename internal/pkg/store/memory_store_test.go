package store

import (
	"context"
	"errors"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMemoryStore_Init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantError  bool
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name:      "WHEN cold store error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.ColdStore.EXPECT().LoadAll().Return(nil, assert.AnError)
			},
		},
		{
			name: "WHEN no errors THEN ok",
			hookBefore: func(mock *mocks.Mock) {
				mock.ColdStore.EXPECT().LoadAll().Return(nil, nil)
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)

			configuration := &config.Configuration{
				MemoryStore: &config.MemoryStoreConfiguration{},
			}
			store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)
			err := store.Init()

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMemoryStore_CheckAvailability(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)

	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{},
	}
	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

	err := store.CheckAvailability(context.Background())
	require.NoError(t, err)
}

func TestMemoryStore_Load(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}

	tests := []struct {
		name       string
		args       *args
		want       *domain.URLItem
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name: "WHEN unknown key THEN nil",
			args: &args{
				key: "foo",
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.ColdStore.EXPECT().LoadAll().Return([]*domain.ColdStoreEntry{}, nil)
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN has key THEN ok",
			args: &args{
				key: "foo",
			},
			want: &domain.URLItem{
				URL: "http://foo.bar",
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.ColdStore.EXPECT().LoadAll().Return([]*domain.ColdStoreEntry{
					{
						Key:   "foo",
						Value: "http://foo.bar",
					},
				}, nil)
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)

			configuration := &config.Configuration{
				MemoryStore: &config.MemoryStoreConfiguration{},
			}
			store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)
			err := store.Init()
			require.NoError(t, err)

			got, err := store.Load(context.Background(), tt.args.key)
			require.Equal(t, tt.want, got)
			require.NoError(t, err)
		})
	}
}

func TestMemoryStore_LoadAllByUserID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)

	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{},
	}
	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

	items, err := store.LoadAllByUserID(context.Background(), uuid.New())
	require.Nil(t, items)
	require.NoError(t, err)
}

func TestMemoryStore_Save(t *testing.T) {
	t.Parallel()

	type args struct {
		key   string
		value string
	}

	defaultConfiguration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{
				KeyLength:        8,
				KeyMaxLength:     100,
				KeyMaxIterations: 10,
			},
		},
	}

	tests := []struct {
		name        string
		args        *args
		hookBefore  func(mock *mocks.Mock, args *args) *config.Configuration
		checkResult func(key string, err error, args *args)
	}{
		{
			name: "GIVEN already existing value WHEN no errors THEN duplicate error",
			args: &args{
				key:   "foo",
				value: "http://foo.bar",
			},
			hookBefore: func(mock *mocks.Mock, args *args) *config.Configuration {
				mock.ColdStore.EXPECT().LoadAll().Return([]*domain.ColdStoreEntry{
					{
						Key:   args.key,
						Value: args.value,
					},
				}, nil)
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				return defaultConfiguration
			},
			checkResult: func(key string, err error, args *args) {
				var duplicateURLError *DuplicateURLError
				require.True(t, errors.As(err, &duplicateURLError))
				require.Equal(t, args.key, duplicateURLError.Key)
				require.Equal(t, args.value, duplicateURLError.URL)
			},
		},
		{
			name: "WHEN save to cold store error THEN error",
			args: &args{
				value: "http://foo.bar",
			},
			hookBefore: func(mock *mocks.Mock, args *args) *config.Configuration {
				// Init.
				mock.ColdStore.EXPECT().LoadAll().Return([]*domain.ColdStoreEntry{}, nil)
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				// Save.
				mock.ColdStore.EXPECT().Save(gomock.Any()).DoAndReturn(func(entry *domain.ColdStoreEntry) error {
					require.GreaterOrEqual(t, len(entry.Key), 1)
					require.Equal(t, args.value, entry.Value)
					return assert.AnError
				})
				return defaultConfiguration
			},
			checkResult: func(key string, err error, _ *args) {
				require.Error(t, err)
			},
		},
		{
			name: "GIVEN all short keys already exist WHEN no errors THEN saved with doubled key",
			args: &args{
				value: "http://foo.bar",
			},
			hookBefore: func(mock *mocks.Mock, args *args) *config.Configuration {
				// Init.
				entries := lo.Map([]rune(alphabet), func(ch rune, _ int) *domain.ColdStoreEntry {
					return &domain.ColdStoreEntry{
						Key:   string(ch),
						Value: string(ch),
					}
				})
				mock.ColdStore.EXPECT().LoadAll().Return(entries, nil)
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				// Save.
				mock.ColdStore.EXPECT().Save(gomock.Any()).DoAndReturn(func(entry *domain.ColdStoreEntry) error {
					require.GreaterOrEqual(t, len(entry.Key), 1)
					require.Equal(t, args.value, entry.Value)
					return nil
				})
				return &config.Configuration{
					MemoryStore: &config.MemoryStoreConfiguration{
						KeyStoreConfiguration: config.KeyStoreConfiguration{
							KeyLength:        1,
							KeyMaxLength:     2,
							KeyMaxIterations: 1,
						},
					},
				}
			},
			checkResult: func(key string, err error, _ *args) {
				require.Len(t, key, 2)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)

			configuration := tt.hookBefore(mock, tt.args)
			store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)
			{
				err := store.Init()
				require.NoError(t, err)
			}

			key, err := store.Save(context.Background(), tt.args.value, uuid.New())
			tt.checkResult(key, err, tt.args)
		})
	}
}

func TestMemoryStore_SaveBatch(t *testing.T) {
	t.Parallel()

	type args struct {
		items []*domain.BatchRequestItem
	}

	defaultConfiguration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{
				KeyLength:        8,
				KeyMaxLength:     100,
				KeyMaxIterations: 10,
			},
		},
	}
	correlationID := uuid.NewString()

	tests := []struct {
		name        string
		args        *args
		hookBefore  func(mock *mocks.Mock, args *args)
		checkResult func(items []*domain.BatchResponseItem, err error)
	}{
		{
			name: "WHEN save error THEN error",
			args: &args{
				items: []*domain.BatchRequestItem{
					{
						OriginalURL: "http://foo.bar",
					},
				},
			},
			hookBefore: func(mock *mocks.Mock, args *args) {
				mock.ColdStore.EXPECT().Save(gomock.Any()).DoAndReturn(func(entry *domain.ColdStoreEntry) error {
					require.GreaterOrEqual(t, len(entry.Key), defaultConfiguration.MemoryStore.KeyLength)
					require.Equal(t, args.items[0].OriginalURL, entry.Value)
					return assert.AnError
				})
			},
			checkResult: func(items []*domain.BatchResponseItem, err error) {
				require.Nil(t, items)
				require.Error(t, err)
			},
		},
		{
			name: "WHEN no error THEN ok",
			args: &args{
				items: []*domain.BatchRequestItem{
					{
						CorrelationID: correlationID,
						OriginalURL:   "http://foo.bar",
					},
				},
			},
			hookBefore: func(mock *mocks.Mock, args *args) {
				mock.ColdStore.EXPECT().Save(gomock.Any()).DoAndReturn(func(entry *domain.ColdStoreEntry) error {
					require.GreaterOrEqual(t, len(entry.Key), defaultConfiguration.MemoryStore.KeyLength)
					require.Equal(t, args.items[0].OriginalURL, entry.Value)
					return nil
				})
			},
			checkResult: func(items []*domain.BatchResponseItem, err error) {
				require.Len(t, items, 1)
				require.Equal(t, correlationID, items[0].CorrelationID)
				require.GreaterOrEqual(t, len(items[0].Key), defaultConfiguration.MemoryStore.KeyLength)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock, tt.args)

			store := NewMemoryStore(mock.ColdStore, defaultConfiguration, mock.Logger)
			items, err := store.SaveBatch(context.Background(), tt.args.items, uuid.New())
			tt.checkResult(items, err)
		})
	}
}

func TestMemoryStore_DeleteBatch(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)

	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{},
	}
	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

	err := store.DeleteBatch(context.Background(), []string{}, uuid.New())
	require.NoError(t, err)
}

func TestMemoryStore_GetStatistics(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)

	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{},
	}
	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

	// Act-assert.
	urlsCount, usersCount, err := store.GetStatistics(context.Background())
	require.Zero(t, urlsCount)
	require.Zero(t, usersCount)
	require.NoError(t, err)
}
