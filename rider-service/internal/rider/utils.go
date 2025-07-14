package rider

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const riderIDKey contextKey = "rider_id"

func generateUUID() string {
	return uuid.New().String()
}

func generateJWT(riderID string, status string) (string, error) {
	claims := jwt.MapClaims{
		"rider_id": riderID,
		"status":   status,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func validateJWT(tokenString string) (jwt.MapClaims, error) {
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

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := validateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		riderID, ok := claims["rider_id"].(string)
		if !ok {
			http.Error(w, "Invalid token claims: missing rider_id", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), riderIDKey, riderID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRiderID(r *http.Request) string {
	if v := r.Context().Value(riderIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
