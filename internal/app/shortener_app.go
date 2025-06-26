package app

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/store"
	"github.com/aleffnull/shortener/models"
)

type ShortenerApp struct {
	storage       store.Store
	coldStorage   store.ColdStore
	configuration *config.Configuration
}

var _ App = (*ShortenerApp)(nil)

func NewShortenerApp(storage store.Store, coldStorage store.ColdStore, configuration *config.Configuration) App {
	return &ShortenerApp{
		storage:       storage,
		coldStorage:   coldStorage,
		configuration: configuration,
	}
}

func (s *ShortenerApp) Init(ctx context.Context) error {
	entries, err := s.coldStorage.LoadAll()
	if err != nil {
		return fmt.Errorf("Init, coldStorage.LoadAll failed: %w", err)
	}

	for _, entry := range entries {
		s.storage.PreSave(entry.Key, entry.Value)
	}

	log := logger.LoggerFromContext(ctx)
	log.Infof("Loaded %v entries from cold storage", len(entries))

	return nil
}

func (s *ShortenerApp) GetURL(key string) (string, bool) {
	url, ok := s.storage.Load(key)
	return url, ok
}

func (s *ShortenerApp) ShortenURL(request *models.ShortenRequest) (*models.ShortenResponse, error) {
	longURL := request.URL
	key, err := s.storage.Save(longURL)
	if err != nil {
		return nil, fmt.Errorf("saving to storage failed: %w", err)
	}

	coldStoreEntry := &store.ColdStoreEntry{
		Key:   key,
		Value: longURL,
	}
	err = s.coldStorage.Save(coldStoreEntry)
	if err != nil {
		return nil, fmt.Errorf("saving to cold storage failed: %w", err)
	}

	shortURL, err := url.JoinPath(s.configuration.BaseURL, key)
	if err != nil {
		return nil, fmt.Errorf("URL joining failed: %w", err)
	}

	return &models.ShortenResponse{
		Result: shortURL,
	}, nil
}
