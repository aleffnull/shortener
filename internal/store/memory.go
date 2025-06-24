package store

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"

	"github.com/aleffnull/shortener/internal/config"
)

type MemoryStore struct {
	storeMap      map[string]string
	configuration *config.MemoryStoreConfiguration
	mutex         sync.RWMutex
}

var _ Store = (*MemoryStore)(nil)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func NewMemoryStore(configuration *config.Configuration) Store {
	return &MemoryStore{
		storeMap:      make(map[string]string),
		configuration: configuration.MemoryStore,
	}
}

func (ms *MemoryStore) Load(key string) (string, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	value, ok := ms.storeMap[key]
	return value, ok
}

func (ms *MemoryStore) Save(value string) (string, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	key, err := ms.getUniqueKey()
	if err != nil {
		return "", fmt.Errorf("failed to save value: %w", err)
	}

	ms.storeMap[key] = value
	return key, nil
}

func (ms *MemoryStore) PreSave(key, value string) {
	// Called only in main goroutine, so no need for mutex locking.
	ms.storeMap[key] = value
}

func (ms *MemoryStore) getUniqueKey() (string, error) {
	length := ms.configuration.KeyLength
	i := 0

	for length <= ms.configuration.KeyMaxLength {
		key := randomString(length)
		_, exists := ms.storeMap[key]
		if !exists {
			return key, nil
		}

		i++
		if i >= ms.configuration.KeyMaxIterations {
			length *= 2
			i = 0
		}
	}

	return "", errors.New("failed to generate unique key")
}

func randomString(length int) string {
	var arr = make([]byte, length)
	for i := range arr {
		arr[i] = alphabet[rand.IntN(len(alphabet))]
	}

	return string(arr)
}
