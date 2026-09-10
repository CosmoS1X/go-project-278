package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

func ParseRangeParam(c *gin.Context, errKey string) (offset, limit int32, ok bool) {
	rangeParam := c.Query("range")
	if rangeParam == "" {
		return 0, DefaultLimit, true
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
	if limit > MaxLimit {
		c.JSON(http.StatusBadRequest, gin.H{errKey: fmt.Sprintf("range exceeds max limit of %d", MaxLimit)})
		return 0, 0, false
	}

	return start, limit, true
}
