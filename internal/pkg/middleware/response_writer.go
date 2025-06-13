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

func (writer *ResponseWriter) Write(data []byte) (int, error) {
	size, err := writer.ResponseWriter.Write(data)
	writer.responseData.size += size
	return size, err
}

func (writer *ResponseWriter) WriteHeader(statusCode int) {
	writer.ResponseWriter.WriteHeader(statusCode)
	writer.responseData.status = statusCode
}

func (writer *ResponseWriter) GetStatus() int {
	return writer.responseData.status
}

func (writer *ResponseWriter) GetSize() int {
	return writer.responseData.size
}
