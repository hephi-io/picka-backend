package rider

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"rider-service/internal/db"
	"rider-service/internal/models"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// POST /riders/signup
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Vehicle  string `json:"vehicle"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	id := generateUUID()
	_, err = db.DB.Exec(`INSERT INTO riders (id, name, phone, password_hash, vehicle, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, req.Name, req.Phone, string(hash), req.Vehicle, "available", time.Now())
	if err != nil {
		http.Error(w, "Error creating rider", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// POST /riders/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	var id, hash, name, vehicle, status string
	err := db.DB.QueryRow(`SELECT id, password_hash, name, vehicle, status FROM riders WHERE phone = $1`, req.Phone).Scan(&id, &hash, &name, &vehicle, &status)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Error fetching rider", http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := generateJWT(id, status)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// GET /riders/profile
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	riderID := GetRiderID(r)
	var rider models.RiderProfile
	err := db.DB.QueryRow(
		`SELECT id, user_id, phone, vehicle_type, status, location, created_at FROM rider_profiles WHERE user_id = $1`,
		riderID,
	).Scan(&rider.ID, &rider.UserID, &rider.Phone, &rider.VehicleType, &rider.Status, &rider.Location, &rider.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Rider not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error fetching profile", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rider)
}

// GET /riders/assignments
func AssignmentsHandler(w http.ResponseWriter, r *http.Request) {
	riderID := GetRiderID(r)
	rows, err := db.DB.Query(`SELECT id, rider_id, shipment_id, status, updated_at FROM assignments WHERE rider_id = $1`, riderID)
	if err != nil {
		http.Error(w, "Error fetching assignments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	assignments := []models.Assignment{}
	for rows.Next() {
		var a models.Assignment
		err := rows.Scan(&a.ID, &a.RiderID, &a.ShipmentID, &a.Status, &a.UpdatedAt)
		if err != nil {
			continue
		}
		assignments = append(assignments, a)
	}
	json.NewEncoder(w).Encode(assignments)
}

// PATCH /riders/assignments/:id/status
func UpdateAssignmentStatusHandler(w http.ResponseWriter, r *http.Request) {
	riderID := GetRiderID(r)
	vars := mux.Vars(r)
	id := vars["id"]
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	var shipmentID string
	_ = db.DB.QueryRow(`SELECT shipment_id FROM assignments WHERE id = $1 AND rider_id = $2`, id, riderID).Scan(&shipmentID)
	_, err := db.DB.Exec(`UPDATE assignments SET status = $1, updated_at = $2 WHERE id = $3 AND rider_id = $4`, req.Status, time.Now(), id, riderID)
	if err != nil {
		http.Error(w, "Error updating assignment", http.StatusInternalServerError)
		return
	}
	// Publish event to NATS
	_ = PublishShipmentUpdated(shipmentID, req.Status, riderID)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": req.Status})
}

// POST /rider/profile
func CreateRiderProfileHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string `json:"user_id"`
		Phone       string `json:"phone"`
		VehicleType string `json:"vehicle_type"`
		Status      string `json:"status"`
		Location    string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		// Optionally extract from JWT
		riderID := GetRiderID(r)
		if riderID == "" {
			http.Error(w, "Missing user_id", http.StatusBadRequest)
			return
		}
		req.UserID = riderID
	}
	id := generateUUID()
	_, err := db.DB.Exec(`INSERT INTO rider_profiles (id, user_id, phone, vehicle_type, status, location, created_at) VALUES ($1, $2, $3, $4, $5, ST_GeogFromText($6), $7)`,
		id, req.UserID, req.Phone, req.VehicleType, req.Status, req.Location, time.Now())
	if err != nil {
		http.Error(w, "Error creating rider profile", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "user_id": req.UserID})
}

// Helper: generateUUID (use github.com/google/uuid)
/* func generateUUID() string {
	return "TODO"
} */

// Helper: generateJWT (placeholder)
/* func generateJWT(riderID string) (string, error) {
	return "TODO", nil
} */

// Helper: GetRiderID (from JWT context, placeholder)
/* func GetRiderID(r *http.Request) string {
	return "TODO"
} */
