package visits

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	visits []LinkVisit
}

const sampleIP = "1.2.3.4"

func (f *fakeRepository) RecordVisit(_ context.Context, linkID int64, referer, ip, userAgent string, status int32) error {
	f.visits = append(f.visits, LinkVisit{
		ID:        int64(len(f.visits) + 1),
		LinkID:    linkID,
		IP:        ip,
		Referer:   referer,
		UserAgent: userAgent,
		Status:    status,
		CreatedAt: time.Now(),
	})
	return nil
}

func (f *fakeRepository) ListVisits(_ context.Context, offset, limit int32) ([]LinkVisit, int64, error) {
	total := int64(len(f.visits))
	if int64(offset) >= total {
		return []LinkVisit{}, total, nil
	}
	start := int(offset)
	end := int(offset) + int(limit)
	if int64(end) > total {
		end = int(total)
	}
	return f.visits[start:end], total, nil
}

func newTestHandler(repo Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(repo)

	router := gin.New()
	router.GET("/api/link_visits", handler.ListVisits)

	return router
}

func doRequest(t *testing.T, router *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, path, http.NoBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func TestListVisits(t *testing.T) {
	now := time.Now()
	fake := &fakeRepository{visits: []LinkVisit{
		{ID: 1, LinkID: 1, IP: sampleIP, Referer: "https://google.com", UserAgent: "Mozilla", Status: 302, CreatedAt: now},
	}}
	router := newTestHandler(fake)

	w := doRequest(t, router, "/api/link_visits")
	require.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"ip":"1.2.3.4"`)
	assert.Contains(t, w.Body.String(), `"reffer":"https://google.com"`)
	assert.Contains(t, w.Body.String(), `"status":302`)
	assert.Equal(t, "link_visits 0-0/1", w.Header().Get("Content-Range"))
}

func TestListVisitsWithRange(t *testing.T) {
	now := time.Now()
	visits := make([]LinkVisit, 5)
	for i := range visits {
		visits[i] = LinkVisit{ID: int64(i + 1), LinkID: 1, IP: fmt.Sprintf("1.0.0.%d", i+1), Status: 302, CreatedAt: now}
	}
	fake := &fakeRepository{visits: visits}
	router := newTestHandler(fake)

	w := doRequest(t, router, "/api/link_visits?range=[0,2]")
	require.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"ip":"1.0.0.1"`)
	assert.Contains(t, w.Body.String(), `"ip":"1.0.0.2"`)
	assert.Contains(t, w.Body.String(), `"ip":"1.0.0.3"`)
	assert.NotContains(t, w.Body.String(), `"ip":"1.0.0.4"`)
	assert.Equal(t, "link_visits 0-2/5", w.Header().Get("Content-Range"))
}

func TestListVisitsInvalidRange(t *testing.T) {
	fake := &fakeRepository{}
	router := newTestHandler(fake)

	w := doRequest(t, router, "/api/link_visits?range=abc")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListVisitsRangeOutOfBounds(t *testing.T) {
	fake := &fakeRepository{visits: []LinkVisit{{ID: 1, LinkID: 1, IP: sampleIP, Status: 302, CreatedAt: time.Now()}}}
	router := newTestHandler(fake)

	w := doRequest(t, router, "/api/link_visits?range=[100,110]")
	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "link_visits 100-100/1", w.Header().Get("Content-Range"))
}
