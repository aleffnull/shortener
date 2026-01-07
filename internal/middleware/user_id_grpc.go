package middleware

import (
	"context"

	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UserIDInterceptor(
	ctx context.Context,
	request any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
	authorizationService service.AuthorizationService,
	log logger.Logger,
) (any, error) {
	log.Infof("Got GRPC request to %v", info.FullMethod)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "No metadata in incoming context")
	}

	auth := md.Get(useIDAuthorizationMetadataKey)
	if len(auth) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "No authorization token in incoming context")
	}

	userID, err := authorizationService.GetUserIDFromToken(auth[0])
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user ID from authorization token: %v", err)
	}

	userIDContext := context.WithValue(ctx, userIDContextKey, userID)
	return handler(userIDContext, request)
}
