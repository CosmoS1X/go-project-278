package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CosmoS1X/go-project-278/internal/config"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
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

	_, err = pool.Exec(t.Context(), "TRUNCATE links RESTART IDENTITY")
	require.NoError(t, err)

	return pool
}

func TestPing(t *testing.T) {
	pool := newTestPool(t)

	cfg, err := config.Load()
	require.NoError(t, err)

	router := NewRouter(pool, cfg)

	req, err := http.NewRequest(http.MethodGet, "/ping", http.NoBody)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
