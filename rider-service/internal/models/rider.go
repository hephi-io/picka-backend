package models

import (
	"time"
)

type RiderProfile struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Phone       string    `json:"phone"`
	VehicleType string    `json:"vehicle_type"`
	Status      string    `json:"status"`
	Location    string    `json:"location"` // WKT or GeoJSON string for now
	CreatedAt   time.Time `json:"created_at"`
}
