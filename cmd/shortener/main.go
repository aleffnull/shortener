package main

import (
	"log"
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"
)

func main() {
	configuration, err := config.GetConfiguration()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	log.Printf("using configuration: %+v\n", configuration)
	shortenerApp := app.NewShortenerApp()
	handler := app.NewHandler(configuration, shortenerApp)

	router := chi.NewRouter()
	router.Use(middleware.AllowContentType(mimetype.TextPlain))
	router.Get("/{key}", func(response http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		handler.HandleGetRequest(response, key)
	})
	router.Post("/", func(response http.ResponseWriter, request *http.Request) {
		handler.HandlePostRequest(response, request)
	})

	err = http.ListenAndServe(configuration.ServerAddress, router)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
