package app

import (
	"encoding/json"
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

func TestInternalHandler_HandleStatsRequest(t *testing.T) {
	t.Parallel()

	type args struct {
		responseWriter http.ResponseWriter
	}
	type want struct {
		statusCode int
	}

	tests := []struct {
		name       string
		args       *args
		want       *want
		hookBefore func(mock *mocks.Mock)
		hookAfter  func(result *http.Response)
	}{
		{
			name: "WHEN app error THEN internal error",
			args: &args{
				responseWriter: httptest.NewRecorder(),
			},
			want: &want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetStatistics(gomock.Any()).Return(nil, assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN response write error THEN internal error",
			args: &args{
				responseWriter: utils.NewFaultyResponseWriter(),
			},
			want: &want{
				statusCode: http.StatusInternalServerError,
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetStatistics(gomock.Any()).Return(&models.Statistics{
					UrlsCount:  1,
					UsersCount: 2,
				}, nil)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name: "WHEN no errors THEN ok",
			args: &args{
				responseWriter: httptest.NewRecorder(),
			},
			want: &want{
				statusCode: http.StatusOK,
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetStatistics(gomock.Any()).Return(&models.Statistics{
					UrlsCount:  1,
					UsersCount: 2,
				}, nil)
			},
			hookAfter: func(result *http.Response) {
				var statistics models.Statistics
				err := json.NewDecoder(result.Body).Decode(&statistics)
				require.NoError(t, err)
				require.Equal(t, 1, statistics.UrlsCount)
				require.Equal(t, 2, statistics.UsersCount)
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
			handler := NewInternalHandler(mock.App, mock.Logger)
			request := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)

			// Act.
			handler.HandleStatsRequest(tt.args.responseWriter, request)

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
