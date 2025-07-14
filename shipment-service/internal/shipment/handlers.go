package shipment

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"shipment-service/internal/db"
	"shipment-service/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// POST /shipments
func CreateShipmentHandler(w http.ResponseWriter, r *http.Request) {
	var shipment models.Shipment
	if err := json.NewDecoder(r.Body).Decode(&shipment); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	vendorID := GetVendorID(r)
	if vendorID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id := uuid.New().String()
	createdAt := time.Now().Format(time.RFC3339)
	query := `INSERT INTO shipments (id, vendor_id, product_id, quantity, destination, notes, status, pickup_time, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.DB.Exec(query, id, vendorID, shipment.ProductID, shipment.Quantity, shipment.Destination, shipment.Notes, "pending", shipment.PickupTime, createdAt)
	if err != nil {
		http.Error(w, "Error creating shipment", http.StatusInternalServerError)
		return
	}
	shipment.ID = id
	shipment.VendorID = vendorID
	shipment.Status = "pending"
	shipment.CreatedAt = createdAt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shipment)
}

// GET /shipments
func ListShipmentsHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := db.DB.Query(`SELECT id, vendor_id, product_id, quantity, destination, notes, status, pickup_time, created_at FROM shipments WHERE vendor_id = $1`, vendorID)
	if err != nil {
		http.Error(w, "Error fetching shipments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	shipments := []models.Shipment{}
	for rows.Next() {
		var s models.Shipment
		err := rows.Scan(&s.ID, &s.VendorID, &s.ProductID, &s.Quantity, &s.Destination, &s.Notes, &s.Status, &s.PickupTime, &s.CreatedAt)
		if err != nil {
			continue
		}
		shipments = append(shipments, s)
	}
	json.NewEncoder(w).Encode(shipments)
}

// GET /shipments/{id}
func GetShipmentHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	var s models.Shipment
	err := db.DB.QueryRow(`SELECT id, vendor_id, product_id, quantity, destination, notes, status, pickup_time, created_at FROM shipments WHERE id = $1 AND vendor_id = $2`, id, vendorID).
		Scan(&s.ID, &s.VendorID, &s.ProductID, &s.Quantity, &s.Destination, &s.Notes, &s.Status, &s.PickupTime, &s.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Shipment not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error fetching shipment", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(s)
}

// PATCH /shipments/{id}
func UpdateShipmentStatusHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	type statusPayload struct {
		Status string `json:"status"`
	}
	var payload statusPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// Only update if the shipment belongs to the vendor
	res, err := db.DB.Exec(`UPDATE shipments SET status = $1 WHERE id = $2 AND vendor_id = $3`, payload.Status, id, vendorID)
	if err != nil {
		http.Error(w, "Error updating shipment status", http.StatusInternalServerError)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "Shipment not found or forbidden", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": payload.Status})
}
