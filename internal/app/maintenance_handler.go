package app

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/utils"
)

// MaintenanceHandler структура обработчиков служебных запросов.
type MaintenanceHandler struct {
	shortener App
	logger    logger.Logger
}

// NewMaintenanceHandler Конструктор.
func NewMaintenanceHandler(shortener App, logger logger.Logger) *MaintenanceHandler {
	return &MaintenanceHandler{
		shortener: shortener,
		logger:    logger,
	}
}

// HandlePingRequest обработчик запроса проверки работоспособности.
func (h *MaintenanceHandler) HandlePingRequest(response http.ResponseWriter, request *http.Request) {
	err := h.shortener.CheckStore(request.Context())
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	response.WriteHeader(http.StatusOK)
}
