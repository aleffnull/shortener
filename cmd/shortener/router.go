package main

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
)

func handleRequest(response http.ResponseWriter, request *http.Request, shortener *app.ShortenerApp) {
	if request.Method == http.MethodGet {
		app.HandleGetRequest(response, request, shortener)
	} else if request.Method == http.MethodPost {
		app.HandlePostRequest(response, request, shortener)
	} else {
		app.HandleInvalidMethod(response)
	}
}
