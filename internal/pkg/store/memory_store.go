package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/models"
)

type MemoryStore struct {
	keyStore
	coldStore     ColdStore
	configuration *config.MemoryStoreConfiguration
	logger        logger.Logger
	storeMap      map[string]string
	mutex         sync.RWMutex
}

var _ Store = (*MemoryStore)(nil)

func NewMemoryStore(coldStore ColdStore, configuration *config.Configuration, logger logger.Logger) Store {
	store := &MemoryStore{
		keyStore: keyStore{
			configuration: &configuration.MemoryStore.KeyStoreConfiguration,
		},
		coldStore:     coldStore,
		configuration: configuration.MemoryStore,
		logger:        logger,
		storeMap:      make(map[string]string),
	}

	return store
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

func (s *MemoryStore) Load(_ context.Context, key string) (string, bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.storeMap[key]
	return value, ok, nil
}

func (s *MemoryStore) Save(ctx context.Context, value string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.saveValue(ctx, value)
}

func (s *MemoryStore) SaveBatch(ctx context.Context, requestItems []*models.BatchRequestItem) ([]*models.BatchResponseItem, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	responseItems := make([]*models.BatchResponseItem, 0, len(requestItems))
	for _, requestItem := range requestItems {
		key, err := s.saveValue(ctx, requestItem.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("SaveBatch, saveValue failed: %w", err)
		}

		responseItems = append(responseItems, &models.BatchResponseItem{
			CorelationID: requestItem.CorelationID,
			Key:          key,
		})
	}

	return responseItems, nil
}

func (s *MemoryStore) saveValue(ctx context.Context, value string) (string, error) {
	// Save to hot store.
	key, err := s.saveWithUniqueKey(ctx, value, s.saver)
	if err != nil {
		return "", fmt.Errorf("MemoryStore.Save, saveWithUniqueKey failed: %w", err)
	}

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

func (s *MemoryStore) saver(_ context.Context, key, value string) (bool, error) {
	_, exists := s.storeMap[key]
	if exists {
		return true, nil
	}

	s.storeMap[key] = value
	return false, nil
}
