package links

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const errKey = "error"

type Handler struct {
	repo    Repository
	baseURL string
}

func NewHandler(repo Repository, baseURL string) *Handler {
	return &Handler{
		repo:    repo,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

func (h *Handler) List(c *gin.Context) {
	offset, limit, ok := parseRangeParam(c)
	if !ok {
		return
	}

	items, total, err := h.repo.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to list links"})
		return
	}

	if int64(offset) >= total && total > 0 {
		c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", offset, offset, total))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	resp := make([]response, 0, len(items))
	for _, item := range items {
		resp = append(resp, h.toResponse(item))
	}

	end := int64(offset) + int64(len(items)) - 1
	if len(items) == 0 {
		end = int64(offset)
	}
	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", offset, end, total))
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "invalid request body"})
		return
	}

	var (
		originalURL = strings.TrimSpace(req.OriginalURL)
		shortName   = strings.TrimSpace(req.ShortName)
	)
	if originalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "original_url is required"})
		return
	}

	if shortName == "" {
		generated, err := h.repo.GenerateShortName(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to generate short name"})
			return
		}
		shortName = generated
	}

	item, err := h.repo.Create(c.Request.Context(), originalURL, shortName)
	if err != nil {
		if errors.Is(err, ErrShortNameTaken) {
			c.JSON(http.StatusConflict, gin.H{errKey: "short name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, h.toResponse(item))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{errKey: "link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to get link"})
		return
	}

	c.JSON(http.StatusOK, h.toResponse(item))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "invalid request body"})
		return
	}

	var (
		originalURL = strings.TrimSpace(req.OriginalURL)
		shortName   = strings.TrimSpace(req.ShortName)
	)
	if originalURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "original_url is required"})
		return
	}
	if shortName == "" {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "short_name is required"})
		return
	}

	item, err := h.repo.Update(c.Request.Context(), id, originalURL, shortName)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{errKey: "link not found"})
		case errors.Is(err, ErrShortNameTaken):
			c.JSON(http.StatusConflict, gin.H{errKey: "short name already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to update link"})
		}
		return
	}

	c.JSON(http.StatusOK, h.toResponse(item))
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to delete link"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) toResponse(item Link) response {
	return response{
		ID:          item.ID,
		OriginalURL: item.OriginalURL,
		ShortName:   item.ShortName,
		ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, item.ShortName),
		CreatedAt:   item.CreatedAt,
	}
}

const (
	defaultLimit = 10
	maxLimit     = 100
)

func parseRangeParam(c *gin.Context) (offset, limit int32, ok bool) {
	rangeParam := c.Query("range")
	if rangeParam == "" {
		return 0, defaultLimit, true
	}

	var pair [2]int32
	if err := json.Unmarshal([]byte(rangeParam), &pair); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "invalid range parameter"})
		return 0, 0, false
	}

	start, end := pair[0], pair[1]
	if start < 0 || end < start {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "invalid range: start must be >= 0 and end must be >= start"})
		return 0, 0, false
	}

	limit = end - start + 1
	if limit > maxLimit {
		c.JSON(http.StatusBadRequest, gin.H{errKey: fmt.Sprintf("range exceeds max limit of %d", maxLimit)})
		return 0, 0, false
	}

	return start, limit, true
}

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "invalid id"})
		return 0, false
	}
	return id, true
}
