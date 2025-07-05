package app

import (
	"context"

	"github.com/aleffnull/shortener/models"
)

type App interface {
	Init() error
	Shutdown()
	GetURL(key string) (string, bool)
	ShortenURL(request *models.ShortenRequest) (*models.ShortenResponse, error)
	CheckStore(context.Context) error
}
