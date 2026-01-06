package grpc

import (
	"context"

	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"github.com/aleffnull/shortener/models"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ShortenerService) ListUserURLs(
	ctx context.Context,
	_ *emptypb.Empty,
) (*api.UserURLsResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	items, err := s.shortener.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	urlData := lo.Map(items, func(item *models.UserURLsResponseItem, _ int) *api.URLData {
		return &api.URLData{
			ShortUrl:    item.ShortURL,
			OriginalUrl: item.OriginalURL,
		}
	})

	return &api.UserURLsResponse{
		Url: urlData,
	}, nil
}
