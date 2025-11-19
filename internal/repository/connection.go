package repository

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
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Connection interface {
	Init(ctx context.Context) error
	Shutdown()
	Ping(ctx context.Context) error
	QueryRow(ctx context.Context, result any, sql string, args ...any) error
	QueryRows(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) error
	ExecTx(ctx context.Context, tx *sql.Tx, sql string, args ...any) error
	DoInTx(ctx context.Context, action func(*sql.Tx) error) error
}

type connectionImpl struct {
	configuration *config.DatabaseStoreConfiguration
	logger        logger.Logger
	db            *sql.DB
}

var _ Connection = (*connectionImpl)(nil)

func NewConnection(configuration *config.Configuration, logger logger.Logger) Connection {
	connection := &connectionImpl{
		configuration: configuration.DatabaseStore,
		logger:        logger,
	}

	return connection
}

func (c *connectionImpl) Init(ctx context.Context) error {
	if !c.configuration.IsDatabaseEnabled() {
		c.logger.Infof("Database is disabled")
		return nil
	}

	db, err := sql.Open("pgx", c.configuration.DataSourceName)
	if err != nil {
		return fmt.Errorf("connectionImpl.Init, sql.Open failed: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("connectionImpl.Init, postgres.WithInstance failed: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("connectionImpl.Init, migrate.NewWithDatabaseInstance failed: %w", err)
	}

	m.Log = logger.NewMigrateLogger(c.logger)
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("connectionImpl.Init, m.Up failed: %w", err)
	}

	c.db = db
	c.logger.Infof("Database connection initialized")

	return nil
}

func (c *connectionImpl) Shutdown() {
	if c.db == nil {
		return
	}

	c.db.Close()
}

func (c *connectionImpl) Ping(ctx context.Context) error {
	if c.db == nil {
		return nil
	}

	return c.db.PingContext(ctx)
}

func (c *connectionImpl) QueryRow(ctx context.Context, result any, sql string, args ...any) error {
	if c.db == nil {
		return nil
	}

	err := c.db.QueryRowContext(ctx, sql, args...).Scan(result)
	if err != nil {
		return fmt.Errorf("connectionImpl.QueryRow, Row.Scan failed: %w", err)
	}

	return nil
}

func (c *connectionImpl) QueryRows(ctx context.Context, sql string, args ...any) (*sql.Rows, error) {
	if c.db == nil {
		return nil, nil
	}

	rows, err := c.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("connectionImpl.QueryRows, db.QueryContext failed: %w", err)
	}

	return rows, nil
}

func (c *connectionImpl) Exec(ctx context.Context, sql string, args ...any) error {
	if c.db == nil {
		return nil
	}

	_, err := c.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("connectionImpl.Exec, db.ExecContext failed: %w", err)
	}

	return nil
}

func (c *connectionImpl) ExecTx(ctx context.Context, tx *sql.Tx, sql string, args ...any) error {
	if c.db == nil {
		return nil
	}

	_, err := tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("connectionImpl.InsertRow, pool.Exec failed: %w", err)
	}

	return nil
}

func (c *connectionImpl) DoInTx(ctx context.Context, action func(*sql.Tx) error) error {
	if c.db == nil {
		return nil
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("connectionImpl.DoInTx, pool.Begin failed: %w", err)
	}

	if err = action(tx); err != nil {
		originalError := fmt.Errorf("connectionImpl.DoInTx, action failed: %w", err)
		if txErr := tx.Rollback(); txErr != nil {
			return errors.Join(originalError, fmt.Errorf("connectionImpl.DoInTx, tx.Rollback failed: %w", txErr))
		}

		return originalError
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("connectionImpl.DoInTx, tx.Commit failed: %w", err)
	}

	return nil
}
