// be/internal/controller/campaign.go

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

// CampaignController handles campaign-related HTTP requests
type CampaignController struct {
	campaignService service.CampaignService
	logger          zerolog.Logger
}

// NewCampaignController creates a new campaign controller
func NewCampaignController(campaignService service.CampaignService, logger zerolog.Logger) *CampaignController {
	return &CampaignController{
		campaignService: campaignService,
		logger:          logger,
	}
}

// GetCampaigns handles getting all campaigns
func (c *CampaignController) GetCampaigns(w http.ResponseWriter, r *http.Request) {
	// Get campaigns
	campaigns, err := c.campaignService.GetCampaigns()
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get campaigns")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get campaigns")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, campaigns)
}

// GetCampaign handles getting a specific campaign
func (c *CampaignController) GetCampaign(w http.ResponseWriter, r *http.Request) {
	// Get campaign ID from URL
	campaignID := chi.URLParam(r, "id")
	if campaignID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	// Get campaign
	campaign, err := c.campaignService.GetCampaignByID(campaignID)
	if err != nil {
		c.logger.Error().Err(err).Str("campaignID", campaignID).Msg("Failed to get campaign")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get campaign")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, campaign)
}

// GetChapter handles getting a specific chapter
func (c *CampaignController) GetChapter(w http.ResponseWriter, r *http.Request) {
	// Get chapter ID from URL
	chapterID := chi.URLParam(r, "id")
	if chapterID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Chapter ID is required")
		return
	}

	// Get chapter
	chapter, err := c.campaignService.GetChapterByID(chapterID)
	if err != nil {
		c.logger.Error().Err(err).Str("chapterID", chapterID).Msg("Failed to get chapter")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get chapter")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, chapter)
}

// GetMission handles getting a specific mission
func (c *CampaignController) GetMission(w http.ResponseWriter, r *http.Request) {
	// Get mission ID from URL
	missionID := chi.URLParam(r, "id")
	if missionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Mission ID is required")
		return
	}

	// Get mission
	mission, err := c.campaignService.GetMissionByID(missionID)
	if err != nil {
		c.logger.Error().Err(err).Str("missionID", missionID).Msg("Failed to get mission")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get mission")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, mission)
}

// GetBranch handles getting a specific branch
func (c *CampaignController) GetBranch(w http.ResponseWriter, r *http.Request) {
	// Get branch ID from URL
	branchID := chi.URLParam(r, "id")
	if branchID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Branch ID is required")
		return
	}

	// Get branch
	branch, err := c.campaignService.GetBranchByID(branchID)
	if err != nil {
		c.logger.Error().Err(err).Str("branchID", branchID).Msg("Failed to get branch")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get branch")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, branch)
}

// GetPlayerProgress handles getting a player's campaign progress
func (c *CampaignController) GetPlayerProgress(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get campaign ID from URL
	campaignID := chi.URLParam(r, "id")
	if campaignID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	// Get progress
	progress, err := c.campaignService.GetPlayerCampaignProgress(playerID, campaignID)
	if err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("campaignID", campaignID).Msg("Failed to get player campaign progress")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get player campaign progress")
		return
	}

	// If no progress, return empty object
	if progress == nil {
		util.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"started": false,
		})
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, progress)
}

// StartCampaign handles starting a campaign
func (c *CampaignController) StartCampaign(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get campaign ID from URL
	campaignID := chi.URLParam(r, "id")
	if campaignID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	// Start campaign
	progress, err := c.campaignService.StartCampaign(playerID, campaignID)
	if err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("campaignID", campaignID).Msg("Failed to start campaign")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to start campaign")
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		progress,
		util.GameMessageTypeSuccess,
		"Campaign started successfully!",
	)
}

// GetCurrentMission handles getting a player's current mission
func (c *CampaignController) GetCurrentMission(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get campaign ID from URL
	campaignID := chi.URLParam(r, "id")
	if campaignID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	// Get current mission
	mission, err := c.campaignService.GetCurrentMission(playerID, campaignID)
	if err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("campaignID", campaignID).Msg("Failed to get current mission")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get current mission")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, mission)
}

// SelectBranch handles selecting a branch for a mission
func (c *CampaignController) SelectBranch(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get mission ID from URL
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Mission ID is required")
		return
	}

	// Parse request body
	var request struct {
		BranchID string `json:"branchId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Select branch
	if err := c.campaignService.SelectBranch(playerID, missionID, request.BranchID); err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("missionID", missionID).Str("branchID", request.BranchID).Msg("Failed to select branch")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to select branch")
		return
	}

	// Get the branch for the response
	branch, err := c.campaignService.GetBranchByID(request.BranchID)
	if err != nil {
		c.logger.Error().Err(err).Str("branchID", request.BranchID).Msg("Failed to get branch")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get branch")
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		branch,
		util.GameMessageTypeSuccess,
		"Branch selected successfully!",
	)
}

// be/internal/controller/campaign.go (continued)

// GetPOIsByBranch handles getting POIs for a branch
func (c *CampaignController) GetPOIsByBranch(w http.ResponseWriter, r *http.Request) {
	// Get branch ID from URL
	branchID := chi.URLParam(r, "id")
	if branchID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Branch ID is required")
		return
	}

	// Get POIs
	pois, err := c.campaignService.GetPOIsByBranchID(branchID)
	if err != nil {
		c.logger.Error().Err(err).Str("branchID", branchID).Msg("Failed to get POIs")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get POIs")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, pois)
}

// GetPOI handles getting a specific POI
func (c *CampaignController) GetPOI(w http.ResponseWriter, r *http.Request) {
	// Get POI ID from URL
	poiID := chi.URLParam(r, "id")
	if poiID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "POI ID is required")
		return
	}

	// Get POI
	poi, err := c.campaignService.GetPOIByID(poiID)
	if err != nil {
		c.logger.Error().Err(err).Str("poiID", poiID).Msg("Failed to get POI")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get POI")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, poi)
}

// GetPOIDialogues handles getting dialogues for a POI
func (c *CampaignController) GetPOIDialogues(w http.ResponseWriter, r *http.Request) {
	// Get POI ID from URL
	poiID := chi.URLParam(r, "id")
	if poiID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "POI ID is required")
		return
	}

	// Get dialogues
	dialogues, err := c.campaignService.GetDialoguesByPOIID(poiID)
	if err != nil {
		c.logger.Error().Err(err).Str("poiID", poiID).Msg("Failed to get dialogues")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get dialogues")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, dialogues)
}

// InteractWithPOI handles interaction with a POI
func (c *CampaignController) InteractWithPOI(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get POI ID from URL
	poiID := chi.URLParam(r, "id")
	if poiID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "POI ID is required")
		return
	}

	// Parse request body
	var request struct {
		InteractionType string `json:"interactionType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Convert interaction type
	interactionType := model.InteractionType(request.InteractionType)

	// Interact with POI
	dialogue, resourceEffect, err := c.campaignService.InteractWithPOI(playerID, poiID, interactionType)
	if err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("poiID", poiID).Str("interactionType", request.InteractionType).Msg("Failed to interact with POI")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to interact with POI")
		return
	}

	// Return success response
	response := map[string]interface{}{
		"dialogue":       dialogue,
		"resourceEffect": resourceEffect,
	}

	util.RespondWithJSON(w, http.StatusOK, response)
}

// CompletePOI handles marking a POI as completed
func (c *CampaignController) CompletePOI(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get POI ID from URL
	poiID := chi.URLParam(r, "id")
	if poiID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "POI ID is required")
		return
	}

	// Complete POI
	if err := c.campaignService.CompletePOI(playerID, poiID); err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("poiID", poiID).Msg("Failed to complete POI")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to complete POI")
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		map[string]bool{"success": true},
		util.GameMessageTypeSuccess,
		"POI completed successfully!",
	)
}

// GetOperationsByBranch handles getting operations for a branch
func (c *CampaignController) GetOperationsByBranch(w http.ResponseWriter, r *http.Request) {
	// Get branch ID from URL
	branchID := chi.URLParam(r, "id")
	if branchID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Branch ID is required")
		return
	}

	// Get operations
	operations, err := c.campaignService.GetOperationsByBranchID(branchID)
	if err != nil {
		c.logger.Error().Err(err).Str("branchID", branchID).Msg("Failed to get operations")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get operations")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, operations)
}

// CompleteOperation handles marking an operation as completed
func (c *CampaignController) CompleteOperation(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get operation ID from URL
	operationID := chi.URLParam(r, "id")
	if operationID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Operation ID is required")
		return
	}

	// Parse request body
	var request struct {
		AttemptID string `json:"attemptId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Complete operation
	if err := c.campaignService.CompleteOperation(playerID, operationID, request.AttemptID); err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("operationID", operationID).Str("attemptID", request.AttemptID).Msg("Failed to complete operation")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to complete operation")
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		map[string]bool{"success": true},
		util.GameMessageTypeSuccess,
		"Operation completed successfully!",
	)
}

// CheckBranchCompletion handles checking if a branch is complete
func (c *CampaignController) CheckBranchCompletion(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get branch ID from URL
	branchID := chi.URLParam(r, "id")
	if branchID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Branch ID is required")
		return
	}

	// Check branch completion
	complete, err := c.campaignService.CheckBranchCompletion(playerID, branchID)
	if err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("branchID", branchID).Msg("Failed to check branch completion")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to check branch completion")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, map[string]bool{"complete": complete})
}

// CompleteBranch handles completing a branch
func (c *CampaignController) CompleteBranch(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get mission ID and branch ID from URL
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Mission ID is required")
		return
	}

	branchID := chi.URLParam(r, "branchId")
	if branchID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Branch ID is required")
		return
	}

	// Complete branch
	if err := c.campaignService.CompleteBranch(playerID, missionID, branchID); err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("missionID", missionID).Str("branchID", branchID).Msg("Failed to complete branch")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to complete branch")
		return
	}

	// Get player's progress to return updated state
	// First need to get campaign ID
	mission, err := c.campaignService.GetMissionByID(missionID)
	if err != nil {
		c.logger.Error().Err(err).Str("missionID", missionID).Msg("Failed to get mission")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get mission")
		return
	}

	// Get the chapter to find campaign ID
	chapter, err := c.campaignService.GetChapterByID(mission.ChapterID)
	if err != nil {
		c.logger.Error().Err(err).Str("chapterID", mission.ChapterID).Msg("Failed to get chapter")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get chapter")
		return
	}

	// Get updated progress
	progress, err := c.campaignService.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		c.logger.Error().Err(err).Str("playerID", playerID).Str("campaignID", chapter.CampaignID).Msg("Failed to get player campaign progress")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get player campaign progress")
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		progress,
		util.GameMessageTypeSuccess,
		"Branch completed successfully!",
	)
}

// GetChaptersByCampaign handles getting chapters for a campaign
func (c *CampaignController) GetChaptersByCampaign(w http.ResponseWriter, r *http.Request) {
	// Get campaign ID from URL
	campaignID := chi.URLParam(r, "id")
	if campaignID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Campaign ID is required")
		return
	}

	// Get chapters
	chapters, err := c.campaignService.GetChaptersByCampaignID(campaignID)
	if err != nil {
		c.logger.Error().Err(err).Str("campaignID", campaignID).Msg("Failed to get chapters")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get chapters")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, chapters)
}

// GetMissionsByChapter handles getting missions for a chapter
func (c *CampaignController) GetMissionsByChapter(w http.ResponseWriter, r *http.Request) {
	// Get chapter ID from URL
	chapterID := chi.URLParam(r, "id")
	if chapterID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Chapter ID is required")
		return
	}

	// Get missions
	missions, err := c.campaignService.GetMissionsByChapterID(chapterID)
	if err != nil {
		c.logger.Error().Err(err).Str("chapterID", chapterID).Msg("Failed to get missions")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get missions")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, missions)
}

// GetBranchesByMission handles getting branches for a mission
func (c *CampaignController) GetBranchesByMission(w http.ResponseWriter, r *http.Request) {
	// Get mission ID from URL
	missionID := chi.URLParam(r, "id")
	if missionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Mission ID is required")
		return
	}

	// Get branches
	branches, err := c.campaignService.GetBranchesByMissionID(missionID)
	if err != nil {
		c.logger.Error().Err(err).Str("missionID", missionID).Msg("Failed to get branches")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get branches")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, branches)
}


