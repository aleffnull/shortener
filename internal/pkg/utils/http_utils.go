package utils

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/pkg/logger"
)

func HandleServerError(response http.ResponseWriter, err error, logger logger.Logger) {
	logger.Errorf("Server error: %v", err)
	http.Error(response, "Internal server error", http.StatusInternalServerError)
}

func HandleRequestError(response http.ResponseWriter, err error, logger logger.Logger) {
	logger.Errorf("Request error: %v", err)
	http.Error(response, err.Error(), http.StatusBadRequest)
}

func HandleUnauthorized(response http.ResponseWriter, message string, logger logger.Logger) {
	logger.Warnf("Unauthorized access: %v", message)
	http.Error(response, "Unauthorized", http.StatusUnauthorized)
}

func HandleForbidden(response http.ResponseWriter, message string, logger logger.Logger) {
	logger.Warnf("Access forbidden: %v", message)
	http.Error(response, "Forbidden", http.StatusForbidden)
}
