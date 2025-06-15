package app

import (
	"fmt"
	"net/url"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/store"
	"github.com/aleffnull/shortener/models"
)

type ShortenerApp struct {
	storage       store.Store
	configuration *config.Configuration
}

var _ App = (*ShortenerApp)(nil)

func NewShortenerApp(storage store.Store, configuration *config.Configuration) App {
	return &ShortenerApp{
		storage:       storage,
		configuration: configuration,
	}
}

func (shortener *ShortenerApp) GetURL(key string) (string, bool) {
	url, ok := shortener.storage.Load(key)
	return url, ok
}

func (shortener *ShortenerApp) ShortenURL(request *models.ShortenRequest) (*models.ShortenResponse, error) {
	key, err := shortener.storage.Save(request.URL)
	if err != nil {
		return nil, fmt.Errorf("saving to storage failed: %w", err)
	}

	shortURL, err := url.JoinPath(shortener.configuration.BaseURL, key)
	if err != nil {
		return nil, fmt.Errorf("URL joining failed: %w", err)
	}

	return &models.ShortenResponse{
		Result: shortURL,
	}, nil
}
