package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/repository"
)

type StoreManager interface {
	Init() error
	Shutdown()
	CheckAvailability(context.Context) error
}

type DataStore interface {
	Load(context.Context, string) (*domain.URLItem, error)
	LoadAllByUserID(context.Context, uuid.UUID) ([]*domain.KeyOriginalURLItem, error)
	Save(context.Context, string, uuid.UUID) (string, error)
	SaveBatch(context.Context, []*domain.BatchRequestItem, uuid.UUID) ([]*domain.BatchResponseItem, error)
	DeleteBatch(context.Context, []string, uuid.UUID) error
}

type Store interface {
	StoreManager
	DataStore
}

type ColdStore interface {
	LoadAll() ([]*domain.ColdStoreEntry, error)
	Save(*domain.ColdStoreEntry) error
}

func NewStore(
	connection repository.Connection,
	coldStore ColdStore,
	configuration *config.Configuration,
	logger logger.Logger,
) Store {
	if configuration.DatabaseStore.IsDatabaseEnabled() {
		return NewDatabaseStore(connection, configuration, logger)
	}

	return NewMemoryStore(coldStore, configuration, logger)
}
