package shipment

import (
	"encoding/json"
	"fmt"
	"os"
	"shipment-service/internal/db"

	"github.com/nats-io/nats.go"
)

type ShipmentEvent struct {
	Event      string `json:"event"`
	ShipmentID string `json:"shipment_id"`
	Status     string `json:"status"`
	RiderID    string `json:"rider_id"`
	Timestamp  string `json:"timestamp"`
}

func StartNATSSubscriber() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		fmt.Println("NATS connection error:", err)
		return
	}

	nc.Subscribe("shipment.events", func(m *nats.Msg) {
		var event ShipmentEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			fmt.Println("NATS event unmarshal error:", err)
			return
		}
		if event.Event == "shipment.updated" && event.ShipmentID != "" && event.Status != "" {
			_, err := db.DB.Exec(`UPDATE shipments SET status = $1 WHERE id = $2`, event.Status, event.ShipmentID)
			if err != nil {
				fmt.Println("Failed to update shipment status:", err)
			} else {
				fmt.Printf("Shipment %s status updated to %s via NATS\n", event.ShipmentID, event.Status)
			}
		}
	})

	fmt.Println("NATS subscriber for shipment events started.")
	select {} // Block forever
}
