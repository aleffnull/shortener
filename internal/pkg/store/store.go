package store

import (
	"context"

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
