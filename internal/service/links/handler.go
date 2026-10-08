package links

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/CosmoS1X/go-project-278/internal/httpapi"
)

type VisitRecorder interface {
	RecordVisit(ctx context.Context, linkID int64, referer, ip, userAgent string, status int32) error
}

type Handler struct {
	repo     Repository
	recorder VisitRecorder
	baseURL  string
}

func NewHandler(repo Repository, recorder VisitRecorder, baseURL string) *Handler {
	return &Handler{
		repo:     repo,
		recorder: recorder,
		baseURL:  strings.TrimSuffix(baseURL, "/"),
	}
}

func (h *Handler) List(c *gin.Context) {
	offset, limit, ok := httpapi.ParseRangeParam(c)
	if !ok {
		return
	}

	links, total, err := h.repo.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to list links"})
		return
	}

	if int64(offset) >= total && total > 0 {
		c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", offset, offset, total))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	resp := make([]response, 0, len(links))
	for _, link := range links {
		resp = append(resp, h.toResponse(link))
	}

	end := int64(offset) + int64(len(links)) - 1
	if len(links) == 0 {
		end = int64(offset)
	}
	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", offset, end, total))
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	var req request
	if !httpapi.BindAndValidate(c, &req) {
		return
	}

	originalURL := strings.TrimSpace(req.OriginalURL)
	shortName := strings.TrimSpace(req.ShortName)

	if shortName == "" {
		generated, err := h.repo.GenerateShortName(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to generate short name"})
			return
		}
		shortName = generated
	}

	link, err := h.repo.Create(c.Request.Context(), originalURL, shortName)
	if errors.Is(err, ErrShortNameTaken) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": ErrShortNameTaken.Error()}})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, h.toResponse(link))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	link, err := h.repo.GetByID(c.Request.Context(), id)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{httpapi.ErrKey: ErrNotFound.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to get link"})
		return
	}

	c.JSON(http.StatusOK, h.toResponse(link))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req request
	if !httpapi.BindAndValidate(c, &req) {
		return
	}

	originalURL := strings.TrimSpace(req.OriginalURL)
	shortName := strings.TrimSpace(req.ShortName)

	link, err := h.repo.Update(c.Request.Context(), id, originalURL, shortName)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{httpapi.ErrKey: ErrNotFound.Error()})
		case errors.Is(err, ErrShortNameTaken):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": ErrShortNameTaken.Error()}})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to update link"})
		}
		return
	}

	c.JSON(http.StatusOK, h.toResponse(link))
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to delete link"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Redirect(c *gin.Context) {
	codeParam := c.Param("code")

	link, err := h.repo.GetByShortName(c.Request.Context(), codeParam)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{httpapi.ErrKey: ErrNotFound.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to get link"})
		return
	}

	status := int32(http.StatusFound)

	if err := h.recorder.RecordVisit(c.Request.Context(), link.ID, c.Request.Referer(), c.ClientIP(), c.Request.UserAgent(), status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{httpapi.ErrKey: "failed to record visit"})
		return
	}

	c.Redirect(int(status), link.OriginalURL)
}

func (h *Handler) toResponse(link Link) response {
	return response{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
		ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, link.ShortName),
		CreatedAt:   link.CreatedAt,
	}
}

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{httpapi.ErrKey: "invalid id"})
		return 0, false
	}
	return id, true
}
