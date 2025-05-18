// internal/service/campaign.go

package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
)

// CampaignService handles campaign-related business logic
type CampaignService interface {
	// Campaign and mission management
	GetAllCampaigns() ([]model.Campaign, error)
	GetCampaignByID(campaignID string) (*model.Campaign, error)
	GetChapterByID(chapterID string) (*model.Chapter, error)
	GetMissionByID(missionID string) (*model.Mission, error)

	// Player progress tracking
	GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error)
	GetPlayerCampaignProgresses(playerID string) ([]model.PlayerCampaignProgress, error)
	GetPlayerMissionProgress(playerID string, missionID string) (*model.PlayerMissionProgress, error)

	// Campaign interactions
	StartCampaign(playerID string, campaignID string) (*model.PlayerCampaignProgress, error)
	StartMission(playerID string, missionID string) (*model.PlayerMissionProgress, error)
	CompleteMission(playerID string, missionID string, choiceID string) (*model.MissionCompleteResult, error)
	CheckMissionRequirements(playerID string, missionID string) (bool, []string, error)

	// POI management
	GetActivePlayerPOIs(playerID string) ([]model.PlayerPOI, error)
	ActivatePlayerPOI(playerID string, templateID string) (*model.PlayerPOI, error)
	CompletePlayerPOI(playerID string, playerPOIID string) error
	GetActivePOIsForMission(playerID string, missionID string) ([]model.PlayerPOI, error)

	// Mission Operation management
	GetActivePlayerMissionOperations(playerID string) ([]model.PlayerMissionOperation, error)
	ActivatePlayerMissionOperation(playerID string, templateID string) (*model.PlayerMissionOperation, error)
	CompletePlayerMissionOperation(playerID string, playerOpID string) error
	GetActiveOperationsForMission(playerID string, missionID string) ([]model.PlayerMissionOperation, error)

	// Player action tracking
	TrackPlayerAction(playerID string, actionType string, actionValue string) error
	CheckChoiceCompletion(playerID string, missionID string, choiceID string) (bool, error)
	ActivateChoice(playerID string, missionID string, choiceID string) error

	// Initialization
	LoadCampaigns(dirPath string) error

	// Provider interface implementations
	GetInjectedOperations(playerID string, regionID *string) ([]model.Operation, error)
	GetInjectedHotspots(playerID string, regionID *string) ([]model.Hotspot, error)
}

// MissionCompleteResult contains the results of completing a mission
type MissionCompleteResult struct {
	Success     bool                         `json:"success"`
	Rewards     model.MissionRewards         `json:"rewards"`
	NextMission *model.Mission               `json:"nextMission,omitempty"`
	Progress    *model.PlayerMissionProgress `json:"progress"`
	Message     string                       `json:"message"`
}

type campaignService struct {
	campaignRepo      repository.CampaignRepository
	playerRepo        repository.PlayerRepository
	playerService     PlayerService
	operationsService OperationsService
	territoryService  TerritoryService
	sseService        SSEService
	logger            zerolog.Logger
}

// NewCampaignService creates a new campaign service
func NewCampaignService(
	campaignRepo repository.CampaignRepository,
	playerRepo repository.PlayerRepository,
	playerService PlayerService,
	operationsService OperationsService,
	territoryService TerritoryService,
	sseService SSEService,
	logger zerolog.Logger,
) CampaignService {
	return &campaignService{
		campaignRepo:      campaignRepo,
		playerRepo:        playerRepo,
		playerService:     playerService,
		operationsService: operationsService,
		territoryService:  territoryService,
		sseService:        sseService,
		logger:            logger,
	}
}

// GetAllCampaigns retrieves all available campaigns
func (s *campaignService) GetAllCampaigns() ([]model.Campaign, error) {
	return s.campaignRepo.GetAllCampaigns()
}

// GetCampaignByID retrieves a campaign by ID
func (s *campaignService) GetCampaignByID(campaignID string) (*model.Campaign, error) {
	return s.campaignRepo.GetCampaignByID(campaignID)
}

// GetChapterByID retrieves a chapter by ID
func (s *campaignService) GetChapterByID(chapterID string) (*model.Chapter, error) {
	return s.campaignRepo.GetChapterByID(chapterID)
}

// GetMissionByID retrieves a mission by ID
func (s *campaignService) GetMissionByID(missionID string) (*model.Mission, error) {
	return s.campaignRepo.GetMissionByID(missionID)
}

// GetPlayerCampaignProgress retrieves a player's progress for a campaign
func (s *campaignService) GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error) {
	return s.campaignRepo.GetPlayerCampaignProgress(playerID, campaignID)
}

// GetPlayerCampaignProgresses retrieves all campaign progress for a player
func (s *campaignService) GetPlayerCampaignProgresses(playerID string) ([]model.PlayerCampaignProgress, error) {
	return s.campaignRepo.GetPlayerCampaignProgresses(playerID)
}

// GetPlayerMissionProgress retrieves a player's progress for a mission
func (s *campaignService) GetPlayerMissionProgress(playerID string, missionID string) (*model.PlayerMissionProgress, error) {
	// Get basic progress info from database
	progress, err := s.campaignRepo.GetPlayerMissionProgress(playerID, missionID)
	if err != nil {
		return nil, err
	}

	if progress == nil {
		return nil, nil
	}

	// Get the mission
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return nil, err
	}

	// Initialize objectives array
	progress.Objectives = []model.MissionObjective{}

	// If there's an active choice, get its conditions, POIs, and operations
	if progress.CurrentActiveChoice != "" {
		// Add condition objectives
		conditions, err := s.campaignRepo.GetPlayerCompletionConditions(playerID, progress.CurrentActiveChoice)
		if err == nil {
			for _, condition := range conditions {
				description := s.getConditionDescription(condition)
				objective := model.MissionObjective{
					Type:        "condition",
					Description: description,
					Target:      condition.RequiredValue,
					IsCompleted: condition.IsCompleted,
					CompletedAt: condition.CompletedAt,
				}
				progress.Objectives = append(progress.Objectives, objective)
			}
		}

		// Add POI objectives
		pois, err := s.campaignRepo.GetPlayerPOIsByMission(playerID, missionID)
		if err == nil {
			for _, poi := range pois {
				if poi.ChoiceID == progress.CurrentActiveChoice && poi.IsActive {
					objective := model.MissionObjective{
						Type:        "poi",
						Description: fmt.Sprintf("Visit %s", poi.Name),
						Target:      poi.ID,
						IsCompleted: poi.IsCompleted,
						CompletedAt: poi.CompletedAt,
					}
					progress.Objectives = append(progress.Objectives, objective)
				}
			}
		}

		// Add operation objectives
		operations, err := s.campaignRepo.GetPlayerMissionOperationsByMission(playerID, missionID)
		if err == nil {
			for _, op := range operations {
				if op.ChoiceID == progress.CurrentActiveChoice && op.IsActive {
					objective := model.MissionObjective{
						Type:        "operation",
						Description: fmt.Sprintf("Complete operation: %s", op.Name),
						Target:      op.ID,
						IsCompleted: op.IsCompleted,
						CompletedAt: op.CompletedAt,
					}
					progress.Objectives = append(progress.Objectives, objective)
				}
			}
		}
	} else {
		// No active choice yet, check if any choices can be activated automatically
		// This is for initial mission objectives
		for _, choice := range mission.Choices {
			templates, err := s.campaignRepo.GetConditionTemplatesByChoice(choice.ID)
			if err != nil || len(templates) == 0 {
				continue
			}

			// Add initial condition for each choice
			firstCondition := templates[0]
			objective := model.MissionObjective{
				Type:        "initial_condition",
				Description: s.getTemplateConditionDescription(firstCondition),
				Target:      firstCondition.RequiredValue,
				IsCompleted: false,
			}
			progress.Objectives = append(progress.Objectives, objective)
		}
	}

	// Determine if all objectives are completed
	progress.CanComplete = len(progress.Objectives) > 0
	for _, obj := range progress.Objectives {
		if !obj.IsCompleted {
			progress.CanComplete = false
			break
		}
	}

	return progress, nil
}

// Helper function to get condition descriptions
func (s *campaignService) getConditionDescription(condition model.PlayerCompletionCondition) string {
	switch condition.Type {
	case "travel":
		return fmt.Sprintf("Travel to the %s region", s.getRegionName(condition.RequiredValue))
	case "territory":
		parts := strings.Split(condition.RequiredValue, "_")
		if len(parts) > 1 {
			action := parts[0]
			target := strings.Join(parts[1:], "_")
			switch action {
			case "takeover":
				return fmt.Sprintf("Take control of %s", s.getHotspotName(target))
			case "extortion":
				return fmt.Sprintf("Extort money from %s", s.getHotspotName(target))
			default:
				return fmt.Sprintf("Perform %s on %s", action, s.getHotspotName(target))
			}
		}
		return condition.RequiredValue
	case "operation":
		return fmt.Sprintf("Complete an operation of type: %s", condition.RequiredValue)
	default:
		return condition.RequiredValue
	}
}

// Helper functions to get names
func (s *campaignService) getRegionName(regionID string) string {
	region, err := s.territoryService.GetRegionByID(regionID)
	if err != nil || region == nil {
		return regionID
	}
	return region.Name
}

func (s *campaignService) getHotspotName(hotspotID string) string {
	hotspot, err := s.territoryService.GetHotspotByID(hotspotID)
	if err != nil || hotspot == nil {
		return hotspotID
	}
	return hotspot.Name
}

// Helper function to get descriptions for condition templates
func (s *campaignService) getTemplateConditionDescription(condition model.ConditionTemplate) string {
	switch condition.Type {
	case "travel":
		return fmt.Sprintf("Travel to the %s region", s.getRegionName(condition.RequiredValue))

	case "territory":
		parts := strings.Split(condition.RequiredValue, "_")
		if len(parts) > 1 {
			action := parts[0]
			target := strings.Join(parts[1:], "_")

			switch action {
			case "takeover":
				return fmt.Sprintf("Take control of %s", s.getHotspotName(target))
			case "extortion":
				return fmt.Sprintf("Extort money from %s", s.getHotspotName(target))
			case "defend":
				return fmt.Sprintf("Defend your control of %s", s.getHotspotName(target))
			case "collection":
				return fmt.Sprintf("Collect income from %s", s.getHotspotName(target))
			default:
				return fmt.Sprintf("Perform %s action on %s", action, s.getHotspotName(target))
			}
		}
		return fmt.Sprintf("Perform territory action: %s", condition.RequiredValue)

	case "operation":
		switch condition.RequiredValue {
		case "carjacking":
			return "Complete a carjacking operation"
		case "goods_smuggling":
			return "Complete a goods smuggling operation"
		case "drug_trafficking":
			return "Complete a drug trafficking operation"
		case "official_bribing":
			return "Complete an official bribing operation"
		case "intelligence_gathering":
			return "Complete an intelligence gathering operation"
		case "crew_recruitment":
			return "Complete a crew recruitment operation"
		default:
			return fmt.Sprintf("Complete an operation of type: %s", condition.RequiredValue)
		}

	case "market":
		parts := strings.Split(condition.RequiredValue, "_")
		if len(parts) > 1 {
			action := parts[0]
			resource := strings.Join(parts[1:], "_")

			switch action {
			case "buy":
				return fmt.Sprintf("Buy %s from the market", s.formatResourceName(resource))
			case "sell":
				return fmt.Sprintf("Sell %s on the market", s.formatResourceName(resource))
			default:
				return fmt.Sprintf("Perform %s market action for %s", action, s.formatResourceName(resource))
			}
		}
		return fmt.Sprintf("Perform market action: %s", condition.RequiredValue)

	case "resource":
		parts := strings.Split(condition.RequiredValue, "_")
		if len(parts) == 2 {
			action := parts[0]
			resource := parts[1]

			amount := 0
			if condition.AdditionalValue != "" {
				amount, _ = strconv.Atoi(condition.AdditionalValue)
			}

			if amount > 0 {
				switch action {
				case "acquire":
					return fmt.Sprintf("Acquire %d %s", amount, s.formatResourceName(resource))
				case "spend":
					return fmt.Sprintf("Spend %d %s", amount, s.formatResourceName(resource))
				default:
					return fmt.Sprintf("%s %d %s", action, amount, s.formatResourceName(resource))
				}
			} else {
				switch action {
				case "acquire":
					return fmt.Sprintf("Acquire some %s", s.formatResourceName(resource))
				case "spend":
					return fmt.Sprintf("Spend some %s", s.formatResourceName(resource))
				default:
					return fmt.Sprintf("%s %s", action, s.formatResourceName(resource))
				}
			}
		}
		return fmt.Sprintf("Perform resource action: %s", condition.RequiredValue)

	case "mission":
		return fmt.Sprintf("Complete mission: %s", s.getMissionName(condition.RequiredValue))

	case "poi":
		return fmt.Sprintf("Visit the %s", s.getPOIName(condition.RequiredValue))

	default:
		return condition.RequiredValue
	}
}

// Helper function to format resource names nicely
func (s *campaignService) formatResourceName(resource string) string {
	switch resource {
	case "crew":
		return "crew members"
	case "money":
		return "money"
	case "weapons":
		return "weapons"
	case "vehicles":
		return "vehicles"
	case "respect":
		return "respect"
	case "influence":
		return "influence"
	case "heat":
		return "heat"
	default:
		return resource
	}
}

// Helper function to get POI name
func (s *campaignService) getPOIName(poiID string) string {
	// Try to get from active POIs first
	pois, err := s.campaignRepo.GetActivePlayerPOIs("")
	if err == nil {
		for _, poi := range pois {
			if poi.TemplateID == poiID || poi.ID == poiID {
				return poi.Name
			}
		}
	}

	// Try to get from templates
	templates, err := s.campaignRepo.GetAllPOITemplates()
	if err == nil {
		for _, template := range templates {
			if template.ID == poiID {
				return template.Name
			}
		}
	}

	return "location"
}

// Helper function to get mission name
func (s *campaignService) getMissionName(missionID string) string {
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil || mission == nil {
		return "mission"
	}
	return mission.Title
}

// StartCampaign initiates a campaign for a player
func (s *campaignService) StartCampaign(playerID string, campaignID string) (*model.PlayerCampaignProgress, error) {
	// Check if player exists
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	// Get campaign
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return nil, errors.New("campaign not found")
	}

	// Check if player meets requirements for campaign
	_ = player
	/*
		if campaign.RequiredLevel > 0 {
			// Check player title/level
			titleRank := getTitleRank(player.Title)
			reqRank := campaign.RequiredLevel

			if titleRank < reqRank {
				return nil, fmt.Errorf("player requires at least level %d to start this campaign", reqRank)
			}
		}
	*/

	// Check if player already has progress for this campaign
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, campaignID)
	if err != nil {
		return nil, err
	}

	if progress != nil {
		// Already started, return existing progress
		return progress, nil
	}

	// Create new progress
	now := time.Now()
	progress = &model.PlayerCampaignProgress{
		PlayerID:         playerID,
		CampaignID:       campaignID,
		CurrentChapterID: campaign.InitialChapterID,
		IsCompleted:      false,
		LastUpdated:      now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Set the initial mission for the initial chapter
	initialChapter, err := s.campaignRepo.GetChapterByID(campaign.InitialChapterID)
	if err != nil {
		return nil, errors.New("error retrieving initial chapter")
	}

	if len(initialChapter.Missions) > 0 {
		// Sort missions by order and get the first one
		initialMission := initialChapter.Missions[0]
		progress.CurrentMissionID = initialMission.ID

		// Create mission progress for the initial mission
		missionProgress := &model.PlayerMissionProgress{
			PlayerID:  playerID,
			MissionID: initialMission.ID,
			Status:    "not_started",
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Save mission progress
		if err := s.campaignRepo.SavePlayerMissionProgress(missionProgress); err != nil {
			return nil, errors.New("error saving mission progress")
		}

		// Unlock the first mission
		initialMission.IsLocked = false
		if err := s.campaignRepo.GetDB().Model(&model.Mission{}).Where("id = ?", initialMission.ID).Update("is_locked", false).Error; err != nil {
			s.logger.Error().Err(err).Msg("Failed to unlock initial mission")
		}
	}

	// Save campaign progress
	if err := s.campaignRepo.SavePlayerCampaignProgress(progress); err != nil {
		return nil, errors.New("error saving campaign progress")
	}

	// Add notification to player
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   fmt.Sprintf("You've started the '%s' campaign. Your journey begins now.", campaign.Title),
		Type:      util.NotificationTypeSystem,
		Timestamp: now,
		Read:      false,
	}
	if err := s.playerRepo.AddNotification(notification); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add campaign start notification")
	}

	// Get the campaign for notification and SSE
	campaign, err = s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get campaign for notification")
	}

	// Send SSE update for campaign start
	if campaign != nil {
		data := map[string]interface{}{
			"campaign": map[string]interface{}{
				"id":        campaignID,
				"title":     campaign.Title,
				"progress":  progress,
				"startedAt": progress.CreatedAt.Format(time.RFC3339),
			},
		}
		s.SendCampaignUpdate(playerID, "campaign_started", data)
	}

	return progress, nil
}

// StartMission begins a mission for a player
func (s *campaignService) StartMission(playerID string, missionID string) (*model.PlayerMissionProgress, error) {
	// Check if player exists
	_, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	// Get mission
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	// Check if player already has progress for this mission
	progress, err := s.campaignRepo.GetPlayerMissionProgress(playerID, missionID)
	if err != nil {
		return nil, err
	}

	// For new missions, try to automatically activate the first appropriate choice
	if progress == nil || progress.Status == "not_started" {
		// Get the mission
		mission, err := s.campaignRepo.GetMissionByID(missionID)
		if err != nil {
			return nil, errors.New("mission not found")
		}

		// Get player location to check for location-based conditions
		player, err := s.playerRepo.GetPlayerByID(playerID)
		if err == nil && player.CurrentRegionID != nil {
			currentRegionID := *player.CurrentRegionID

			// Check each choice to see if it has a travel condition matching the player's location
			for _, choice := range mission.Choices {
				templates, err := s.campaignRepo.GetConditionTemplatesByChoice(choice.ID)
				if err != nil || len(templates) == 0 {
					continue
				}

				// Check if the first condition is a travel condition matching the player's location
				if len(templates) > 0 && templates[0].Type == "travel" &&
					templates[0].RequiredValue == currentRegionID {
					// Automatically activate this choice
					if err := s.ActivateChoice(playerID, missionID, choice.ID); err != nil {
						s.logger.Error().Err(err).Msg("Failed to auto-activate choice based on location")
					} else {
						s.logger.Info().Msg("Auto-activated choice based on player location")
					}
					break
				}
			}
		}
	}

	if progress != nil && (progress.Status == "in_progress" || progress.Status == "completed") {
		// Already started or completed, return existing progress
		return progress, nil
	}

	// Check requirements before starting mission
	meetsRequirements, failedRequirements, err := s.CheckMissionRequirements(playerID, missionID)
	if err != nil {
		return nil, err
	}

	if !meetsRequirements {
		reqStr := "Missing requirements: " + fmt.Sprintf("%v", failedRequirements)
		return nil, errors.New(reqStr)
	}

	// Create or update mission progress
	now := time.Now()

	if progress == nil {
		// Create new progress
		progress = &model.PlayerMissionProgress{
			PlayerID:  playerID,
			MissionID: missionID,
			Status:    "in_progress",
			StartedAt: &now,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		// Update existing progress
		progress.Status = "in_progress"
		progress.StartedAt = &now
		progress.UpdatedAt = now
	}

	// Save mission progress
	if err := s.campaignRepo.SavePlayerMissionProgress(progress); err != nil {
		return nil, errors.New("error saving mission progress")
	}

	// Update campaign progress with current mission
	campaign, err := s.getCampaignForMission(missionID)
	if err == nil && campaign != nil {
		campaignProgress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, campaign.ID)
		if err == nil && campaignProgress != nil {
			campaignProgress.CurrentMissionID = missionID
			campaignProgress.LastUpdated = now
			campaignProgress.UpdatedAt = now

			if err := s.campaignRepo.SavePlayerCampaignProgress(campaignProgress); err != nil {
				s.logger.Error().Err(err).Msg("Failed to update campaign progress with current mission")
			}
		}
	}

	// Add notification to player
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   fmt.Sprintf("Mission '%s' started. Complete the objectives to advance.", mission.Title),
		Type:      util.NotificationTypeCampaign,
		Timestamp: now,
		Read:      false,
	}
	if err := s.playerRepo.AddNotification(notification); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add mission start notification")
	}

	// Get the mission for notification and SSE
	mission, err = s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get mission for notification")
	}

	// Send SSE update for mission start
	if mission != nil {
		s.SendMissionUpdate(playerID, mission, progress)
	}

	return progress, nil
}

// CompleteMission completes a mission for a player
func (s *campaignService) CompleteMission(playerID string, missionID string, choiceID string) (*model.MissionCompleteResult, error) {
	// Check if player exists
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	// Get mission
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	// Get mission progress
	progress, err := s.campaignRepo.GetPlayerMissionProgress(playerID, missionID)
	if err != nil {
		return nil, err
	}

	// Check if mission is in progress
	if progress == nil || progress.Status != "in_progress" {
		return nil, errors.New("mission not in progress")
	}

	// Check if mission has been completed already
	if progress.Status == "completed" {
		return nil, errors.New("mission already completed")
	}

	// Initialize rewards with mission rewards
	rewards := mission.Rewards
	var nextMissionID string
	choiceText := ""

	// If a choice was made, get additional rewards and next mission from that choice
	if choiceID != "" {
		var selectedChoice *model.MissionChoice
		for _, choice := range mission.Choices {
			if choice.ID == choiceID {
				selectedChoice = &choice
				break
			}
		}

		if selectedChoice == nil {
			return nil, errors.New("invalid choice")
		}

		// Check if player meets choice requirements
		if !s.playerMeetsRequirements(player, selectedChoice.Requirements) {
			return nil, errors.New("player does not meet choice requirements")
		}

		// Add choice rewards to mission rewards
		rewards.Money += selectedChoice.Rewards.Money
		rewards.Crew += selectedChoice.Rewards.Crew
		rewards.Weapons += selectedChoice.Rewards.Weapons
		rewards.Vehicles += selectedChoice.Rewards.Vehicles
		rewards.Respect += selectedChoice.Rewards.Respect
		rewards.Influence += selectedChoice.Rewards.Influence
		rewards.HeatReduction += selectedChoice.Rewards.HeatReduction

		// Handle any special reward unlocks
		if selectedChoice.Rewards.UnlockChapterID != "" {
			s.unlockChapter(playerID, selectedChoice.Rewards.UnlockChapterID)
		}
		if selectedChoice.Rewards.UnlockMissionID != "" {
			s.unlockMission(playerID, selectedChoice.Rewards.UnlockMissionID)
		}
		if selectedChoice.Rewards.UnlockHotspotID != "" {
			s.unlockHotspot(playerID, selectedChoice.Rewards.UnlockHotspotID)
		}

		// Set next mission from choice
		nextMissionID = selectedChoice.NextMissionID
		choiceText = selectedChoice.Text

		// Save the choice that was made
		progress.ChoiceID = choiceID
	} else if mission.Choices == nil || len(mission.Choices) == 0 {
		// If no choices and mission has a default next mission
		// Look for a direct connection in the chapter
		missions, err := s.campaignRepo.GetMissionsByChapterID(mission.ChapterID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get missions for chapter")
		} else {
			// Find the next mission in sequence
			for i, m := range missions {
				if m.ID == mission.ID && i+1 < len(missions) {
					nextMissionID = missions[i+1].ID
					break
				}
			}
		}
	}

	// Apply rewards to player
	resourceUpdates := make(map[string]int)

	if rewards.Money > 0 {
		resourceUpdates["money"] = rewards.Money
	}
	if rewards.Crew > 0 {
		resourceUpdates["crew"] = rewards.Crew
	}
	if rewards.Weapons > 0 {
		resourceUpdates["weapons"] = rewards.Weapons
	}
	if rewards.Vehicles > 0 {
		resourceUpdates["vehicles"] = rewards.Vehicles
	}
	if rewards.Respect > 0 {
		resourceUpdates["respect"] = rewards.Respect
	}
	if rewards.Influence > 0 {
		resourceUpdates["influence"] = rewards.Influence
	}
	if rewards.HeatReduction > 0 {
		resourceUpdates["heat"] = -rewards.HeatReduction
	}

	if len(resourceUpdates) > 0 {
		if err := s.playerService.UpdatePlayerResources(playerID, resourceUpdates); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update player resources")
		}
	}

	// Handle any mission-specific unlocks
	if rewards.UnlockChapterID != "" {
		s.unlockChapter(playerID, rewards.UnlockChapterID)
	}
	if rewards.UnlockMissionID != "" {
		s.unlockMission(playerID, rewards.UnlockMissionID)
	}
	if rewards.UnlockHotspotID != "" {
		s.unlockHotspot(playerID, rewards.UnlockHotspotID)
	}

	// Update mission progress
	now := time.Now()
	progress.Status = "completed"
	progress.CompletedAt = &now
	progress.UpdatedAt = now

	if err := s.campaignRepo.SavePlayerMissionProgress(progress); err != nil {
		return nil, errors.New("error saving mission progress")
	}

	// Get the mission for SSE updates
	mission, err = s.campaignRepo.GetMissionByID(missionID)
	if err == nil {
		// Send mission completed SSE update
		s.SendMissionUpdate(playerID, mission, progress)
	}

	// If there's a next mission, unlock it and update campaign progress
	var nextMission *model.Mission
	if nextMissionID != "" {
		// Fetch next mission
		nextMission, err = s.campaignRepo.GetMissionByID(nextMissionID)
		if err == nil && nextMission != nil {
			// Unlock the next mission
			nextMission.IsLocked = false
			if err := s.campaignRepo.GetDB().Model(&model.Mission{}).Where("id = ?", nextMissionID).Update("is_locked", false).Error; err != nil {
				s.logger.Error().Err(err).Msg("Failed to unlock next mission")
			}

			// Create mission progress for the next mission
			nextProgress := &model.PlayerMissionProgress{
				PlayerID:  playerID,
				MissionID: nextMissionID,
				Status:    "not_started",
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.campaignRepo.SavePlayerMissionProgress(nextProgress); err != nil {
				s.logger.Error().Err(err).Msg("Failed to create progress for next mission")
			}

			// Update campaign progress
			campaign, err := s.getCampaignForMission(missionID)
			if err == nil && campaign != nil {
				campaignProgress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, campaign.ID)
				if err == nil && campaignProgress != nil {
					// Check if we need to change chapter
					if nextMission.ChapterID != mission.ChapterID {
						campaignProgress.CurrentChapterID = nextMission.ChapterID
					}

					campaignProgress.CurrentMissionID = nextMissionID
					campaignProgress.LastUpdated = now
					campaignProgress.UpdatedAt = now

					if err := s.campaignRepo.SavePlayerCampaignProgress(campaignProgress); err != nil {
						s.logger.Error().Err(err).Msg("Failed to update campaign progress")
					}
				}
			}

			// If there's a next mission, also send SSE for that
			nextProgress, err = s.campaignRepo.GetPlayerMissionProgress(playerID, nextMission.ID)
			if err == nil && nextProgress != nil {
				s.SendMissionUpdate(playerID, nextMission, nextProgress)
			}

		}
	} else {
		// No next mission - check if we've completed the chapter
		missions, err := s.campaignRepo.GetMissionsByChapterID(mission.ChapterID)
		if err == nil {
			allCompleted := true
			for _, m := range missions {
				// Skip the current mission since we just completed it
				if m.ID == missionID {
					continue
				}

				missionProgress, err := s.campaignRepo.GetPlayerMissionProgress(playerID, m.ID)
				if err != nil || missionProgress == nil || missionProgress.Status != "completed" {
					allCompleted = false
					break
				}
			}

			if allCompleted {
				// Chapter completed
				campaign, err := s.getCampaignForMission(missionID)
				if err == nil && campaign != nil {
					campaignProgress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, campaign.ID)
					if err == nil && campaignProgress != nil {
						// Check if there are more chapters
						// Find next chapter
						var nextChapter *model.Chapter
						for i, ch := range campaign.Chapters {
							if ch.ID == mission.ChapterID && i+1 < len(campaign.Chapters) {
								nextChapter = &campaign.Chapters[i+1]
								break
							}
						}

						if nextChapter != nil {
							// Unlock next chapter
							s.unlockChapter(playerID, nextChapter.ID)

							// Get first mission in next chapter
							nextChapterMissions, err := s.campaignRepo.GetMissionsByChapterID(nextChapter.ID)
							if err == nil && len(nextChapterMissions) > 0 {
								// Update campaign progress to next chapter/mission
								campaignProgress.CurrentChapterID = nextChapter.ID
								campaignProgress.CurrentMissionID = nextChapterMissions[0].ID
								campaignProgress.LastUpdated = now
								campaignProgress.UpdatedAt = now

								if err := s.campaignRepo.SavePlayerCampaignProgress(campaignProgress); err != nil {
									s.logger.Error().Err(err).Msg("Failed to update campaign progress")
								}

								// Unlock first mission in next chapter
								s.unlockMission(playerID, nextChapterMissions[0].ID)

								// Create mission progress for first mission in next chapter
								nextProgress := &model.PlayerMissionProgress{
									PlayerID:  playerID,
									MissionID: nextChapterMissions[0].ID,
									Status:    "not_started",
									CreatedAt: now,
									UpdatedAt: now,
								}
								if err := s.campaignRepo.SavePlayerMissionProgress(nextProgress); err != nil {
									s.logger.Error().Err(err).Msg("Failed to create progress for next mission")
								}

								nextMission = &nextChapterMissions[0]
							}
						} else {
							// No more chapters - campaign completed
							campaignProgress.IsCompleted = true
							campaignProgress.CompletedAt = &now
							campaignProgress.LastUpdated = now
							campaignProgress.UpdatedAt = now

							if err := s.campaignRepo.SavePlayerCampaignProgress(campaignProgress); err != nil {
								s.logger.Error().Err(err).Msg("Failed to update campaign progress")
							}

							// Add notification for completing the campaign
							notification := &model.Notification{
								PlayerID:  playerID,
								Message:   fmt.Sprintf("Congratulations! You've completed the '%s' campaign.", campaign.Title),
								Type:      util.NotificationTypeCampaign,
								Timestamp: now,
								Read:      false,
							}
							if err := s.playerRepo.AddNotification(notification); err != nil {
								s.logger.Error().Err(err).Msg("Failed to add campaign completion notification")
							}

							// This is where we need to add the SSE campaign completion notification
							campaignData := map[string]interface{}{
								"campaignID":  campaign.ID,
								"title":       campaign.Title,
								"isCompleted": true,
								"completedAt": now.Format(time.RFC3339),
							}
							s.SendCampaignUpdate(playerID, "campaign_completed", campaignData)
						}
					}
				}
			}
		}
	}

	// Add notification for mission completion
	message := fmt.Sprintf("Mission '%s' completed successfully. ", mission.Title)
	if choiceText != "" {
		message += fmt.Sprintf("You chose: %s. ", choiceText)
	}

	if rewards.Money > 0 || rewards.Crew > 0 || rewards.Weapons > 0 || rewards.Vehicles > 0 ||
		rewards.Respect > 0 || rewards.Influence > 0 || rewards.HeatReduction > 0 {
		message += "Rewards received."
	}

	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   message,
		Type:      util.NotificationTypeCampaign,
		Timestamp: now,
		Read:      false,
	}
	if err := s.playerRepo.AddNotification(notification); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add mission completion notification")
	}

	// Return the result
	return &model.MissionCompleteResult{
		Success:     true,
		Rewards:     rewards,
		NextMission: nextMission,
		Progress:    progress,
		Message:     message,
	}, nil
}

// CheckMissionRequirements checks if a player meets the requirements for a mission
func (s *campaignService) CheckMissionRequirements(playerID string, missionID string) (bool, []string, error) {
	// Get player
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return false, nil, errors.New("player not found")
	}

	// Get mission
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return false, nil, errors.New("mission not found")
	}

	// Check mission requirements
	return s.checkPlayerRequirements(player, mission.Requirements)
}

// GetActivePlayerPOIs retrieves all active POIs for a player
func (s *campaignService) GetActivePlayerPOIs(playerID string) ([]model.PlayerPOI, error) {
	return s.campaignRepo.GetActivePlayerPOIs(playerID)
}

// ActivatePlayerPOI activates a POI for a player
func (s *campaignService) ActivatePlayerPOI(playerID string, templateID string) (*model.PlayerPOI, error) {
	return s.campaignRepo.ActivatePlayerPOI(playerID, templateID)
}

// CompletePlayerPOI marks a POI as completed
func (s *campaignService) CompletePlayerPOI(playerID string, playerPOIID string) error {
	// Get the POI
	playerPOI, err := s.campaignRepo.GetPlayerPOI(playerPOIID)
	if err != nil {
		return err
	}

	// Check if the POI belongs to the player
	if playerPOI.PlayerID != playerID {
		return errors.New("POI not owned by player")
	}

	// Check if POI is active
	if !playerPOI.IsActive {
		return errors.New("POI is not active")
	}

	// Check if POI is already completed
	if playerPOI.IsCompleted {
		return errors.New("POI is already completed")
	}

	// Mark as completed
	if err := s.campaignRepo.CompletePlayerPOI(playerID, playerPOIID); err != nil {
		return err
	}

	// Get the updated POI after completing
	updatedPOI, err := s.campaignRepo.GetPlayerPOI(playerPOIID)
	if err != nil {
		return err
	}

	// Send SSE update for POI completion
	s.SendPOICompletedUpdate(playerID, updatedPOI)

	// Check if this POI completion contributes to a choice completion
	if playerPOI.ChoiceID != "" {
		completed, err := s.CheckChoiceCompletion(playerID, playerPOI.MissionID, playerPOI.ChoiceID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to check choice completion")
		} else if completed {
			// If choice completed, mission has been auto-completed via the choice
			// This will already have sent appropriate SSE notifications
			s.logger.Info().Str("choiceID", playerPOI.ChoiceID).Msg("Choice completed through POI completion")
		}
	}

	return nil
}

// GetActivePlayerMissionOperations retrieves all active mission operations for a player
func (s *campaignService) GetActivePlayerMissionOperations(playerID string) ([]model.PlayerMissionOperation, error) {
	return s.campaignRepo.GetActivePlayerMissionOperations(playerID)
}

// ActivatePlayerMissionOperation activates a mission operation for a player
func (s *campaignService) ActivatePlayerMissionOperation(playerID string, templateID string) (*model.PlayerMissionOperation, error) {
	return s.campaignRepo.ActivatePlayerMissionOperation(playerID, templateID)
}

// CompletePlayerMissionOperation marks a mission operation as completed
func (s *campaignService) CompletePlayerMissionOperation(playerID string, playerOpID string) error {
	// Get the mission operation
	playerOp, err := s.campaignRepo.GetPlayerMissionOperation(playerOpID)
	if err != nil {
		return err
	}

	// Check if the operation belongs to the player
	if playerOp.PlayerID != playerID {
		return errors.New("operation not owned by player")
	}

	// Check if operation is active
	if !playerOp.IsActive {
		return errors.New("operation is not active")
	}

	// Check if operation is already completed
	if playerOp.IsCompleted {
		return errors.New("operation is already completed")
	}

	// Mark as completed
	if err := s.campaignRepo.CompletePlayerMissionOperation(playerID, playerOpID); err != nil {
		return err
	}

	// Get the updated operation after completing
	updatedOp, err := s.campaignRepo.GetPlayerMissionOperation(playerOpID)
	if err != nil {
		return err
	}

	// Send SSE update for operation completion
	s.SendOperationCompletedUpdate(playerID, updatedOp)

	// Check if this operation completion contributes to a choice completion
	if playerOp.ChoiceID != "" {
		completed, err := s.CheckChoiceCompletion(playerID, playerOp.MissionID, playerOp.ChoiceID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to check choice completion")
		} else if completed {
			// If choice completed, mission has been auto-completed via the choice
			// This will already have sent appropriate SSE notifications
			s.logger.Info().Str("choiceID", playerOp.ChoiceID).Msg("Choice completed through operation completion")
		}
	}

	return nil
}

// TrackPlayerAction processes player actions for mission progress
func (s *campaignService) TrackPlayerAction(playerID string, actionType string, actionValue string) error {
	// Get active mission progress for the player
	var activeProgress []model.PlayerMissionProgress
	if err := s.campaignRepo.GetDB().
		Where("player_id = ? AND status = ?", playerID, "in_progress").
		Find(&activeProgress).Error; err != nil {
		return err
	}

	// Check each active mission progress
	for _, progress := range activeProgress {
		// If there's a current active choice, check if this action contributes to it
		if progress.CurrentActiveChoice != "" {
			// Get player conditions for this choice
			conditions, err := s.campaignRepo.GetPlayerCompletionConditions(playerID, progress.CurrentActiveChoice)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get player choice conditions")
				continue
			}

			// Check if this action satisfies any of the conditions
			for _, condition := range conditions {
				if !condition.IsCompleted &&
					condition.Type == actionType &&
					condition.RequiredValue == actionValue {

					// Check if we need sequential order
					mission, err := s.campaignRepo.GetMissionByID(progress.MissionID)
					if err != nil {
						s.logger.Error().Err(err).Msg("Failed to get mission")
						continue
					}

					// Find the choice
					var choice *model.MissionChoice
					for _, c := range mission.Choices {
						if c.ID == progress.CurrentActiveChoice {
							choice = &c
							break
						}
					}

					if choice == nil {
						s.logger.Error().Msg("Failed to find choice")
						continue
					}

					// If sequential order is required, check if this is the next condition
					if choice.SequentialOrder {
						// Get all conditions ordered by index
						allConditions := conditions
						// Check if any earlier condition is not completed
						canComplete := true
						for _, c := range allConditions {
							if c.OrderIndex < condition.OrderIndex && !c.IsCompleted {
								// Earlier condition not completed, can't complete this one yet
								canComplete = false
								break
							}
						}

						if !canComplete {
							continue
						}
					}

					// Mark the condition as completed
					if err := s.campaignRepo.CompletePlayerCompletionCondition(playerID, condition.ID); err != nil {
						s.logger.Error().Err(err).Msg("Failed to complete condition")
						continue
					}

					// Get updated condition
					updatedCondition, err := s.campaignRepo.GetPlayerCompletionCondition(condition.ID)
					if err != nil {
						s.logger.Error().Err(err).Msg("Failed to get updated condition")
					} else {
						// Send SSE update for condition completion
						s.SendConditionCompletedUpdate(playerID, updatedCondition.ID)
					}

					// Check if all conditions are now completed for this choice
					allCompleted, _ := s.CheckChoiceCompletion(playerID, progress.MissionID, choice.ID)
					if allCompleted {
						// This might have completed the mission via the choice
						s.logger.Info().Str("missionID", progress.MissionID).Str("choiceID", choice.ID).Msg("Choice conditions completed")
					}
				}
			}
		} else {
			// No active choice, see if this action could activate any choice
			mission, err := s.campaignRepo.GetMissionByID(progress.MissionID)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get mission")
				continue
			}

			// Check each choice in the mission
			for _, choice := range mission.Choices {
				// Get conditions templates for this choice
				conditionTemplates, err := s.campaignRepo.GetConditionTemplatesByChoice(choice.ID)
				if err != nil {
					s.logger.Error().Err(err).Msg("Failed to get choice condition templates")
					continue
				}

				// If the first condition matches this action, activate the choice
				for _, condition := range conditionTemplates {
					if condition.OrderIndex == 0 &&
						condition.Type == actionType &&
						condition.RequiredValue == actionValue {

						// Activate this choice for the player
						if err := s.ActivateChoice(playerID, mission.ID, choice.ID); err != nil {
							s.logger.Error().Err(err).Msg("Failed to activate choice")
						}
						break
					}
				}
			}
		}
	}

	return nil
}

// CheckChoiceCompletion checks if all conditions for a choice are completed
func (s *campaignService) CheckChoiceCompletion(playerID string, missionID string, choiceID string) (bool, error) {
	// Get the player's conditions for this choice
	conditions, err := s.campaignRepo.GetPlayerCompletionConditions(playerID, choiceID)
	if err != nil {
		return false, err
	}

	// Check if all conditions are completed
	allCompleted := true
	for _, condition := range conditions {
		if !condition.IsCompleted {
			allCompleted = false
			break
		}
	}

	// If all conditions are completed, complete the choice
	if allCompleted && len(conditions) > 0 {
		// Get mission progress
		progress, err := s.campaignRepo.GetPlayerMissionProgress(playerID, missionID)
		if err != nil {
			return false, err
		}

		if progress == nil {
			return false, errors.New("mission progress not found")
		}

		// Update mission progress to use this choice
		now := time.Now()

		// Only complete the mission if it hasn't been completed already
		if progress.Status != "completed" {
			progress.ChoiceID = choiceID
			progress.Status = "completed"
			progress.CompletedAt = &now
			progress.UpdatedAt = now

			if err := s.campaignRepo.SavePlayerMissionProgress(progress); err != nil {
				return false, err
			}

			// Get the choice and mission to find the next mission
			mission, err := s.campaignRepo.GetMissionByID(missionID)
			if err != nil {
				return false, err
			}

			// Send mission completion SSE update
			s.SendMissionUpdate(playerID, mission, progress)

			var choice *model.MissionChoice
			for _, c := range mission.Choices {
				if c.ID == choiceID {
					choice = &c
					break
				}
			}

			if choice == nil {
				return false, errors.New("choice not found")
			}

			// If there's a next mission, start it
			if choice.NextMissionID != "" {
				nextMissionProgress, err := s.StartMission(playerID, choice.NextMissionID)
				if err != nil {
					s.logger.Error().Err(err).Msg("Failed to start next mission")
				} else {
					// Get the next mission
					nextMission, err := s.campaignRepo.GetMissionByID(choice.NextMissionID)
					if err != nil {
						s.logger.Error().Err(err).Msg("Failed to get next mission")
					} else {
						// Send SSE update for next mission
						s.SendMissionUpdate(playerID, nextMission, nextMissionProgress)
					}
				}
			}

			// Add notification
			notification := &model.Notification{
				PlayerID:  playerID,
				Message:   fmt.Sprintf("You've successfully completed the objectives for mission '%s'.", mission.Title),
				Type:      util.NotificationTypeCampaign,
				Timestamp: now,
				Read:      false,
			}
			if err := s.playerRepo.AddNotification(notification); err != nil {
				s.logger.Error().Err(err).Msg("Failed to add choice completion notification")
			}
		}

		return true, nil
	}

	return allCompleted, nil
}

// ActivateChoice activates a choice for a player
func (s *campaignService) ActivateChoice(playerID string, missionID string, choiceID string) error {
	// Get the player's mission progress
	progress, err := s.campaignRepo.GetPlayerMissionProgress(playerID, missionID)
	if err != nil {
		return err
	}

	if progress == nil {
		return errors.New("mission progress not found")
	}

	// Make sure the mission is in progress
	if progress.Status != "in_progress" {
		return errors.New("mission is not in progress")
	}

	// Make sure there's no active choice already
	if progress.CurrentActiveChoice != "" {
		return errors.New("player already has an active choice for this mission")
	}

	// Set the active choice
	progress.CurrentActiveChoice = choiceID
	if err := s.campaignRepo.SavePlayerMissionProgress(progress); err != nil {
		return err
	}

	// Create player-specific conditions from the choice conditions
	conditionTemplates, err := s.campaignRepo.GetConditionTemplatesByChoice(choiceID)
	if err != nil {
		return err
	}

	for _, template := range conditionTemplates {
		playerCondition := &model.PlayerCompletionCondition{
			TemplateID:      template.ID,
			PlayerID:        playerID,
			ChoiceID:        choiceID,
			Type:            template.Type,
			RequiredValue:   template.RequiredValue,
			AdditionalValue: template.AdditionalValue,
			OrderIndex:      template.OrderIndex,
			IsCompleted:     false,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := s.campaignRepo.CreatePlayerCompletionCondition(playerCondition); err != nil {
			return err
		}
	}

	// Activate any POIs related to this choice
	poiTemplates, err := s.campaignRepo.GetPOITemplatesByChoice(choiceID)
	if err != nil {
		return err
	}

	for _, template := range poiTemplates {
		playerPOI, err := s.campaignRepo.ActivatePlayerPOI(playerID, template.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to activate POI")
			// Continue with other POIs even if one fails
			continue
		}

		// Send SSE update for activated POI
		s.SendPOIActivatedUpdate(playerID, playerPOI)
	}

	// Activate any mission operations related to this choice
	operationTemplates, err := s.campaignRepo.GetOperationTemplatesByChoice(choiceID)
	if err != nil {
		return err
	}

	for _, template := range operationTemplates {
		playerOp, err := s.campaignRepo.ActivatePlayerMissionOperation(playerID, template.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to activate mission operation")
			// Continue with other operations even if one fails
			continue
		}

		// Send SSE update for activated operation
		s.SendOperationActivatedUpdate(playerID, playerOp)
	}

	// Add notification for choice activation
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   fmt.Sprintf("New objectives available for your current mission."),
		Type:      util.NotificationTypeCampaign,
		Timestamp: time.Now(),
		Read:      false,
	}
	if err := s.playerRepo.AddNotification(notification); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add choice activation notification")
	}

	// Send SSE update for choice activation
	s.SendChoiceActivatedUpdate(playerID, missionID, choiceID)

	return nil
}

// Helper function to check if a player meets requirements
func (s *campaignService) checkPlayerRequirements(player *model.Player, requirements model.MissionRequirements) (bool, []string, error) {
	var failedRequirements []string

	// Check resources
	if requirements.Money > 0 && player.Money < requirements.Money {
		failedRequirements = append(failedRequirements, fmt.Sprintf("Need $%d (you have $%d)", requirements.Money, player.Money))
	}
	if requirements.Crew > 0 && player.Crew < requirements.Crew {
		failedRequirements = append(failedRequirements, fmt.Sprintf("Need %d crew (you have %d)", requirements.Crew, player.Crew))
	}
	if requirements.Weapons > 0 && player.Weapons < requirements.Weapons {
		failedRequirements = append(failedRequirements, fmt.Sprintf("Need %d weapons (you have %d)", requirements.Weapons, player.Weapons))
	}
	if requirements.Vehicles > 0 && player.Vehicles < requirements.Vehicles {
		failedRequirements = append(failedRequirements, fmt.Sprintf("Need %d vehicles (you have %d)", requirements.Vehicles, player.Vehicles))
	}

	// Check player attributes
	if requirements.Respect > 0 && player.Respect < requirements.Respect {
		failedRequirements = append(failedRequirements, fmt.Sprintf("Need %d respect (you have %d)", requirements.Respect, player.Respect))
	}
	if requirements.Influence > 0 && player.Influence < requirements.Influence {
		failedRequirements = append(failedRequirements, fmt.Sprintf("Need %d influence (you have %d)", requirements.Influence, player.Influence))
	}
	if requirements.MaxHeat > 0 && player.Heat > requirements.MaxHeat {
		failedRequirements = append(failedRequirements, fmt.Sprintf("Heat too high (max %d, you have %d)", requirements.MaxHeat, player.Heat))
	}

	// Check title requirement
	if requirements.MinTitle != "" {
		if !meetsMinimumTitle(player.Title, requirements.MinTitle) {
			failedRequirements = append(failedRequirements, fmt.Sprintf("Need to be a %s or higher", requirements.MinTitle))
		}
	}

	// Check territory control requirement
	if requirements.ControlledHotspots > 0 {
		controlledCount, err := s.playerRepo.GetControlledHotspotsCount(player.ID)
		if err != nil {
			return false, nil, err
		}

		if controlledCount < requirements.ControlledHotspots {
			failedRequirements = append(failedRequirements, fmt.Sprintf("Need to control %d businesses (you control %d)", requirements.ControlledHotspots, controlledCount))
		}
	}

	// Check region requirement
	if requirements.RegionID != "" && player.CurrentRegionID != nil && *player.CurrentRegionID != requirements.RegionID {
		region, err := s.territoryService.GetRegionByID(requirements.RegionID)
		if err == nil {
			failedRequirements = append(failedRequirements, fmt.Sprintf("Need to be in %s region", region.Name))
		} else {
			failedRequirements = append(failedRequirements, "Need to be in the required region")
		}
	}

	// Return result
	return len(failedRequirements) == 0, failedRequirements, nil
}

// Helper function to check if player meets requirements
func (s *campaignService) playerMeetsRequirements(player *model.Player, requirements model.MissionRequirements) bool {
	meets, _, _ := s.checkPlayerRequirements(player, requirements)
	return meets
}

// Helper function to unlock a chapter
func (s *campaignService) unlockChapter(playerID string, chapterID string) error {
	// Unlock chapter
	err := s.campaignRepo.GetDB().Model(&model.Chapter{}).Where("id = ?", chapterID).Update("is_locked", false).Error
	if err != nil {
		s.logger.Error().Err(err).Str("chapterID", chapterID).Msg("Failed to unlock chapter")
		return err
	}

	// Get chapter for notification
	chapter, err := s.campaignRepo.GetChapterByID(chapterID)
	if err != nil {
		return err
	}

	// Add notification
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   fmt.Sprintf("New chapter unlocked: %s", chapter.Title),
		Type:      util.NotificationTypeCampaign,
		Timestamp: time.Now(),
		Read:      false,
	}
	if err := s.playerRepo.AddNotification(notification); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add chapter unlock notification")
	}

	return nil
}

// Helper function to unlock a mission
func (s *campaignService) unlockMission(playerID string, missionID string) error {
	// Unlock mission
	err := s.campaignRepo.GetDB().Model(&model.Mission{}).Where("id = ?", missionID).Update("is_locked", false).Error
	if err != nil {
		s.logger.Error().Err(err).Str("missionID", missionID).Msg("Failed to unlock mission")
		return err
	}

	// Get mission for notification
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return err
	}

	// Add notification
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   fmt.Sprintf("New mission available: %s", mission.Title),
		Type:      util.NotificationTypeCampaign,
		Timestamp: time.Now(),
		Read:      false,
	}
	if err := s.playerRepo.AddNotification(notification); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add mission unlock notification")
	}

	return nil
}

// Helper function to unlock a hotspot
func (s *campaignService) unlockHotspot(playerID string, hotspotID string) error {
	// This would be implemented based on your territory system
	// For now just a placeholder
	s.logger.Info().Str("playerID", playerID).Str("hotspotID", hotspotID).Msg("Unlocking hotspot")
	return nil
}

// Helper function to get campaign for a mission
func (s *campaignService) getCampaignForMission(missionID string) (*model.Campaign, error) {
	// Get the mission to find its chapter
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return nil, err
	}

	// Get the chapter to find the campaign
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return nil, err
	}

	// Get the campaign
	return s.campaignRepo.GetCampaignByID(chapter.CampaignID)
}

// LoadCampaigns loads campaigns from YAML files
func (s *campaignService) LoadCampaigns(dirPath string) error {
	return s.campaignRepo.LoadCampaignsFromYAML(dirPath)
}

// GetInjectedOperations implements OperationsProvider interface
func (s *campaignService) GetInjectedOperations(playerID string, regionID *string) ([]model.Operation, error) {
	var result []model.Operation

	// Get active operations from player's active missions
	activeOps, err := s.campaignRepo.GetActivePlayerMissionOperations(playerID)
	if err != nil {
		return nil, err
	}

	for _, op := range activeOps {
		if !op.IsActive || op.IsCompleted {
			continue
		}

		// Create an operation that can be used by the operations service
		operation := model.Operation{
			ID:          op.ID,
			Name:        op.Name,
			Description: op.Description,
			Type:        op.OperationType,
			// Set this to true to differentiate campaign ops
			IsSpecial: true,
			IsActive:  true,
			IsLocked:  false,
			// Set appropriate region IDs if the operation is region-specific
			RegionIDs:    []string{},
			Requirements: model.OperationRequirements{
				// Set appropriate requirements
			},
			Resources:   op.Resources,
			Rewards:     op.Rewards,
			Risks:       op.Risks,
			Duration:    op.Duration,
			SuccessRate: op.SuccessRate,
			// Set an appropriate available until date (e.g., 24 hours from now)
			AvailableUntil: time.Now().Add(24 * time.Hour),
			// Add campaign-specific metadata
			Metadata: map[string]interface{}{
				"isCampaignOperation": true,
				"missionID":           op.MissionID,
				"choiceID":            op.ChoiceID,
			},
		}

		// If region filtering is applied, only include operations for that region
		if regionID != nil {
			// Add region check logic if campaign operations are region-specific
			// For now, include all operations
		}

		result = append(result, operation)
	}

	return result, nil
}

// GetInjectedHotspots implements HotspotProvider interface
func (s *campaignService) GetInjectedHotspots(playerID string, regionID *string) ([]model.Hotspot, error) {
	var result []model.Hotspot

	// Get active POIs from player's active missions
	activePOIs, err := s.campaignRepo.GetActivePlayerPOIs(playerID)
	if err != nil {
		return nil, err
	}

	for _, poi := range activePOIs {
		if !poi.IsActive || poi.IsCompleted {
			continue
		}

		// Create a hotspot that can be used by the territory service
		hotspot := model.Hotspot{
			ID:           poi.ID,
			Name:         poi.Name,
			CityID:       poi.LocationID, // Use the location ID as the city ID if it's a city
			Type:         "campaign_poi", // Use a special type for campaign POIs
			BusinessType: "campaign",
			IsLegal:      true,
			Income:       0, // No income for campaign POIs
			// Add campaign-specific metadata
			Metadata: map[string]interface{}{
				"isCampaignPOI": true,
				"missionID":     poi.MissionID,
				"choiceID":      poi.ChoiceID,
				"locationType":  poi.LocationType,
				"description":   poi.Description,
			},
		}

		// If region filtering is applied, only include hotspots for that region
		if regionID != nil && poi.LocationType == "region" && poi.LocationID != *regionID {
			continue
		}

		result = append(result, hotspot)
	}

	return result, nil
}

// GetActivePOIsForMission gets all active POIs for a mission
func (s *campaignService) GetActivePOIsForMission(playerID string, missionID string) ([]model.PlayerPOI, error) {
	return s.campaignRepo.GetPlayerPOIsByMission(playerID, missionID)
}

// GetActiveOperationsForMission gets all active operations for a mission
func (s *campaignService) GetActiveOperationsForMission(playerID string, missionID string) ([]model.PlayerMissionOperation, error) {
	return s.campaignRepo.GetPlayerMissionOperationsByMission(playerID, missionID)
}

// Helper function to get title rank
func getTitleRank(title string) int {
	titleRanks := map[string]int{
		util.PlayerTitleAssociate:   1,
		util.PlayerTitleSoldier:     2,
		util.PlayerTitleCapo:        3,
		util.PlayerTitleUnderboss:   4,
		util.PlayerTitleConsigliere: 5,
		util.PlayerTitleBoss:        6,
		util.PlayerTitleGodfather:   7,
	}

	rank, ok := titleRanks[title]
	if !ok {
		return 0
	}
	return rank
}

// Helper function to check minimum title requirement
// func meetsMinimumTitle(playerTitle string, requiredTitle string) bool {
// 	playerRank := getTitleRank(playerTitle)
// 	requiredRank := getTitleRank(requiredTitle)

// 	return playerRank >= requiredRank
// }
