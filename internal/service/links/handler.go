package links

import (
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
	items, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to list links"})
		return
	}

	resp := make([]response, 0, len(items))
	for _, item := range items {
		resp = append(resp, h.toResponse(item))
	}

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

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{errKey: "invalid id"})
		return 0, false
	}
	return id, true
}
