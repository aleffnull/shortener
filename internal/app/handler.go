package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
)

func HandleGetRequest(response http.ResponseWriter, key string, shortener *ShortenerApp) {
	value, ok := shortener.GetURL(key)
	if !ok {
		http.Error(response, "Key was not found", http.StatusBadRequest)
		return
	}

	response.Header().Set(headers.Location, value)
	response.WriteHeader(http.StatusTemporaryRedirect)
}

func HandlePostRequest(response http.ResponseWriter, request *http.Request, shortener *ShortenerApp) {
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
	shortPath, err := url.JoinPath(shortener.GetBaseURL(), key)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.TextPlain)
	response.WriteHeader(http.StatusCreated)
	fmt.Fprint(response, shortPath)
}
