package links

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	links          []Link
	shortName      string
	generateErr    error
	createErr      error
	uniqueErr      bool
	getErr         error
	getNotFoundErr bool
}

const (
	sampleShortName   = "aaa"
	sampleOriginalURL = "https://a.com"
)

func (f *fakeRepository) List(_ context.Context) ([]Link, error) {
	return f.links, nil
}

func (f *fakeRepository) GetByID(_ context.Context, id int64) (Link, error) {
	if f.getNotFoundErr {
		return Link{}, ErrNotFound
	}
	if f.getErr != nil {
		return Link{}, f.getErr
	}
	for _, l := range f.links {
		if l.ID == id {
			return l, nil
		}
	}
	return Link{}, ErrNotFound
}

func (f *fakeRepository) Create(_ context.Context, originalURL, shortName string) (Link, error) {
	if f.uniqueErr {
		return Link{}, ErrShortNameTaken
	}
	if f.createErr != nil {
		return Link{}, f.createErr
	}
	item := Link{
		ID:          int64(len(f.links) + 1),
		OriginalURL: originalURL,
		ShortName:   shortName,
		CreatedAt:   time.Now(),
	}
	f.links = append(f.links, item)
	return item, nil
}

func (f *fakeRepository) Update(_ context.Context, id int64, originalURL, shortName string) (Link, error) {
	if f.uniqueErr {
		return Link{}, ErrShortNameTaken
	}
	for i, l := range f.links {
		if l.ID == id {
			f.links[i] = Link{ID: id, OriginalURL: originalURL, ShortName: shortName, CreatedAt: l.CreatedAt}
			return f.links[i], nil
		}
	}
	return Link{}, ErrNotFound
}

func (f *fakeRepository) Delete(_ context.Context, id int64) error {
	for i, l := range f.links {
		if l.ID == id {
			f.links = append(f.links[:i], f.links[i+1:]...)
			return nil
		}
	}
	return nil
}

func (f *fakeRepository) GenerateShortName(_ context.Context) (string, error) {
	if f.generateErr != nil {
		return "", f.generateErr
	}
	return f.shortName, nil
}

func newTestHandler(repo Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(repo, "https://short.io/r")

	router := gin.New()
	api := router.Group("/api/links")
	api.GET("", handler.List)
	api.POST("", handler.Create)
	api.GET("/:id", handler.Get)
	api.PUT("/:id", handler.Update)
	api.DELETE("/:id", handler.Delete)

	return router
}

func doRequest(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var req *http.Request
	var err error

	if body != "" {
		req, err = http.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, path, http.NoBody)
	}
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func TestCreateLink(t *testing.T) {
	fake := &fakeRepository{shortName: "gen123"}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodPost, "/api/links", `{"original_url": "https://example.com/long"}`)
	require.Equal(t, http.StatusCreated, w.Code)

	assert.Contains(t, w.Body.String(), `"id":1`)
	assert.Contains(t, w.Body.String(), `"original_url":"https://example.com/long"`)
	assert.Contains(t, w.Body.String(), `"short_name":"gen123"`)
	assert.Contains(t, w.Body.String(), `"short_url":"https://short.io/r/gen123"`)
	assert.Contains(t, w.Body.String(), `"created_at"`)
}

func TestCreateLinkWithProvidedShortName(t *testing.T) {
	fake := &fakeRepository{}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodPost, "/api/links", `{"original_url": "https://example.com", "short_name": "exmpl"}`)
	require.Equal(t, http.StatusCreated, w.Code)

	assert.Contains(t, w.Body.String(), `"short_name":"exmpl"`)
	assert.Contains(t, w.Body.String(), `"short_url":"https://short.io/r/exmpl"`)
}

func TestCreateLinkMissingOriginalURL(t *testing.T) {
	fake := &fakeRepository{}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodPost, "/api/links", `{"original_url": "  "}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLinkShortNameTaken(t *testing.T) {
	fake := &fakeRepository{uniqueErr: true}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodPost, "/api/links", `{"original_url": "https://example.com", "short_name": "dup"}`)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGenerateShortNameError(t *testing.T) {
	fake := &fakeRepository{generateErr: errors.New("boom")}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodPost, "/api/links", `{"original_url": "https://example.com"}`)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetLink(t *testing.T) {
	fake := &fakeRepository{links: []Link{{ID: 1, OriginalURL: sampleOriginalURL, ShortName: sampleShortName, CreatedAt: time.Now()}}}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodGet, "/api/links/1", "")
	require.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"id":1`)
	assert.Contains(t, w.Body.String(), `"short_url":"https://short.io/r/aaa"`)
}

func TestGetLinkNotFound(t *testing.T) {
	fake := &fakeRepository{}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodGet, "/api/links/999", "")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetLinkInvalidID(t *testing.T) {
	fake := &fakeRepository{}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodGet, "/api/links/abc", "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListLinks(t *testing.T) {
	fake := &fakeRepository{links: []Link{{ID: 1, OriginalURL: sampleOriginalURL, ShortName: sampleShortName, CreatedAt: time.Now()}}}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodGet, "/api/links", "")
	require.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"id":1`)
}

func TestUpdateLink(t *testing.T) {
	fake := &fakeRepository{links: []Link{{ID: 1, OriginalURL: sampleOriginalURL, ShortName: sampleShortName, CreatedAt: time.Now()}}}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodPut, "/api/links/1", `{"original_url": "https://new.com", "short_name": "new1"}`)
	require.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"original_url":"https://new.com"`)
	assert.Contains(t, w.Body.String(), `"short_name":"new1"`)
}

func TestUpdateLinkNotFound(t *testing.T) {
	fake := &fakeRepository{}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodPut, "/api/links/999", `{"original_url": "https://new.com", "short_name": "new1"}`)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteLink(t *testing.T) {
	fake := &fakeRepository{links: []Link{{ID: 1, OriginalURL: sampleOriginalURL, ShortName: sampleShortName, CreatedAt: time.Now()}}}
	router := newTestHandler(fake)

	w := doRequest(t, router, http.MethodDelete, "/api/links/1", "")
	assert.Equal(t, http.StatusNoContent, w.Code)
}
