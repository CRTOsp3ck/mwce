// internal/controller/auth.go

package controller

import (
	"encoding/json"
	"net/http"

	"mwce-be/internal/middleware"
	"mwce-be/internal/model"
	"mwce-be/internal/service"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	authService service.AuthService
	logger      zerolog.Logger
}

// NewAuthController creates a new auth controller
func NewAuthController(authService service.AuthService, logger zerolog.Logger) *AuthController {
	return &AuthController{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var request model.RegisterRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if request.Name == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if request.Email == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if request.Password == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	if request.Password != request.ConfirmPassword {
		util.RespondWithError(w, http.StatusBadRequest, "Passwords do not match")
		return
	}

	// if request.Territory == "" {
	// 	util.RespondWithError(w, http.StatusBadRequest, "Starting territory is required")
	// 	return
	// }

	// Register the user
	response, err := c.authService.Register(request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Registration failed")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusCreated, response)
}

// Login handles user login
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var request model.LoginRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if request.Email == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if request.Password == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Authenticate the user
	response, err := c.authService.Login(request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Login failed")
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, response)
}

// Validate handles token validation
func (c *AuthController) Validate(w http.ResponseWriter, r *http.Request) {
	// The AuthMiddleware has already verified the token at this point
	// Just return a success response with the player ID
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Token is valid",
		"user_id": userID,
	})
}
