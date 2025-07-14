package models

type Shipment struct {
	ID          string `json:"id"`
	VendorID    string `json:"vendor_id"`
	ProductID   string `json:"product_id"`
	Quantity    int    `json:"quantity"`
	Destination string `json:"destination"`
	Notes       string `json:"notes,omitempty"`
	Status      string `json:"status"`
	PickupTime  string `json:"pickup_time,omitempty"`
	CreatedAt   string `json:"created_at"`
}
