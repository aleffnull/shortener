package audit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEndpointReceiver_AddEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantError  bool
		hookBefore func(mock *mocks.Mock) *config.Configuration
	}{
		{
			name: "WHEN no audit url THEN do nothing",
			hookBefore: func(_ *mocks.Mock) *config.Configuration {
				return &config.Configuration{}
			},
		},
		{
			name:      "WHEN invalid audit url THEN error",
			wantError: true,
			hookBefore: func(_ *mocks.Mock) *config.Configuration {
				return &config.Configuration{
					AuditURL: ":::\\::",
				}
			},
		},
		{
			name: "WHEN no error THEN ok",
			hookBefore: func(mock *mocks.Mock) *config.Configuration {
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				t.Cleanup(server.Close)
				return &config.Configuration{
					AuditURL: server.URL,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			configuration := tt.hookBefore(mock)
			receiver := NewEndpointReceiver(configuration, mock.Logger)

			// Act.
			err := receiver.AddEvent(&domain.AuditEvent{
				Timestamp: domain.AuditFormattedTime(time.Date(2025, 1, 2, 18, 0, 0, 0, time.UTC)),
				Action:    domain.AuditActionShorten,
				UserID:    uuid.New(),
				URL:       "http://foo.bar",
			})

			// Assert.
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
