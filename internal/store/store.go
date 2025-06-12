package store

type Store interface {
	Load(key string) (string, bool)
	Save(value string) (string, error)
	Set(key, value string)
}
