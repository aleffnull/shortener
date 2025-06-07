package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
)

func HandleGetRequest(response http.ResponseWriter, request *http.Request, shortener *ShortenerApp) {
	key := strings.TrimPrefix(request.URL.Path, "/")
	if len(key) == 0 {
		http.Error(response, "Key is required", http.StatusBadRequest)
		return
	}

	value, ok := shortener.GetURL(key)
	if !ok {
		http.Error(response, "Key was not found", http.StatusBadRequest)
		return
	}

	response.Header().Set(headers.Location, value)
	response.WriteHeader(http.StatusTemporaryRedirect)
}

func HandlePostRequest(response http.ResponseWriter, request *http.Request, shortener *ShortenerApp) {
	if !strings.HasPrefix(request.Header.Get(headers.ContentType), mimetype.TextPlain) {
		errorString := fmt.Sprintf("Only %v content type is allowed", mimetype.TextPlain)
		http.Error(response, errorString, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}

	bodyStr := string(body)
	if len(body) == 0 {
		http.Error(response, "Body is required", http.StatusBadRequest)
		return
	}

	key := shortener.SaveURL(bodyStr)

	response.Header().Set(headers.ContentType, mimetype.TextPlain)
	response.WriteHeader(http.StatusCreated)
	fmt.Fprintf(response, "http://localhost:8080/%v", key)
}

func HandleInvalidMethod(response http.ResponseWriter) {
	errorString := fmt.Sprintf("Only %v and %v requests are allowed", http.MethodGet, http.MethodPost)
	http.Error(response, errorString, http.StatusBadRequest)
}
