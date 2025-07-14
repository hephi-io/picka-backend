package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"vendor-service/internal/db"
	"vendor-service/internal/vendor"
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

	// Start NATS subscriber for shipment updates
	go vendor.StartNATSSubscriber()

	r := mux.NewRouter()

	// Secure vendor routes with JWT middleware
	vendorRouter := r.PathPrefix("/vendor").Subrouter()
	vendorRouter.Use(vendor.JWTVendorMiddleware)
	vendorRouter.HandleFunc("/register", vendor.RegisterVendorHandler).Methods("POST")
	vendorRouter.HandleFunc("/profile", vendor.GetVendorProfileHandler).Methods("GET")
	vendorRouter.HandleFunc("/profile", vendor.UpdateVendorProfileHandler).Methods("PUT")
	vendorRouter.HandleFunc("/products", vendor.AddProductHandler).Methods("POST")
	vendorRouter.HandleFunc("/products", vendor.ListProductsHandler).Methods("GET")
	vendorRouter.HandleFunc("/products/{id}", vendor.UpdateProductHandler).Methods("PUT")
	vendorRouter.HandleFunc("/products/{id}", vendor.DeleteProductHandler).Methods("DELETE")
	vendorRouter.HandleFunc("/delivery-request", vendor.CreateDeliveryRequestHandler).Methods("POST")
	vendorRouter.HandleFunc("/deliveries", vendor.ListDeliveriesHandler).Methods("GET")
	vendorRouter.HandleFunc("/summary", vendor.VendorSummaryHandler).Methods("GET")
	// New endpoints to be added below

	fmt.Println("Vendor service running on port 8083")
	http.ListenAndServe(":8083", r)
}
