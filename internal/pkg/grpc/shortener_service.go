package grpc

import (
	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"github.com/aleffnull/shortener/internal/service"
)

type ShortenerService struct {
	shortener    app.App
	auditService service.AuditService
	api.UnimplementedShortenerServiceServer
}

var _ api.ShortenerServiceServer = (*ShortenerService)(nil)

func NewShortenerService(shortener app.App, auditService service.AuditService) *ShortenerService {
	return &ShortenerService{
		shortener:    shortener,
		auditService: auditService,
	}
}
