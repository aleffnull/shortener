package main

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"
)

func main() {
	shortener := app.NewShortenerApp()

	router := chi.NewRouter()
	router.Use(middleware.AllowContentType(mimetype.TextPlain))
	router.Get("/{key}", func(response http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		app.HandleGetRequest(response, key, shortener)
	})
	router.Post("/", func(response http.ResponseWriter, request *http.Request) {
		app.HandlePostRequest(response, request, shortener)
	})

	err := http.ListenAndServe("localhost:8080", router)
	if err != nil {
		panic(err)
	}
}
