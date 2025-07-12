package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/models"
	"github.com/go-http-utils/headers"
	"github.com/go-playground/validator/v10"
	"github.com/ldez/mimetype"
	"github.com/samber/lo"
)

type Handler struct {
	shortener App
	logger    logger.Logger
}

func NewHandler(shortener App, logger logger.Logger) *Handler {
	return &Handler{
		shortener: shortener,
		logger:    logger,
	}
}

func (h *Handler) HandleGetRequest(response http.ResponseWriter, request *http.Request, key string) {
	value, ok, err := h.shortener.GetURL(request.Context(), key)
	if err != nil {
		h.handleServerError(response, err)
		return
	}
	if !ok {
		h.handleRequestError(response, errors.New("key was not found"))
		return
	}

	response.Header().Set(headers.Location, value)
	response.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) HandlePostRequest(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		h.handleServerError(response, err)
		return
	}

	longURL := string(body)
	if len(body) == 0 {
		h.handleRequestError(response, errors.New("body is required"))
		return
	}

	shortenRequest := &models.ShortenRequest{
		URL: longURL,
	}
	shortenerResponse, err := h.shortener.ShortenURL(request.Context(), shortenRequest)
	if err != nil {
		h.handleServerError(response, err)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.TextPlain)
	response.WriteHeader(lo.Ternary(shortenerResponse.IsDuplicate, http.StatusConflict, http.StatusCreated))
	fmt.Fprint(response, shortenerResponse.Result)
}

func (h *Handler) HandleAPIRequest(response http.ResponseWriter, request *http.Request) {
	var shortenRequest models.ShortenRequest
	if err := json.NewDecoder(request.Body).Decode(&shortenRequest); err != nil {
		h.handleRequestError(response, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(shortenRequest); err != nil {
		h.handleRequestError(response, err)
		return
	}

	shortenerResponse, err := h.shortener.ShortenURL(request.Context(), &shortenRequest)
	if err != nil {
		h.handleServerError(response, err)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	response.WriteHeader(lo.Ternary(shortenerResponse.IsDuplicate, http.StatusConflict, http.StatusCreated))

	if err = json.NewEncoder(response).Encode(shortenerResponse); err != nil {
		h.handleServerError(response, err)
		return
	}
}

func (h *Handler) HandleAPIBatchRequest(response http.ResponseWriter, request *http.Request) {
	var requestItems []*models.ShortenBatchRequestItem
	if err := json.NewDecoder(request.Body).Decode(&requestItems); err != nil {
		h.handleRequestError(response, err)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Var(requestItems, "omitempty,dive"); err != nil {
		h.handleRequestError(response, err)
		return
	}

	shortenerBatchResponse, err := h.shortener.ShortenURLBatch(request.Context(), requestItems)
	if err != nil {
		h.handleServerError(response, err)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	response.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(response).Encode(shortenerBatchResponse); err != nil {
		h.handleServerError(response, err)
		return
	}
}

func (h *Handler) HandlePingRequest(response http.ResponseWriter, request *http.Request) {
	err := h.shortener.CheckStore(request.Context())
	if err != nil {
		h.handleServerError(response, err)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (h *Handler) handleServerError(response http.ResponseWriter, err error) {
	h.logger.Errorf("Server error: %v", err)
	http.Error(response, "Internal server error", http.StatusInternalServerError)
}

func (h *Handler) handleRequestError(response http.ResponseWriter, err error) {
	h.logger.Errorf("Request error: %v", err)
	http.Error(response, err.Error(), http.StatusBadRequest)
}
