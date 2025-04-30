package api

import (
	"errors"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type StatsHandler struct {
	service service.StatsService
	logger  *logrus.Logger
}

func NewStatsHandler(s service.StatsService, l *logrus.Logger) *StatsHandler {
	return &StatsHandler{service: s, logger: l}
}

// GetStats handles requests for fridge statistics (Gin version).
func (h *StatsHandler) GetStats(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	stats, err := h.service.GetFridgeStatistics(userID)
	if err != nil {
		_ = c.Error(err) // Push error for central handler
		return
	}

	c.JSON(http.StatusOK, stats)
}
