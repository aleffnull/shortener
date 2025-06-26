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

func (s *MemoryStore) Load(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.storeMap[key]
	return value, ok
}

func (s *MemoryStore) Save(value string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key, err := s.getUniqueKey()
	if err != nil {
		return "", fmt.Errorf("failed to save value: %w", err)
	}

	s.storeMap[key] = value
	return key, nil
}

func (s *MemoryStore) PreSave(key, value string) {
	// Called only in main goroutine, so no need for mutex locking.
	s.storeMap[key] = value
}

func (s *MemoryStore) getUniqueKey() (string, error) {
	length := s.configuration.KeyLength
	i := 0

	for length <= s.configuration.KeyMaxLength {
		key := randomString(length)
		_, exists := s.storeMap[key]
		if !exists {
			return key, nil
		}

		i++
		if i >= s.configuration.KeyMaxIterations {
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
