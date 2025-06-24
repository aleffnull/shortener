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

func (sa *ShortenerApp) Init(ctx context.Context) error {
	entries, err := sa.coldStorage.LoadAll()
	if err != nil {
		return fmt.Errorf("Init, coldStorage.LoadAll failed: %w", err)
	}

	for _, entry := range entries {
		sa.storage.PreSave(entry.Key, entry.Value)
	}

	log := logger.LoggerFromContext(ctx)
	log.Infof("Loaded %v entries from cold storage", len(entries))

	return nil
}

func (sa *ShortenerApp) GetURL(key string) (string, bool) {
	url, ok := sa.storage.Load(key)
	return url, ok
}

func (sa *ShortenerApp) ShortenURL(request *models.ShortenRequest) (*models.ShortenResponse, error) {
	longURL := request.URL
	key, err := sa.storage.Save(longURL)
	if err != nil {
		return nil, fmt.Errorf("saving to storage failed: %w", err)
	}

	coldStoreEntry := &store.ColdStoreEntry{
		Key:   key,
		Value: longURL,
	}
	err = sa.coldStorage.Save(coldStoreEntry)
	if err != nil {
		return nil, fmt.Errorf("saving to cold storage failed: %w", err)
	}

	shortURL, err := url.JoinPath(sa.configuration.BaseURL, key)
	if err != nil {
		return nil, fmt.Errorf("URL joining failed: %w", err)
	}

	return &models.ShortenResponse{
		Result: shortURL,
	}, nil
}
