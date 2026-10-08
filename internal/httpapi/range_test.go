package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRangeRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/range", func(c *gin.Context) {
		offset, limit, ok := ParseRangeParam(c)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"offset": offset, "limit": limit})
	})

	return router
}

func doRangeRequest(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, path, http.NoBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	newRangeRouter().ServeHTTP(w, req)

	return w
}

func TestParseRangeParamDefault(t *testing.T) {
	w := doRangeRequest(t, "/range")
	require.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"offset":0`)
	assert.Contains(t, w.Body.String(), `"limit":10`)
}

func TestParseRangeParamValid(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		offset int32
		limit  int32
	}{
		{name: "single cell", query: "[0,0]", offset: 0, limit: 1},
		{name: "window", query: "[3,7]", offset: 3, limit: 5},
		{name: "max limit", query: "[0,99]", offset: 0, limit: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doRangeRequest(t, "/range?range="+tt.query)
			require.Equal(t, http.StatusOK, w.Code)

			assert.Contains(t, w.Body.String(), fmt.Sprintf(`"offset":%d`, tt.offset))
			assert.Contains(t, w.Body.String(), fmt.Sprintf(`"limit":%d`, tt.limit))
		})
	}
}

func TestParseRangeParamInvalidJSON(t *testing.T) {
	w := doRangeRequest(t, "/range?range=abc")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid range parameter")
}

func TestParseRangeParamNegativeStart(t *testing.T) {
	w := doRangeRequest(t, "/range?range=[-1,2]")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// JSONEq compares semantically: gin HTML-escapes ">" when rendering JSON.
	assert.JSONEq(t, `{"error":"invalid range: start must be >= 0 and end must be >= start"}`, w.Body.String())
}

func TestParseRangeParamEndBeforeStart(t *testing.T) {
	w := doRangeRequest(t, "/range?range=[5,3]")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid range: start must be >= 0 and end must be >= start"}`, w.Body.String())
}

func TestParseRangeParamExceedsMaxLimit(t *testing.T) {
	w := doRangeRequest(t, "/range?range=[0,100]")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "range exceeds max limit of 100")
}
