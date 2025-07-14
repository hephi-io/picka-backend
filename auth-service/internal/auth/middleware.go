package auth

import (
	"net/http"
	"strings"

	"auth-service/internal/utils"
)

func AdminJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			http.Error(w, "Forbidden: admin only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
