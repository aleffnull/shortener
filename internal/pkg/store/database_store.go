package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/database"
	pkg_errors "github.com/aleffnull/shortener/internal/pkg/errors"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseStore struct {
	keyStore
	connection    database.Connection
	configuration *config.DatabaseStoreConfiguration
	logger        logger.Logger
}

type executorFunc func(ctx context.Context, sql string, args ...any) error

var _ Store = (*DatabaseStore)(nil)

func NewDatabaseStore(connection database.Connection, configuration *config.Configuration, logger logger.Logger) Store {
	store := &DatabaseStore{
		keyStore: keyStore{
			configuration: &configuration.DatabaseStore.KeyStoreConfiguration,
		},
		connection:    connection,
		configuration: configuration.DatabaseStore,
		logger:        logger,
	}

	return store
}

func (s *DatabaseStore) Init() error {
	return nil
}

func (s *DatabaseStore) Shutdown() {
	//
}

func (s *DatabaseStore) CheckAvailability(ctx context.Context) error {
	err := s.connection.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database not available: %w", err)
	}

	return nil
}

func (s *DatabaseStore) Load(ctx context.Context, key string) (*models.URLItem, error) {
	rows, err := s.connection.QueryRows(
		ctx,
		"select original_url, is_deleted from urls where url_key = $1",
		key,
	)
	if err != nil {
		return nil, fmt.Errorf("DatabaseStore.Load, connection.QueryRows failed: %w", err)
	}

	defer rows.Close()

	var item *models.URLItem
	for rows.Next() {
		item = &models.URLItem{}
		err = rows.Scan(&item.URL, &item.IsDeleted)
		if err != nil {
			return nil, fmt.Errorf("DatabaseStore.Load, rows.Scan failed: %w", err)
		}

		// It should be only one item.
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("DatabaseStore.Load, rows.Err failed: %w", err)
	}

	return item, nil
}

func (s *DatabaseStore) LoadAllByUserID(ctx context.Context, userID uuid.UUID) ([]*models.KeyOriginalURLItem, error) {
	rows, err := s.connection.QueryRows(
		ctx,
		"select url_key, original_url from urls where user_id = $1",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("DatabaseStore.LoadAllByUserID, connection.QueryRows failed: %w", err)
	}

	defer rows.Close()

	items := []*models.KeyOriginalURLItem{}
	for rows.Next() {
		item := &models.KeyOriginalURLItem{}
		err = rows.Scan(&item.URLKey, &item.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("DatabaseStore.LoadAllByUserID, rows.Scan failed: %w", err)
		}

		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("DatabaseStore.LoadAllByUserID, rows.Err failed: %w", err)
	}

	return items, nil
}

func (s *DatabaseStore) Save(ctx context.Context, value string, userID uuid.UUID) (string, error) {
	key, err := s.saveWithUniqueKey(ctx, value, func(ctx context.Context, key, value string) (bool, error) {
		return s.saver(ctx, s.connection.Exec, key, value, userID)
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

func (s *DatabaseStore) SaveBatch(ctx context.Context, requestItems []*models.BatchRequestItem, userID uuid.UUID) ([]*models.BatchResponseItem, error) {
	responseItems := make([]*models.BatchResponseItem, 0, len(requestItems))
	err := s.connection.DoInTx(
		ctx,
		func(tx *sql.Tx) error {
			for _, requestItem := range requestItems {
				key, err := s.saveWithUniqueKey(ctx, requestItem.OriginalURL, func(ctx context.Context, key, value string) (bool, error) {
					executor := func(ctx context.Context, sql string, args ...any) error {
						return s.connection.ExecTx(ctx, tx, sql, args...)
					}
					return s.saver(ctx, executor, key, value, userID)
				})
				if err != nil {
					return fmt.Errorf("DatabaseStore.SaveBatch, s.saveWithUniqueKey failed: %w", err)
				}

				responseItems = append(responseItems, &models.BatchResponseItem{
					CorelationID: requestItem.CorelationID,
					Key:          key,
				})
			}

			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("DatabaseStore.SaveBatch, connection.DoInTx failed: %w", err)
	}

	return responseItems, nil
}

func (s *DatabaseStore) DeleteBatch(ctx context.Context, keys []string, userID uuid.UUID) error {
	err := s.connection.Exec(
		ctx,
		"update urls set is_deleted = true where url_key = any($1) and user_id = $2",
		keys,
		userID)
	if err != nil {
		return fmt.Errorf("DatabaseStore.DeleteBatch, connection.Exec failed: %w", err)
	}

	return nil
}

func (s *DatabaseStore) saver(ctx context.Context, executor executorFunc, key, value string, userID uuid.UUID) (bool, error) {
	err := executor(
		ctx,
		"insert into urls (url_key, original_url, user_id) values ($1, $2, $3)",
		key, value, userID.String(),
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

	return false, nil
}

func (s *DatabaseStore) getExistingKeyByValue(ctx context.Context, value string) (string, error) {
	var key string
	err := s.connection.QueryRow(
		ctx,
		&key,
		"select url_key from urls where original_url = $1",
		value,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("getExistingKeyByValue, row.Scan: failed to find existing key %v", key)
		}

		return "", fmt.Errorf("getExistingKeyByValue, row.Scan failed: %w", err)
	}

	return key, nil
}
