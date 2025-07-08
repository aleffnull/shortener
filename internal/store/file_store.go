package store

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/models"
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

func (s *FileStore) LoadAll() ([]*models.ColdStoreEntry, error) {
	// Called only during startup, so no need for mutex locking.

	if _, err := os.Stat(s.configuration.FilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*models.ColdStoreEntry{}, nil
		} else {
			return nil, fmt.Errorf("LoadAll, os.Stat failed: %w", err)
		}
	}

	file, err := os.Open(s.configuration.FilePath)
	if err != nil {
		return nil, fmt.Errorf("LoadAll, os.Open failed: %w", err)
	}

	defer file.Close()

	entries := []*models.ColdStoreEntry{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Bytes()
		entry := &models.ColdStoreEntry{}
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

func (s *FileStore) Save(entry *models.ColdStoreEntry) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("Save, json.Marshal failed: %w", err)
	}

	file, err := os.OpenFile(s.configuration.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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
