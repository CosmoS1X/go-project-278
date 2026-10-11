package httpapi

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteRangeResponse(c *gin.Context, resource string, offset int32, count int, total int64) bool {
	if int64(offset) >= total && total > 0 {
		c.Header("Content-Range", fmt.Sprintf("%s %d-%d/%d", resource, offset, offset, total))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return false
	}

	end := int64(offset) + int64(count) - 1
	if count == 0 {
		end = int64(offset)
	}
	c.Header("Content-Range", fmt.Sprintf("%s %d-%d/%d", resource, offset, end, total))
	return true
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{ErrKey: msg})
}

func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{ErrKey: msg})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{ErrKey: msg})
}

func ValidationError(c *gin.Context, field, msg string) {
	ValidationErrors(c, map[string]string{field: msg})
}

func ValidationErrors(c *gin.Context, errs map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
}
