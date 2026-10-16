package testutil

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmoS1X/go-project-278/internal/storage/sqlc"
)

func NewClosedDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", "postgres://closed")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	return db
}

type ErrDBTX struct {
	DB *sql.DB
}

func (e *ErrDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return e.DB.ExecContext(ctx, query, args...)
}

func (e *ErrDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return e.DB.PrepareContext(ctx, query)
}

func (e *ErrDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return e.DB.QueryContext(ctx, query, args...)
}

func (e *ErrDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return e.DB.QueryRowContext(ctx, query, args...)
}

type CountFailTX struct {
	sqlc.DBTX
	Closed *sql.DB
}

func (c *CountFailTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.Closed.QueryRowContext(ctx, query, args...)
}
