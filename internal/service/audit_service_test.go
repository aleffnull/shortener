package service

import (
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/audit"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type auditReceiver struct {
	events      []*domain.AuditEvent
	returnError bool
}

func (r *auditReceiver) AddEvent(event *domain.AuditEvent) error {
	r.events = append(r.events, event)
	if r.returnError {
		return assert.AnError
	}

	return nil
}

func TestAuditService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hookBefore func(mock *mocks.Mock) *auditReceiver
	}{
		{
			name: "WHEN error in receiver THEN error logged",
			hookBefore: func(mock *mocks.Mock) *auditReceiver {
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(21)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any())
				return &auditReceiver{
					returnError: true,
				}
			},
		},
		{
			name: "WHEN no errors THEN ok",
			hookBefore: func(mock *mocks.Mock) *auditReceiver {
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(21)
				return &auditReceiver{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			receiver := tt.hookBefore(mock)
			service := NewAuditService([]audit.Receiver{receiver}, mock.Logger)
			event := &domain.AuditEvent{
				Timestamp: domain.AuditFormattedTime(time.Date(2025, 1, 2, 18, 0, 0, 0, time.UTC)),
				Action:    domain.AuditActionFollow,
				UserID:    uuid.New(),
				URL:       "http://foo.bar",
			}

			// Act.
			service.Init()
			service.AuditEvent(event)

			// Ждем, пока отработает receiver и logger.
			time.Sleep(500 * time.Millisecond)
			service.Shutdown()

			// Ждем остановки всех воркеров.
			time.Sleep(500 * time.Millisecond)

			// Assert.
			require.Len(t, receiver.events, 1)
			require.Equal(t, event, receiver.events[0])
		})
	}
}
