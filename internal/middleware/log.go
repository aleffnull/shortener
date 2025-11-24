package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-http-utils/headers"

	"github.com/aleffnull/shortener/internal/pkg/logger"
)

func LogHandler(handlerFunc http.HandlerFunc, logger logger.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()
		responseWriter := NewResponseWriter(writer)
		handlerFunc(responseWriter, request)
		duration := time.Since(startTime)

		status := responseWriter.GetStatus()
		if status == http.StatusCreated || status == http.StatusOK {
			return
		}

		sb := &strings.Builder{}

		fmt.Fprintf(
			sb,
			"URL: %v, method: %v, time: %v, status: %v, response size %v bytes",
			request.URL,
			request.Method,
			duration,
			responseWriter.GetStatus(),
			responseWriter.GetSize())

		requestEncoding := request.Header.Get(headers.ContentEncoding)
		if requestEncoding != "" {
			fmt.Fprintf(sb, ", request encoding: %v", requestEncoding)
		}

		responseEncoding := responseWriter.Header().Get(headers.ContentEncoding)
		if responseEncoding != "" {
			fmt.Fprintf(sb, ", response encoding: %v", responseEncoding)
		}

		logger.Infof(sb.String())
	}
}
