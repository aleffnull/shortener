package grpc

import (
	"context"
	"time"

	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"github.com/aleffnull/shortener/models"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ShortenerService) ShortenURL(
	ctx context.Context,
	request *api.URLShortenRequest,
) (*api.URLShortenResponse, error) {
	shortenRequest := &models.ShortenRequest{
		URL: request.GetUrl(),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(shortenRequest); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	shortenerResponse, err := s.shortener.ShortenURL(ctx, shortenRequest, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err.Error())
	}

	if !shortenerResponse.IsDuplicate {
		s.auditService.AuditEvent(&domain.AuditEvent{
			Timestamp: domain.AuditFormattedTime(time.Now()),
			Action:    domain.AuditActionShorten,
			UserID:    userID,
			URL:       shortenRequest.URL,
		})
	}

	return &api.URLShortenResponse{
		Result: shortenerResponse.Result,
	}, nil
}
