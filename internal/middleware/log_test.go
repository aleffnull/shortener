package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLogHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hookBefore func(mock *mocks.Mock) (int, http.ResponseWriter, *http.Request)
	}{
		{
			name: "WHEN http status Created THEN do nothing",
			hookBefore: func(_ *mocks.Mock) (int, http.ResponseWriter, *http.Request) {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				return http.StatusCreated, recorder, request
			},
		},
		{
			name: "WHEN http status OK THEN do nothing",
			hookBefore: func(_ *mocks.Mock) (int, http.ResponseWriter, *http.Request) {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				return http.StatusOK, recorder, request
			},
		},
		{
			name: "WHEN http status internal error THEN log",
			hookBefore: func(mock *mocks.Mock) (int, http.ResponseWriter, *http.Request) {
				mock.Logger.EXPECT().Infof(gomock.Any()).DoAndReturn(func(str string, _ ...any) {
					require.NotEmpty(t, str)
				})
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, "/api/foo", nil)
				recorder.Header().Add(headers.ContentEncoding, "gzip")
				request.Header.Add(headers.ContentEncoding, "gzip")
				return http.StatusInternalServerError, recorder, request
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			statusCode, recorder, request := tt.hookBefore(mock)

			logHandler := LogHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
				fmt.Fprint(w, 42)
			}), mock.Logger)

			// Act.
			logHandler.ServeHTTP(recorder, request)
		})
	}
}
