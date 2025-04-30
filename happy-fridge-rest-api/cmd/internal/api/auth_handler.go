package api

import (
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	service service.AuthService
	logger  *logrus.Logger
}

func NewAuthHandler(s service.AuthService, l *logrus.Logger) *AuthHandler {
	return &AuthHandler{service: s, logger: l}
}

// Register handles user registration requests.
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if !BindJSONOrAbort(c, &req) {
		return
	}

	resp, err := h.service.Register(req.Login, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles user login requests.
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.AuthRequest

	if !BindJSONOrAbort(c, &req) {
		return
	}

	resp, err := h.service.Login(req.Login, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
