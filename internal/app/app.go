package app

import (
	"context"

	"github.com/aleffnull/shortener/models"
)

type App interface {
	Init() error
	Shutdown()
	GetURL(context.Context, string) (string, bool, error)
	ShortenURL(context.Context, *models.ShortenRequest) (*models.ShortenResponse, error)
	ShortenURLBatch(context.Context, []*models.ShortenBatchRequestItem) ([]*models.ShortenBatchResponseItem, error)
	CheckStore(context.Context) error
}
