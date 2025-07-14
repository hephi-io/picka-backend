package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"shipment-service/internal/db"
	"shipment-service/internal/shipment"
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

	// Start NATS subscriber for async shipment status updates
	go shipment.StartNATSSubscriber()

	r := mux.NewRouter()

	r.HandleFunc("/shipments", shipment.CreateShipmentHandler).Methods("POST")
	r.HandleFunc("/shipments", shipment.ListShipmentsHandler).Methods("GET")
	r.HandleFunc("/shipments/{id}", shipment.GetShipmentHandler).Methods("GET")
	r.HandleFunc("/shipments/{id}", shipment.UpdateShipmentStatusHandler).Methods("PATCH")

	fmt.Println("Shipment service running on port 8082")
	http.ListenAndServe(":8082", r)
}
