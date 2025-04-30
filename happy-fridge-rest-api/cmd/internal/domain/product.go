package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Product aligns with the 'products' table schema
type Product struct {
	ID            int64           `json:"id" db:"id"`
	FridgeID      int64           `json:"fridge_id" db:"fridge_id"`
	Name          string          `json:"name" db:"name" validate:"required,max=255"`
	Category      ProductCategory `json:"category" db:"category" validate:"required"`
	Quantity      float64         `json:"quantity" db:"quantity" validate:"required,gt=0"`
	Unit          ProductUnit     `json:"unit" db:"unit" validate:"required"`
	ExpiryDate    int64           `json:"-" db:"expiry_date"`
	DaysLeft      int             `json:"days_left"`
	ExpiryDateStr string          `json:"expiry_date"`
}

// ProductCreateRequest is used when adding a new product
type ProductCreateRequest struct {
	Name       string          `json:"name" validate:"required,max=255"`
	Category   ProductCategory `json:"category" validate:"required,oneof=Meat Milk Vegetables Fruit Sweets Bakery Seafood Cooked Drinks Other"` // Use ENUM values
	Quantity   float64         `json:"quantity" validate:"required,gt=0"`
	Unit       ProductUnit     `json:"unit" validate:"required,oneof=pcs kg l g ml"`        // Use ENUM values
	ExpiryDate string          `json:"expiry_date" validate:"required,datetime=2006-01-02"` // Expect YYYY-MM-DD format
}

// ProductUpdateRequest is used when editing a product
type ProductUpdateRequest struct {
	Name       *string          `json:"name,omitempty" validate:"omitempty,max=255"` // Use pointers for optional fields
	Category   *ProductCategory `json:"category,omitempty" validate:"omitempty,oneof=Meat Milk Vegetables Fruit Sweets Bakery Seafood Cooked Drinks Other"`
	Quantity   *float64         `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	Unit       *ProductUnit     `json:"unit,omitempty" validate:"omitempty,oneof=pcs kg l g ml"`
	ExpiryDate *string          `json:"expiry_date,omitempty" validate:"omitempty,datetime=2006-01-02"` // Expect YYYY-MM-DD format
}

// ProductListParams holds query parameters for fetching products.
type ProductListParams struct {
	FridgeID int64
	Category *ProductCategory
	Search   *string
	SortBy   string
}

// Validate checks the product struct fields based on validation tags.
func (p *Product) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// CalculateDaysLeft computes the difference between expiry date and today.
func (p *Product) CalculateDaysLeft() {
	expiryTime := time.Unix(p.ExpiryDate, 0)

	now := time.Now()

	year, month, day := now.Date()
	todayStart := time.Date(year, month, day, 0, 0, 0, 0, now.Location()) // Полночь сегодня

	expiryYear, expiryMonth, expiryDay := expiryTime.Date()
	expiryDayStart := time.Date(expiryYear, expiryMonth, expiryDay, 0, 0, 0, 0, expiryTime.Location()) // Полночь в день истечения срока

	if expiryDayStart.Before(todayStart) {
		duration := expiryDayStart.Sub(todayStart)
		p.DaysLeft = int(duration.Hours() / 24)

	} else if expiryDayStart.Equal(todayStart) {
		p.DaysLeft = 0
	} else {
		duration := expiryDayStart.Sub(todayStart)
		p.DaysLeft = int(duration.Hours() / 24)
	}
	p.ExpiryDateStr = expiryTime.Format("2006-01-02")
}

// IsExpired checks if the product's expiry date has passed
func (p *Product) IsExpired() bool {
	// Compare expiry timestamp with current time's timestamp
	return p.ExpiryDate < time.Now().Unix()
}
