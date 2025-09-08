package store

import (
	"context"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/database"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/models"
	"github.com/google/uuid"
)

type StoreManager interface {
	Init() error
	Shutdown()
	CheckAvailability(context.Context) error
}

type DataStore interface {
	Load(context.Context, string) (*models.URLItem, error)
	LoadAllByUserID(context.Context, uuid.UUID) ([]*models.KeyOriginalURLItem, error)
	Save(context.Context, string, uuid.UUID) (string, error)
	SaveBatch(context.Context, []*models.BatchRequestItem, uuid.UUID) ([]*models.BatchResponseItem, error)
	DeleteBatch(context.Context, []string, uuid.UUID) error
}

type Store interface {
	StoreManager
	DataStore
}

type ColdStore interface {
	LoadAll() ([]*models.ColdStoreEntry, error)
	Save(*models.ColdStoreEntry) error
}

func NewStore(
	connection database.Connection,
	coldStore ColdStore,
	configuration *config.Configuration,
	logger logger.Logger,
) Store {
	if configuration.DatabaseStore.IsDatabaseEnabled() {
		return NewDatabaseStore(connection, configuration, logger)
	}

	return NewMemoryStore(coldStore, configuration, logger)
}
