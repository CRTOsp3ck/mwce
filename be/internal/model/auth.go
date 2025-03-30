// internal/model/auth.go

package model

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
	Avatar          string `json:"avatar"`
	Territory       string `json:"territory" binding:"required"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	Token  string `json:"token"`
	Player Player `json:"player"`
}
