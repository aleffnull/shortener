package app

import (
	"context"

	"github.com/aleffnull/shortener/models"
	"github.com/google/uuid"
)

type App interface {
	Init(ctx context.Context) error
	Shutdown()
	GetURL(context.Context, string) (string, bool, error)
	GetUserURLs(context.Context, uuid.UUID) ([]*models.UserURLsResponseItem, error)
	ShortenURL(context.Context, *models.ShortenRequest, uuid.UUID) (*models.ShortenResponse, error)
	ShortenURLBatch(context.Context, []*models.ShortenBatchRequestItem, uuid.UUID) ([]*models.ShortenBatchResponseItem, error)
	CheckStore(context.Context) error
}
