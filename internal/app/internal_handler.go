package app

import (
	"encoding/json"
	"net/http"

	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/go-http-utils/headers"
	"github.com/ldez/mimetype"
)

type InternalHandler struct {
	shortener App
	logger    logger.Logger
}

func NewInternalHandler(shortener App, logger logger.Logger) *InternalHandler {
	return &InternalHandler{
		shortener: shortener,
		logger:    logger,
	}
}

func (h *InternalHandler) HandleStatsRequest(response http.ResponseWriter, request *http.Request) {
	stats, err := h.shortener.GetStatistics(request.Context())
	if err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}

	response.Header().Set(headers.ContentType, mimetype.ApplicationJSON)
	response.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(response).Encode(stats); err != nil {
		utils.HandleServerError(response, err, h.logger)
		return
	}
}
