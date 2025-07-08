package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aleffnull/shortener/models"
	"github.com/go-http-utils/headers"
	"github.com/go-playground/validator/v10"
	"github.com/ldez/mimetype"
)

type Handler struct {
	shortener App
}

func NewHandler(shortener App) *Handler {
	return &Handler{
		shortener: shortener,
	}
}

func (h *Handler) HandleGetRequest(response http.ResponseWriter, key string) {
	value, ok := h.shortener.GetURL(key)
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

	longURL := string(body)
	if len(body) == 0 {
		http.Error(response, "Body is required", http.StatusBadRequest)
		return
	}

	shortenRequest := &models.ShortenRequest{
		URL: longURL,
	}
	shortenerResponse, err := h.shortener.ShortenURL(shortenRequest)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.TextPlain)
	response.WriteHeader(http.StatusCreated)
	fmt.Fprint(response, shortenerResponse.Result)
}

func (h *Handler) HandleAPIRequest(response http.ResponseWriter, request *http.Request) {
	var shortenRequest models.ShortenRequest
	if err := json.NewDecoder(request.Body).Decode(&shortenRequest); err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(shortenRequest); err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	shortenerResponse, err := h.shortener.ShortenURL(&shortenRequest)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	response.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(response).Encode(shortenerResponse); err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandlePingRequest(response http.ResponseWriter, request *http.Request) {
	err := h.shortener.CheckStore(request.Context())
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
}
