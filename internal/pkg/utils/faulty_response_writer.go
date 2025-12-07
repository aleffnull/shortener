package utils

import (
	"net/http"

	"github.com/stretchr/testify/assert"
)

type FaultyResponseWriter struct {
	StatusCode int
}

var _ http.ResponseWriter = (*FaultyResponseWriter)(nil)

func NewFaultyResponseWriter() *FaultyResponseWriter {
	return &FaultyResponseWriter{}
}

func (w *FaultyResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *FaultyResponseWriter) Write([]byte) (int, error) {
	return 0, assert.AnError
}

func (w *FaultyResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}
