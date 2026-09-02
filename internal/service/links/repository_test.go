package links

import (
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

func newTestRepository(t *testing.T) Repository {
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

	return NewRepository(sqlc.New(tx))
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
