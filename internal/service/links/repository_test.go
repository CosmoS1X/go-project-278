package links

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmoS1X/go-project-278/internal/storage/sqlc"
)

func newTestRepositoryTx(t *testing.T) (Repository, *sql.Tx) {
	t.Helper()

	_ = godotenv.Load(filepath.Join("..", "..", "..", ".env"))

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(t.Context(), databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(t.Context()))

	db := stdlib.OpenDBFromPool(pool)
	tx, err := db.Begin()
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	return NewRepository(sqlc.New(tx)), tx
}

func newTestRepository(t *testing.T) Repository {
	t.Helper()

	repo, _ := newTestRepositoryTx(t)
	return repo
}

type errDBTX struct {
	db *sql.DB
}

func (e *errDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return e.db.ExecContext(ctx, query, args...)
}

func (e *errDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return e.db.PrepareContext(ctx, query)
}

func (e *errDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return e.db.QueryContext(ctx, query, args...)
}

func (e *errDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return e.db.QueryRowContext(ctx, query, args...)
}

func newClosedDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", "postgres://closed")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	return db
}

func newStubRepository(t *testing.T) Repository {
	t.Helper()

	return NewRepository(sqlc.New(&errDBTX{db: newClosedDB(t)}))
}

type countFailTX struct {
	sqlc.DBTX
	closed *sql.DB
}

func (c *countFailTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.closed.QueryRowContext(ctx, query, args...)
}

func TestRepositoryCreateAndGetByID(t *testing.T) {
	repo := newTestRepository(t)

	created, err := repo.Create(t.Context(), "https://example.com/long", "exmpl")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/long", created.OriginalURL)
	assert.Equal(t, "exmpl", created.ShortName)
	assert.NotZero(t, created.ID)
	assert.False(t, created.CreatedAt.IsZero())

	got, err := repo.GetByID(t.Context(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.OriginalURL, got.OriginalURL)
}

func TestRepositoryGetByIDNotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.GetByID(t.Context(), 999)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestRepositoryCreateDuplicateShortName(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.Create(t.Context(), "https://a.com", "dup1")
	require.NoError(t, err)

	_, err = repo.Create(t.Context(), "https://b.com", "dup1")
	assert.True(t, errors.Is(err, ErrShortNameTaken))
}

func TestRepositoryList(t *testing.T) {
	repo := newTestRepository(t)

	created1, err := repo.Create(t.Context(), "https://a.com", "aaa")
	require.NoError(t, err)
	created2, err := repo.Create(t.Context(), "https://b.com", "bbb")
	require.NoError(t, err)

	items, total, err := repo.List(t.Context(), 0, 100)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(2))
	assert.True(t, containsLink(items, created1.ID))
	assert.True(t, containsLink(items, created2.ID))
}

func containsLink(items []Link, id int64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func TestRepositoryUpdate(t *testing.T) {
	repo := newTestRepository(t)

	created, err := repo.Create(t.Context(), "https://a.com", "aaa")
	require.NoError(t, err)

	updated, err := repo.Update(t.Context(), created.ID, "https://new.com", "new1")
	require.NoError(t, err)
	assert.Equal(t, "https://new.com", updated.OriginalURL)
	assert.Equal(t, "new1", updated.ShortName)

	got, err := repo.GetByID(t.Context(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "https://new.com", got.OriginalURL)
}

func TestRepositoryUpdateNotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.Update(t.Context(), 999, "https://a.com", "aaa")
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestRepositoryDelete(t *testing.T) {
	repo := newTestRepository(t)

	created, err := repo.Create(t.Context(), "https://a.com", "aaa")
	require.NoError(t, err)

	require.NoError(t, repo.Delete(t.Context(), created.ID))

	_, err = repo.GetByID(t.Context(), created.ID)
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestRepositoryGenerateShortName(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.Create(t.Context(), "https://a.com", "aaaaaaaa")
	require.NoError(t, err)

	name, err := repo.GenerateShortName(t.Context())
	require.NoError(t, err)
	assert.Len(t, name, shortNameLength)
	assert.NotEqual(t, "aaaaaaaa", name)
}

func TestRepositoryGetByShortName(t *testing.T) {
	repo := newTestRepository(t)

	created, err := repo.Create(t.Context(), "https://example.com", "test1")
	require.NoError(t, err)

	got, err := repo.GetByShortName(t.Context(), "test1")
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.OriginalURL, got.OriginalURL)
}

func TestRepositoryGetByShortNameNotFound(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.GetByShortName(t.Context(), "nonexistent")
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestRepositoryUpdateDuplicateShortName(t *testing.T) {
	repo := newTestRepository(t)

	_, err := repo.Create(t.Context(), "https://a.com", "takens")
	require.NoError(t, err)

	b, err := repo.Create(t.Context(), "https://b.com", "other1")
	require.NoError(t, err)

	_, err = repo.Update(t.Context(), b.ID, "https://b.com", "takens")
	assert.True(t, errors.Is(err, ErrShortNameTaken))
}

func TestRepositoryErrorsOnClosedDB(t *testing.T) {
	repo := newStubRepository(t)

	t.Run("list", func(t *testing.T) {
		_, _, err := repo.List(t.Context(), 0, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get links")
	})

	t.Run("get by id", func(t *testing.T) {
		_, err := repo.GetByID(t.Context(), 1)
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNotFound)
		assert.Contains(t, err.Error(), "get link by id")
	})

	t.Run("create", func(t *testing.T) {
		_, err := repo.Create(t.Context(), "https://a.com", "create1")
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrShortNameTaken)
		assert.Contains(t, err.Error(), "create link")
	})

	t.Run("update", func(t *testing.T) {
		_, err := repo.Update(t.Context(), 1, "https://a.com", "update1")
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNotFound)
		assert.NotErrorIs(t, err, ErrShortNameTaken)
		assert.Contains(t, err.Error(), "update link")
	})

	t.Run("delete", func(t *testing.T) {
		err := repo.Delete(t.Context(), 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete link")
	})

	t.Run("get by short name", func(t *testing.T) {
		_, err := repo.GetByShortName(t.Context(), "some1")
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNotFound)
		assert.Contains(t, err.Error(), "get link by short name")
	})

	t.Run("generate short name", func(t *testing.T) {
		_, err := repo.GenerateShortName(t.Context())
		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNotFound)
	})
}

func TestRepositoryListCountError(t *testing.T) {
	_, tx := newTestRepositoryTx(t)
	repo := NewRepository(sqlc.New(&countFailTX{
		DBTX:   tx,
		closed: newClosedDB(t),
	}))

	_, _, err := repo.List(t.Context(), 0, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count links")
}
