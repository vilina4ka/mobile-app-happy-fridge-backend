package auth

import (
	"errors" // Import errors package
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for a given user ID.
func GenerateJWT(userID int64, secretKey string, duration time.Duration) (string, *domain.AppError) {
	expirationTime := time.Now().Add(duration)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "happy-fridge-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", domain.ErrInternalServer(fmt.Errorf("failed to sign JWT token: %w", err))
	}

	return tokenString, nil
}

// ValidateJWT checks the validity of a token string and returns the claims if valid.
func ValidateJWT(tokenString string, secretKey string) (*Claims, *domain.AppError) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, domain.ErrUnauthorized() // Invalid signature
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.NewAppError(err, "Token has expired", 401)
		}
		return nil, domain.NewAppError(err, fmt.Sprintf("Invalid token: %v", err), 401) // Generic invalid token with details
	}

	if !token.Valid {
		return nil, domain.ErrUnauthorized() // Token is not valid for other reasons
	}
	return claims, nil
}
