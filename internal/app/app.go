package app

import (
	"context"

	"github.com/aleffnull/shortener/models"
)

type App interface {
	Init(context.Context) error
	GetURL(key string) (string, bool)
	ShortenURL(request *models.ShortenRequest) (*models.ShortenResponse, error)
}
