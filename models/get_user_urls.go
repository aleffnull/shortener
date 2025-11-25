package models

// UserURLsResponseItem ответ на запрос получения всех URL пользователя.
type UserURLsResponseItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
