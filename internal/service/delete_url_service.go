package service

import (
	"context"
	"sync"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/store"
)

type DeleteURLsService interface {
	Init()
	Shutdown()
	Delete(request *domain.DeleteURLsRequest)
}

type deleteURLsServiceImpl struct {
	logger            logger.Logger
	storage           store.Store
	deleteURLsChannel chan *domain.DeleteURLsRequest
	quitChannel       chan struct{}
}

var _ DeleteURLsService = (*deleteURLsServiceImpl)(nil)

func NewDeleteURLsService(storage store.Store, logger logger.Logger) DeleteURLsService {
	return &deleteURLsServiceImpl{
		storage:           storage,
		logger:            logger,
		deleteURLsChannel: make(chan *domain.DeleteURLsRequest, 100),
		quitChannel:       make(chan struct{}),
	}
}

func (i *deleteURLsServiceImpl) Init() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)

		requests := []*domain.DeleteURLsRequest{}
		for {
			select {
			case <-i.quitChannel:
				// Завершаем работу.
				return
			case request := <-i.deleteURLsChannel:
				// Накапливаем сообщения.
				requests = append(requests, request)
			case <-ticker.C:
				// Пробуем выполнить удаление при каждом срабатывании таймера.
				if len(requests) == 0 {
					continue
				}

				requests = i.doDeleteURLs(requests)
			}
		}
	}()
}

func (i *deleteURLsServiceImpl) Shutdown() {
	close(i.quitChannel)
	close(i.deleteURLsChannel)
}

func (i *deleteURLsServiceImpl) Delete(request *domain.DeleteURLsRequest) {
	i.deleteURLsChannel <- request
}

func (i *deleteURLsServiceImpl) doDeleteURLs(requests []*domain.DeleteURLsRequest) []*domain.DeleteURLsRequest {
	// Ограничиваем количество одновременных запросов к базе.
	semaphore := make(chan struct{}, 5)
	waitGroup := sync.WaitGroup{}
	// Если что-то не смогли удалить, нужно это сохранить.
	nonDeletedChannel := make(chan *domain.DeleteURLsRequest, len(requests))

	for _, request := range requests {
		semaphore <- struct{}{}
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()
			defer func() { <-semaphore }()

			if err := i.storage.DeleteBatch(context.Background(), request.Keys, request.UserID); err != nil {
				i.logger.Errorf("DeleteURLsService.deletePendingURLs, storage.DeleteBatch failed: %v", err)
				nonDeletedChannel <- request
			}
		}()
	}

	waitGroup.Wait()
	close(nonDeletedChannel)

	// Восстановим то, что не смогли удалить.
	notDeletedRequests := []*domain.DeleteURLsRequest{}
	for request := range nonDeletedChannel {
		notDeletedRequests = append(notDeletedRequests, request)
	}

	return notDeletedRequests
}
