package domain

import (
	"database/sql"

	"github.com/go-playground/validator/v10"
)

type User struct {
	ID             int64          `json:"id" db:"id"`
	FridgeCode     sql.NullString `json:"fridge_code,omitempty" db:"fridge_code"`
	Login          string         `json:"login" db:"login" validate:"required,min=3,max=50"`
	Password       string         `json:"password,omitempty" validate:"omitempty,min=8,max=40"`
	HashedPassword string         `json:"-" db:"hashed_password"`
}

// AuthRequest is used for login requests
type AuthRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse contains the token upon successful login/registration
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=40"`
}

func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}
