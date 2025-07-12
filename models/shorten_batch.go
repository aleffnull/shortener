package models

type ShortenBatchRequestItem struct {
	CorelationID string `json:"correlation_id" validate:"required"`
	OriginalURL  string `json:"original_url" validate:"required,url"`
}

type ShortenBatchResponseItem struct {
	CorelationID string `json:"correlation_id"`
	ShortURL     string `json:"short_url"`
}
