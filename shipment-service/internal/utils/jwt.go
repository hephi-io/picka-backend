package utils

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func RequireVendorRole(claims jwt.MapClaims) error {
	role, ok := claims["role"].(string)
	if !ok || role != "vendor" {
		return errors.New("forbidden: vendor role required")
	}
	return nil
}
