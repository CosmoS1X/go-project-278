package visits

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmoS1X/go-project-278/internal/storage/sqlc"
	"github.com/CosmoS1X/go-project-278/internal/testutil"
)

func newTestRepository(t *testing.T) (Repository, sqlc.DBTX) {
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

func createTestLink(t *testing.T, db sqlc.DBTX, shortName string) int64 {
	t.Helper()

	link, err := sqlc.New(db).CreateLink(t.Context(), sqlc.CreateLinkParams{
		OriginalUrl: "https://example.com",
		ShortName:   shortName,
	})
	require.NoError(t, err)

	return link.ID
}

func TestRepositoryRecordVisit(t *testing.T) {
	repo, db := newTestRepository(t)
	linkID := createTestLink(t, db, "visit_rec1")

	require.NoError(t, repo.RecordVisit(t.Context(), linkID, "https://google.com", sampleIP, "Mozilla/5.0", 302))

	visits, total, err := repo.List(t.Context(), 0, 100)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))

	var found bool
	for _, v := range visits {
		if v.LinkID != linkID || v.IP != sampleIP {
			continue
		}
		assert.Equal(t, "Mozilla/5.0", v.UserAgent)
		assert.Equal(t, "https://google.com", v.Referer)
		assert.Equal(t, int32(302), v.Status)
		assert.False(t, v.CreatedAt.IsZero())
		found = true
	}
	assert.True(t, found, "expected visit with link_id=%d and ip=1.2.3.4", linkID)
}

func TestRepositoryListVisits(t *testing.T) {
	repo, db := newTestRepository(t)
	linkID := createTestLink(t, db, "visit_lst1")

	require.NoError(t, repo.RecordVisit(t.Context(), linkID, "https://a.com", "10.0.0.1", "Chrome", 302))
	require.NoError(t, repo.RecordVisit(t.Context(), linkID, "https://b.com", "10.0.0.2", "Firefox", 302))

	visits, total, err := repo.List(t.Context(), 0, 100)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(2))

	var count int
	for _, v := range visits {
		if v.LinkID == linkID {
			count++
		}
	}
	assert.GreaterOrEqual(t, count, 2)
}

func TestRepositoryListVisitsError(t *testing.T) {
	repo := NewRepository(sqlc.New(&testutil.ErrDBTX{DB: testutil.NewClosedDB(t)}))

	_, _, err := repo.List(t.Context(), 0, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get link visits")
}

func TestRepositoryListVisitsCountError(t *testing.T) {
	_, db := newTestRepository(t)
	repo := NewRepository(sqlc.New(&testutil.CountFailTX{
		DBTX:   db,
		Closed: testutil.NewClosedDB(t),
	}))

	_, _, err := repo.List(t.Context(), 0, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count link visits")
}

func TestRepositoryRecordVisitUnknownLink(t *testing.T) {
	repo, _ := newTestRepository(t)

	err := repo.RecordVisit(t.Context(), 9_000_000_000, "https://google.com", sampleIP, "Mozilla/5.0", 302)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "record visit")
}
