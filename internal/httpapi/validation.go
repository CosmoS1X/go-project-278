package httpapi

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var registerJSONTagNames = sync.OnceFunc(func() {
	validate, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
})

func BindAndValidate(c *gin.Context, obj any) bool {
	registerJSONTagNames()

	if err := c.ShouldBindJSON(obj); err != nil {
		if validationErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			errs := make(map[string]string, len(validationErrs))
			for _, fe := range validationErrs {
				errs[fe.Field()] = fe.Error()
			}
			ValidationErrors(c, errs)
			return false
		}
		BadRequest(c, "invalid request")
		return false
	}
	return true
}
