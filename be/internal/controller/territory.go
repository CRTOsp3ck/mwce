// internal/controller/territory.go

package controller

import (
	"encoding/json"
	"net/http"

	"mwce-be/internal/middleware"
	"mwce-be/internal/model"
	"mwce-be/internal/service"
	"mwce-be/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// TerritoryController handles territory-related HTTP requests
type TerritoryController struct {
	territoryService service.TerritoryService
	logger           zerolog.Logger
}

// NewTerritoryController creates a new territory controller
func NewTerritoryController(territoryService service.TerritoryService, logger zerolog.Logger) *TerritoryController {
	return &TerritoryController{
		territoryService: territoryService,
		logger:           logger,
	}
}

// GetRegions handles getting all regions
func (c *TerritoryController) GetRegions(w http.ResponseWriter, r *http.Request) {
	// Get all regions
	regions, err := c.territoryService.GetAllRegions()
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get regions")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get regions")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, regions)
}

// GetRegion handles getting a specific region
func (c *TerritoryController) GetRegion(w http.ResponseWriter, r *http.Request) {
	// Get region ID from URL
	regionID := chi.URLParam(r, "id")
	if regionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Region ID is required")
		return
	}

	// Get the region
	region, err := c.territoryService.GetRegionByID(regionID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get region")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get region")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, region)
}

// GetDistricts handles getting all districts or districts in a region
func (c *TerritoryController) GetDistricts(w http.ResponseWriter, r *http.Request) {
	// Check if region ID is provided as a query parameter
	regionID := r.URL.Query().Get("regionId")

	var districts []model.District
	var err error

	if regionID != "" {
		// Get districts in the specified region
		districts, err = c.territoryService.GetDistrictsByRegionID(regionID)
	} else {
		// Get all districts
		districts, err = c.territoryService.GetAllDistricts()
	}

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get districts")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get districts")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, districts)
}

// GetDistrict handles getting a specific district
func (c *TerritoryController) GetDistrict(w http.ResponseWriter, r *http.Request) {
	// Get district ID from URL
	districtID := chi.URLParam(r, "id")
	if districtID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "District ID is required")
		return
	}

	// Get the district
	district, err := c.territoryService.GetDistrictByID(districtID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get district")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get district")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, district)
}

// GetCities handles getting all cities or cities in a district
func (c *TerritoryController) GetCities(w http.ResponseWriter, r *http.Request) {
	// Check if district ID is provided as a query parameter
	districtID := r.URL.Query().Get("districtId")

	var cities []model.City
	var err error

	if districtID != "" {
		// Get cities in the specified district
		cities, err = c.territoryService.GetCitiesByDistrictID(districtID)
	} else {
		// Get all cities
		cities, err = c.territoryService.GetAllCities()
	}

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get cities")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get cities")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, cities)
}

// GetCity handles getting a specific city
func (c *TerritoryController) GetCity(w http.ResponseWriter, r *http.Request) {
	// Get city ID from URL
	cityID := chi.URLParam(r, "id")
	if cityID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "City ID is required")
		return
	}

	// Get the city
	city, err := c.territoryService.GetCityByID(cityID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get city")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get city")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, city)
}

// GetHotspots handles getting all hotspots or hotspots in a city
func (c *TerritoryController) GetHotspots(w http.ResponseWriter, r *http.Request) {
	// Check if city ID is provided as a query parameter
	cityID := r.URL.Query().Get("cityId")

	var hotspots []model.Hotspot
	var err error

	if cityID != "" {
		// Get hotspots in the specified city
		hotspots, err = c.territoryService.GetHotspotsByCity(cityID)
	} else {
		// Get all hotspots
		hotspots, err = c.territoryService.GetAllHotspots()
	}

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get hotspots")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get hotspots")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, hotspots)
}

// GetHotspot handles getting a specific hotspot
func (c *TerritoryController) GetHotspot(w http.ResponseWriter, r *http.Request) {
	// Get hotspot ID from URL
	hotspotID := chi.URLParam(r, "id")
	if hotspotID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Hotspot ID is required")
		return
	}

	// Get the hotspot
	hotspot, err := c.territoryService.GetHotspotByID(hotspotID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get hotspot")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get hotspot")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, hotspot)
}

// GetControlledHotspots handles getting hotspots controlled by the player
func (c *TerritoryController) GetControlledHotspots(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get controlled hotspots
	hotspots, err := c.territoryService.GetControlledHotspots(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get controlled hotspots")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get controlled hotspots")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, hotspots)
}

// GetRecentActions handles getting recent territory actions
func (c *TerritoryController) GetRecentActions(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get recent actions
	actions, err := c.territoryService.GetRecentActions(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get recent actions")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get recent actions")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, actions)
}

// PerformAction handles performing a territory action
func (c *TerritoryController) PerformAction(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get action type from URL
	actionType := chi.URLParam(r, "action")
	if actionType == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Action type is required")
		return
	}

	// Validate action type
	validActionTypes := []string{
		util.TerritoryActionTypeExtortion,
		util.TerritoryActionTypeTakeover,
		util.TerritoryActionTypeCollection,
		util.TerritoryActionTypeDefend,
	}

	isValidAction := false
	for _, validType := range validActionTypes {
		if actionType == validType {
			isValidAction = true
			break
		}
	}

	if !isValidAction {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid action type")
		return
	}

	// Parse request body
	var request model.PerformActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if request.HotspotID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Hotspot ID is required")
		return
	}

	// Perform the action
	result, err := c.territoryService.PerformAction(playerID, actionType, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to perform action")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response with game message
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		result,
		func() string {
			// result.Success ? util.GameMessageTypeSuccess : util.GameMessageTypeError,
			if result.Success {
				return util.GameMessageTypeSuccess
			}
			return util.GameMessageTypeError
		}(),
		result.Message,
	)
}

// CollectHotspotIncome handles collecting income from a specific hotspot
func (c *TerritoryController) CollectHotspotIncome(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get hotspot ID from URL
	hotspotID := chi.URLParam(r, "id")
	if hotspotID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Hotspot ID is required")
		return
	}

	// Collect income from the hotspot
	response, err := c.territoryService.CollectHotspotIncome(playerID, hotspotID)
	if err != nil {
		c.logger.Error().Err(err).
			Str("playerID", playerID).
			Str("hotspotID", hotspotID).
			Msg("Failed to collect hotspot income")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response with game message
	messageType := util.GameMessageTypeSuccess
	if response.CollectedAmount <= 0 {
		messageType = util.GameMessageTypeInfo
	}

	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		response,
		messageType,
		response.Message,
	)

	// Get updated hotspot
	updatedHotspot, err := c.territoryService.GetHotspotByID(hotspotID)
	if err == nil {
		// Send SSE event to notify about the hotspot update
		c.territoryService.GetSSEService().SendEventToPlayer(
			playerID,
			"hotspot_updated",
			map[string]interface{}{
				"hotspot": updatedHotspot,
			},
		)
	}
}

// CollectAllHotspotIncome handles collecting income from all controlled hotspots
func (c *TerritoryController) CollectAllHotspotIncome(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Collect income from all hotspots
	response, err := c.territoryService.CollectAllHotspotIncome(playerID)
	if err != nil {
		c.logger.Error().Err(err).
			Str("playerID", playerID).
			Msg("Failed to collect all hotspot income")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response with game message
	messageType := util.GameMessageTypeSuccess
	if response.CollectedAmount <= 0 {
		messageType = util.GameMessageTypeInfo
	}

	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		response,
		messageType,
		response.Message,
	)

	// Get all controlled hotspots
	controlledHotspots, err := c.territoryService.GetControlledHotspots(playerID)
	if err == nil {
		// Send SSE event to notify about all hotspot updates
		c.territoryService.GetSSEService().SendEventToPlayer(
			playerID,
			"hotspots_updated",
			map[string]interface{}{
				"hotspots": controlledHotspots,
			},
		)
	}
}
