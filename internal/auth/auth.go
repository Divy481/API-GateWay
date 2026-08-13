package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoToken     = errors.New("no token provided")
	ErrInvalidToken = errors.New("invalid token")
)

// Validator defines how tokens are validated
type Validator struct {
	SecretKey []byte
}

// NewValidator creates a new JWT validator
func NewValidator(secret string) *Validator {
	return &Validator{
		SecretKey: []byte(secret),
	}
}

// Validate checks the request for a valid JWT
func (v *Validator) Validate(r *http.Request) (*jwt.Token, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrNoToken
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, ErrInvalidToken
	}

	tokenString := parts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.SecretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}
