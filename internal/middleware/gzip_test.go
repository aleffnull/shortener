package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
)

func TestGzipHandler(t *testing.T) {
	t.Parallel()

	compressableString := strings.Repeat("Q", 1000)

	tests := []struct {
		name       string
		hookBefore func() *http.Request
		hookAfter  func(statusCode int, resultBody []byte)
	}{
		{
			name: "GIVEN no gzipping in request and response THEN ok",
			hookBefore: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/api/foo", bytes.NewReader([]byte(compressableString)))
			},
			hookAfter: func(statusCode int, resultBody []byte) {
				require.Equal(t, http.StatusOK, statusCode)
				require.Equal(t, compressableString, string(resultBody))
			},
		},
		{
			name: "GIVEN gzip in request WHEN error durung request reading THEN internal error",
			hookBefore: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/api/foo", utils.AFaultyReader)
				request.Header.Add(headers.ContentEncoding, "gzip")
				return request
			},
			hookAfter: func(statusCode int, resultBody []byte) {
				require.Equal(t, http.StatusInternalServerError, statusCode)
				require.NotEmpty(t, resultBody)
			},
		},
		{
			name: "GIVEN gzip in request WHEN no errors THEN uncompressed",
			hookBefore: func() *http.Request {
				buf := &bytes.Buffer{}
				gzipWriter := gzip.NewWriter(buf)
				_, err := gzipWriter.Write([]byte(compressableString))
				require.NoError(t, err)
				require.NoError(t, gzipWriter.Close())

				request := httptest.NewRequest(http.MethodPost, "/api/foo", bytes.NewReader(buf.Bytes()))
				request.Header.Add(headers.ContentEncoding, "gzip")
				return request
			},
			hookAfter: func(statusCode int, resultBody []byte) {
				require.Equal(t, http.StatusOK, statusCode)
				require.Equal(t, compressableString, string(resultBody))
			},
		},
		{
			name: "GIVEN gzip in Accept-Encoding header WHEN no errors THEN compressed",
			hookBefore: func() *http.Request {
				request := httptest.NewRequest(http.MethodPost, "/api/foo", bytes.NewReader([]byte(compressableString)))
				request.Header.Add(headers.AcceptEncoding, "gzip")
				return request
			},
			hookAfter: func(statusCode int, resultBody []byte) {
				require.Equal(t, http.StatusOK, statusCode)
				gzipReader, err := gzip.NewReader(bytes.NewReader(resultBody))
				require.NoError(t, err)
				defer gzipReader.Close()
				body, err := io.ReadAll(gzipReader)
				require.NoError(t, err)
				require.Equal(t, compressableString, string(body))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			gzipHandler := GzipHandler(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				w.WriteHeader(http.StatusOK)
				_, err = w.Write(body)
				require.NoError(t, err)
			})
			recorder := httptest.NewRecorder()
			request := tt.hookBefore()

			// Act.
			gzipHandler(recorder, request)

			// Assert.
			result := recorder.Result()
			defer result.Body.Close()

			resultBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			tt.hookAfter(result.StatusCode, resultBody)
		})
	}
}
