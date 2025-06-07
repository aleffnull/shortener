package store

import (
	"math/rand/v2"
	"sync"
)

type MemoryStore struct {
	storeMap map[string]string
	mutex    sync.RWMutex
}

var _ Store = &MemoryStore{}

const (
	alphabet     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lettersCount = 8
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

func (ms *MemoryStore) Save(value string) string {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	var key string
	exists := true
	for exists {
		key = randomString(lettersCount)
		_, exists = ms.storeMap[key]
	}

	ms.storeMap[key] = value
	return key
}

func randomString(length int) string {
	var arr = make([]byte, length)
	for i := range arr {
		arr[i] = alphabet[rand.IntN(len(alphabet))]
	}

	return string(arr)
}
