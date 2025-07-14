package rider

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"rider-service/internal/db"

	"github.com/nats-io/nats.go"
)

type UserCreatedEvent struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

func StartNATSSubscriber() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	nc.Subscribe("user.created", func(m *nats.Msg) {
		var event UserCreatedEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Printf("Invalid user.created event: %v", err)
			return
		}
		if event.Role != "rider" {
			return
		}
		_, err := db.DB.Exec(`INSERT INTO rider_profiles (id, user_id, phone, vehicle_type, status, location, created_at) VALUES (gen_random_uuid(), $1, $2, '', 'offline', NULL, $3) ON CONFLICT (user_id) DO NOTHING`,
			event.UserID, event.Phone, time.Now())
		if err != nil {
			log.Printf("Failed to create rider profile for user %s: %v", event.UserID, err)
		} else {
			log.Printf("Created rider profile for user %s", event.UserID)
		}
	})
}

type ShipmentEvent struct {
	Event      string `json:"event"`
	ShipmentID string `json:"shipment_id"`
	Status     string `json:"status"`
	RiderID    string `json:"rider_id"`
	Timestamp  string `json:"timestamp"`
}

func PublishShipmentUpdated(shipmentID, status, riderID string) error {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return err
	}
	defer nc.Close()
	event := ShipmentEvent{
		Event:      "shipment.updated",
		ShipmentID: shipmentID,
		Status:     status,
		RiderID:    riderID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(event)
	return nc.Publish("shipment.events", data)
}
