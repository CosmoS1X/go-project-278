package links

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/CosmoS1X/url-shortener/internal/httpapi"
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

	items, total, err := h.repo.List(c.Request.Context(), offset, limit)
	if err != nil {
		httpapi.ServerError(c, err, "failed to list links")
		return
	}

	resp := make([]response, 0, len(items))
	for _, item := range items {
		resp = append(resp, h.toResponse(item))
	}

	if !httpapi.WriteRangeResponse(c, "links", offset, len(items), total) {
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	var req request
	if !httpapi.BindAndValidate(c, &req) {
		return
	}

	shortName := req.ShortName
	if shortName == "" {
		generated, err := h.repo.GenerateShortName(c.Request.Context())
		if err != nil {
			httpapi.ServerError(c, err, "failed to generate short name")
			return
		}
		shortName = generated
	}

	link, err := h.repo.Create(c.Request.Context(), req.OriginalURL, shortName)
	if errors.Is(err, ErrShortNameTaken) {
		httpapi.ValidationError(c, "short_name", ErrShortNameTaken.Error())
		return
	}
	if err != nil {
		httpapi.ServerError(c, err, "failed to create link")
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
		httpapi.NotFound(c, ErrNotFound.Error())
		return
	}
	if err != nil {
		httpapi.ServerError(c, err, "failed to get link")
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

	link, err := h.repo.Update(c.Request.Context(), id, req.OriginalURL, req.ShortName)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httpapi.NotFound(c, ErrNotFound.Error())
		case errors.Is(err, ErrShortNameTaken):
			httpapi.ValidationError(c, "short_name", ErrShortNameTaken.Error())
		default:
			httpapi.ServerError(c, err, "failed to update link")
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
		if errors.Is(err, ErrNotFound) {
			httpapi.NotFound(c, ErrNotFound.Error())
			return
		}
		httpapi.ServerError(c, err, "failed to delete link")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Redirect(c *gin.Context) {
	codeParam := c.Param("code")

	link, err := h.repo.GetByShortName(c.Request.Context(), codeParam)
	if errors.Is(err, ErrNotFound) {
		httpapi.NotFound(c, ErrNotFound.Error())
		return
	}
	if err != nil {
		httpapi.ServerError(c, err, "failed to get link")
		return
	}

	if err := h.recorder.RecordVisit(c.Request.Context(), link.ID, c.Request.Referer(), c.ClientIP(), c.Request.UserAgent(), http.StatusFound); err != nil {
		httpapi.ServerError(c, err, "failed to record visit")
		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL)
}

func (h *Handler) toResponse(link Link) response {
	return response{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
		ShortURL:    fmt.Sprintf("%s/r/%s", h.baseURL, link.ShortName),
		CreatedAt:   link.CreatedAt,
	}
}

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid id")
		return 0, false
	}
	return id, true
}
