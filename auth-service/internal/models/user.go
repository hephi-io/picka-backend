package models

type SignUpRequest struct {
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password"`
}

type User struct {
	ID       string
	Email    string
	Phone    string
	Password string
	Role     string
}
