package models

// ShortenBatchRequestItem элемент пакетного запроса на сокращение URL.
type ShortenBatchRequestItem struct {
	CorrelationID string `json:"correlation_id" validate:"required"`
	OriginalURL   string `json:"original_url" validate:"required,url"`
}

// ShortenBatchResponseItem элемент ответа на пакетный запрос сокращения URL.
type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
