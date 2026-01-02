package store

import (
	"context"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDatabaseStore_Init(t *testing.T) {
	t.Parallel()

	// Arrange.
	ctrl := gomock.NewController(t)
	mock := mocks.NewMock(ctrl)
	configuration := &config.Configuration{
		DatabaseStore: &config.DatabaseStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{},
		},
	}
	store := NewDatabaseStore(mock.Connection, configuration, mock.Logger)

	// Act.
	store.Init()
}

func TestDatabaseStore_CheckAvailability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantError  bool
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name:      "WHEN ping error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.Connection.EXPECT().Ping(gomock.Any()).Return(assert.AnError)
			},
		},
		{
			name: "WHEN no error THEN ok",
			hookBefore: func(mock *mocks.Mock) {
				mock.Connection.EXPECT().Ping(gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)
			configuration := &config.Configuration{
				DatabaseStore: &config.DatabaseStoreConfiguration{
					KeyStoreConfiguration: config.KeyStoreConfiguration{},
				},
			}
			store := NewDatabaseStore(mock.Connection, configuration, mock.Logger)

			// Act.
			err := store.CheckAvailability(context.Background())

			// Assert.
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseStore_Load(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}

	tests := []struct {
		name       string
		args       *args
		want       *domain.URLItem
		wantError  bool
		hookBefore func(mock *mocks.Mock, args *args)
	}{
		{
			name: "WHEN connection error THEN error",
			args: &args{
				key: "foo",
			},
			wantError: true,
			hookBefore: func(mock *mocks.Mock, args *args) {
				mock.Connection.EXPECT().
					QueryRows(
						gomock.Any(),
						"select original_url, user_id, is_deleted from urls where url_key = $1",
						args.key,
					).
					Return(nil, assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock, tt.args)
			configuration := &config.Configuration{
				DatabaseStore: &config.DatabaseStoreConfiguration{
					KeyStoreConfiguration: config.KeyStoreConfiguration{},
				},
			}
			store := NewDatabaseStore(mock.Connection, configuration, mock.Logger)

			// Act.
			item, err := store.Load(context.Background(), tt.args.key)
			require.Equal(t, tt.want, item)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseStore_DeleteBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantError  bool
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name:      "WHEN conection error THEN error",
			wantError: true,
			hookBefore: func(mock *mocks.Mock) {
				mock.Connection.EXPECT().
					Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
		},
		{
			name: "WHEN no error THEN ok",
			hookBefore: func(mock *mocks.Mock) {
				mock.Connection.EXPECT().
					Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)
			configuration := &config.Configuration{
				DatabaseStore: &config.DatabaseStoreConfiguration{
					KeyStoreConfiguration: config.KeyStoreConfiguration{},
				},
			}
			store := NewDatabaseStore(mock.Connection, configuration, mock.Logger)

			// Act.
			err := store.DeleteBatch(context.Background(), []string{"foo"}, uuid.New())
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
