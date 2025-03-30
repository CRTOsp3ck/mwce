package controller

import (
	"context"
	"net/http"

	"mwce-be/internal/service"

	"github.com/rs/zerolog"
)

// SSEController handles SSE connections
type SSEController struct {
	authService service.AuthService
	sseService  service.SSEService
	logger      zerolog.Logger
}

// NewSSEController creates a new SSE controller
func NewSSEController(
	authService service.AuthService,
	sseService service.SSEService,
	logger zerolog.Logger) *SSEController {
	return &SSEController{
		authService: authService,
		sseService:  sseService,
		logger:      logger,
	}
}

// HandleConnection establishes an SSE connection
func (c *SSEController) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter instead of Authorization header
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	// Validate token and get player ID
	playerID, err := c.authService.ValidateToken(token) // You'll need to inject authService
	if err != nil {
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	// Add player ID to request context
	ctx := context.WithValue(r.Context(), "userID", playerID)

	// Handle SSE connection with the context containing player ID
	c.sseService.HandleConnection(w, r.WithContext(ctx))
}
