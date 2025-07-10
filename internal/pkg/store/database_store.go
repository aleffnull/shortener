package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aleffnull/shortener/internal/config"
	pkg_errors "github.com/aleffnull/shortener/internal/pkg/errors"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/samber/lo"
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
		"select original_url from urls where url_key = $1",
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
	key, err := s.saveWithUniqueKey(ctx, value, func(ctx context.Context, key, value string) (bool, error) {
		return s.saver(ctx, s.db, nil, key, value)
	})

	if err != nil {
		var duplicateURLError *pkg_errors.DuplicateURLError
		if errors.As(err, &duplicateURLError) {
			existingKey, err := s.getExistingKeyByValue(ctx, value)
			if err != nil {
				return "", fmt.Errorf("Save, getExistingKeyByValue error: %w", err)
			}
			duplicateURLError.Key = existingKey
			return "", duplicateURLError
		} else {
			return "", fmt.Errorf("Save, doInTx failed: %w", err)
		}
	}

	return key, nil
}

func (s *DatabaseStore) SaveBatch(ctx context.Context, requestItems []*models.BatchRequestItem) ([]*models.BatchResponseItem, error) {
	responseItems, err := doInTx(ctx, s.db, func(ctx context.Context, tx *sql.Tx) ([]*models.BatchResponseItem, error) {
		responseItems := make([]*models.BatchResponseItem, 0, len(requestItems))
		for _, requestItem := range requestItems {
			key, err := s.saveWithUniqueKey(ctx, requestItem.OriginalURL, func(ctx context.Context, key, value string) (bool, error) {
				return s.saver(ctx, nil, tx, key, value)
			})
			if err != nil {
				return nil, fmt.Errorf("SaveBatch, saveWithUniqueKey failed: %w", err)
			}
			responseItems = append(responseItems, &models.BatchResponseItem{
				CorelationID: requestItem.CorelationID,
				Key:          key,
			})
		}
		return responseItems, nil
	})

	if err != nil {
		return nil, fmt.Errorf("SaveBatch, doInTx failed: %w", err)
	}

	return responseItems, nil
}

func (s *DatabaseStore) saver(ctx context.Context, db *sql.DB, tx *sql.Tx, key, value string) (bool, error) {
	executor := lo.Ternary(tx == nil, db.ExecContext, tx.ExecContext)
	result, err := executor(
		ctx,
		"insert into urls (url_key, original_url) values ($1, $2)",
		key, value,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "urls_original_url_unique" {
				return false, pkg_errors.NewDuplicateURLError("", value)
			}
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

func (s *DatabaseStore) getExistingKeyByValue(ctx context.Context, value string) (string, error) {
	row := s.db.QueryRowContext(
		ctx,
		"select url_key from urls where original_url = $1",
		value,
	)

	var key string
	if err := row.Scan(&key); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("getExistingKeyByValue, row.Scan: failed to find existing key %v", key)
		}

		return "", fmt.Errorf("getExistingKeyByValue, row.Scan failed: %w", err)
	}

	return key, nil
}

func doInTx[T any](ctx context.Context, db *sql.DB, worker func(context.Context, *sql.Tx) (T, error)) (T, error) {
	var emptyResult T

	tx, err := db.Begin()
	if err != nil {
		return emptyResult, fmt.Errorf("doInTx, db.Begin failed: %w", err)
	}

	result, err := worker(ctx, tx)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, fmt.Errorf("doInTx, tx.Rollback failed: %w", txErr))
		}
		return emptyResult, fmt.Errorf("doInTx, saveWithUniqueKey failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return emptyResult, fmt.Errorf("doInTx, tx.Commit failed: %w", err)
	}

	return result, nil
}
