package model

import (
	"github.com/go-playground/validator/v10"
)

func passwordValidation(fl validator.FieldLevel) bool {
	var user *User
	switch parent := fl.Parent().Interface().(type) {
	case *User:
		user = parent
	case User:
		user = &parent
	default:
		return false
	}

	if user.EncryptedPassword != "" {
		return true
	}

	pass := fl.Field().String()
	if len(pass) < 8 || len(pass) > 40 {
		return false
	}
	return true
}
