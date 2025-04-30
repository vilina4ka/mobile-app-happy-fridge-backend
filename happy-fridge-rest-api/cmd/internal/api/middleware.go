package api

import (
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/auth"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const UserIDContextKey = "userID"

// GinErrorHandler is a centralized error handler middleware for Gin.
func GinErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr, ok := err.(*domain.AppError); ok {
				logger.Warnf("API Error: Status=%d, Path=%s, Message=%s, Details=%v", appErr.HTTPStatus, c.Request.URL.Path, appErr.Message, appErr.OriginalError)
				apiErr := domain.APIError{
					Status:  appErr.HTTPStatus,
					Message: appErr.Message,
				}
				c.AbortWithStatusJSON(appErr.HTTPStatus, apiErr)
			} else {
				logger.Errorf("Internal Server Error: Path=%s, Error=%v", c.Request.URL.Path, err)
				apiErr := domain.APIError{
					Status:  http.StatusInternalServerError,
					Message: "Internal Server Error",
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, apiErr)
			}
		}
	}
}

// AuthMiddlewareGin creates a Gin middleware function that requires JWT authentication.
func AuthMiddlewareGin(jwtSecret string, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Authorization header missing")
			_ = c.Error(domain.ErrUnauthorized())
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logger.Warnf("Invalid Authorization header format: %s", authHeader)
			_ = c.Error(domain.NewAppError(nil, "Invalid Authorization header format", http.StatusUnauthorized))
			c.Abort()
			return
		}
		tokenString := parts[1]

		claims, err := auth.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			logger.Warnf("JWT validation failed: %v", err)
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set(UserIDContextKey, claims.UserID)
		logger.Debugf("User %d authenticated successfully for path %s", claims.UserID, c.Request.URL.Path)

		c.Next()
	}
}

// GetUserIDFromContextGin retrieves the user ID from Gin's context.
func GetUserIDFromContextGin(c *gin.Context) (int64, bool) {
	userIDAny, exists := c.Get(UserIDContextKey)
	if !exists {
		return 0, false
	}
	userID, ok := userIDAny.(int64)
	return userID, ok
}

// LoggingMiddlewareGin provides basic request/response logging for Gin.
func LoggingMiddlewareGin(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		logger.Infof("--> %s %s", c.Request.Method, path)
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := ""
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		logger.Infof("<-- %s %s %d %s (%s) | %s %s",
			method,
			path,
			statusCode,
			http.StatusText(statusCode),
			latency,
			clientIP,
			errorMessage,
		)
	}
}

// CORSMiddlewareGin provides basic CORS headers for Gin.
func CORSMiddlewareGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Adjust for production
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
