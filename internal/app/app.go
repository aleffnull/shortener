package app

import (
	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/store"
)

type ShortenerApp struct {
	configuration config.Configuration
	storage       store.Store
}

func NewShortenerApp(configuration config.Configuration) *ShortenerApp {
	return &ShortenerApp{
		configuration: configuration,
		storage:       store.NewMemoryStore(),
	}
}

func (shortener *ShortenerApp) GetBaseURL() string {
	return shortener.configuration.BaseURL
}

func (shortener *ShortenerApp) GetURL(key string) (string, bool) {
	url, ok := shortener.storage.Load(key)
	return url, ok
}

func (shortener *ShortenerApp) SaveURL(url string) string {
	return shortener.storage.Save(url)
}

func (shortener *ShortenerApp) SetKeyAndURL(key, url string) {
	shortener.storage.Set(key, url)
}
