package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"rider-service/internal/db"
	"rider-service/internal/rider"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Initialize database
	db.Init()
	db.RunMigrations(db.DB)

	r := mux.NewRouter()

	r.HandleFunc("/riders/signup", rider.SignupHandler).Methods("POST")
	r.HandleFunc("/riders/login", rider.LoginHandler).Methods("POST")

	// Protected endpoints (JWT middleware)
	protected := r.PathPrefix("/riders").Subrouter()
	protected.Use(rider.JWTMiddleware)
	protected.HandleFunc("/profile", rider.ProfileHandler).Methods("GET")
	protected.HandleFunc("/assignments", rider.AssignmentsHandler).Methods("GET")
	protected.HandleFunc("/assignments/{id}/status", rider.UpdateAssignmentStatusHandler).Methods("PATCH")

	fmt.Println("Rider service running on port 8084")
	http.ListenAndServe(":8084", r)
}
