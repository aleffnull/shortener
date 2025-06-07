package store

type Store interface {
	Load(key string) (string, bool)
	Save(value string) string
	Set(key, value string)
}
