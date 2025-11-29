package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/logger"
)

type MemoryStore struct {
	keyStore
	coldStore     ColdStore
	configuration *config.MemoryStoreConfiguration
	logger        logger.Logger
	keyToValueMap map[string]string
	valueToKeyMap map[string]string
	mutex         *sync.RWMutex
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
		keyToValueMap: make(map[string]string),
		valueToKeyMap: make(map[string]string),
		mutex:         &sync.RWMutex{},
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
		s.keyToValueMap[entry.Key] = entry.Value
		s.valueToKeyMap[entry.Value] = entry.Key
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

func (s *MemoryStore) Load(_ context.Context, key string) (*domain.URLItem, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, ok := s.keyToValueMap[key]
	if !ok {
		return nil, nil
	}

	return &domain.URLItem{
		URL: value,
	}, nil
}

func (s *MemoryStore) LoadAllByUserID(context.Context, uuid.UUID) ([]*domain.KeyOriginalURLItem, error) {
	return nil, nil
}

func (s *MemoryStore) Save(ctx context.Context, value string, _ uuid.UUID) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.saveValue(ctx, value)
}

func (s *MemoryStore) SaveBatch(ctx context.Context, requestItems []*domain.BatchRequestItem, _ uuid.UUID) ([]*domain.BatchResponseItem, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	responseItems := make([]*domain.BatchResponseItem, 0, len(requestItems))
	for _, requestItem := range requestItems {
		key, err := s.saveValue(ctx, requestItem.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("SaveBatch, saveValue failed: %w", err)
		}

		responseItems = append(responseItems, &domain.BatchResponseItem{
			CorelationID: requestItem.CorelationID,
			Key:          key,
		})
	}

	return responseItems, nil
}

func (s *MemoryStore) DeleteBatch(context.Context, []string, uuid.UUID) error {
	return nil
}

func (s *MemoryStore) saveValue(ctx context.Context, value string) (string, error) {
	// Save to hot store.
	key, err := s.saveWithUniqueKey(ctx, value, s.saver)
	if err != nil {
		return "", fmt.Errorf("MemoryStore.Save, saveWithUniqueKey failed: %w", err)
	}

	// Save to cold store.
	coldStoreEntry := &domain.ColdStoreEntry{
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
	if _, exists := s.keyToValueMap[key]; exists {
		return true, nil
	}

	if existingKey, ok := s.valueToKeyMap[value]; ok {
		return false, NewDuplicateURLError(existingKey, value)
	}

	s.keyToValueMap[key] = value
	s.valueToKeyMap[value] = key
	return false, nil
}
