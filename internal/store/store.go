package store

type Store interface {
	Load(key string) (string, bool)
	Save(value string) (string, error)
	PreSave(key, value string)
}

type ColdStoreEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ColdStore interface {
	LoadAll() ([]*ColdStoreEntry, error)
	Save(entry *ColdStoreEntry) error
}
