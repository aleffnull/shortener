package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aleffnull/shortener/internal/store"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeValue  = "text/plain"

	locationHeader = "Location"
)

var storage = store.NewMemoryStore()

func HandleGetRequest(response http.ResponseWriter, request *http.Request) {
	key := strings.TrimPrefix(request.URL.Path, "/")
	if len(key) == 0 {
		http.Error(response, "Key is required", http.StatusBadRequest)
		return
	}

	value, ok := storage.Load(key)
	if !ok {
		http.Error(response, "Key was not found", http.StatusBadRequest)
		return
	}

	response.Header().Set(locationHeader, value)
	response.WriteHeader(http.StatusTemporaryRedirect)
}

func HandlePostRequest(response http.ResponseWriter, request *http.Request) {
	if !strings.HasPrefix(request.Header.Get(contentTypeHeader), contentTypeValue) {
		errorString := fmt.Sprintf("Only %v content type is allowed", contentTypeValue)
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

	key := storage.Save(bodyStr)

	response.Header().Set(contentTypeHeader, contentTypeValue)
	response.WriteHeader(http.StatusCreated)
	fmt.Fprintf(response, "http://localhost:8080/%v", key)
}

func HandleInvalidMethod(response http.ResponseWriter) {
	errorString := fmt.Sprintf("Only %v and %v requests are allowed", http.MethodGet, http.MethodPost)
	http.Error(response, errorString, http.StatusBadRequest)
}
