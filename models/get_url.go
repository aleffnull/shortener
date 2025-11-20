package models

import "github.com/google/uuid"

type GetURLResponseItem struct {
	URL       string
	UserID    uuid.UUID
	IsDeleted bool
}
