package domain

import (
	"github.com/go-playground/validator/v10"
)

// Fridge aligns with the 'fridges' table schema.
type Fridge struct {
	ID                     int64  `json:"id" db:"id"`
	UserID                 int64  `json:"user_id" db:"user_id"`
	Code                   string `json:"code" db:"code"`
	Level                  int64  `json:"level" db:"level"`
	ConsumedOnTime         int64  `json:"consumed_on_time" db:"consumed_on_time"`
	ConsumedAll            int64  `json:"consumed_all" db:"consumed_all"`
	ConsumedMeatExp        int64  `json:"consumed_meat_exp" db:"consumed_meat_exp"`
	ConsumedMilkExp        int64  `json:"consumed_milk_exp" db:"consumed_milk_exp"`
	ConsumedVegetablesExp  int64  `json:"consumed_vegetables_exp" db:"consumed_vegetables_exp"`
	ConsumedFruitExp       int64  `json:"consumed_fruit_exp" db:"consumed_fruit_exp"`
	ConsumedSweetsExp      int64  `json:"consumed_sweets_exp" db:"consumed_sweets_exp"`
	ConsumedBakeryExp      int64  `json:"consumed_backery_exp" db:"consumed_backery_exp"`
	ConsumedSeafoodExp     int64  `json:"consumed_seafood_exp" db:"consumed_seafood_exp"`
	ConsumedCookedExp      int64  `json:"consumed_cooked_exp" db:"consumed_cooked_exp"`
	ConsumedOtherExp       int64  `json:"consumed_other_exp" db:"consumed_other_exp"`
	IsOneDayNotification   bool   `json:"is_one_day_notification" db:"is_one_day_notification"`
	IsThreeDayNotification bool   `json:"is_three_day_notification" db:"is_three_day_notification"`
	IsFiveDayNotification  bool   `json:"is_five_day_notification" db:"is_five_day_notification"`
	AchHelloWorld          bool   `json:"ach_hello_world" db:"ach_hello_world"`
	AchEquator             bool   `json:"ach_equator" db:"ach_equator"`
	AchSalad               bool   `json:"ach_salad" db:"ach_salad"`
}

// FridgeStats represents the data needed for the stats screen
type FridgeStats struct {
	ConsumedOnTimePercentage  float64            `json:"consumed_on_time_percentage"`
	ExpiredByCategory         map[string]int64   `json:"expired_by_category"`
	ExpiredCategoryPercentage map[string]float64 `json:"expired_category_percentage"`
}

type UpdateNotificationSettingsRequest struct {
	IsOneDay   *bool `json:"is_one_day_notification"`
	IsThreeDay *bool `json:"is_three_day_notification"`
	IsFiveDay  *bool `json:"is_five_day_notification"`
}

// Validate checks the fridge struct fields based on validation tags.
func (f *Fridge) Validate() error {
	validate := validator.New()
	return validate.Struct(f)
}
