package app

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmoS1X/go-project-278/internal/config"
)

func newTestDB(t *testing.T) (*sql.Tx, *config.Config) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	if _, err := os.Stat("../../.env"); err == nil {
		if err := godotenv.Load("../../.env"); err != nil {
			t.Fatalf("load .env: %v", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.DatabaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(t.Context(), cfg.DatabaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(t.Context()))

	db := stdlib.OpenDBFromPool(pool)
	tx, err := db.Begin()
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	return tx, cfg
}

func TestPing(t *testing.T) {
	tx, cfg := newTestDB(t)

	router := NewRouter(tx, cfg)

	req, err := http.NewRequest(http.MethodGet, "/ping", http.NoBody)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
