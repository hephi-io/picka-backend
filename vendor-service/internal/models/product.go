package models

type Product struct {
	ID          int     `json:"id"`
	VendorID    int     `json:"vendor_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
