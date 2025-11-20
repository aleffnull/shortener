package audit

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Action int

const (
	ActionUnknown Action = iota
	ActionShorten
	ActionFollow
)

type FormattedTime time.Time

type Event struct {
	Timestamp FormattedTime `json:"ts"`
	Action    Action        `json:"action"`
	UserID    uuid.UUID     `json:"user_id"`
	URL       string        `json:"url"`
}

func (action Action) MarshalJSON() ([]byte, error) {
	var str string
	switch action {
	case ActionUnknown:
		str = "unknown"
	case ActionShorten:
		str = "shorten"
	case ActionFollow:
		str = "follow"
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}

	return fmt.Appendf(nil, "\"%v\"", str), nil
}

func (t FormattedTime) MarshalJSON() ([]byte, error) {
	unix := time.Time(t).Unix()
	return fmt.Appendf(nil, "\"%v\"", unix), nil
}
