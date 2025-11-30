package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/audit"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/aleffnull/shortener/internal/service"
	"github.com/aleffnull/shortener/models"
	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
	"github.com/samber/lo"
)

// SimpleAPIHandler структура обработчиков запросов простого (не REST) API.
type SimpleAPIHandler struct {
	shortener    App
	auditService service.AuditService
	logger       logger.Logger
}

// NewSimpleAPIHandler Конструктор.
func NewSimpleAPIHandler(shortener App, auditService service.AuditService, logger logger.Logger) *SimpleAPIHandler {
	return &SimpleAPIHandler{
		shortener:    shortener,
		auditService: auditService,
		logger:       logger,
	}
}

// HandleGetRequest обработчик GET-запроса.
func (h *SimpleAPIHandler) HandleGetRequest(response http.ResponseWriter, request *http.Request, key string) {
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

	h.auditService.AuditEvent(&audit.Event{
		Timestamp: audit.FormattedTime(time.Now()),
		Action:    audit.ActionFollow,
		UserID:    item.UserID,
		URL:       item.URL,
	})
}

// HandlePostRequest Обработчик POST-запроса.
func (h *SimpleAPIHandler) HandlePostRequest(response http.ResponseWriter, request *http.Request) {
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
		h.auditService.AuditEvent(&audit.Event{
			Timestamp: audit.FormattedTime(time.Now()),
			Action:    audit.ActionShorten,
			UserID:    userID,
			URL:       longURL,
		})
	}
}
