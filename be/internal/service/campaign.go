// internal/service/campaign.go

package service

import (
	"errors"
	"fmt"
	"time"

	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// CampaignService handles campaign-related business logic
type CampaignService interface {
	GetAllCampaigns() ([]model.Campaign, error)
	GetCampaignByID(campaignID string) (*model.Campaign, error)
	GetChapterByID(chapterID string) (*model.Chapter, error)
	GetMissionByID(missionID string) (*model.Mission, error)
	GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error)
	GetPlayerCampaignProgresses(playerID string) ([]model.PlayerCampaignProgress, error)
	GetPlayerMissionProgress(playerID string, missionID string) (*model.PlayerMissionProgress, error)
	StartCampaign(playerID string, campaignID string) (*model.PlayerCampaignProgress, error)
	StartMission(playerID string, missionID string) (*model.PlayerMissionProgress, error)
	CompleteMission(playerID string, missionID string, choiceID string) (*MissionCompleteResult, error)
	CheckMissionRequirements(playerID string, missionID string) (bool, []string, error)
	LoadCampaigns(dirPath string) error

	// POI methods
	GetActivePOIs(playerID string) ([]model.POI, error)
	ActivatePOIForPlayer(poiID string, playerID string) error
	CompletePOI(poiID string, playerID string) error

	// Mission Operation methods
	GetActiveMissionOperations(playerID string) ([]model.MissionOperation, error)
	ActivateMissionOperationForPlayer(operationID string, playerID string) error
	CompleteMissionOperation(operationID string, playerID string) error

	// Player action tracking
	TrackPlayerAction(playerID string, actionType string, actionValue string) error
	CheckChoiceCompletion(playerID string, missionID string, choiceID string) (bool, error)
	ActivateChoice(playerID string, missionID string, choiceID string) error
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
	return s.campaignRepo.GetPlayerMissionProgress(playerID, missionID)
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
	if campaign.RequiredLevel > 0 {
		// Check player title/level
		titleRank := getTitleRank(player.Title)
		reqRank := campaign.RequiredLevel

		if titleRank < reqRank {
			return nil, fmt.Errorf("player requires at least level %d to start this campaign", reqRank)
		}
	}

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
	campaignProgress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, mission.ChapterID)
	if err == nil && campaignProgress != nil {
		campaignProgress.CurrentMissionID = missionID
		campaignProgress.LastUpdated = now
		campaignProgress.UpdatedAt = now

		if err := s.campaignRepo.SavePlayerCampaignProgress(campaignProgress); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update campaign progress with current mission")
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

	return progress, nil
}

// CompleteMission completes a mission for a player
func (s *campaignService) CompleteMission(playerID string, missionID string, choiceID string) (*MissionCompleteResult, error) {
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
			campaignProgress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, mission.ChapterID)
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
				campaignProgress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, mission.ChapterID)
				if err == nil && campaignProgress != nil {
					// Check if there are more chapters
					campaign, err := s.campaignRepo.GetCampaignByID(campaignProgress.CampaignID)
					if err == nil {
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
	return &MissionCompleteResult{
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

// POI Service Methods

func (s *campaignService) GetActivePOIs(playerID string) ([]model.POI, error) {
	return s.campaignRepo.GetActivePlayerPOIs(playerID)
}

func (s *campaignService) ActivatePOIForPlayer(poiID string, playerID string) error {
	// First check if the POI exists
	poi, err := s.campaignRepo.GetPOIByID(poiID)
	if err != nil {
		return err
	}

	// Make sure it's not already activated
	if poi.IsActive && poi.PlayerID == playerID {
		return errors.New("POI is already activated for this player")
	}

	// Activate the POI
	if err = s.campaignRepo.ActivatePOI(poiID, playerID); err != nil {
		return err
	}

	// Add SSE event
	if s.sseService != nil {
		// Get the activated POI
		poi, err := s.campaignRepo.GetPOIByID(poiID)
		if err == nil {
			s.sseService.SendEventToPlayer(playerID, "campaign_poi_updated", map[string]interface{}{
				"poi":         poi,
				"isActivated": true,
				"timestamp":   time.Now().Format(time.RFC3339),
			})
		}
	}

	return nil
}

func (s *campaignService) CompletePOI(poiID string, playerID string) error {
	// First check if the POI exists and is active for this player
	poi, err := s.campaignRepo.GetPOIByID(poiID)
	if err != nil {
		return err
	}

	if !poi.IsActive || poi.PlayerID != playerID {
		return errors.New("POI is not active for this player")
	}

	if poi.IsCompleted {
		return errors.New("POI is already completed")
	}

	// Mark the POI as completed
	if err := s.campaignRepo.CompletePOI(poiID, playerID); err != nil {
		return err
	}

	// Get the mission and choice this POI is related to
	mission, err := s.campaignRepo.GetMissionByID(poi.MissionID)
	if err != nil {
		return err
	}

	// Check if this POI completion contributes to a choice completion
	if poi.ChoiceID != "" {
		s.CheckChoiceCompletion(playerID, mission.ID, poi.ChoiceID)
	}

	// Add SSE event
	if s.sseService != nil {
		// Get the updated POI
		poi, err := s.campaignRepo.GetPOIByID(poiID)
		if err == nil {
			s.sseService.SendEventToPlayer(playerID, "campaign_poi_updated", map[string]interface{}{
				"poi":         poi,
				"isCompleted": true,
				"timestamp":   time.Now().Format(time.RFC3339),
			})
		}
	}

	return nil
}

// Mission Operation Service Methods
func (s *campaignService) GetActiveMissionOperations(playerID string) ([]model.MissionOperation, error) {
	return s.campaignRepo.GetActivePlayerMissionOperations(playerID)
}

func (s *campaignService) ActivateMissionOperationForPlayer(operationID string, playerID string) error {
	// First check if the operation exists
	operation, err := s.campaignRepo.GetMissionOperationByID(operationID)
	if err != nil {
		return err
	}

	// Make sure it's not already activated
	if operation.IsActive && operation.PlayerID == playerID {
		return errors.New("operation is already activated for this player")
	}

	// Activate the operation
	if err = s.campaignRepo.ActivateMissionOperation(operationID, playerID); err != nil {
		return nil
	}

	// Add SSE event
	if s.sseService != nil {
		// Get the activated operation
		operation, err := s.campaignRepo.GetMissionOperationByID(operationID)
		if err == nil {
			s.sseService.SendEventToPlayer(playerID, "campaign_operation_updated", map[string]interface{}{
				"operation":   operation,
				"isActivated": true,
				"timestamp":   time.Now().Format(time.RFC3339),
			})
		}
	}

	return nil
}

func (s *campaignService) CompleteMissionOperation(operationID string, playerID string) error {
	// First check if the operation exists and is active for this player
	operation, err := s.campaignRepo.GetMissionOperationByID(operationID)
	if err != nil {
		return err
	}

	if !operation.IsActive || operation.PlayerID != playerID {
		return errors.New("operation is not active for this player")
	}

	if operation.IsCompleted {
		return errors.New("operation is already completed")
	}

	// Mark the operation as completed
	if err := s.campaignRepo.CompleteMissionOperation(operationID, playerID); err != nil {
		return err
	}

	// Get the mission and choice this operation is related to
	mission, err := s.campaignRepo.GetMissionByID(operation.MissionID)
	if err != nil {
		return err
	}

	// Check if this operation completion contributes to a choice completion
	if operation.ChoiceID != "" {
		s.CheckChoiceCompletion(playerID, mission.ID, operation.ChoiceID)
	}

	// Add SSE event
	if s.sseService != nil {
		// Get the updated operation
		operation, err := s.campaignRepo.GetMissionOperationByID(operationID)
		if err == nil {
			s.sseService.SendEventToPlayer(playerID, "campaign_operation_updated", map[string]interface{}{
				"operation":   operation,
				"isCompleted": true,
				"timestamp":   time.Now().Format(time.RFC3339),
			})
		}
	}

	return nil
}

// Player Action Tracking Service Methods
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
			// Get the choice
			choice, err := s.GetMissionChoiceByID(progress.CurrentActiveChoice)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get active choice")
				continue
			}

			// Get the conditions for this choice
			conditions, err := s.campaignRepo.GetPlayerCompletionConditions(playerID, choice.ID)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get choice conditions")
				continue
			}

			// Check if this action satisfies any of the conditions
			for _, condition := range conditions {
				if !condition.IsCompleted &&
					condition.Type == actionType &&
					condition.RequiredValue == actionValue {

					// If sequential order is required, check if this is the next condition
					if choice.SequentialOrder {
						// Get all conditions ordered by index
						allConditions, err := s.campaignRepo.GetPlayerCompletionConditions(playerID, choice.ID)
						if err != nil {
							s.logger.Error().Err(err).Msg("Failed to get all choice conditions")
							continue
						}

						// Check if any earlier condition is not completed
						for _, c := range allConditions {
							if c.OrderIndex < condition.OrderIndex && !c.IsCompleted {
								// Earlier condition not completed, can't complete this one yet
								continue
							}
						}
					}

					// Mark the condition as completed
					if err := s.campaignRepo.CompleteCondition(condition.ID, playerID); err != nil {
						s.logger.Error().Err(err).Msg("Failed to complete condition")
						continue
					}

					// Send SSE event for condition completion
					if s.sseService != nil {
						eventData := map[string]interface{}{
							"actionType":         actionType,
							"actionValue":        actionValue,
							"choiceId":           choice.ID,
							"conditionCompleted": true,
							"missionId":          progress.MissionID,
							"timestamp":          time.Now().Format(time.RFC3339),
						}
						s.sseService.SendEventToPlayer(playerID, "campaign_action_tracked", eventData)
					}

					// Check if all conditions are now completed for this choice
					allCompleted, _ := s.CheckChoiceCompletion(playerID, progress.MissionID, choice.ID)
					if allCompleted && s.sseService != nil {
						// Send choice completion event
						eventData := map[string]interface{}{
							"missionId":   progress.MissionID,
							"choiceId":    choice.ID,
							"isCompleted": true,
							"timestamp":   time.Now().Format(time.RFC3339),
						}
						s.sseService.SendEventToPlayer(playerID, "campaign_choice_updated", eventData)
					}
				}
			}
		} else {
			// No active choice, see if this action could activate any choice
			mission, err := s.GetMissionByID(progress.MissionID)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get mission")
				continue
			}

			// Check each choice in the mission
			for _, choice := range mission.Choices {
				// Get conditions for this choice
				conditions, err := s.campaignRepo.GetCompletionConditions(choice.ID)
				if err != nil {
					s.logger.Error().Err(err).Msg("Failed to get choice conditions")
					continue
				}

				// If the first condition matches this action, activate the choice
				for _, condition := range conditions {
					if condition.OrderIndex == 0 &&
						condition.Type == actionType &&
						condition.RequiredValue == actionValue {

						// Activate this choice for the player
						s.ActivateChoice(playerID, mission.ID, choice.ID)

						// Send SSE event for choice activation
						if s.sseService != nil {
							eventData := map[string]interface{}{
								"missionId":   mission.ID,
								"choiceId":    choice.ID,
								"isActivated": true,
								"timestamp":   time.Now().Format(time.RFC3339),
							}
							s.sseService.SendEventToPlayer(playerID, "campaign_choice_updated", eventData)
						}
						break
					}
				}
			}
		}
	}

	return nil
}

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
	if allCompleted {
		// Get mission progress
		progress, err := s.GetPlayerMissionProgress(playerID, missionID)
		if err != nil {
			return false, err
		}

		// Update mission progress to use this choice
		progress.ChoiceID = choiceID
		progress.Status = "completed"
		now := time.Now()
		progress.CompletedAt = &now

		if err := s.campaignRepo.SavePlayerMissionProgress(progress); err != nil {
			return false, err
		}

		// Get the choice to find the next mission
		choice, err := s.GetMissionChoiceByID(choiceID)
		if err != nil {
			return false, err
		}

		// If there's a next mission, start it
		if choice.NextMissionID != "" {
			s.StartMission(playerID, choice.NextMissionID)
		}

		return true, nil
	}

	return false, nil
}

func (s *campaignService) ActivateChoice(playerID string, missionID string, choiceID string) error {
	// Get the player's mission progress
	progress, err := s.GetPlayerMissionProgress(playerID, missionID)
	if err != nil {
		return err
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
	conditions, err := s.campaignRepo.GetCompletionConditions(choiceID)
	if err != nil {
		return err
	}

	for _, condition := range conditions {
		playerCondition := model.CompletionCondition{
			ChoiceID:        choiceID,
			Type:            condition.Type,
			RequiredValue:   condition.RequiredValue,
			AdditionalValue: condition.AdditionalValue,
			OrderIndex:      condition.OrderIndex,
			IsCompleted:     false,
			PlayerID:        playerID,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := s.campaignRepo.SaveCompletionCondition(&playerCondition); err != nil {
			return err
		}
	}

	// Activate any POIs related to this choice
	pois, err := s.campaignRepo.GetPOIsByChoice(choiceID)
	if err != nil {
		return err
	}

	for _, poi := range pois {
		if err := s.campaignRepo.ActivatePOI(poi.ID, playerID); err != nil {
			return err
		}
	}

	// Activate any mission operations related to this choice
	operations, err := s.campaignRepo.GetMissionOperationsByChoice(choiceID)
	if err != nil {
		return err
	}

	for _, operation := range operations {
		if err := s.campaignRepo.ActivateMissionOperation(operation.ID, playerID); err != nil {
			return err
		}
	}

	return nil
}

// Helper function to get a choice by ID
func (s *campaignService) GetMissionChoiceByID(id string) (*model.MissionChoice, error) {
	var choice model.MissionChoice
	if err := s.campaignRepo.GetDB().Where("id = ?", id).First(&choice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("choice not found")
		}
		return nil, err
	}
	return &choice, nil
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

// Helper function to check if player meets mission requirements
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

// LoadCampaigns loads campaigns from YAML files
func (s *campaignService) LoadCampaigns(dirPath string) error {
	return s.campaignRepo.LoadCampaignsFromYAML(dirPath)
}
