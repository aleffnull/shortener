package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"github.com/aleffnull/shortener/internal/repository"
	"github.com/aleffnull/shortener/internal/service"
	"github.com/aleffnull/shortener/models"
)

type ShortenerApp struct {
	connection        repository.Connection
	storage           store.Store
	deleteURLsService service.DeleteURLsService
	auditService      service.AuditService
	logger            logger.Logger
	parameters        parameters.AppParameters
	configuration     *config.Configuration
}

var _ App = (*ShortenerApp)(nil)

func NewShortenerApp(
	connection repository.Connection,
	storage store.Store,
	deleteURLsService service.DeleteURLsService,
	auditService service.AuditService,
	logger logger.Logger,
	parameters parameters.AppParameters,
	configuration *config.Configuration,
) App {
	return &ShortenerApp{
		connection:        connection,
		storage:           storage,
		deleteURLsService: deleteURLsService,
		auditService:      auditService,
		logger:            logger,
		parameters:        parameters,
		configuration:     configuration,
	}
}

func (s *ShortenerApp) Init(ctx context.Context) error {
	if err := s.storage.Init(); err != nil {
		return fmt.Errorf("ShortenerApp.Init, storage.Init failed: %w", err)
	}

	if err := s.connection.Init(ctx); err != nil {
		return fmt.Errorf("ShortenerApp.Init, connection.Init failed: %w", err)
	}

	if err := s.parameters.Init(ctx); err != nil {
		return fmt.Errorf("ShortenerApp.Init, parameters.Init failed: %w", err)
	}

	s.auditService.Init()
	s.deleteURLsService.Init()

	return nil
}

func (s *ShortenerApp) Shutdown() {
	s.deleteURLsService.Shutdown()
	s.auditService.Shutdown()
	s.connection.Shutdown()
}

func (s *ShortenerApp) GetURL(ctx context.Context, key string) (*models.GetURLResponseItem, error) {
	item, err := s.storage.Load(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("ShortenerApp.GetURL, storage.Load failed: %w", err)
	}

	if item == nil {
		return nil, nil
	}

	return &models.GetURLResponseItem{
		URL:       item.URL,
		UserID:    item.UserID,
		IsDeleted: item.IsDeleted,
	}, nil
}

func (s *ShortenerApp) GetUserURLs(ctx context.Context, userID uuid.UUID) ([]*models.UserURLsResponseItem, error) {
	items, err := s.storage.LoadAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ShortenerApp.GetUserURLs, storage.LoadAllByUserID failed: %w", err)
	}

	responseItems := []*models.UserURLsResponseItem{}
	for _, item := range items {
		shortURL, err := url.JoinPath(s.configuration.BaseURL, item.URLKey)
		if err != nil {
			return nil, fmt.Errorf("ShortenerApp.GetUserURLs, url.JoinPath: %w", err)
		}

		responseItems = append(responseItems, &models.UserURLsResponseItem{
			ShortURL:    shortURL,
			OriginalURL: item.OriginalURL,
		})
	}

	return responseItems, nil
}

func (s *ShortenerApp) ShortenURL(ctx context.Context, request *models.ShortenRequest, userID uuid.UUID) (*models.ShortenResponse, error) {
	longURL := request.URL
	key, err := s.storage.Save(ctx, longURL, userID)

	isDuplicate := false
	if err != nil {
		var duplicateURLError *store.DuplicateURLError
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

func (s *ShortenerApp) ShortenURLBatch(ctx context.Context, requestItems []*models.ShortenBatchRequestItem, userID uuid.UUID) ([]*models.ShortenBatchResponseItem, error) {
	if len(requestItems) == 0 {
		return []*models.ShortenBatchResponseItem{}, nil
	}

	requestModels := lo.Map(requestItems, func(item *models.ShortenBatchRequestItem, _ int) *domain.BatchRequestItem {
		return &domain.BatchRequestItem{
			CorrelationID: item.CorrelationID,
			OriginalURL:   item.OriginalURL,
		}
	})
	responseModels, err := s.storage.SaveBatch(ctx, requestModels, userID)
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
			CorrelationID: responseModel.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return responseItems, nil
}

func (s *ShortenerApp) DeleteURLs(keys []string, userID uuid.UUID) {
	s.deleteURLsService.Delete(&domain.DeleteURLsRequest{
		Keys:   keys,
		UserID: userID,
	})
}

func (s *ShortenerApp) CheckStore(ctx context.Context) error {
	err := s.storage.CheckAvailability(ctx)
	if err != nil {
		return fmt.Errorf("store is not available: %w", err)
	}

	return nil
}
