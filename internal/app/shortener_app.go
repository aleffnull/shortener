package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/aleffnull/shortener/internal/config"
	pkg_errors "github.com/aleffnull/shortener/internal/pkg/errors"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	pkg_models "github.com/aleffnull/shortener/internal/pkg/models"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"github.com/aleffnull/shortener/models"
	"github.com/samber/lo"
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

func (s *ShortenerApp) GetURL(ctx context.Context, key string) (string, bool, error) {
	return s.storage.Load(ctx, key)
}

func (s *ShortenerApp) ShortenURL(ctx context.Context, request *models.ShortenRequest) (*models.ShortenResponse, error) {
	longURL := request.URL
	key, err := s.storage.Save(ctx, longURL)

	isDuplicate := false
	if err != nil {
		var duplicateURLError *pkg_errors.DuplicateURLError
		if errors.As(err, &duplicateURLError) {
			s.logger.Infof("Duplicate error: %v", duplicateURLError)
			key = duplicateURLError.Key
			isDuplicate = true
		} else {
			return nil, fmt.Errorf("ShortenURL, storage.Save failed: %w", err)
		}
	}

	shortURL, err := url.JoinPath(s.configuration.BaseURL, key)
	if err != nil {
		return nil, fmt.Errorf("ShortenURL, url.JoinPath: %w", err)
	}

	return &models.ShortenResponse{
		Result:      shortURL,
		IsDuplicate: isDuplicate,
	}, nil
}

func (s *ShortenerApp) ShortenURLBatch(ctx context.Context, requestItems []*models.ShortenBatchRequestItem) ([]*models.ShortenBatchResponseItem, error) {
	if len(requestItems) == 0 {
		return []*models.ShortenBatchResponseItem{}, nil
	}

	requestModels := lo.Map(requestItems, func(item *models.ShortenBatchRequestItem, _ int) *pkg_models.BatchRequestItem {
		return &pkg_models.BatchRequestItem{
			CorelationID: item.CorelationID,
			OriginalURL:  item.OriginalURL,
		}
	})
	responseModels, err := s.storage.SaveBatch(ctx, requestModels)
	if err != nil {
		return nil, fmt.Errorf("ShortenURLBatch, storage.SaveBatch: %w", err)
	}

	responseItems := make([]*models.ShortenBatchResponseItem, 0, len(requestItems))
	for _, responseModel := range responseModels {
		shortURL, err := url.JoinPath(s.configuration.BaseURL, responseModel.Key)
		if err != nil {
			return nil, fmt.Errorf("ShortenURL, url.JoinPath: %w", err)
		}

		responseItems = append(responseItems, &models.ShortenBatchResponseItem{
			CorelationID: responseModel.CorelationID,
			ShortURL:     shortURL,
		})
	}

	return responseItems, nil
}

func (s *ShortenerApp) CheckStore(ctx context.Context) error {
	err := s.storage.CheckAvailability(ctx)
	if err != nil {
		return fmt.Errorf("store is not available: %w", err)
	}

	return nil
}
