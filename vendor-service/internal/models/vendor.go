package models

import "time"

type VendorProfile struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	BusinessName    string    `json:"business_name"`
	BusinessAddress string    `json:"business_address"`
	LogoURL         string    `json:"logo_url"`
	CreatedAt       time.Time `json:"created_at"`
}
