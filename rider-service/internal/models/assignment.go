package models

import "time"

type Assignment struct {
	ID         string    `json:"id"`
	RiderID    string    `json:"rider_id"`
	ShipmentID string    `json:"shipment_id"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
}
