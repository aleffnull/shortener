package domain

import "github.com/google/uuid"

type ColdStoreEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type BatchRequestItem struct {
	CorrelationID string
	OriginalURL   string
}

type BatchResponseItem struct {
	CorrelationID string
	Key           string
}

type KeyOriginalURLItem struct {
	URLKey      string
	OriginalURL string
}

type URLItem struct {
	URL       string
	UserID    uuid.UUID
	IsDeleted bool
}

type DeleteURLsRequest struct {
	Keys   []string
	UserID uuid.UUID
}
