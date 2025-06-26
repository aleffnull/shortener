package config

import "fmt"

type FileStoreConfiguration struct {
	FilePath string `env:"FILE_STORAGE_PATH" validate:"required"`
}

func (c *FileStoreConfiguration) String() string {
	return fmt.Sprintf("&FileStoreConfiguration{FilePath:%v}", c.FilePath)
}
