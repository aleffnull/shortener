package audit

import "github.com/aleffnull/shortener/internal/domain"

type Receiver interface {
	AddEvent(event *domain.AuditEvent) error
}
