package app

import (
	"encoding/json"
	"net/http"

	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
)

// UserHandler структура обработчиков запросов, относящихся к пользователю.
type UserHandler struct {
	shortener App
	logger    logger.Logger
}

// NewUserHandler Конструктор.
func NewUserHandler(shortener App, logger logger.Logger) *UserHandler {
	return &UserHandler{
		shortener: shortener,
		logger:    logger,
	}
}

// HandleGetUserURLsRequest обработчик запроса получения всех URL пользователя.
func (h *UserHandler) HandleGetUserURLsRequest(response http.ResponseWriter, request *http.Request) {
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

// HandleBatchDeleteRequest обработчик запроса пакетного удаления.
func (h *UserHandler) HandleBatchDeleteRequest(response http.ResponseWriter, request *http.Request) {
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
