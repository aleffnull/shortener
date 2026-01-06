package middleware

import (
	"context"
	"testing"

	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUserIDInterceptor(t *testing.T) {
	t.Parallel()

	const authToken = "auth"
	userID := uuid.New()

	tests := []struct {
		name       string
		hookBefore func(mock *mocks.Mock) (context.Context, grpc.UnaryHandler)
		hookAfter  func(err error)
	}{
		{
			name: "WHEN no metadata in context THEN internal error",
			hookBefore: func(mock *mocks.Mock) (context.Context, grpc.UnaryHandler) {
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				return context.Background(), nil
			},
			hookAfter: func(err error) {
				code, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, code.Code())
			},
		},
		{
			name: "WHEN no authorization info in metadata THEN invalid argument error",
			hookBefore: func(mock *mocks.Mock) (context.Context, grpc.UnaryHandler) {
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				return metadata.NewIncomingContext(context.Background(), metadata.MD{}), nil
			},
			hookAfter: func(err error) {
				code, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, code.Code())
			},
		},
		{
			name: "WHEN authorization service error THEN internal error",
			hookBefore: func(mock *mocks.Mock) (context.Context, grpc.UnaryHandler) {
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(authToken).Return(uuid.UUID{}, assert.AnError)
				md := metadata.New(map[string]string{useIDAuthorizationMetadataKey: authToken})
				return metadata.NewIncomingContext(context.Background(), md), nil
			},
			hookAfter: func(err error) {
				code, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, code.Code())
			},
		},
		{
			name: "WHEN no errors THEN ok",
			hookBefore: func(mock *mocks.Mock) (context.Context, grpc.UnaryHandler) {
				mock.Logger.EXPECT().Infof(gomock.Any(), gomock.Any())
				mock.AuthorizationService.EXPECT().GetUserIDFromToken(authToken).Return(userID, nil)

				md := metadata.New(map[string]string{useIDAuthorizationMetadataKey: authToken})
				ctx := metadata.NewIncomingContext(context.Background(), md)

				handler := func(ctx context.Context, request any) (any, error) {
					require.Equal(t, userID, ctx.Value(userIDContextKey))
					return request, nil
				}

				return ctx, handler
			},
			hookAfter: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			ctrl := gomock.NewController(t)
			mock := mocks.NewMock(ctrl)
			ctx, handler := tt.hookBefore(mock)

			// Act-assert.
			_, err := UserIDInterceptor(
				ctx,
				nil,
				&grpc.UnaryServerInfo{},
				handler,
				mock.AuthorizationService,
				mock.Logger,
			)
			tt.hookAfter(err)
		})
	}
}
