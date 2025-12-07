package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/pkg/audit"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/aleffnull/shortener/models"
	"github.com/go-http-utils/headers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSimpleAPIHandler_HandleGetRequest(t *testing.T) {
	t.Parallel()

	type want struct {
		statusCode int
		headers    map[string]string
		emptyBody  bool
	}

	const fullURL = "http://bar.buz"

	tests := []struct {
		name       string
		key        string
		want       want
		hookBefore func(key string, mock *mocks.Mock)
	}{
		{
			name: "WHEN favicon.ico THEN not found",
			key:  "favicon.ico",
			want: want{
				statusCode: http.StatusNotFound,
				emptyBody:  true,
			},
		},
		{
			name: "WHEN app error THEN internal error",
			key:  "foo",
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(key string, mock *mocks.Mock) {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN unknown key THEN bad request",
			key:  "foo",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(key string, mock *mocks.Mock) {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(nil, nil)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN key deleted THEN gone",
			key:  "foo",
			want: want{
				statusCode: http.StatusGone,
				emptyBody:  true,
			},
			hookBefore: func(key string, mock *mocks.Mock) {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(&models.GetURLResponseItem{
					IsDeleted: true,
				}, nil)
			},
		},
		{
			name: "WHEN existing key THEN redirect",
			key:  "foo",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				headers: map[string]string{
					headers.Location: fullURL,
				},
				emptyBody: true,
			},
			hookBefore: func(key string, mock *mocks.Mock) {
				mock.App.EXPECT().
					GetURL(gomock.Any(), key).
					Return(&models.GetURLResponseItem{URL: fullURL}, nil)
				mock.AuditService.EXPECT().AuditEvent(gomock.Any()).DoAndReturn(func(event *audit.Event) {
					require.LessOrEqual(t, event.Timestamp, time.Now())
					require.Equal(t, audit.ActionFollow, event.Action)
					require.Equal(t, uuid.UUID{}, event.UserID)
					require.Equal(t, fullURL, event.URL)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/foo", nil)
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			handler := NewSimpleAPIHandler(mock.App, mock.AuditService, mock.Logger)
			if tt.hookBefore != nil {
				tt.hookBefore(tt.key, mock)
			}

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
	t.Parallel()

	const (
		longURL  = "http://foo.bar"
		shortURL = "http://localhost/abc"
	)

	type want struct {
		statusCode  int
		validateURL bool
	}

	tests := []struct {
		name       string
		want       want
		hookBefore func(mock *mocks.Mock) io.Reader
	}{
		{
			name: "WHEN read body error THEN internal error",
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) io.Reader {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return utils.AFaultyReader
			},
		},
		{
			name: "WHEN no body THEN bad request",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(mock *mocks.Mock) io.Reader {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return strings.NewReader("")
			},
		},
		{
			name: "WHEN app error THEN internal error",
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) io.Reader {
				shortenRequest := &models.ShortenRequest{
					URL: longURL,
				}
				mock.App.EXPECT().ShortenURL(gomock.Any(), shortenRequest, gomock.Any()).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return strings.NewReader(longURL)
			},
		},
		{
			name: "GIVEN not duplicate WHEN valid request THEN created",
			want: want{
				statusCode:  http.StatusCreated,
				validateURL: true,
			},
			hookBefore: func(mock *mocks.Mock) io.Reader {
				shortenRequest := &models.ShortenRequest{
					URL: longURL,
				}
				mock.App.EXPECT().ShortenURL(gomock.Any(), shortenRequest, gomock.Any()).Return(&models.ShortenResponse{
					Result: shortURL,
				}, nil)
				mock.AuditService.EXPECT().AuditEvent(gomock.Any()).DoAndReturn(func(event *audit.Event) {
					require.LessOrEqual(t, event.Timestamp, time.Now())
					require.Equal(t, audit.ActionShorten, event.Action)
					require.Equal(t, uuid.UUID{}, event.UserID)
					require.Equal(t, longURL, event.URL)
				})
				return strings.NewReader(longURL)
			},
		},
		{
			name: "GIVEN duplicate WHEN valid request THEN conflict",
			want: want{
				statusCode:  http.StatusConflict,
				validateURL: true,
			},
			hookBefore: func(mock *mocks.Mock) io.Reader {
				shortenRequest := &models.ShortenRequest{
					URL: longURL,
				}
				mock.App.EXPECT().ShortenURL(gomock.Any(), shortenRequest, gomock.Any()).Return(&models.ShortenResponse{
					Result:      shortURL,
					IsDuplicate: true,
				}, nil)
				return strings.NewReader(longURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			body := tt.hookBefore(mock)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/", body)
			handler := NewSimpleAPIHandler(mock.App, mock.AuditService, mock.Logger)

			// Act.
			handler.HandlePostRequest(recorder, request)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.want.statusCode, result.StatusCode)

			defer result.Body.Close()
			resultBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			if tt.want.validateURL {
				_, err = url.ParseRequestURI(string(resultBody))
				require.NoError(t, err)
			}
		})
	}
}
