package models

type ColdStoreEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type BatchRequestItem struct {
	CorelationID string
	OriginalURL  string
}

type BatchResponseItem struct {
	CorelationID string
	Key          string
}

type KeyOriginalURLItem struct {
	URLKey      string
	OriginalURL string
}

type URLItem struct {
	URL       string
	IsDeleted bool
}
