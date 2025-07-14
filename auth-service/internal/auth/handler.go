package auth

import (
	"auth-service/internal/db"
	"auth-service/internal/models"
	"auth-service/internal/utils"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"log"
	"time"

	"database/sql"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
)

var natsConn *nats.Conn

func InitNATS() {
	var err error
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	natsConn, err = nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
}

type UserCreatedEvent struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name,omitempty"`
}

func publishUserCreatedEvent(event UserCreatedEvent) {
	if natsConn == nil {
		return
	}
	data, _ := json.Marshal(event)
	natsConn.Publish("user.created", data)
}

// POST /auth/vendor/signup
func VendorSignupHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SignUpRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" || strings.TrimSpace(req.BusinessName) == "" {
		http.Error(w, "Email, password, first_name, last_name, and business_name required", http.StatusBadRequest)
		return
	}
	req.Role = "vendor"

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var userID string
	err = db.DB.QueryRow(`INSERT INTO users (email, password, role, first_name, last_name, business_name) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`, req.Email, hashedPassword, req.Role, req.FirstName, req.LastName, req.BusinessName).Scan(&userID)
	if err != nil {
		http.Error(w, "Error creating vendor", http.StatusInternalServerError)
		return
	}

	publishUserCreatedEvent(UserCreatedEvent{
		UserID:       userID,
		Role:         req.Role,
		Email:        req.Email,
		Phone:        req.Phone,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BusinessName: req.BusinessName,
	})

	// Send verification email if email is present
	if req.Email != "" {
		token, _ := utils.GenerateRandomToken(32)
		expiry := time.Now().Add(24 * time.Hour)
		_, _ = db.DB.Exec(`UPDATE users SET email_verification_token = $1, email_verification_expires = $2 WHERE id = $3`, token, expiry, userID)
		verifyURL := os.Getenv("FRONTEND_URL") + "/verify-email?token=" + token
		subject := "Verify your email"
		body := "Click the link to verify your email: " + verifyURL
		_ = utils.SendEmail(req.Email, subject, body)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"user_id": userID, "role": req.Role})
}

// POST /auth/vendor/login
func VendorLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	var user models.User
	var isActive bool
	err := db.DB.QueryRow(`SELECT id, password, role, is_active FROM users WHERE email = $1 AND role = 'vendor'`, req.Email).Scan(&user.ID, &user.Password, &user.Role, &isActive)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !isActive {
		http.Error(w, "Account is inactive", http.StatusForbidden)
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Update last_login
	_, _ = db.DB.Exec(`UPDATE users SET last_login = NOW() WHERE id = $1`, user.ID)

	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token, "user_id": user.ID, "role": user.Role})
}

// POST /auth/rider/signup
func RiderSignupHandler(w http.ResponseWriter, r *http.Request) {
	var req models.SignUpRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if strings.TrimSpace(req.Phone) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.FirstName) == "" || strings.TrimSpace(req.LastName) == "" {
		http.Error(w, "Phone, password, first_name, and last_name required", http.StatusBadRequest)
		return
	}
	req.Role = "rider"

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	var userID string
	err = db.DB.QueryRow(`INSERT INTO users (phone, password, role, first_name, last_name) VALUES ($1, $2, $3, $4, $5) RETURNING id`, req.Phone, hashedPassword, req.Role, req.FirstName, req.LastName).Scan(&userID)
	if err != nil {
		http.Error(w, "Error creating rider", http.StatusInternalServerError)
		return
	}

	publishUserCreatedEvent(UserCreatedEvent{
		UserID:    userID,
		Role:      req.Role,
		Email:     req.Email,
		Phone:     req.Phone,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"user_id": userID, "role": req.Role})
}

// POST /auth/rider/login
func RiderLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if strings.TrimSpace(req.Phone) == "" || strings.TrimSpace(req.Password) == "" {
		http.Error(w, "Phone and password required", http.StatusBadRequest)
		return
	}

	var user models.User
	var isActive bool
	err := db.DB.QueryRow(`SELECT id, password, role, is_active FROM users WHERE phone = $1 AND role = 'rider'`, req.Phone).Scan(&user.ID, &user.Password, &user.Role, &isActive)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !isActive {
		http.Error(w, "Account is inactive", http.StatusForbidden)
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Update last_login
	_, _ = db.DB.Exec(`UPDATE users SET last_login = NOW() WHERE id = $1`, user.ID)

	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token, "user_id": user.ID, "role": user.Role})
}

// POST /auth/request-password-reset
func RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}
	var userID string
	err := db.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, req.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Do not reveal if email exists
		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	token, _ := utils.GenerateRandomToken(32)
	expiry := time.Now().Add(1 * time.Hour)
	_, _ = db.DB.Exec(`UPDATE users SET password_reset_token = $1, reset_token_expires = $2 WHERE id = $3`, token, expiry, userID)
	resetURL := os.Getenv("FRONTEND_URL") + "/reset-password?token=" + token
	subject := "Password Reset Request"
	body := "Click the link to reset your password: " + resetURL
	_ = utils.SendEmail(req.Email, subject, body)
	w.WriteHeader(http.StatusOK)
}

// POST /auth/reset-password
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" || req.Password == "" {
		http.Error(w, "Token and password required", http.StatusBadRequest)
		return
	}
	var userID string
	var expires time.Time
	err := db.DB.QueryRow(`SELECT id, reset_token_expires FROM users WHERE password_reset_token = $1`, req.Token).Scan(&userID, &expires)
	if err == sql.ErrNoRows || time.Now().After(expires) {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	hash, _ := utils.HashPassword(req.Password)
	_, _ = db.DB.Exec(`UPDATE users SET password = $1, password_reset_token = NULL, reset_token_expires = NULL WHERE id = $2`, hash, userID)
	w.WriteHeader(http.StatusOK)
}

// POST /auth/verify-email
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "Token required", http.StatusBadRequest)
		return
	}
	var userID string
	var expires time.Time
	err := db.DB.QueryRow(`SELECT id, email_verification_expires FROM users WHERE email_verification_token = $1`, req.Token).Scan(&userID, &expires)
	if err == sql.ErrNoRows || time.Now().After(expires) {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	_, _ = db.DB.Exec(`UPDATE users SET email_verified = TRUE, email_verification_token = NULL, email_verification_expires = NULL WHERE id = $1`, userID)
	w.WriteHeader(http.StatusOK)
}

// PATCH /admin/users/{user_id}/status
func AdminUpdateUserStatusHandler(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromJWT(r) // helper to extract admin user_id from JWT
	vars := mux.Vars(r)
	userID := vars["user_id"]
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	_, err := db.DB.Exec(`UPDATE users SET is_active = $1 WHERE id = $2`, req.IsActive, userID)
	if err != nil {
		http.Error(w, "Error updating user status", http.StatusInternalServerError)
		return
	}
	log.Printf("[AUDIT] admin %s set is_active=%v for user %s at %s", adminID, req.IsActive, userID, time.Now().Format(time.RFC3339))
	json.NewEncoder(w).Encode(map[string]interface{}{"user_id": userID, "is_active": req.IsActive})
}

// POST /admin/users/{user_id}/reset-password
func AdminResetUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	adminID := getAdminIDFromJWT(r)
	vars := mux.Vars(r)
	userID := vars["user_id"]
	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	_, err = db.DB.Exec(`UPDATE users SET password = $1 WHERE id = $2`, hash, userID)
	if err != nil {
		http.Error(w, "Error resetting password", http.StatusInternalServerError)
		return
	}
	log.Printf("[AUDIT] admin %s reset password for user %s at %s", adminID, userID, time.Now().Format(time.RFC3339))
	json.NewEncoder(w).Encode(map[string]interface{}{"user_id": userID, "reset": true})
}

// Helper to extract admin user_id from JWT
func getAdminIDFromJWT(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	tokenString := strings.TrimPrefix(header, "Bearer ")
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		return ""
	}
	userID, _ := claims["user_id"].(string)
	return userID
}
