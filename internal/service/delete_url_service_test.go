package service

import (
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeleteURLsService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hookBefore func(mock *mocks.Mock) *domain.DeleteURLsRequest
	}{
		{
			name: "WHEN no requests THEN ok",
			hookBefore: func(_ *mocks.Mock) *domain.DeleteURLsRequest {
				return nil
			},
		},
		{
			name: "WHEN storage error THEN logged",
			hookBefore: func(mock *mocks.Mock) *domain.DeleteURLsRequest {
				request := &domain.DeleteURLsRequest{
					Keys:   []string{"foo"},
					UserID: uuid.New(),
				}
				mock.Store.EXPECT().DeleteBatch(gomock.Any(), request.Keys, request.UserID).Return(assert.AnError)
				mock.Logger.EXPECT().Errorf(gomock.Any(), gomock.Any())
				return request
			},
		},
		{
			name: "WHEN no error THEN ok",
			hookBefore: func(mock *mocks.Mock) *domain.DeleteURLsRequest {
				request := &domain.DeleteURLsRequest{
					Keys:   []string{"foo"},
					UserID: uuid.New(),
				}
				mock.Store.EXPECT().DeleteBatch(gomock.Any(), request.Keys, request.UserID).Return(nil)
				return request
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			request := tt.hookBefore(mock)
			service := NewDeleteURLsService(mock.Store, mock.Logger)

			// Act.
			service.Init()
			if request != nil {
				service.Delete(request)
			}

			// Ждем, пока отработает удаление по таймеру.
			time.Sleep(4 * time.Second)
			service.Shutdown()

			// Ждем завершения горутины удаления.
			time.Sleep(500 * time.Millisecond)
		})
	}
}
