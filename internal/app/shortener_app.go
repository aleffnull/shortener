package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/database"
	pkg_errors "github.com/aleffnull/shortener/internal/pkg/errors"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	pkg_models "github.com/aleffnull/shortener/internal/pkg/models"
	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"github.com/aleffnull/shortener/models"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type ShortenerApp struct {
	connection        database.Connection
	storage           store.Store
	logger            logger.Logger
	parameters        parameters.AppParameters
	configuration     *config.Configuration
	deleteURLsChannel chan deleteURLsRequest
	quitChannel       chan struct{}
}

type deleteURLsRequest struct {
	keys   []string
	userID uuid.UUID
}

var _ App = (*ShortenerApp)(nil)

func NewShortenerApp(
	connection database.Connection,
	storage store.Store,
	logger logger.Logger,
	parameters parameters.AppParameters,
	configuration *config.Configuration,
) App {
	return &ShortenerApp{
		connection:        connection,
		storage:           storage,
		logger:            logger,
		parameters:        parameters,
		configuration:     configuration,
		deleteURLsChannel: make(chan deleteURLsRequest, 100),
		quitChannel:       make(chan struct{}),
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

	go s.deletePendingURLs()

	return nil
}

func (s *ShortenerApp) Shutdown() {
	close(s.quitChannel)
	close(s.deleteURLsChannel)
	s.storage.Shutdown()
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

func (s *ShortenerApp) ShortenURLBatch(ctx context.Context, requestItems []*models.ShortenBatchRequestItem, userID uuid.UUID) ([]*models.ShortenBatchResponseItem, error) {
	if len(requestItems) == 0 {
		return []*models.ShortenBatchResponseItem{}, nil
	}

	requestModels := lo.Map(requestItems, func(item *models.ShortenBatchRequestItem, _ int) *pkg_models.BatchRequestItem {
		return &pkg_models.BatchRequestItem{
			CorelationID: item.CorelationID,
			OriginalURL:  item.OriginalURL,
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
			CorelationID: responseModel.CorelationID,
			ShortURL:     shortURL,
		})
	}

	return responseItems, nil
}

func (s *ShortenerApp) DeleteURLs(keys []string, userID uuid.UUID) {
	s.deleteURLsChannel <- deleteURLsRequest{
		keys:   keys,
		userID: userID,
	}
}

func (s *ShortenerApp) CheckStore(ctx context.Context) error {
	err := s.storage.CheckAvailability(ctx)
	if err != nil {
		return fmt.Errorf("store is not available: %w", err)
	}

	return nil
}

func (s *ShortenerApp) deletePendingURLs() {
	ticker := time.NewTicker(3 * time.Second)

	requests := []deleteURLsRequest{}
	for {
		select {
		case <-s.quitChannel:
			// Завершаем работу.
			return
		case request := <-s.deleteURLsChannel:
			// Накапливаем сообщения.
			requests = append(requests, request)
		case <-ticker.C:
			// Пробуем выполнить удаление при каждом срабатывании таймера.
			if len(requests) == 0 {
				continue
			}

			requests = s.doDeleteURLs(requests)
		}
	}
}

func (s *ShortenerApp) doDeleteURLs(requests []deleteURLsRequest) []deleteURLsRequest {
	// Ограничиваем количество одновременных запросов к базе.
	semaphore := make(chan struct{}, 5)
	waitGroup := sync.WaitGroup{}
	// Если что-то не смогли удалить, нужно это сохранить.
	nonDeletedChannel := make(chan deleteURLsRequest, len(requests))

	for _, request := range requests {
		semaphore <- struct{}{}
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()
			defer func() { <-semaphore }()

			if err := s.storage.DeleteBatch(context.Background(), request.keys, request.userID); err != nil {
				s.logger.Errorf("ShortenerApp.deletePendingURLs, storage.DeleteBatch failed: %v", err)
				nonDeletedChannel <- request
			}
		}()
	}

	waitGroup.Wait()
	close(nonDeletedChannel)

	// Восстановим то, что не смогли удалить.
	notDeletedRequests := []deleteURLsRequest{}
	for request := range nonDeletedChannel {
		notDeletedRequests = append(notDeletedRequests, request)
	}

	return notDeletedRequests
}
