package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"github.com/aleffnull/shortener/models"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestExpandURL(t *testing.T) {
	t.Parallel()

	const (
		fullURL = "http://foo.bar"
		key     = "abc"
	)

	type want struct {
		code     *codes.Code
		response *api.URLExpandResponse
	}

	tests := []struct {
		name       string
		want       *want
		hookBefore func(mock *mocks.Mock) *api.URLExpandRequest
	}{
		{
			name: "WHEN no key in request THEN invalid argument error",
			want: &want{
				code: lo.ToPtr(codes.InvalidArgument),
			},
			hookBefore: func(_ *mocks.Mock) *api.URLExpandRequest {
				return &api.URLExpandRequest{}
			},
		},
		{
			name: "WHEN app error THEN internal error",
			want: &want{
				code: lo.ToPtr(codes.Internal),
			},
			hookBefore: func(mock *mocks.Mock) *api.URLExpandRequest {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(nil, assert.AnError)
				return &api.URLExpandRequest{
					Id: key,
				}
			},
		},
		{
			name: "WHEN nil item THEN not found error",
			want: &want{
				code: lo.ToPtr(codes.NotFound),
			},
			hookBefore: func(mock *mocks.Mock) *api.URLExpandRequest {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(nil, nil)
				return &api.URLExpandRequest{
					Id: key,
				}
			},
		},
		{
			name: "WHEN item deleted THEN not found error",
			want: &want{
				code: lo.ToPtr(codes.NotFound),
			},
			hookBefore: func(mock *mocks.Mock) *api.URLExpandRequest {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(&models.GetURLResponseItem{
					IsDeleted: true,
				}, nil)
				return &api.URLExpandRequest{
					Id: key,
				}
			},
		},
		{
			name: "WHEN error THEN ok",
			want: &want{
				response: &api.URLExpandResponse{
					Result: fullURL,
				},
			},
			hookBefore: func(mock *mocks.Mock) *api.URLExpandRequest {
				mock.App.EXPECT().GetURL(gomock.Any(), key).Return(&models.GetURLResponseItem{
					URL: fullURL,
				}, nil)
				mock.AuditService.EXPECT().AuditEvent(gomock.Any()).DoAndReturn(func(event *domain.AuditEvent) {
					require.LessOrEqual(t, event.Timestamp, time.Now())
					require.Equal(t, domain.AuditActionFollow, event.Action)
					require.Equal(t, uuid.UUID{}, event.UserID)
					require.Equal(t, fullURL, event.URL)
				})
				return &api.URLExpandRequest{
					Id: key,
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
			request := tt.hookBefore(mock)
			service := NewShortenerService(mock.App, mock.AuditService)

			// Act-assert.
			response, err := service.ExpandURL(context.Background(), request)
			if tt.want.code == nil {
				require.NoError(t, err)
				require.Equal(t, tt.want.response, response)
			} else {
				require.Error(t, err)
				code, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, *tt.want.code, code.Code())
			}
		})
	}
}
