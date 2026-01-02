package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuditAction int

const (
	AudotActionUnknown AuditAction = iota
	AuditActionShorten
	AuditActionFollow
)

type AuditFormattedTime time.Time

type AuditEvent struct {
	Timestamp AuditFormattedTime `json:"ts"`
	Action    AuditAction        `json:"action"`
	UserID    uuid.UUID          `json:"user_id"`
	URL       string             `json:"url"`
}

func (action AuditAction) MarshalJSON() ([]byte, error) {
	var str string
	switch action {
	case AudotActionUnknown:
		str = "unknown"
	case AuditActionShorten:
		str = "shorten"
	case AuditActionFollow:
		str = "follow"
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}

	return fmt.Appendf(nil, "\"%v\"", str), nil
}

func (t AuditFormattedTime) MarshalJSON() ([]byte, error) {
	unix := time.Time(t).Unix()
	return fmt.Appendf(nil, "\"%v\"", unix), nil
}
