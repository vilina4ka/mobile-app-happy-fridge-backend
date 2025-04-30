package api

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BindJSONOrAbort binds JSON request body and handles potential errors, aborting the request.
func BindJSONOrAbort(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			_ = c.Error(domain.ErrValidationFailed(verr))
			c.Abort()
			return false
		}
		_ = c.Error(domain.NewAppError(err, "Invalid request body", http.StatusBadRequest))
		c.Abort()
		return false
	}
	return true // Binding successful
}

// GetIDParamGin extracts an ID from the URL path parameters using Gin.
func GetIDParamGin(c *gin.Context, paramName string) (int64, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		_ = c.Error(domain.ErrBadRequest(fmt.Sprintf("Invalid '%s' parameter in URL path: must be a positive integer", paramName)))
		c.Abort()
		return 0, false
	}
	return id, true
}

// AbortWithError pushes a domain.AppError onto the Gin context and aborts.
func AbortWithError(c *gin.Context, err *domain.AppError) {
	if err == nil {
		return // Should not happen if called correctly
	}
	_ = c.Error(err)
	c.Abort()
}
