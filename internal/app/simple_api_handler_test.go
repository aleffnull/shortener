package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/models"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSimpleAPIHandler_HandleGetRequest(t *testing.T) {
	type want struct {
		statusCode int
		headers    map[string]string
		emptyBody  bool
	}

	tests := []struct {
		name       string
		key        string
		want       want
		hookBefore func(key string, mock *mocks.Mock)
	}{
		{
			name: "unknown key",
			key:  "foo",
			want: want{
				statusCode: http.StatusBadRequest,
				emptyBody:  false,
			},
			hookBefore: func(key string, mock *mocks.Mock) {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(nil, nil)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "existing key",
			key:  "foo",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				headers: map[string]string{
					headers.Location: "http://bar.buz",
				},
				emptyBody: true,
			},
			hookBefore: func(key string, mock *mocks.Mock) {
				mock.App.EXPECT().
					GetURL(gomock.Any(), key).
					Return(&models.GetURLResponseItem{URL: "http://bar.buz"}, nil)
				mock.AuditService.EXPECT().AuditEvent(gomock.Any())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/foo", nil)
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			handler := NewSimpleAPIHandler(mock.App, mock.AuditService, mock.Logger)
			tt.hookBefore(tt.key, mock)

			// Act.
			handler.HandleGetRequest(recorder, request, tt.key)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)

			if len(tt.want.headers) != 0 {
				for header, value := range tt.want.headers {
					require.Equal(t, value, result.Header.Get(header))
				}
			}

			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			if tt.want.emptyBody {
				require.Empty(t, body)
			} else {
				require.NotEmpty(t, body)
			}
		})
	}
}

func TestHandler_HandlePostRequest(t *testing.T) {
	type want struct {
		statusCode  int
		validateURL bool
	}

	tests := []struct {
		name       string
		longURL    string
		want       want
		hookBefore func(longURL string, mock *mocks.Mock)
	}{
		{
			name: "no body",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(longURL string, mock *mocks.Mock) {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name:    "app error",
			longURL: "http://foo.bar",
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(longURL string, mock *mocks.Mock) {
				shortenRequest := &models.ShortenRequest{
					URL: longURL,
				}
				mock.App.EXPECT().ShortenURL(gomock.Any(), shortenRequest, gomock.Any()).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name:    "valid request",
			longURL: "http://foo.bar",
			want: want{
				statusCode:  http.StatusCreated,
				validateURL: true,
			},
			hookBefore: func(longURL string, mock *mocks.Mock) {
				shortenRequest := &models.ShortenRequest{
					URL: longURL,
				}
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
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.longURL))
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			handler := NewSimpleAPIHandler(mock.App, mock.AuditService, mock.Logger)
			if tt.hookBefore != nil {
				tt.hookBefore(tt.longURL, mock)
			}

			// Act.
			handler.HandlePostRequest(recorder, request)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)

			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			if tt.want.validateURL {
				_, err = url.ParseRequestURI(string(body))
				require.NoError(t, err)
			}
		})
	}
}
