package models

type ShortenRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type ShortenResponse struct {
	Result      string `json:"result"`
	IsDuplicate bool   `json:"-"`
}
