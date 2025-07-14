package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	auth "auth-service/internal/auth"
	"auth-service/internal/db"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Initialize database
	db.Init()
	auth.InitNATS()

	r := mux.NewRouter()

	r.HandleFunc("/auth/vendor/signup", auth.VendorSignupHandler).Methods("POST")
	r.HandleFunc("/auth/vendor/login", auth.VendorLoginHandler).Methods("POST")
	r.HandleFunc("/auth/rider/signup", auth.RiderSignupHandler).Methods("POST")
	r.HandleFunc("/auth/rider/login", auth.RiderLoginHandler).Methods("POST")
	r.HandleFunc("/auth/request-password-reset", auth.RequestPasswordResetHandler).Methods("POST")
	r.HandleFunc("/auth/reset-password", auth.ResetPasswordHandler).Methods("POST")
	r.HandleFunc("/auth/verify-email", auth.VerifyEmailHandler).Methods("POST")

	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(auth.AdminJWTMiddleware)
	adminRouter.HandleFunc("/users/{user_id}/status", auth.AdminUpdateUserStatusHandler).Methods("PATCH")
	adminRouter.HandleFunc("/users/{user_id}/reset-password", auth.AdminResetUserPasswordHandler).Methods("POST")

	fmt.Println("Auth service running on port 8081")
	http.ListenAndServe(":8081", r)
}
