package models

// ShortenRequest Запрос на сокращение URL.
type ShortenRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// ShortenResponse сокращенный URL.
type ShortenResponse struct {
	Result      string `json:"result"`
	IsDuplicate bool   `json:"-"`
}
