// internal/controller/travel.go

package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mwce-be/internal/middleware"
	"mwce-be/internal/model"
	"mwce-be/internal/service"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
)

// TravelController handles travel-related HTTP requests
type TravelController struct {
	travelService service.TravelService
	logger        zerolog.Logger
}

// NewTravelController creates a new travel controller
func NewTravelController(travelService service.TravelService, logger zerolog.Logger) *TravelController {
	return &TravelController{
		travelService: travelService,
		logger:        logger,
	}
}

// GetAvailableRegions handles getting regions available for travel
func (c *TravelController) GetAvailableRegions(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get available regions
	regions, err := c.travelService.GetAvailableRegions(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get available regions")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get available regions")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, regions)
}

// GetCurrentRegion handles getting a player's current region
func (c *TravelController) GetCurrentRegion(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get current region
	region, err := c.travelService.GetCurrentRegion(playerID)
	if err != nil {
		// If player has no current region, return null instead of error
		if err.Error() == "player has no current region" {
			util.RespondWithJSON(w, http.StatusOK, nil)
			return
		}

		c.logger.Error().Err(err).Msg("Failed to get current region")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get current region")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, region)
}

// Travel handles player travel to a new region
func (c *TravelController) Travel(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request model.TravelRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if request.RegionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Region ID is required")
		return
	}

	// Perform travel
	result, err := c.travelService.Travel(playerID, request.RegionID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to travel")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Determine message type based on success
	messageType := util.GameMessageTypeSuccess
	if !result.Success {
		messageType = util.GameMessageTypeWarning
	}

	// Return success response with game message
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		result,
		messageType,
		result.Message,
	)
}

// GetTravelHistory handles getting a player's travel history
func (c *TravelController) GetTravelHistory(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get limit from query parameters, default to 10
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get travel history
	history, err := c.travelService.GetTravelHistory(playerID, limit)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get travel history")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get travel history")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, history)
}
