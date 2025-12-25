package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/utils"
	"github.com/aleffnull/shortener/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_HandleGetUserURLsRequest(t *testing.T) {
	t.Parallel()

	const (
		fullURL  = "http://foo.bar"
		shortURL = "http://localhost/abc3"
	)

	type args struct {
		responseWriter http.ResponseWriter
	}

	type want struct {
		statusCode int
	}

	tests := []struct {
		name       string
		args       args
		want       want
		hookBefore func(mock *mocks.Mock)
		hookAfter  func(result *http.Response)
	}{
		{
			name: "WHEN app error THEN internal error",
			args: args{
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN no items THEN no content",
			args: args{
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusNoContent,
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return([]*models.UserURLsResponseItem{}, nil)
			},
		},
		{
			name: "WHEN response write error THEN internal error",
			args: args{
				responseWriter: utils.NewFaultyResponseWriter(),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return([]*models.UserURLsResponseItem{
					{
						ShortURL:    shortURL,
						OriginalURL: fullURL,
					},
				}, nil)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN no errors THEN ok",
			args: args{
				responseWriter: httptest.NewRecorder(),
			},
			want: want{
				statusCode: http.StatusOK,
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return([]*models.UserURLsResponseItem{
					{
						ShortURL:    shortURL,
						OriginalURL: fullURL,
					},
				}, nil)
			},
			hookAfter: func(result *http.Response) {
				var items []*models.UserURLsResponseItem
				err := json.NewDecoder(result.Body).Decode(&items)
				require.NoError(t, err)

				require.Len(t, items, 1)
				require.Equal(t, shortURL, items[0].ShortURL)
				require.Equal(t, fullURL, items[0].OriginalURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			tt.hookBefore(mock)

			handler := NewUserHandler(mock.App, mock.Logger)
			request := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

			// Act.
			handler.HandleGetUserURLsRequest(tt.args.responseWriter, request)

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

func TestUserHandler_HandleBatchDeleteRequest(t *testing.T) {
	t.Parallel()

	type want struct {
		statusCode int
	}

	tests := []struct {
		name       string
		want       want
		hookBefore func(mock *mocks.Mock) io.Reader
	}{
		{
			name: "WHEN read body error THEN bad request",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			hookBefore: func(mock *mocks.Mock) io.Reader {
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return utils.AFaultyReader
			},
		},
		{
			name: "WHEN no errors THEN ok",
			want: want{
				statusCode: http.StatusAccepted,
			},
			hookBefore: func(mock *mocks.Mock) io.Reader {
				keys := []string{"foo", "bar"}
				mock.App.EXPECT().DeleteURLs(keys, gomock.Any())
				data, err := json.Marshal(keys)
				require.NoError(t, err)
				return bytes.NewReader(data)
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
			handler := NewUserHandler(mock.App, mock.Logger)
			request := httptest.NewRequest(http.MethodDelete, "/api/user/urls", body)

			// Act.
			handler.HandleBatchDeleteRequest(recorder, request)

			// Assert.
			result := recorder.Result()
			defer result.Body.Close()

			require.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}
