package vendor

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"vendor-service/internal/db"
	"vendor-service/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Helper: generateUUID
func generateUUID() string {
	return uuid.New().String()
}

// POST /vendor/register
func RegisterVendorHandler(w http.ResponseWriter, r *http.Request) {
	var vendor models.VendorProfile
	if err := json.NewDecoder(r.Body).Decode(&vendor); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	query := `INSERT INTO vendor_profiles (user_id, business_name, business_address) VALUES ($1, $2, $3) RETURNING id`
	err := db.DB.QueryRow(query, vendor.UserID, vendor.BusinessName, vendor.BusinessAddress).Scan(&vendor.ID)
	if err != nil {
		http.Error(w, "Error registering vendor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vendor)
}

// GET /vendor/profile
func GetVendorProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetVendorID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var vendor models.VendorProfile
	query := `SELECT id, user_id, business_name, business_address FROM vendor_profiles WHERE user_id = $1`
	err := db.DB.QueryRow(query, strconv.Itoa(userID)).Scan(&vendor.ID, &vendor.UserID, &vendor.BusinessName, &vendor.BusinessAddress)
	if err == sql.ErrNoRows {
		http.Error(w, "Vendor not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error fetching vendor profile", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(vendor)
}

// PUT /vendor/profile
func UpdateVendorProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetVendorID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var vendor models.VendorProfile
	if err := json.NewDecoder(r.Body).Decode(&vendor); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	query := `UPDATE vendor_profiles SET business_name=$1, business_address=$2 WHERE user_id=$3 RETURNING id, user_id, business_name, business_address`
	err := db.DB.QueryRow(query, vendor.BusinessName, vendor.BusinessAddress, strconv.Itoa(userID)).Scan(&vendor.ID, &vendor.UserID, &vendor.BusinessName, &vendor.BusinessAddress)
	if err == sql.ErrNoRows {
		http.Error(w, "Vendor not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error updating vendor profile", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(vendor)
}

// POST /vendor/profile
func CreateVendorProfileHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID          string `json:"user_id"`
		BusinessName    string `json:"business_name"`
		BusinessAddress string `json:"business_address"`
		LogoURL         string `json:"logo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		vendorID := GetVendorID(r)
		if vendorID == 0 {
			http.Error(w, "Missing user_id", http.StatusBadRequest)
			return
		}
		req.UserID = strconv.Itoa(vendorID)
	}
	id := generateUUID()
	_, err := db.DB.Exec(`INSERT INTO vendor_profiles (id, user_id, business_name, business_address, logo_url, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, req.UserID, req.BusinessName, req.BusinessAddress, req.LogoURL, time.Now())
	if err != nil {
		http.Error(w, "Error creating vendor profile", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "user_id": req.UserID})
}

// POST /vendor/products
func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	query := `INSERT INTO vendor.products (vendor_id, name, description, price) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.DB.QueryRow(query, vendorID, product.Name, product.Description, product.Price).Scan(&product.ID)
	if err != nil {
		http.Error(w, "Error adding product", http.StatusInternalServerError)
		return
	}
	product.VendorID = vendorID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// GET /vendor/products
func ListProductsHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := db.DB.Query(`SELECT id, vendor_id, name, description, price FROM vendor.products WHERE vendor_id = $1`, vendorID)
	if err != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.VendorID, &p.Name, &p.Description, &p.Price)
		if err != nil {
			continue
		}
		products = append(products, p)
	}
	json.NewEncoder(w).Encode(products)
}

// PUT /vendor/products/{id}
func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	query := `UPDATE vendor.products SET name=$1, description=$2, price=$3 WHERE id=$4 AND vendor_id=$5 RETURNING id, vendor_id, name, description, price`
	err = db.DB.QueryRow(query, product.Name, product.Description, product.Price, id, vendorID).Scan(&product.ID, &product.VendorID, &product.Name, &product.Description, &product.Price)
	if err == sql.ErrNoRows {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error updating product", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(product)
}

// DELETE /vendor/products/{id}
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	query := `DELETE FROM vendor.products WHERE id=$1 AND vendor_id=$2`
	result, err := db.DB.Exec(query, id, vendorID)
	if err != nil {
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /vendor/delivery-request
func CreateDeliveryRequestHandler(w http.ResponseWriter, r *http.Request) {
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	resp, err := forwardToShipmentService("POST", "/shipments", payload, jwt)
	if err != nil {
		http.Error(w, "Error contacting shipment service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// GET /vendor/deliveries
func ListDeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	resp, err := forwardToShipmentService("GET", "/shipments", nil, jwt)
	if err != nil {
		http.Error(w, "Error contacting shipment service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// GET /vendor/summary
func VendorSummaryHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := GetVendorID(r)
	if vendorID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Count products
	var totalProducts int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM vendor.products WHERE vendor_id = $1`, vendorID).Scan(&totalProducts)
	if err != nil {
		http.Error(w, "Error counting products", http.StatusInternalServerError)
		return
	}
	// Get deliveries from shipment-service
	jwt := r.Header.Get("Authorization")
	resp, err := forwardToShipmentService("GET", "/shipments", nil, jwt)
	if err != nil {
		http.Error(w, "Error contacting shipment service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	var deliveries []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&deliveries)
	pending := 0
	deliveredToday := 0
	today := time.Now().Format("2006-01-02")
	for _, d := range deliveries {
		if d["status"] == "pending" {
			pending++
		}
		if created, ok := d["created_at"].(string); ok && len(created) >= 10 && created[:10] == today && d["status"] == "delivered" {
			deliveredToday++
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_products":     totalProducts,
		"pending_deliveries": pending,
		"delivered_today":    deliveredToday,
	})
}
