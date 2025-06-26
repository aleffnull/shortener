package store

import (
	"testing"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_Load_UnknownKey(t *testing.T) {
	// Arrange.
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyLength:        8,
			KeyMaxLength:     100,
			KeyMaxIterations: 10,
		},
	}
	store := NewMemoryStore(configuration)

	// Act.
	value, ok := store.Load("foo")

	// Assert.
	require.Empty(t, value)
	require.False(t, ok)
}

func TestMemoryStore_SaveAndLoad(t *testing.T) {
	// Arrange.
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyLength:        8,
			KeyMaxLength:     100,
			KeyMaxIterations: 10,
		},
	}
	store := NewMemoryStore(configuration)

	// Act.
	key, err := store.Save("foo")
	value, ok := store.Load(key)

	// Assert.
	require.NoError(t, err)
	require.NotEmpty(t, key)
	require.True(t, ok)
	require.Equal(t, "foo", value)
}

func TestMemoryStore_PreSaveAndLoad(t *testing.T) {
	// Arrange.
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyLength:        8,
			KeyMaxLength:     100,
			KeyMaxIterations: 10,
		},
	}
	store := NewMemoryStore(configuration)

	// Act.
	store.PreSave("key", "foo")
	value, ok := store.Load("key")

	// Assert.
	require.True(t, ok)
	require.Equal(t, "foo", value)
}

func TestMemoryStore_Save_NotUniqueKey(t *testing.T) {
	// Arrange.
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyLength:        1,
			KeyMaxLength:     1,
			KeyMaxIterations: 1,
		},
	}
	store := NewMemoryStore(configuration)

	// Act.
	var err error
	for range 100 {
		_, err = store.Save("foo")
		if err != nil {
			break
		}
	}

	// Assert.
	require.Error(t, err)
}

func TestMemoryStore_Save_KeyLengthIsDoubled(t *testing.T) {
	// Arrange.
	configuration := &config.Configuration{
		MemoryStore: &config.MemoryStoreConfiguration{
			KeyLength:        1,
			KeyMaxLength:     10,
			KeyMaxIterations: 1,
		},
	}
	store := NewMemoryStore(configuration)

	// Act.
	var key string
	var err error
	for range 100 {
		key, err = store.Save("foo")
		require.NoError(t, err)
		if len(key) > 1 {
			break
		}
	}

	// Assert.
	require.Len(t, key, 2)
}
