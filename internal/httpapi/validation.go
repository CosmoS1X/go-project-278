package httpapi

import (
	"encoding/json"
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

type normalizer interface {
	Normalize()
}

func BindAndValidate(c *gin.Context, obj any) bool {
	registerJSONTagNames()

	if err := json.NewDecoder(c.Request.Body).Decode(obj); err != nil {
		BadRequest(c, "invalid request")
		return false
	}

	if n, ok := obj.(normalizer); ok {
		n.Normalize()
	}

	if err := binding.Validator.ValidateStruct(obj); err != nil {
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
