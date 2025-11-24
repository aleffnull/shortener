package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/repository"
)

type StoreManager interface {
	Init() error
	Shutdown()
	CheckAvailability(context.Context) error
}

type DataStore interface {
	Load(context.Context, string) (*URLItem, error)
	LoadAllByUserID(context.Context, uuid.UUID) ([]*KeyOriginalURLItem, error)
	Save(context.Context, string, uuid.UUID) (string, error)
	SaveBatch(context.Context, []*BatchRequestItem, uuid.UUID) ([]*BatchResponseItem, error)
	DeleteBatch(context.Context, []string, uuid.UUID) error
}

type Store interface {
	StoreManager
	DataStore
}

type ColdStore interface {
	LoadAll() ([]*ColdStoreEntry, error)
	Save(*ColdStoreEntry) error
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
