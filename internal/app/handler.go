package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/go-playground/validator/v10"
	"github.com/ldez/mimetype"
	"github.com/samber/lo"

	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/audit"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/aleffnull/shortener/models"
)

type Handler struct {
	shortener      App
	parameters     parameters.AppParameters
	logger         logger.Logger
	auditReceivers []audit.Receiver
}

func NewHandler(
	shortener App,
	parameters parameters.AppParameters,
	logger logger.Logger,
	auditReceivers []audit.Receiver,
) *Handler {
	return &Handler{
		shortener:      shortener,
		parameters:     parameters,
		logger:         logger,
		auditReceivers: auditReceivers,
	}
}

func (h *Handler) HandleGetRequest(response http.ResponseWriter, request *http.Request, key string) {
	if key == "favicon.ico" {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	item, err := h.shortener.GetURL(request.Context(), key)
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}
	if item == nil {
		utils.HandleRequestError(response, fmt.Errorf("key was not found: '%v'", key), h.logger)
		return
	}

	if item.IsDeleted {
		response.WriteHeader(http.StatusGone)
		return
	}

	response.Header().Set(headers.Location, item.URL)
	response.WriteHeader(http.StatusTemporaryRedirect)

	h.notifyAudit(&audit.Event{
		Timestamp: audit.FormattedTime(time.Now()),
		Action:    audit.ActionFollow,
		UserID:    item.UserID,
		URL:       item.URL,
	})
}

func (h *Handler) HandleGetUserURLsRequest(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := middleware.GetUserIDFromContext(ctx)
	items, err := h.shortener.GetUserURLs(ctx, userID)
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	if len(items) == 0 {
		response.WriteHeader(http.StatusNoContent)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	response.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(response).Encode(items); err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}
}

func (h *Handler) HandlePostRequest(response http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	longURL := string(body)
	if len(body) == 0 {
		utils.HandleRequestError(response, errors.New("body is required"), h.logger)
		return
	}

	shortenRequest := &models.ShortenRequest{
		URL: longURL,
	}
	ctx := request.Context()
	userID := middleware.GetUserIDFromContext(ctx)
	shortenerResponse, err := h.shortener.ShortenURL(ctx, shortenRequest, userID)
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.TextPlain)
	response.WriteHeader(lo.Ternary(shortenerResponse.IsDuplicate, http.StatusConflict, http.StatusCreated))
	fmt.Fprint(response, shortenerResponse.Result)

	if !shortenerResponse.IsDuplicate {
		h.notifyAudit(&audit.Event{
			Timestamp: audit.FormattedTime(time.Now()),
			Action:    audit.ActionShorten,
			UserID:    userID,
			URL:       longURL,
		})
	}
}

func (h *Handler) HandleAPIRequest(response http.ResponseWriter, request *http.Request) {
	var shortenRequest models.ShortenRequest
	if err := json.NewDecoder(request.Body).Decode(&shortenRequest); err != nil {
		utils.HandleRequestError(response, err, h.logger)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(shortenRequest); err != nil {
		utils.HandleRequestError(response, err, h.logger)
		return
	}

	ctx := request.Context()
	userID := middleware.GetUserIDFromContext(ctx)
	shortenerResponse, err := h.shortener.ShortenURL(ctx, &shortenRequest, userID)
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	response.WriteHeader(lo.Ternary(shortenerResponse.IsDuplicate, http.StatusConflict, http.StatusCreated))

	if err = json.NewEncoder(response).Encode(shortenerResponse); err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	if !shortenerResponse.IsDuplicate {
		h.notifyAudit(&audit.Event{
			Timestamp: audit.FormattedTime(time.Now()),
			Action:    audit.ActionShorten,
			UserID:    userID,
			URL:       shortenRequest.URL,
		})
	}
}

func (h *Handler) HandleAPIBatchRequest(response http.ResponseWriter, request *http.Request) {
	var requestItems []*models.ShortenBatchRequestItem
	if err := json.NewDecoder(request.Body).Decode(&requestItems); err != nil {
		utils.HandleRequestError(response, err, h.logger)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Var(requestItems, "omitempty,dive"); err != nil {
		utils.HandleRequestError(response, err, h.logger)
		return
	}

	ctx := request.Context()
	userID := middleware.GetUserIDFromContext(ctx)
	shortenerBatchResponse, err := h.shortener.ShortenURLBatch(ctx, requestItems, userID)
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	response.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(response).Encode(shortenerBatchResponse); err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}
}

func (h *Handler) HandleBatchDeleteRequest(response http.ResponseWriter, request *http.Request) {
	var keys []string
	if err := json.NewDecoder(request.Body).Decode(&keys); err != nil {
		utils.HandleRequestError(response, err, h.logger)
		return
	}

	ctx := request.Context()
	userID := middleware.GetUserIDFromContext(ctx)
	h.shortener.DeleteURLs(keys, userID)

	response.WriteHeader(http.StatusAccepted)
}

func (h *Handler) HandlePingRequest(response http.ResponseWriter, request *http.Request) {
	err := h.shortener.CheckStore(request.Context())
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	response.WriteHeader(http.StatusOK)
}

func (h *Handler) notifyAudit(event *audit.Event) {
	for _, receiver := range h.auditReceivers {
		if err := receiver.AddEvent(event); err != nil {
			h.logger.Errorf("Audit error: %v", err)
		}
	}
}
