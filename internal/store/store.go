package store

import (
	"context"

	"github.com/aleffnull/shortener/internal/pkg/models"
)

type Store interface {
	Init() error
	Shutdown()
	CheckAvailability(context.Context) error

	Load(key string) (string, bool)
	Save(value string) (string, error)
}

type ColdStore interface {
	LoadAll() ([]*models.ColdStoreEntry, error)
	Save(entry *models.ColdStoreEntry) error
}
