package api

import (
	"errors"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	service service.UserService
	logger  *logrus.Logger
}

func NewUserHandler(s service.UserService, l *logrus.Logger) *UserHandler {
	return &UserHandler{service: s, logger: l}
}

// GetProfile handles requests to retrieve the current user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	user, err := h.service.GetUserProfile(userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword handles requests to change the current user's password (Gin version).
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	var req domain.ChangePasswordRequest
	if !BindJSONOrAbort(c, &req) {
		return
	}

	if err := h.service.ChangePassword(userID, req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"}) // Use gin.H for simple map
}

// DeleteAccount handles requests to delete the current user's account (Gin version).
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	if err := h.service.DeleteUser(userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
