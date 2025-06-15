package app

import "github.com/aleffnull/shortener/models"

type App interface {
	GetURL(key string) (string, bool)
	ShortenURL(request *models.ShortenRequest) (*models.ShortenResponse, error)
}
