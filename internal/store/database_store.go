package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aleffnull/shortener/internal/config"
)

type DatabaseStore struct {
	configuration *config.DatabaseStoreConfiguration
	db            *sql.DB
}

var _ Store = (*DatabaseStore)(nil)

func NewDatabaseStore(configuration *config.Configuration) Store {
	return &DatabaseStore{
		configuration: configuration.DatabaseStore,
	}
}

func (s *DatabaseStore) Init() error {
	if !s.configuration.IsDatabaseEnabled() {
		return nil
	}

	db, err := sql.Open("pgx", s.configuration.DataSourceName)
	if err != nil {
		return fmt.Errorf("initDatabase, sql.Open failed: %w", err)
	}

	s.db = db
	return nil
}

func (s *DatabaseStore) Shutdown() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *DatabaseStore) CheckAvailability(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("database not available: %w", err)
	}

	return nil
}

func (s *DatabaseStore) Load(key string) (string, bool) {
	return "", false
}

func (s *DatabaseStore) Save(value string) (string, error) {
	return "", nil
}
