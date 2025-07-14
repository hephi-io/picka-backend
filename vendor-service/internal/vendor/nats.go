package vendor

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"vendor-service/internal/db"

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
		if event.Role != "vendor" {
			return
		}
		_, err := db.DB.Exec(`INSERT INTO vendor_profiles (id, user_id, business_name, business_address, logo_url, created_at) VALUES (gen_random_uuid(), $1, $2, '', '', $3) ON CONFLICT (user_id) DO NOTHING`,
			event.UserID, event.BusinessName, time.Now())
		if err != nil {
			log.Printf("Failed to create vendor profile for user %s: %v", event.UserID, err)
		} else {
			log.Printf("Created vendor profile for user %s", event.UserID)
		}
	})
}
