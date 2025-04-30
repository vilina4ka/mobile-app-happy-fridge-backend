package api

import (
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NewRouterGin creates and configures the main application router using Gin.
func NewRouterGin(logger *logrus.Logger, services *service.Service, jwtSecret string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(LoggingMiddlewareGin(logger))
	router.Use(CORSMiddlewareGin())
	router.Use(GinErrorHandler(logger))

	authHandler := NewAuthHandler(services.Auth, logger)
	userHandler := NewUserHandler(services.User, logger)
	fridgeHandler := NewFridgeHandler(services.Fridge, logger)
	productHandler := NewProductHandler(services.Product, logger)
	statsHandler := NewStatsHandler(services.Stats, logger)

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	apiGroup := router.Group("/api")
	apiGroup.Use(AuthMiddlewareGin(jwtSecret, logger))
	{
		userGroup := apiGroup.Group("/user")
		{
			userGroup.GET("/profile", userHandler.GetProfile)
			userGroup.PATCH("/password", userHandler.ChangePassword)
			userGroup.DELETE("/account", userHandler.DeleteAccount)
		}

		fridgeGroup := apiGroup.Group("/fridge")
		{
			fridgeGroup.GET("", fridgeHandler.GetFridgeInfo)
			fridgeGroup.PATCH("/notifications", fridgeHandler.UpdateNotifications)
		}

		productGroup := apiGroup.Group("/products")
		{
			productGroup.POST("", productHandler.AddProduct)
			productGroup.GET("", productHandler.ListProducts) // Query params: ?category=...&search=...
			productGroup.GET("/:productId", productHandler.GetProduct)
			productGroup.PATCH("/:productId", productHandler.UpdateProduct)
			productGroup.DELETE("/:productId", productHandler.DeleteProduct)
			productGroup.POST("/:productId/consume", productHandler.ConsumeProduct)   // Action POST
			productGroup.PATCH("/:productId/quantity", productHandler.UpdateQuantity) // Update quantity
		}

		calendarGroup := apiGroup.Group("/calendar")
		{
			calendarGroup.GET("", productHandler.GetCalendar)
		}

		statsGroup := apiGroup.Group("/stats")
		{
			statsGroup.GET("", statsHandler.GetStats)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
