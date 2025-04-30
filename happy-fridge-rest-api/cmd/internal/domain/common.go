package domain

import (
	"time"
)

// APIError represents a standard error response format
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PagingInfo for list responses
type PagingInfo struct {
	TotalItems   int `json:"total_items"`
	ItemsPerPage int `json:"items_per_page"`
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
}

// StandardResponse wraps successful responses
type StandardResponse struct {
	Data   interface{} `json:"data"`
	Paging *PagingInfo `json:"paging,omitempty"`
}

// NullableTime allows handling nullable timestamp columns
type NullableTime struct {
	Time  time.Time
	Valid bool
}

type ProductUnit string

const (
	UnitPcs ProductUnit = "pcs"
	UnitKg  ProductUnit = "kg"
	UnitL   ProductUnit = "l"
)

type ProductCategory string

const (
	CategoryMeat       ProductCategory = "Meat"
	CategoryMilk       ProductCategory = "Milk"
	CategoryVegetables ProductCategory = "Vegetables"
	CategoryFruit      ProductCategory = "Fruit"
	CategorySweets     ProductCategory = "Sweets"
	CategoryBakery     ProductCategory = "Bakery"
	CategorySeafood    ProductCategory = "Seafood"
	CategoryCooked     ProductCategory = "Cooked"
	CategoryOther      ProductCategory = "Other"
)

func (pu ProductUnit) IsCountable() bool {
	return pu == UnitPcs
}

func (pu ProductUnit) IsWeightOrVolume() bool {
	return pu == UnitKg || pu == UnitL
}

// ProductCategory gets the relevant consumed_expired field name based on category.
func (pc ProductCategory) GetConsumedExpFridgeField() string {
	switch pc {
	case CategoryMeat:
		return "consumed_meat_exp"
	case CategoryMilk:
		return "consumed_milk_exp"
	case CategoryVegetables:
		return "consumed_vegetables_exp"
	case CategoryFruit:
		return "consumed_fruit_exp"
	case CategorySweets:
		return "consumed_sweets_exp"
	case CategoryBakery:
		return "consumed_backery_exp"
	case CategorySeafood:
		return "consumed_seafood_exp"
	case CategoryCooked:
		return "consumed_cooked_exp"
	case CategoryOther:
		return "consumed_other_exp"
	}
	return ""
}
