package service

import (
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/repository"
	"time"
)

// AuthService handles user authentication logic.
type AuthService interface {
	Register(login, password string) (*domain.AuthResponse, *domain.AppError)
	Login(login, password string) (*domain.AuthResponse, *domain.AppError)
}

// UserService handles user profile and settings management.
type UserService interface {
	GetUserProfile(userID int64) (*domain.User, *domain.AppError)
	ChangePassword(userID int64, req domain.ChangePasswordRequest) *domain.AppError
	DeleteUser(userID int64) *domain.AppError
}

// FridgeService handles operations related to the fridge itself.
type FridgeService interface {
	GetFridgeByUserID(userID int64) (*domain.Fridge, *domain.AppError)
	UpdateNotificationSettings(userID int64, req domain.UpdateNotificationSettingsRequest) *domain.AppError
	GetFridgeInfo(userID int64) (*FridgeInfoResponse, *domain.AppError)                                                // Combines fridge data + level info
	CheckAndGrantAchievement(fridgeID int64, achievementType string, relatedData interface{}) (bool, *domain.AppError) // Checks and grants achievement, returns true if newly granted
}

// ProductService handles CRUD operations for products and consumption logic.
type ProductService interface {
	AddProduct(userID int64, req domain.ProductCreateRequest) (*domain.Product, *domain.AppError)
	ListProducts(userID int64, params domain.ProductListParams) ([]domain.Product, *domain.AppError)
	GetProductByID(userID, productID int64) (*domain.Product, *domain.AppError) // Ensures user owns product
	UpdateProduct(userID, productID int64, req domain.ProductUpdateRequest) (*domain.Product, *domain.AppError)
	DeleteProduct(userID, productID int64) *domain.AppError
	UpdateProductQuantity(userID, productID int64, change float64) (*domain.Product, *domain.AppError)
	ConsumeProduct(userID, productID int64) *domain.AppError // "Потратить весь продукт"
	GetProductsExpiringOnDate(userID int64, date string) ([]domain.Product, *domain.AppError)
}

// StatsService handles calculating and retrieving fridge statistics.
type StatsService interface {
	GetFridgeStatistics(userID int64) (*domain.FridgeStats, *domain.AppError)
}

// FridgeInfoResponse combines Fridge data with dynamic info like 'products to next level'
type FridgeInfoResponse struct {
	domain.Fridge
	ProductsToNextLevel int64  `json:"products_to_next_level"`
	Description         string `json:"description"`
}

const (
	AchievementHelloWorld = "hello_world"
	AchievementEquator    = "equator"
	AchievementSalad      = "salad"
)

var levelThresholds = map[int64]int64{
	1: 0,
	2: 100,
	3: 1000,
}

func getLevel(consumedCount int64) int64 {
	currentLevel := int64(1)
	levels := []int64{1, 2, 3}
	for _, level := range levels {
		threshold, exists := levelThresholds[level]
		if exists && consumedCount >= threshold {
			currentLevel = level
		} else if exists && consumedCount < threshold {
			break
		}
	}
	return currentLevel
}

func getProductsToNextLevel(consumedCount int64) int64 {
	currentLevel := getLevel(consumedCount)
	nextLevel := currentLevel + 1
	nextThreshold, exists := levelThresholds[nextLevel]
	if !exists {
		return 0 // Max level reached or not defined
	}
	needed := nextThreshold - consumedCount
	if needed < 0 {
		needed = 0
	}
	return needed
}

func getFridgeDescription(level int64) string {
	switch level {
	case 1:
		return "Привет! Сейчас я бэйби-холодильник. Чтобы вырастить меня, корми меня побольше и следи за сроками хранения!"
	case 2:
		return "Ура! Я подрос! Продолжай в том же духе!"
	case 3:
		return "Я уже совсем большой холодильник!"
	default:
		return fmt.Sprintf("Я холодильник %d уровня!", level)
	}
}

// Service is the main struct holding all service implementations.
type Service struct {
	Auth    AuthService
	User    UserService
	Fridge  FridgeService
	Product ProductService
	Stats   StatsService
}

// NewService creates a new Service collection.
func NewService(repo repository.Store, jwtSecret string, jwtDuration time.Duration) *Service {

	authService := NewAuthServiceImpl(repo, jwtSecret, jwtDuration)
	fridgeService := NewFridgeServiceImpl(repo)
	userService := NewUserServiceImpl(repo)
	statsService := NewStatsServiceImpl(repo)
	productService := NewProductServiceImpl(repo, fridgeService)

	return &Service{
		Auth:    authService,
		User:    userService,
		Fridge:  fridgeService,
		Product: productService,
		Stats:   statsService,
	}
}
