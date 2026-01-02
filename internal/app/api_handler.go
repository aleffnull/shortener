package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/aleffnull/shortener/internal/service"
	"github.com/aleffnull/shortener/models"
	"github.com/go-http-utils/headers"
	"github.com/go-playground/validator/v10"
	"github.com/ldez/mimetype"
	"github.com/samber/lo"
)

// APIHandler структура обработчиков запросов основного REST API.
type APIHandler struct {
	shortener    App
	auditService service.AuditService
	logger       logger.Logger
}

// NewAPIHandler Конструктор.
func NewAPIHandler(shortener App, auditService service.AuditService, logger logger.Logger) *APIHandler {
	return &APIHandler{
		shortener:    shortener,
		auditService: auditService,
		logger:       logger,
	}
}

// HandleAPIRequest обработчик POST-запроса с JSON-нагрузкой.
func (h *APIHandler) HandleAPIRequest(response http.ResponseWriter, request *http.Request) {
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
		h.auditService.AuditEvent(&domain.AuditEvent{
			Timestamp: domain.AuditFormattedTime(time.Now()),
			Action:    domain.AuditActionShorten,
			UserID:    userID,
			URL:       shortenRequest.URL,
		})
	}
}

// HandleAPIBatchRequest обработчик запроса пакетного сокращения URL.
func (h *APIHandler) HandleAPIBatchRequest(response http.ResponseWriter, request *http.Request) {
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
