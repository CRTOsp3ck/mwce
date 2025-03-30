// internal/controller/player.go

package controller

import (
	"net/http"

	"mwce-be/internal/middleware"
	"mwce-be/internal/service"
	"mwce-be/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// PlayerController handles player-related HTTP requests
type PlayerController struct {
	playerService service.PlayerService
	logger        zerolog.Logger
}

// NewPlayerController creates a new player controller
func NewPlayerController(playerService service.PlayerService, logger zerolog.Logger) *PlayerController {
	return &PlayerController{
		playerService: playerService,
		logger:        logger,
	}
}

// GetProfile handles getting the player's profile
func (c *PlayerController) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get the player profile
	player, err := c.playerService.GetProfile(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get player profile")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get player profile")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, player)
}

// GetStats handles getting the player's statistics
func (c *PlayerController) GetStats(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get the player stats
	stats, err := c.playerService.GetStats(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get player stats")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get player stats")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, stats)
}

// GetNotifications handles getting the player's notifications
func (c *PlayerController) GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get the player notifications
	notifications, err := c.playerService.GetNotifications(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get player notifications")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get player notifications")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, notifications)
}

// MarkAllNotificationsRead handles marking all notifications as read
func (c *PlayerController) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Mark all notifications as read
	if err := c.playerService.MarkAllNotificationsRead(playerID); err != nil {
		c.logger.Error().Err(err).Msg("Failed to mark all notifications as read")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to mark all notifications as read")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// MarkNotificationRead handles marking a specific notification as read
func (c *PlayerController) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get notification ID from URL
	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Notification ID is required")
		return
	}

	// Mark notification as read
	if err := c.playerService.MarkNotificationRead(notificationID, playerID); err != nil {
		c.logger.Error().Err(err).Msg("Failed to mark notification as read")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// CollectAllPending handles collecting all pending resources
func (c *PlayerController) CollectAllPending(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Collect all pending resources
	response, err := c.playerService.CollectAllPending(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to collect pending resources")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to collect pending resources")
		return
	}

	// Return success response with game message
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		response,
		func() string {
			// response.CollectedAmount > 0 ? util.GameMessageTypeSuccess : util.GameMessageTypeInfo,
			if response.CollectedAmount > 0 {
				return util.GameMessageTypeSuccess
			}
			return util.GameMessageTypeInfo
		}(),
		response.Message,
	)
}
