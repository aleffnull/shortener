package main

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
)

func handleRequest(response http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		app.HandleGetRequest(response, request)
	} else if request.Method == http.MethodPost {
		app.HandlePostRequest(response, request)
	} else {
		app.HandleInvalidMethod(response)
	}
}
