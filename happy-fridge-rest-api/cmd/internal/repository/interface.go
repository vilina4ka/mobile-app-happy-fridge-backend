package repository

import (
	"github.com/jmoiron/sqlx"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
)

// UserRepository defines methods for interacting with user data.
type UserRepository interface {
	Create(user *domain.User) error
	FindByLogin(login string) (*domain.User, error)
	FindByID(id int64) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id int64) error
}

// FridgeRepository defines methods for interacting with fridge data.
type FridgeRepository interface {
	Create(fridge *domain.Fridge) error
	FindByUserID(userID int64) (*domain.Fridge, error)
	FindByID(id int64) (*domain.Fridge, error)
	Update(fridge *domain.Fridge) error
	IncrementCounter(fridgeID int64, fieldName string, value int64) error
	UpdateNotificationSettings(fridgeID int64, settings *domain.UpdateNotificationSettingsRequest) error
	UpdateAchievements(fridgeID int64, achHelloWorld, achEquator, achSalad *bool) error
	GetStats(fridgeID int64) (*domain.Fridge, error)
}

// ProductRepository defines methods for interacting with product data.
type ProductRepository interface {
	Create(product *domain.Product) error
	FindByID(id int64) (*domain.Product, error)
	ListByFridgeID(params domain.ProductListParams) ([]domain.Product, error)
	Update(product *domain.Product) error
	Delete(id int64) error
	UpdateQuantity(productID int64, change float64) (*domain.Product, error)
	CountByFridgeID(fridgeID int64) (int64, error)
	GetProductsExpiringOnDate(fridgeID int64, date int64) ([]domain.Product, error)
	GetProductsExpiringSoon(fridgeID int64, notificationDays []int) ([]domain.Product, error)
}

// Store combines all repositories (useful for dependency injection)
type Store interface {
	User() UserRepository
	Fridge() FridgeRepository
	Product() ProductRepository
	Close() error
	Ping() error
	DB() *sqlx.DB
}
