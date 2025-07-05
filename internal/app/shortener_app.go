package app

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/store"
	"github.com/aleffnull/shortener/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ShortenerApp struct {
	storage       store.Store
	logger        logger.Logger
	configuration *config.Configuration
}

var _ App = (*ShortenerApp)(nil)

func NewShortenerApp(storage store.Store, logger logger.Logger, configuration *config.Configuration) App {
	return &ShortenerApp{
		storage:       storage,
		logger:        logger,
		configuration: configuration,
	}
}

func (s *ShortenerApp) Init() error {
	err := s.storage.Init()
	if err != nil {
		return fmt.Errorf("Init, initStorage failed: %w", err)
	}

	return nil
}

func (s *ShortenerApp) Shutdown() {
	s.storage.Shutdown()
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

	shortURL, err := url.JoinPath(s.configuration.BaseURL, key)
	if err != nil {
		return nil, fmt.Errorf("URL joining failed: %w", err)
	}

	return &models.ShortenResponse{
		Result: shortURL,
	}, nil
}

func (s *ShortenerApp) CheckStore(ctx context.Context) error {
	err := s.storage.CheckAvailability(ctx)
	if err != nil {
		return fmt.Errorf("store is not available: %w", err)
	}

	return nil
}
