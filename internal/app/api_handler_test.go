package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAPIHandler_HandleAPIRequest(t *testing.T) {
	type want struct {
		statusCode  int
		validateURL bool
	}

	tests := []struct {
		name           string
		shortenRequest *models.ShortenRequest
		want           want
		hookBefore     func(shortenRequest *models.ShortenRequest, mock *mocks.Mock)
	}{
		{
			name:           "no body",
			shortenRequest: nil,
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(shortenRequest *models.ShortenRequest, mock *mocks.Mock) {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name:           "invalid request",
			shortenRequest: &models.ShortenRequest{},
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(shortenRequest *models.ShortenRequest, mock *mocks.Mock) {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "app error",
			shortenRequest: &models.ShortenRequest{
				URL: "http://foo.bar",
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(shortenRequest *models.ShortenRequest, mock *mocks.Mock) {
				mock.App.EXPECT().ShortenURL(gomock.Any(), shortenRequest, gomock.Any()).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "valid request",
			shortenRequest: &models.ShortenRequest{
				URL: "http://foo.bar",
			},
			want: want{
				statusCode:  http.StatusCreated,
				validateURL: true,
			},
			hookBefore: func(shortenRequest *models.ShortenRequest, mock *mocks.Mock) {
				mock.App.EXPECT().ShortenURL(gomock.Any(), shortenRequest, gomock.Any()).Return(&models.ShortenResponse{
					Result: "http://localhost/abc",
				}, nil)
				mock.AuditService.EXPECT().AuditEvent(gomock.Any())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var body io.Reader
			if tt.shortenRequest != nil {
				jsonRequest, err := json.Marshal(tt.shortenRequest)
				require.NoError(t, err)
				body = bytes.NewReader(jsonRequest)
			}

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			handler := NewAPIHandler(mock.App, mock.AuditService, mock.Logger)
			if tt.hookBefore != nil {
				tt.hookBefore(tt.shortenRequest, mock)
			}

			// Act.
			handler.HandleAPIRequest(recorder, request)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.validateURL {
				defer result.Body.Close()
				var shortenResponse models.ShortenResponse
				err := json.NewDecoder(result.Body).Decode(&shortenResponse)
				require.NoError(t, err)

				_, err = url.ParseRequestURI(string(shortenResponse.Result))
				require.NoError(t, err)
			}
		})
	}
}
