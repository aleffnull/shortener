package app

import (
	"github.com/aleffnull/shortener/internal/store"
)

type ShortenerApp struct {
	storage store.Store
}

func NewShortenerApp() *ShortenerApp {
	return &ShortenerApp{
		storage: store.NewMemoryStore(),
	}
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
