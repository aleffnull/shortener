package main

import (
	"fmt"
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"
)

func main() {
	err := config.ParseConfiguration()
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	fmt.Printf("Using configuration: %+v\n", config.Current)
	shortener := app.NewShortenerApp(config.Current)

	router := chi.NewRouter()
	router.Use(middleware.AllowContentType(mimetype.TextPlain))
	router.Get("/{key}", func(response http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		app.HandleGetRequest(response, key, shortener)
	})
	router.Post("/", func(response http.ResponseWriter, request *http.Request) {
		app.HandlePostRequest(response, request, shortener)
	})

	err = http.ListenAndServe(config.Current.ServerAddress, router)
	if err != nil {
		panic(err)
	}
}
