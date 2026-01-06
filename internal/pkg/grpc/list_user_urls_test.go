package grpc

import (
	"context"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"github.com/aleffnull/shortener/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestListUserURLs(t *testing.T) {
	t.Parallel()

	const (
		fullURL  = "http://foo.bar"
		shortURL = "http://localhost/abc3"
	)

	type want struct {
		code     *codes.Code
		response *api.UserURLsResponse
	}

	tests := []struct {
		name       string
		want       *want
		hookBefore func(mock *mocks.Mock)
	}{
		{
			name: "WHEN app error THEN internal error",
			want: &want{
				code: lo.ToPtr(codes.Internal),
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
			},
		},
		{
			name: "WHEN no errors THEN ok",
			want: &want{
				response: &api.UserURLsResponse{
					Url: []*api.URLData{
						{
							ShortUrl:    shortURL,
							OriginalUrl: fullURL,
						},
					},
				},
			},
			hookBefore: func(mock *mocks.Mock) {
				mock.App.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return([]*models.UserURLsResponseItem{
					{
						ShortURL:    shortURL,
						OriginalURL: fullURL,
					},
				}, nil)
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
			service := NewShortenerService(mock.App, mock.AuditService)

			// Act-assert.
			response, err := service.ListUserURLs(context.Background(), &emptypb.Empty{})
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
