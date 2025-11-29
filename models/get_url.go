package models

import "github.com/google/uuid"

// GetURLResponseItem ответ на запрос URL.
type GetURLResponseItem struct {
	URL       string
	UserID    uuid.UUID
	IsDeleted bool
}
