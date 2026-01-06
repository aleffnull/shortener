package grpc

import (
	"context"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ShortenerService) ExpandURL(
	ctx context.Context,
	request *api.URLExpandRequest,
) (*api.URLExpandResponse, error) {
	key := request.GetId()
	if len(key) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Key is required")
	}

	item, err := s.shortener.GetURL(ctx, request.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	if item == nil {
		return nil, status.Errorf(codes.NotFound, "Key was not found")
	}

	if item.IsDeleted {
		return nil, status.Errorf(codes.NotFound, "Key was deleted")
	}

	s.auditService.AuditEvent(&domain.AuditEvent{
		Timestamp: domain.AuditFormattedTime(time.Now()),
		Action:    domain.AuditActionFollow,
		UserID:    item.UserID,
		URL:       item.URL,
	})

	return &api.URLExpandResponse{
		Result: item.URL,
	}, nil
}
