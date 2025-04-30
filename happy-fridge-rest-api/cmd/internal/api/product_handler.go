package api

import (
	"errors"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProductHandler struct {
	service service.ProductService
	logger  *logrus.Logger
}

func NewProductHandler(s service.ProductService, l *logrus.Logger) *ProductHandler {
	return &ProductHandler{service: s, logger: l}
}

// AddProduct handles requests to add a new product (Gin version).
func (h *ProductHandler) AddProduct(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	var req domain.ProductCreateRequest
	if !BindJSONOrAbort(c, &req) {
		return
	}

	product, err := h.service.AddProduct(userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, product)
}

// ListProducts handles requests to list products with filtering/searching (Gin version).
func (h *ProductHandler) ListProducts(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	params := domain.ProductListParams{}

	if categoryStr := c.Query("category"); categoryStr != "" {
		category := domain.ProductCategory(categoryStr)
		params.Category = &category
	}
	if searchStr := c.Query("search"); searchStr != "" {
		params.Search = &searchStr
	}

	products, err := h.service.ListProducts(userID, params)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := domain.StandardResponse{Data: products}
	c.JSON(http.StatusOK, resp)
}

// GetProduct handles requests to get a single product by ID (Gin version).
func (h *ProductHandler) GetProduct(c *gin.Context) {
	userID, okUser := GetUserIDFromContextGin(c)
	if !okUser {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}
	productID, okProd := GetIDParamGin(c, "productId")
	if !okProd {
		return // Error handled by helper
	}

	product, err := h.service.GetProductByID(userID, productID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, product)
}

// UpdateProduct handles requests to update a product.
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	userID, okUser := GetUserIDFromContextGin(c)
	if !okUser {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}
	productID, okProd := GetIDParamGin(c, "productId")
	if !okProd {
		return
	}

	var req domain.ProductUpdateRequest
	if !BindJSONOrAbort(c, &req) {
		return
	}

	product, err := h.service.UpdateProduct(userID, productID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct handles requests to delete a product.
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	userID, okUser := GetUserIDFromContextGin(c)
	if !okUser {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}
	productID, okProd := GetIDParamGin(c, "productId")
	if !okProd {
		return
	}

	if err := h.service.DeleteProduct(userID, productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// ConsumeProduct handles requests to consume a product entirely.
func (h *ProductHandler) ConsumeProduct(c *gin.Context) {
	userID, okUser := GetUserIDFromContextGin(c)
	if !okUser {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}
	productID, okProd := GetIDParamGin(c, "productId")
	if !okProd {
		return
	}

	if err := h.service.ConsumeProduct(userID, productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product consumed successfully"})
}

// UpdateQuantity handles requests to increment/decrement product quantity (Gin version).
func (h *ProductHandler) UpdateQuantity(c *gin.Context) {
	userID, okUser := GetUserIDFromContextGin(c)
	if !okUser {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}
	productID, okProd := GetIDParamGin(c, "productId")
	if !okProd {
		return
	}

	var req struct {
		Change float64 `json:"change"` // Expect 1 for +, -1 for -
	}
	if !BindJSONOrAbort(c, &req) {
		return
	}
	if req.Change != 1 && req.Change != -1 {
		AbortWithError(c, domain.ErrBadRequest("Invalid change value, expected 1 or -1"))
		return
	}

	product, err := h.service.UpdateProductQuantity(userID, productID, req.Change)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetCalendar handles requests for the calendar view (Gin version).
func (h *ProductHandler) GetCalendar(c *gin.Context) {
	userID, ok := GetUserIDFromContextGin(c)
	if !ok {
		AbortWithError(c, domain.ErrInternalServer(errors.New("user ID not found in context")))
		return
	}

	dateStr := c.Query("date") // Expect YYYY-MM-DD
	if dateStr == "" {
		AbortWithError(c, domain.ErrBadRequest("Missing required 'date' query parameter (YYYY-MM-DD)"))
		return
	}

	products, err := h.service.GetProductsExpiringOnDate(userID, dateStr)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := domain.StandardResponse{Data: products}
	c.JSON(http.StatusOK, resp)
}
