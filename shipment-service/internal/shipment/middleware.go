package shipment

import (
	"context"
	"net/http"
	"strings"

	"shipment-service/internal/utils"
)

type contextKey string

const vendorIDKey contextKey = "vendor_id"

func JWTVendorMiddleware(next http.Handler) http.Handler {
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
		if err := utils.RequireVendorRole(claims); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		vendorID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Invalid token claims: missing user_id", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), vendorIDKey, vendorID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetVendorID(r *http.Request) string {
	if v := r.Context().Value(vendorIDKey); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
