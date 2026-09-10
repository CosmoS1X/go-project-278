package visits

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/CosmoS1X/go-project-278/internal/httpapi"
)

const errKey = "error"

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) ListVisits(c *gin.Context) {
	offset, limit, ok := httpapi.ParseRangeParam(c, errKey)
	if !ok {
		return
	}

	items, total, err := h.repo.ListVisits(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{errKey: "failed to list visits"})
		return
	}

	if int64(offset) >= total && total > 0 {
		c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", offset, offset, total))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	resp := make([]response, 0, len(items))
	for i := range items {
		resp = append(resp, toResponse(&items[i]))
	}

	end := int64(offset) + int64(len(items)) - 1
	if len(items) == 0 {
		end = int64(offset)
	}
	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", offset, end, total))
	c.JSON(http.StatusOK, resp)
}
