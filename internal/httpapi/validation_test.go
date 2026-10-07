package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleRequest struct {
	Name string `json:"name" binding:"required,min=3"`
}

type hiddenFieldRequest struct {
	Name   string `json:"name" binding:"required"`
	Hidden string `json:"-" binding:"required"`
}

func newValidationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/bind", func(c *gin.Context) {
		var req sampleRequest
		if !BindAndValidate(c, &req) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})
	router.POST("/bind-hidden", func(c *gin.Context) {
		var req hiddenFieldRequest
		if !BindAndValidate(c, &req) {
			return
		}
		c.Status(http.StatusOK)
	})

	return router
}

func doBindRequest(t *testing.T, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	newValidationRouter().ServeHTTP(w, req)

	return w
}

func TestBindAndValidateValid(t *testing.T) {
	w := doBindRequest(t, "/bind", `{"name":"abc"}`)
	require.Equal(t, http.StatusOK, w.Code)

	assert.Contains(t, w.Body.String(), `"name":"abc"`)
}

func TestBindAndValidateMissingRequiredField(t *testing.T) {
	w := doBindRequest(t, "/bind", `{}`)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)

	assert.Contains(t, w.Body.String(), `"errors"`)
	assert.Contains(t, w.Body.String(), `"name"`)
	assert.Contains(t, w.Body.String(), "required")
}

func TestBindAndValidateViolatesConstraint(t *testing.T) {
	w := doBindRequest(t, "/bind", `{"name":"ab"}`)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)

	assert.Contains(t, w.Body.String(), `"errors"`)
	assert.Contains(t, w.Body.String(), `"name"`)
}

func TestBindAndValidateMalformedJSON(t *testing.T) {
	w := doBindRequest(t, "/bind", `{invalid}`)
	require.Equal(t, http.StatusBadRequest, w.Code)

	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestBindAndValidateTypeMismatch(t *testing.T) {
	w := doBindRequest(t, "/bind", `{"name":123}`)
	require.Equal(t, http.StatusBadRequest, w.Code)

	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestBindAndValidateFieldWithoutJSONTag(t *testing.T) {
	w := doBindRequest(t, "/bind-hidden", `{}`)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)

	assert.Contains(t, w.Body.String(), `"errors"`)
	assert.Contains(t, w.Body.String(), `"name"`)
	assert.Contains(t, w.Body.String(), `"Hidden"`)
}
