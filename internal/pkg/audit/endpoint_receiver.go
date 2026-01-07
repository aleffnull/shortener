package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ldez/mimetype"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/domain"
	"github.com/aleffnull/shortener/internal/pkg/logger"
)

type EndpointReceiver struct {
	auditURL string
	logger   logger.Logger
}

var _ Receiver = (*EndpointReceiver)(nil)

func NewEndpointReceiver(configuration *config.Configuration, logger logger.Logger) *EndpointReceiver {
	return &EndpointReceiver{
		auditURL: configuration.AuditURL,
		logger:   logger,
	}
}

func (r *EndpointReceiver) AddEvent(event *domain.AuditEvent) error {
	if len(r.auditURL) == 0 {
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("EndpointReceiver.AddEvent, json.Marshal failed: %w", err)
	}

	response, err := http.Post(r.auditURL, mimetype.ApplicationJSON, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("EndpointReceiver.AddEvent, http.Post failed: %w", err)
	}

	defer response.Body.Close()
	r.logger.Infof("Audit endpoint response status code: %v", response.StatusCode)

	return nil
}
