package store

import (
	"context"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/database"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/models"
)

type Store interface {
	Init() error
	Shutdown()
	CheckAvailability(context.Context) error

	Load(context.Context, string) (string, bool, error)
	Save(context.Context, string) (string, error)
	SaveBatch(context.Context, []*models.BatchRequestItem) ([]*models.BatchResponseItem, error)
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
