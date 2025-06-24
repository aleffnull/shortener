package store

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/aleffnull/shortener/internal/config"
)

type FileStore struct {
	configuration *config.FileStoreConfiguration
	mutex         sync.Mutex
}

var _ ColdStore = (*FileStore)(nil)

func NewFileStore(configuration *config.Configuration) ColdStore {
	return &FileStore{
		configuration: configuration.FileStore,
	}
}

func (fs *FileStore) LoadAll() ([]*ColdStoreEntry, error) {
	// Called only in main goroutine, so no need for mutex locking.

	if _, err := os.Stat(fs.configuration.FilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*ColdStoreEntry{}, nil
		} else {
			return nil, fmt.Errorf("LoadAll, os.Stat failed: %w", err)
		}
	}

	file, err := os.Open(fs.configuration.FilePath)
	if err != nil {
		return nil, fmt.Errorf("LoadAll, os.Open failed: %w", err)
	}

	defer file.Close()

	entries := []*ColdStoreEntry{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()
		entry := &ColdStoreEntry{}
		if err := json.Unmarshal(data, entry); err != nil {
			return nil, fmt.Errorf("LoadAll, json.Unmarshal failed: %w", err)
		}

		entries = append(entries, entry)
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("LoadAll, scanner.Err() returned an error: %w", err)
	}

	return entries, nil
}

func (fs *FileStore) Save(entry *ColdStoreEntry) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("Save, json.Marshal failed: %w", err)
	}

	file, err := os.OpenFile(fs.configuration.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Save, os.OpenFile failed: %w", err)
	}

	defer file.Close()

	data = append(data, '\n')
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Save, file.Write failed: %w", err)
	}

	return nil
}
