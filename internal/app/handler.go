package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
)

type Handler struct {
	configuration *config.Configuration
	shortenerApp  *ShortenerApp
}

func NewHandler(configuration *config.Configuration, shortenerApp *ShortenerApp) *Handler {
	return &Handler{
		configuration: configuration,
		shortenerApp:  shortenerApp,
	}
}

func (h *Handler) HandleGetRequest(response http.ResponseWriter, key string) {
	value, ok := h.shortenerApp.GetURL(key)
	if !ok {
		http.Error(response, "Key was not found", http.StatusBadRequest)
		return
	}

	response.Header().Set(headers.Location, value)
	response.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) HandlePostRequest(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	bodyStr := string(body)
	if len(body) == 0 {
		http.Error(response, "Body is required", http.StatusBadRequest)
		return
	}

	key, err := h.shortenerApp.SaveURL(bodyStr)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	shortPath, err := url.JoinPath(h.configuration.BaseURL, key)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.TextPlain)
	response.WriteHeader(http.StatusCreated)
	fmt.Fprint(response, shortPath)
}
