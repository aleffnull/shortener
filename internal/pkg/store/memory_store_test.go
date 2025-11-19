package store

// import (
// 	"context"
// 	"fmt"
// 	"testing"

// 	"github.com/aleffnull/shortener/internal/config"
// 	"github.com/aleffnull/shortener/internal/pkg/testutils"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/require"
// 	"go.uber.org/mock/gomock"
// )

// func TestMemoryStore_Load_UnknownKey(t *testing.T) {
// 	// Arrange.
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	mock := testutils.NewMock(ctrl)
// 	configuration := &config.Configuration{
// 		MemoryStore: &config.MemoryStoreConfiguration{
// 			KeyStoreConfiguration: config.KeyStoreConfiguration{
// 				KeyLength:        8,
// 				KeyMaxLength:     100,
// 				KeyMaxIterations: 10,
// 			},
// 		},
// 	}
// 	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

// 	// Act.
// 	item, err := store.Load(ctx, "foo")

// 	// Assert.
// 	require.Nil(t, item)
// 	require.NoError(t, err)
// }

// func TestMemoryStore_SaveAndLoad(t *testing.T) {
// 	// Arrange.
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	mock := testutils.NewMock(ctrl)
// 	mock.ColdStore.EXPECT().Save(gomock.Any()).Return(nil)
// 	configuration := &config.Configuration{
// 		MemoryStore: &config.MemoryStoreConfiguration{
// 			KeyStoreConfiguration: config.KeyStoreConfiguration{
// 				KeyLength:        8,
// 				KeyMaxLength:     100,
// 				KeyMaxIterations: 10,
// 			},
// 		},
// 	}
// 	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

// 	// Act.
// 	key, err := store.Save(ctx, "foo", uuid.New())
// 	require.NoError(t, err)
// 	item, err := store.Load(ctx, key)
// 	require.NoError(t, err)

// 	// Assert.
// 	require.NotEmpty(t, key)
// 	require.NotNil(t, item)
// 	require.Equal(t, "foo", item.URL)
// }

// func TestMemoryStore_InitAndLoad(t *testing.T) {
// 	// Arrange.
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	mock := testutils.NewMock(ctrl)
// 	mock.ColdStore.EXPECT().LoadAll().Return([]*ColdStoreEntry{
// 		{
// 			Key:   "key",
// 			Value: "foo",
// 		},
// 	}, nil)
// 	mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
// 	configuration := &config.Configuration{
// 		MemoryStore: &config.MemoryStoreConfiguration{
// 			KeyStoreConfiguration: config.KeyStoreConfiguration{
// 				KeyLength:        8,
// 				KeyMaxLength:     100,
// 				KeyMaxIterations: 10,
// 			},
// 		},
// 	}
// 	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

// 	// Act.
// 	err := store.Init()
// 	require.NoError(t, err)
// 	item, err := store.Load(ctx, "key")
// 	require.NoError(t, err)

// 	// Assert.
// 	require.NotNil(t, item)
// 	require.Equal(t, "foo", item.URL)
// }

// func TestMemoryStore_Save_NotUniqueKey(t *testing.T) {
// 	// Arrange.
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	mock := testutils.NewMock(ctrl)
// 	mock.ColdStore.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()
// 	configuration := &config.Configuration{
// 		MemoryStore: &config.MemoryStoreConfiguration{
// 			KeyStoreConfiguration: config.KeyStoreConfiguration{
// 				KeyLength:        1,
// 				KeyMaxLength:     1,
// 				KeyMaxIterations: 1,
// 			},
// 		},
// 	}
// 	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

// 	// Act.
// 	var err error
// 	for range 100 {
// 		_, err = store.Save(ctx, "foo", uuid.New())
// 		if err != nil {
// 			break
// 		}
// 	}

// 	// Assert.
// 	require.Error(t, err)
// }

// func TestMemoryStore_Save_KeyLengthIsDoubled(t *testing.T) {
// 	// Arrange.
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	mock := testutils.NewMock(ctrl)
// 	mock.ColdStore.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()
// 	configuration := &config.Configuration{
// 		MemoryStore: &config.MemoryStoreConfiguration{
// 			KeyStoreConfiguration: config.KeyStoreConfiguration{
// 				KeyLength:        1,
// 				KeyMaxLength:     10,
// 				KeyMaxIterations: 1,
// 			},
// 		},
// 	}
// 	store := NewMemoryStore(mock.ColdStore, configuration, mock.Logger)

// 	// Act.
// 	var key string
// 	var err error
// 	for i := range 100 {
// 		key, err = store.Save(ctx, fmt.Sprintf("foo%v", i), uuid.New())
// 		require.NoError(t, err)
// 		if len(key) > 1 {
// 			break
// 		}
// 	}

// 	// Assert.
// 	require.Len(t, key, 2)
// }
