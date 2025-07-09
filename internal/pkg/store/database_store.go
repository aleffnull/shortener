package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DatabaseStore struct {
	keyStore
	configuration *config.DatabaseStoreConfiguration
	logger        logger.Logger
	db            *sql.DB
}

var _ Store = (*DatabaseStore)(nil)

func NewDatabaseStore(configuration *config.Configuration, logger logger.Logger) Store {
	store := &DatabaseStore{
		keyStore: keyStore{
			configuration: &configuration.DatabaseStore.KeyStoreConfiguration,
		},
		configuration: configuration.DatabaseStore,
		logger:        logger,
	}
	store.keyStore.saver = store.saver

	return store
}

func (s *DatabaseStore) Init() error {
	if !s.configuration.IsDatabaseEnabled() {
		s.logger.Infof("Database store is disabled")
		return nil
	}

	db, err := sql.Open("pgx", s.configuration.DataSourceName)
	if err != nil {
		return fmt.Errorf("Init, sql.Open failed: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("Init, postgres.WithInstance failed: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("Init, migrate.NewWithDatabaseInstance failed: %w", err)
	}

	m.Log = logger.NewMigrateLogger(s.logger)
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Init, m.Up failed: %w", err)
	}

	s.db = db
	s.logger.Infof("Database store initialized")

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

func (s *DatabaseStore) Load(ctx context.Context, key string) (string, bool, error) {
	row := s.db.QueryRowContext(
		ctx,
		"select origin_url from urls where url_key = $1",
		key,
	)

	var url string
	if err := row.Scan(&url); err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}

		return "", false, fmt.Errorf("Load, row.Scan failed: %w", err)
	}

	return url, true, nil
}

func (s *DatabaseStore) Save(ctx context.Context, value string) (string, error) {
	key, err := s.saveWithUniqueKey(ctx, value)
	if err != nil {
		return "", fmt.Errorf("DatabaseStore.Save, saveWithUniqueKey failed: %w", err)
	}

	return key, nil
}

func (s *DatabaseStore) saver(ctx context.Context, key, value string) (bool, error) {
	result, err := s.db.ExecContext(
		ctx,
		"insert into urls (url_key, origin_url) values ($1, $2)",
		key, value,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Unique key violation
			return true, nil
		}

		return false, fmt.Errorf("saver, db.Exec failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("saver, result.RowsAffected failed: %w", err)
	}
	if rowsAffected != 1 {
		return false, errors.New("saver, failed to insert data")
	}

	return false, nil
}
