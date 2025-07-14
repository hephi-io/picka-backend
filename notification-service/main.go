package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

type ShipmentEvent struct {
	Event      string `json:"event"`
	ShipmentID string `json:"shipment_id"`
	Status     string `json:"status"`
	RiderID    string `json:"rider_id"`
	Timestamp  string `json:"timestamp"`
}

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	subject := "shipment.events"
	_, err = nc.Subscribe(subject, func(m *nats.Msg) {
		var event ShipmentEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Printf("Invalid event data: %v", err)
			return
		}
		if event.Event == "shipment.updated" {
			log.Printf("Notification: Shipment %s status updated to %s by rider %s at %s", event.ShipmentID, event.Status, event.RiderID, event.Timestamp)
		}
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to subject: %v", err)
	}

	log.Printf("Notification service listening for '%s' events...", subject)
	select {} // Block forever
}
