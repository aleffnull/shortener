package store

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
)

type MemoryStore struct {
	storeMap map[string]string
	mutex    sync.RWMutex
}

var _ Store = &MemoryStore{}

const (
	alphabet         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	keyLength        = 8
	keyMaxLength     = 100
	keyMaxIterations = 10
)

func NewMemoryStore() Store {
	return &MemoryStore{
		storeMap: make(map[string]string),
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

func (ms *MemoryStore) Set(key, value string) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.storeMap[key] = value
}

func randomString(length int) string {
	var arr = make([]byte, length)
	for i := range arr {
		arr[i] = alphabet[rand.IntN(len(alphabet))]
	}

	return string(arr)
}

func (ms *MemoryStore) getUniqueKey() (string, error) {
	length := keyLength
	i := 0

	for length < keyMaxLength {
		key := randomString(length)
		_, exists := ms.storeMap[key]
		if !exists {
			return key, nil
		}

		i++
		if i >= keyMaxIterations {
			length *= 2
			i = 0
		}
	}

	return "", errors.New("failed to generate unique key")
}
