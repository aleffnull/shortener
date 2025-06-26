package middleware

import "net/http"

type responseData struct {
	status int
	size   int
}

type ResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func NewResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: rw,
		responseData:   &responseData{},
	}
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.responseData.size += size
	return size, err
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func (w *ResponseWriter) GetStatus() int {
	return w.responseData.status
}

func (w *ResponseWriter) GetSize() int {
	return w.responseData.size
}
