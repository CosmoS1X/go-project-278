package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTimeoutRouter(d time.Duration, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Timeout(d))
	router.GET("/test", handler)

	return router
}

func serveTestRequest(t *testing.T, router *gin.Engine, ctx context.Context) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", http.NoBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func TestTimeoutSetsDeadline(t *testing.T) {
	var (
		deadline    time.Time
		hasDeadline bool
	)
	router := newTimeoutRouter(50*time.Millisecond, func(c *gin.Context) {
		deadline, hasDeadline = c.Request.Context().Deadline()
		c.Status(http.StatusOK)
	})

	w := serveTestRequest(t, router, context.Background())

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, hasDeadline)
	assert.WithinDuration(t, time.Now().Add(50*time.Millisecond), deadline, 40*time.Millisecond)
}

func TestTimeoutRespectsParentDeadline(t *testing.T) {
	var deadline time.Time
	router := newTimeoutRouter(time.Minute, func(c *gin.Context) {
		deadline, _ = c.Request.Context().Deadline()
		c.Status(http.StatusOK)
	})

	parent, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	w := serveTestRequest(t, router, parent)

	require.Equal(t, http.StatusOK, w.Code)
	assert.WithinDuration(t, time.Now().Add(20*time.Millisecond), deadline, 15*time.Millisecond)
}

func TestTimeoutCancelsAfterHandler(t *testing.T) {
	var captured context.Context
	router := newTimeoutRouter(time.Minute, func(c *gin.Context) {
		captured = c.Request.Context()
		c.Status(http.StatusOK)
	})

	w := serveTestRequest(t, router, context.Background())

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, captured)
	assert.ErrorIs(t, captured.Err(), context.Canceled)
}
