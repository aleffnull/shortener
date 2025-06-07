package main

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
)

func main() {
	shortener := app.NewShortenerApp()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		handleRequest(response, request, shortener)
	})

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)
	}
}
