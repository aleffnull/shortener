package store

import (
	"context"
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/models"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mock struct {
	coldStore *mocks.MockColdStore
	logger    *mocks.MockLogger
}

func newMock(ctrl *gomock.Controller) *mock {
	return &mock{
		coldStore: mocks.NewMockColdStore(ctrl),
		logger:    mocks.NewMockLogger(ctrl),
	}
}

func TestMemoryStore_Load_UnknownKey(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := newMock(ctrl)
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{
				KeyLength:        8,
				KeyMaxLength:     100,
				KeyMaxIterations: 10,
			},
		},
	}
	store := NewMemoryStore(mock.coldStore, configuration, mock.logger)

	// Act.
	value, ok, err := store.Load(ctx, "foo")

	// Assert.
	require.Empty(t, value)
	require.False(t, ok)
	require.NoError(t, err)
}

func TestMemoryStore_SaveAndLoad(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := newMock(ctrl)
	mock.coldStore.EXPECT().Save(gomock.Any()).Return(nil)
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{
				KeyLength:        8,
				KeyMaxLength:     100,
				KeyMaxIterations: 10,
			},
		},
	}
	store := NewMemoryStore(mock.coldStore, configuration, mock.logger)

	// Act.
	key, err := store.Save(ctx, "foo")
	require.NoError(t, err)
	value, ok, err := store.Load(ctx, key)
	require.NoError(t, err)

	// Assert.
	require.NotEmpty(t, key)
	require.True(t, ok)
	require.Equal(t, "foo", value)
}

func TestMemoryStore_InitAndLoad(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := newMock(ctrl)
	mock.coldStore.EXPECT().LoadAll().Return([]*models.ColdStoreEntry{
		{
			Key:   "key",
			Value: "foo",
		},
	}, nil)
	mock.logger.EXPECT().Infof(gomock.Any(), gomock.Any())
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{
				KeyLength:        8,
				KeyMaxLength:     100,
				KeyMaxIterations: 10,
			},
		},
	}
	store := NewMemoryStore(mock.coldStore, configuration, mock.logger)

	// Act.
	err := store.Init()
	require.NoError(t, err)
	value, ok, err := store.Load(ctx, "key")
	require.NoError(t, err)

	// Assert.
	require.True(t, ok)
	require.Equal(t, "foo", value)
}

func TestMemoryStore_Save_NotUniqueKey(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := newMock(ctrl)
	mock.coldStore.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{
				KeyLength:        1,
				KeyMaxLength:     1,
				KeyMaxIterations: 1,
			},
		},
	}
	store := NewMemoryStore(mock.coldStore, configuration, mock.logger)

	// Act.
	var err error
	for range 100 {
		_, err = store.Save(ctx, "foo")
		if err != nil {
			break
		}
	}

	// Assert.
	require.Error(t, err)
}

func TestMemoryStore_Save_KeyLengthIsDoubled(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := newMock(ctrl)
	mock.coldStore.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyStoreConfiguration: config.KeyStoreConfiguration{
				KeyLength:        1,
				KeyMaxLength:     10,
				KeyMaxIterations: 1,
			},
		},
	}
	store := NewMemoryStore(mock.coldStore, configuration, mock.logger)

	// Act.
	var key string
	var err error
	for range 100 {
		key, err = store.Save(ctx, "foo")
		require.NoError(t, err)
		if len(key) > 1 {
			break
		}
	}

	// Assert.
	require.Len(t, key, 2)
}
