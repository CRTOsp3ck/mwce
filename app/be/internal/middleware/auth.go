// internal/middleware/auth.go

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"mwce-be/internal/service"
	"mwce-be/internal/util"
)

// Key type is used for context values
type Key string

const (
	// UserIDKey is the key for user ID in context
	UserIDKey Key = "userID"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate verifies the JWT token in the Authorization header
func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			util.RespondWithError(w, http.StatusUnauthorized, "No authorization header provided")
			return
		}

		// Check if the header is in the correct format
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			util.RespondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		// Extract and validate the token
		token := headerParts[1]
		userID, err := am.authService.ValidateToken(token)
		if err != nil {
			fmt.Println(err)
			util.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
