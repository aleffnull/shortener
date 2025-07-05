package store

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/models"
)

type MemoryStore struct {
	coldStore     ColdStore
	configuration *config.MemoryStoreConfiguration
	logger        logger.Logger
	storeMap      map[string]string
	mutex         sync.RWMutex
}

var _ Store = (*MemoryStore)(nil)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func NewMemoryStore(coldStore ColdStore, configuration *config.Configuration, logger logger.Logger) Store {
	return &MemoryStore{
		coldStore:     coldStore,
		configuration: configuration.MemoryStore,
		logger:        logger,
		storeMap:      make(map[string]string),
	}
}

func (s *MemoryStore) Init() error {
	entries, err := s.coldStore.LoadAll()
	if err != nil {
		return fmt.Errorf("InitStorage, coldStorage.LoadAll failed: %w", err)
	}

	// Called only during startup, so no need for mutex locking.
	for _, entry := range entries {
		s.storeMap[entry.Key] = entry.Value
	}

	s.logger.Infof("Loaded %v entries from cold storage", len(entries))

	return nil
}

func (s *MemoryStore) Shutdown() {
	// Do nothing.
}

func (s *MemoryStore) CheckAvailability(context.Context) error {
	return nil
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

	// Save to hot store.
	s.storeMap[key] = value

	// Save to cold store.
	coldStoreEntry := &models.ColdStoreEntry{
		Key:   key,
		Value: value,
	}
	err = s.coldStore.Save(coldStoreEntry)
	if err != nil {
		return key, fmt.Errorf("saving to cold storage failed: %w", err)
	}

	return key, nil
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
