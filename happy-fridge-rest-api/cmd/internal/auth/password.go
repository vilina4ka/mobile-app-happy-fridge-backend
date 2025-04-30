package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
)

// HashPassword generates a bcrypt hash for the given password string.
func HashPassword(password string) (string, *domain.AppError) {
	if password == "" {
		return "", domain.ErrBadRequest("Password cannot be empty")
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", domain.ErrInternalServer(fmt.Errorf("failed to hash password: %w", err))
	}
	return string(hashedBytes), nil
}

// CheckPassword compares a plaintext password with a stored bcrypt hash.
func CheckPassword(hashedPassword, password string) bool {
	if hashedPassword == "" || password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
