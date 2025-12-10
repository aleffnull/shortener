package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
)

func TestMaintenanceHandler_HandlePingRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name:       "WHEN database error THEN internal error",
			statusCode: http.StatusInternalServerError,
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().CheckStore(gomock.Any()).Return(assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
			},
		},
		{
			name:       "WHEN no errors THEN ok",
			statusCode: http.StatusOK,
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().CheckStore(gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/ping", nil)
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			handler := NewMaintenanceHandler(mock.App, mock.Logger)
			tt.hookBefore(mock)

			// Act.
			handler.HandlePingRequest(recorder, request)

			// Assert.
			result := recorder.Result()
			require.Equal(t, tt.statusCode, result.StatusCode)
			defer result.Body.Close()
		})
	}
}
