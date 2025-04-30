package api

import (
	"errors"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type FridgeHandler struct {
	service service.FridgeService
	logger  *logrus.Logger
}

func NewFridgeHandler(s service.FridgeService, l *logrus.Logger) *FridgeHandler {
	return &FridgeHandler{service: s, logger: l}
}

// GetFridgeInfo handles requests for the main fridge screen info.
func (h *FridgeHandler) GetFridgeInfo(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	resp, err := h.service.GetFridgeInfo(userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateNotifications handles requests to update fridge notification settings.
func (h *FridgeHandler) UpdateNotifications(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	var req domain.UpdateNotificationSettingsRequest
	if !BindJSONOrAbort(c, &req) {
		return
	}

	if err := h.service.UpdateNotificationSettings(userID, req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification settings updated"})
}
