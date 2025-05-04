// internal/controller/campaign.go

package controller

import (
	"encoding/json"
	"net/http"

	"mwce-be/internal/middleware"
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

// GetCampaigns handles getting all available campaigns
func (c *CampaignController) GetCampaigns(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get campaigns
	campaigns, err := c.campaignService.GetAllCampaigns()
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get campaigns")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get campaigns")
		return
	}

	// Get player progress for each campaign
	campaignProgress := make(map[string]interface{})
	progresses, err := c.campaignService.GetPlayerCampaignProgresses(playerID)
	if err == nil {
		for _, progress := range progresses {
			campaignProgress[progress.CampaignID] = progress
		}
	}

	// Return success response with both campaigns and progress
	response := map[string]interface{}{
		"campaigns": campaigns,
		"progress":  campaignProgress,
	}

	util.RespondWithJSON(w, http.StatusOK, response)
}

// GetCampaignDetail handles getting a specific campaign with chapters and missions
func (c *CampaignController) GetCampaignDetail(w http.ResponseWriter, r *http.Request) {
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

	// Get campaign with chapters
	campaign, err := c.campaignService.GetCampaignByID(campaignID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get campaign")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get campaign")
		return
	}

	// Get player progress for this campaign
	progress, err := c.campaignService.GetPlayerCampaignProgress(playerID, campaignID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get campaign progress")
	}

	// Get mission progress for all missions in the campaign
	missionProgress := make(map[string]interface{})
	for _, chapter := range campaign.Chapters {
		for _, mission := range chapter.Missions {
			missionProg, err := c.campaignService.GetPlayerMissionProgress(playerID, mission.ID)
			if err == nil && missionProg != nil {
				missionProgress[mission.ID] = missionProg
			}
		}
	}

	// Return success response with campaign, progress, and mission progress
	response := map[string]interface{}{
		"campaign":        campaign,
		"progress":        progress,
		"missionProgress": missionProgress,
	}

	util.RespondWithJSON(w, http.StatusOK, response)
}

// GetChapter handles getting a specific chapter with missions
func (c *CampaignController) GetChapter(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get chapter ID from URL
	chapterID := chi.URLParam(r, "id")
	if chapterID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Chapter ID is required")
		return
	}

	// Get chapter with missions
	chapter, err := c.campaignService.GetChapterByID(chapterID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get chapter")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get chapter")
		return
	}

	// Get mission progress for all missions in the chapter
	missionProgress := make(map[string]interface{})
	for _, mission := range chapter.Missions {
		missionProg, err := c.campaignService.GetPlayerMissionProgress(playerID, mission.ID)
		if err == nil && missionProg != nil {
			missionProgress[mission.ID] = missionProg
		}
	}

	// Return success response with chapter and mission progress
	response := map[string]interface{}{
		"chapter":         chapter,
		"missionProgress": missionProgress,
	}

	util.RespondWithJSON(w, http.StatusOK, response)
}

// GetMission handles getting a specific mission with details
func (c *CampaignController) GetMission(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get mission ID from URL
	missionID := chi.URLParam(r, "id")
	if missionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Mission ID is required")
		return
	}

	// Get mission with choices
	mission, err := c.campaignService.GetMissionByID(missionID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get mission")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get mission")
		return
	}

	// Get mission progress for this mission
	progress, err := c.campaignService.GetPlayerMissionProgress(playerID, missionID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get mission progress")
	}

	// Check requirements
	meetsRequirements, failedRequirements, err := c.campaignService.CheckMissionRequirements(playerID, missionID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to check mission requirements")
	}

	// Return success response with mission, progress, and requirements check
	response := map[string]interface{}{
		"mission":            mission,
		"progress":           progress,
		"meetsRequirements":  meetsRequirements,
		"failedRequirements": failedRequirements,
	}

	util.RespondWithJSON(w, http.StatusOK, response)
}

// StartCampaign handles starting a campaign for a player
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
		c.logger.Error().Err(err).Msg("Failed to start campaign")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		progress,
		util.GameMessageTypeSuccess,
		"Campaign started successfully. Your journey begins now.",
	)
}

// StartMission handles starting a mission for a player
func (c *CampaignController) StartMission(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get mission ID from URL
	missionID := chi.URLParam(r, "id")
	if missionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Mission ID is required")
		return
	}

	// Start mission
	progress, err := c.campaignService.StartMission(playerID, missionID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to start mission")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get mission for message
	mission, err := c.campaignService.GetMissionByID(missionID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get mission")
	}

	missionName := "mission"
	if mission != nil {
		missionName = mission.Title
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		progress,
		util.GameMessageTypeSuccess,
		"Mission '"+missionName+"' started. Complete the objectives to advance.",
	)
}

// CompleteMission handles completing a mission for a player
func (c *CampaignController) CompleteMission(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get mission ID from URL
	missionID := chi.URLParam(r, "id")
	if missionID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Mission ID is required")
		return
	}

	// Parse request body
	var request struct {
		ChoiceID string `json:"choiceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Complete mission
	result, err := c.campaignService.CompleteMission(playerID, missionID, request.ChoiceID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to complete mission")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		result,
		util.GameMessageTypeSuccess,
		result.Message,
	)
}
