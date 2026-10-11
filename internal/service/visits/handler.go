package visits

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/CosmoS1X/go-project-278/internal/httpapi"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(c *gin.Context) {
	offset, limit, ok := httpapi.ParseRangeParam(c)
	if !ok {
		return
	}

	items, total, err := h.repo.List(c.Request.Context(), offset, limit)
	if err != nil {
		httpapi.InternalError(c, "failed to list visits")
		return
	}

	resp := make([]response, 0, len(items))
	for _, item := range items {
		resp = append(resp, response(item))
	}

	if !httpapi.WriteRangeResponse(c, "link_visits", offset, len(items), total) {
		return
	}
	c.JSON(http.StatusOK, resp)
}
