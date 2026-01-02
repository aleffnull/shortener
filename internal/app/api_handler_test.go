package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/aleffnull/shortener/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAPIHandler_HandleAPIRequest(t *testing.T) {
	t.Parallel()

	const (
		fullURL  = "http://foo.bar"
		shortURL = "http://localhost/abc3"
	)

	type args struct {
		shortenRequest *models.ShortenRequest
		responseWriter http.ResponseWriter
	}
	type want struct {
		statusCode int
	}

	tests := []struct {
		name       string
		args       args
		want       want
		hookBefore func(args args, mock *mocks.Mock)
		hookAfter  func(result *http.Response)
	}{
		{
			name: "WHEN no body THEN bad request",
			args: args{
				shortenRequest: nil,
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(args args, mock *mocks.Mock) {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN invalid request THEN bad request",
			args: args{
				shortenRequest: &models.ShortenRequest{},
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(args args, mock *mocks.Mock) {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN app error THEN internal error",
			args: args{
				shortenRequest: &models.ShortenRequest{
					URL: fullURL,
				},
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(args args, mock *mocks.Mock) {
				mock.App.EXPECT().ShortenURL(gomock.Any(), args.shortenRequest, gomock.Any()).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN response write error THEN internal error",
			args: args{
				shortenRequest: &models.ShortenRequest{
					URL: fullURL,
				},
				responseWriter: utils.NewFaultyResponseWriter(),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(args args, mock *mocks.Mock) {
				mock.App.EXPECT().ShortenURL(gomock.Any(), args.shortenRequest, gomock.Any()).Return(&models.ShortenResponse{
					Result: shortURL,
				}, nil)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "GIVEN unique url WHEN valid request THEN created",
			args: args{
				shortenRequest: &models.ShortenRequest{
					URL: fullURL,
				},
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusCreated,
			},
			hookBefore: func(args args, mock *mocks.Mock) {
				mock.App.EXPECT().ShortenURL(gomock.Any(), args.shortenRequest, gomock.Any()).Return(&models.ShortenResponse{
					Result: shortURL,
				}, nil)
				mock.AuditService.EXPECT().AuditEvent(gomock.Any()).DoAndReturn(func(event *domain.AuditEvent) {
					require.LessOrEqual(t, event.Timestamp, time.Now())
					require.Equal(t, domain.AuditActionShorten, event.Action)
					require.Equal(t, uuid.UUID{}, event.UserID)
					require.Equal(t, args.shortenRequest.URL, event.URL)
				})
			},
			hookAfter: func(result *http.Response) {
				var shortenResponse models.ShortenResponse
				err := json.NewDecoder(result.Body).Decode(&shortenResponse)
				require.NoError(t, err)

				actualShortURL, err := url.ParseRequestURI(string(shortenResponse.Result))
				require.NoError(t, err)
				require.Equal(t, shortURL, actualShortURL.String())
			},
		},
		{
			name: "GIVEN duplicate url WHEN valid request THEN conflict",
			args: args{
				shortenRequest: &models.ShortenRequest{
					URL: fullURL,
				},
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusConflict,
			},
			hookBefore: func(args args, mock *mocks.Mock) {
				mock.App.EXPECT().ShortenURL(gomock.Any(), args.shortenRequest, gomock.Any()).Return(&models.ShortenResponse{
					Result:      shortURL,
					IsDuplicate: true,
				}, nil)
			},
			hookAfter: func(result *http.Response) {
				var shortenResponse models.ShortenResponse
				err := json.NewDecoder(result.Body).Decode(&shortenResponse)
				require.NoError(t, err)

				actualShortURL, err := url.ParseRequestURI(string(shortenResponse.Result))
				require.NoError(t, err)
				require.Equal(t, shortURL, actualShortURL.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			var body io.Reader
			if tt.args.shortenRequest != nil {
				jsonRequest, err := json.Marshal(tt.args.shortenRequest)
				require.NoError(t, err)
				body = bytes.NewReader(jsonRequest)
			}

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			handler := NewAPIHandler(mock.App, mock.AuditService, mock.Logger)
			tt.hookBefore(tt.args, mock)

			// Act.
			handler.HandleAPIRequest(tt.args.responseWriter, request)

			// Assert.
			recorder, ok := tt.args.responseWriter.(*httptest.ResponseRecorder)
			if ok {
				result := recorder.Result()
				defer result.Body.Close()

				require.Equal(t, tt.want.statusCode, result.StatusCode)
				if tt.hookAfter != nil {
					tt.hookAfter(result)
				}
			}

			faultyWriter, ok := tt.args.responseWriter.(*utils.FaultyResponseWriter)
			if ok {
				require.Equal(t, tt.want.statusCode, faultyWriter.StatusCode)
			}
		})
	}
}

func TestAPIHandler_HandleAPIBatchRequest(t *testing.T) {
	t.Parallel()

	const (
		fullURL  = "http://foo.bar"
		shortURL = "http://localhost/abc3"
	)

	correlationID := uuid.NewString()

	type want struct {
		statusCode int
	}

	tests := []struct {
		name       string
		want       want
		hookBefore func(mock *mocks.Mock) (io.Reader, http.ResponseWriter)
		hookAfter  func(result *http.Response)
	}{
		{
			name: "WHEN invalid request json THEN bad request",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(mock *mocks.Mock) (io.Reader, http.ResponseWriter) {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return strings.NewReader("invalid json"), httptest.NewRecorder()
			},
		},
		{
			name: "WHEN not valid request THEN bad request",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(mock *mocks.Mock) (io.Reader, http.ResponseWriter) {
				items := []*models.ShortenBatchRequestItem{
					{},
				}
				jsonRequest, err := json.Marshal(items)
				require.NoError(t, err)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return bytes.NewReader(jsonRequest), httptest.NewRecorder()
			},
		},
		{
			name: "WHEN app error THEN internal error",
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) (io.Reader, http.ResponseWriter) {
				items := []*models.ShortenBatchRequestItem{
					{
						CorrelationID: correlationID,
						OriginalURL:   fullURL,
					},
				}
				jsonRequest, err := json.Marshal(items)
				require.NoError(t, err)
				mock.App.EXPECT().ShortenURLBatch(gomock.Any(), items, uuid.UUID{}).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return bytes.NewReader(jsonRequest), httptest.NewRecorder()
			},
		},
		{
			name: "WHEN response write error THEN internal error",
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) (io.Reader, http.ResponseWriter) {
				items := []*models.ShortenBatchRequestItem{
					{
						CorrelationID: correlationID,
						OriginalURL:   fullURL,
					},
				}
				jsonRequest, err := json.Marshal(items)
				require.NoError(t, err)
				mock.App.EXPECT().
					ShortenURLBatch(gomock.Any(), items, uuid.UUID{}).
					Return([]*models.ShortenBatchResponseItem{
						{
							CorrelationID: correlationID,
							ShortURL:      shortURL,
						},
					}, nil)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return bytes.NewReader(jsonRequest), utils.NewFaultyResponseWriter()
			},
		},
		{
			name: "WHEN no error THEN created",
			want: want{
				statusCode: http.StatusCreated,
			},
			hookBefore: func(mock *mocks.Mock) (io.Reader, http.ResponseWriter) {
				items := []*models.ShortenBatchRequestItem{
					{
						CorrelationID: correlationID,
						OriginalURL:   fullURL,
					},
				}
				jsonRequest, err := json.Marshal(items)
				require.NoError(t, err)
				mock.App.EXPECT().
					ShortenURLBatch(gomock.Any(), items, uuid.UUID{}).
					Return([]*models.ShortenBatchResponseItem{
						{
							CorrelationID: correlationID,
							ShortURL:      shortURL,
						},
					}, nil)
				return bytes.NewReader(jsonRequest), httptest.NewRecorder()
			},
			hookAfter: func(result *http.Response) {
				var responseItems []*models.ShortenBatchResponseItem
				err := json.NewDecoder(result.Body).Decode(&responseItems)
				require.NoError(t, err)

				actualShortURL, err := url.ParseRequestURI(responseItems[0].ShortURL)
				require.NoError(t, err)
				require.Equal(t, correlationID, responseItems[0].CorrelationID)
				require.Equal(t, shortURL, actualShortURL.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			body, responseWriter := tt.hookBefore(mock)
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			handler := NewAPIHandler(mock.App, mock.AuditService, mock.Logger)

			// Act.
			handler.HandleAPIBatchRequest(responseWriter, request)

			// Assert.
			recorder, ok := responseWriter.(*httptest.ResponseRecorder)
			if ok {
				result := recorder.Result()
				defer result.Body.Close()

				require.Equal(t, tt.want.statusCode, result.StatusCode)
				if tt.hookAfter != nil {
					tt.hookAfter(result)
				}
			}

			faultyWriter, ok := responseWriter.(*utils.FaultyResponseWriter)
			if ok {
				require.Equal(t, tt.want.statusCode, faultyWriter.StatusCode)
			}
		})
	}
}
