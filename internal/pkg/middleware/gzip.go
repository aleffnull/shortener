package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/go-http-utils/headers"
)

const gzipEncoding = "gzip"

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter io.Writer
}

func (w gzipResponseWriter) Write(p []byte) (int, error) {
	return w.gzipWriter.Write(p)
}

func GzipHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get(headers.ContentEncoding) == gzipEncoding {
			body, err := gzip.NewReader(request.Body)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}

			defer body.Close()
			request.Body = body
		}

		actualWriter := writer
		if strings.Contains(request.Header.Get(headers.AcceptEncoding), gzipEncoding) {
			gzipWriter := gzip.NewWriter(writer)
			defer gzipWriter.Close()

			writer.Header().Set(headers.ContentEncoding, gzipEncoding)
			actualWriter = gzipResponseWriter{
				ResponseWriter: writer,
				gzipWriter:     gzipWriter,
			}
		}

		next.ServeHTTP(actualWriter, request)
	}
}
