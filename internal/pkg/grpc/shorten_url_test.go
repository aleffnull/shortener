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

func TestShortenURL(t *testing.T) {
	t.Parallel()

	const (
		fullURL  = "http://foo.bar"
		shortURL = "http://localhost/abc"
	)

	type want struct {
		code     *codes.Code
		response *api.URLShortenResponse
	}

	tests := []struct {
		name       string
		want       *want
		hookBefore func(mock *mocks.Mock) *api.URLShortenRequest
	}{
		{
			name: "WHEN validation error THEN invalid argument error",
			want: &want{
				code: lo.ToPtr(codes.InvalidArgument),
			},
			hookBefore: func(_ *mocks.Mock) *api.URLShortenRequest {
				return &api.URLShortenRequest{}
			},
		},
		{
			name: "WHEN app error THEN internal error",
			want: &want{
				code: lo.ToPtr(codes.Internal),
			},
			hookBefore: func(mock *mocks.Mock) *api.URLShortenRequest {
				mock.App.EXPECT().ShortenURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
				return &api.URLShortenRequest{
					Url: fullURL,
				}
			},
		},
		{
			name: "GIVEN duplicate url WHEN no errors THEN ok",
			want: &want{
				response: &api.URLShortenResponse{
					Result: shortURL,
				},
			},
			hookBefore: func(mock *mocks.Mock) *api.URLShortenRequest {
				mock.App.EXPECT().
					ShortenURL(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&models.ShortenResponse{
						Result:      shortURL,
						IsDuplicate: true,
					}, nil)
				return &api.URLShortenRequest{
					Url: fullURL,
				}
			},
		},
		{
			name: "GIVEN unique url WHEN no errors THEN ok",
			want: &want{
				response: &api.URLShortenResponse{
					Result: shortURL,
				},
			},
			hookBefore: func(mock *mocks.Mock) *api.URLShortenRequest {
				mock.App.EXPECT().
					ShortenURL(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&models.ShortenResponse{
						Result: shortURL,
					}, nil)
				mock.AuditService.EXPECT().AuditEvent(gomock.Any()).DoAndReturn(func(event *domain.AuditEvent) {
					require.LessOrEqual(t, event.Timestamp, time.Now())
					require.Equal(t, domain.AuditActionShorten, event.Action)
					require.Equal(t, uuid.UUID{}, event.UserID)
					require.Equal(t, fullURL, event.URL)
				})
				return &api.URLShortenRequest{
					Url: fullURL,
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
			response, err := service.ShortenURL(context.Background(), request)
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
