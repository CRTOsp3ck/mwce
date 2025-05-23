// be/internal/service/campaign.go

package service

import (
	"errors"
	"fmt"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/util"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// CampaignService handles campaign-related business logic
type CampaignService interface {
	// Campaign management
	GetCampaigns() ([]model.Campaign, error)
	GetCampaignByID(id string) (*model.Campaign, error)
	GetChaptersByCampaignID(campaignID string) ([]model.Chapter, error)
	GetChapterByID(id string) (*model.Chapter, error)
	GetMissionsByChapterID(chapterID string) ([]model.Mission, error)
	GetMissionByID(id string) (*model.Mission, error)
	GetBranchesByMissionID(missionID string) ([]model.Branch, error)
	GetBranchByID(id string) (*model.Branch, error)

	// Player progress
	GetPlayerCampaignProgress(playerID, campaignID string) (*model.PlayerCampaignProgress, error)
	StartCampaign(playerID, campaignID string) (*model.PlayerCampaignProgress, error)
	GetCurrentMission(playerID, campaignID string) (*model.Mission, error)
	SelectBranch(playerID, missionID, branchID string) error
	CompleteBranch(playerID, missionID, branchID string) error

	// POI interaction
	GetPOIsByBranchID(branchID string) ([]model.CampaignPOI, error)
	GetPOIByID(id string) (*model.CampaignPOI, error)
	GetDialoguesByPOIID(poiID string) ([]model.Dialogue, error)
	InteractWithPOI(playerID, poiID string, interactionType model.InteractionType) (*model.Dialogue, *model.ResourceEffect, error)
	CompletePOI(playerID, poiID string) error

	// Operation management
	GetOperationsByBranchID(branchID string) ([]model.CampaignOperation, error)
	GetOperationByID(id string) (*model.CampaignOperation, error)
	CompleteOperation(playerID, operationID, attemptID string) error

	// Progress checking
	CheckBranchCompletion(playerID, branchID string) (bool, error)
	CheckMissionCompletion(playerID, missionID string) (bool, error)

	// Provider implementations
	GetInjectedHotspots(playerID string, regionID *string) ([]model.Hotspot, error)
	GetInjectedOperations(playerID string, regionID *string) ([]model.Operation, error)
}

type campaignService struct {
	campaignRepo  repository.CampaignRepository
	playerRepo    repository.PlayerRepository
	territoryRepo repository.TerritoryRepository
	playerService PlayerService
	sseService    SSEService
	logger        zerolog.Logger
}

// NewCampaignService creates a new campaign service
func NewCampaignService(
	campaignRepo repository.CampaignRepository,
	playerRepo repository.PlayerRepository,
	territoryRepo repository.TerritoryRepository,
	playerService PlayerService,
	sseService SSEService,
	logger zerolog.Logger,
) CampaignService {
	return &campaignService{
		campaignRepo:  campaignRepo,
		playerRepo:    playerRepo,
		territoryRepo: territoryRepo,
		playerService: playerService,
		sseService:    sseService,
		logger:        logger,
	}
}

// GetCampaigns retrieves all campaigns
func (s *campaignService) GetCampaigns() ([]model.Campaign, error) {
	return s.campaignRepo.GetAllCampaigns()
}

// GetCampaignByID retrieves a campaign by ID
func (s *campaignService) GetCampaignByID(id string) (*model.Campaign, error) {
	return s.campaignRepo.GetCampaignByID(id)
}

// GetChaptersByCampaignID retrieves chapters by campaign ID
func (s *campaignService) GetChaptersByCampaignID(campaignID string) ([]model.Chapter, error) {
	return s.campaignRepo.GetChaptersByCampaignID(campaignID)
}

// GetChapterByID retrieves a chapter by ID
func (s *campaignService) GetChapterByID(id string) (*model.Chapter, error) {
	return s.campaignRepo.GetChapterByID(id)
}

// GetMissionsByChapterID retrieves missions by chapter ID
func (s *campaignService) GetMissionsByChapterID(chapterID string) ([]model.Mission, error) {
	return s.campaignRepo.GetMissionsByChapterID(chapterID)
}

// GetMissionByID retrieves a mission by ID
func (s *campaignService) GetMissionByID(id string) (*model.Mission, error) {
	return s.campaignRepo.GetMissionByID(id)
}

// GetBranchesByMissionID retrieves branches by mission ID
func (s *campaignService) GetBranchesByMissionID(missionID string) ([]model.Branch, error) {
	return s.campaignRepo.GetBranchesByMissionID(missionID)
}

// GetBranchByID retrieves a branch by ID
func (s *campaignService) GetBranchByID(id string) (*model.Branch, error) {
	return s.campaignRepo.GetBranchByID(id)
}

// GetPlayerCampaignProgress retrieves a player's campaign progress
func (s *campaignService) GetPlayerCampaignProgress(playerID, campaignID string) (*model.PlayerCampaignProgress, error) {
	return s.campaignRepo.GetPlayerCampaignProgress(playerID, campaignID)
}

// StartCampaign starts a campaign for a player
func (s *campaignService) StartCampaign(playerID, campaignID string) (*model.PlayerCampaignProgress, error) {
	// Check if campaign exists
	campaign, err := s.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return nil, err
	}

	// Check if player has already started this campaign
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, campaignID)
	if err != nil {
		return nil, err
	}

	if progress != nil {
		return progress, nil // Already started
	}

	// Get first chapter
	chapters, err := s.campaignRepo.GetChaptersByCampaignID(campaignID)
	if err != nil {
		return nil, err
	}

	if len(chapters) == 0 {
		return nil, errors.New("campaign has no chapters")
	}

	// Get first mission
	missions, err := s.campaignRepo.GetMissionsByChapterID(chapters[0].ID)
	if err != nil {
		return nil, err
	}

	if len(missions) == 0 {
		return nil, errors.New("first chapter has no missions")
	}

	// Create new progress
	progress = &model.PlayerCampaignProgress{
		PlayerID:              playerID,
		CampaignID:            campaignID,
		CurrentMissionID:      &missions[0].ID,
		CompletedMissionIDs:   []string{},
		CompletedBranchIDs:    []string{},
		CompletedPOIIDs:       []string{},
		CompletedOperationIDs: []string{},
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.campaignRepo.CreatePlayerCampaignProgress(progress); err != nil {
		return nil, err
	}

	// Send notification
	message := fmt.Sprintf("You have started the '%s' campaign!", campaign.Name)
	s.playerService.AddNotification(playerID, message, util.NotificationTypeCampaign)

	return progress, nil
}

// GetCurrentMission retrieves a player's current mission in a campaign
func (s *campaignService) GetCurrentMission(playerID, campaignID string) (*model.Mission, error) {
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, campaignID)
	if err != nil {
		return nil, err
	}

	if progress == nil {
		return nil, errors.New("player has not started this campaign")
	}

	if progress.CurrentMissionID == nil {
		return nil, errors.New("player has no current mission")
	}

	return s.campaignRepo.GetMissionByID(*progress.CurrentMissionID)
}

// SelectBranch selects a branch for a mission
func (s *campaignService) SelectBranch(playerID, missionID, branchID string) error {
	// Get the mission to find campaign ID
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return err
	}

	// Get the chapter to find campaign ID
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return err
	}

	// Get player's progress
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		return err
	}

	if progress == nil {
		return errors.New("player has not started this campaign")
	}

	// Check if this is the player's current mission
	if progress.CurrentMissionID == nil || *progress.CurrentMissionID != missionID {
		return errors.New("this is not the player's current mission")
	}

	// Verify branch exists
	branch, err := s.campaignRepo.GetBranchByID(branchID)
	if err != nil {
		return err
	}

	if branch.MissionID != missionID {
		return errors.New("branch does not belong to this mission")
	}

	// Set current branch
	progress.CurrentBranchID = &branchID
	progress.UpdatedAt = time.Now()

	return s.campaignRepo.UpdatePlayerCampaignProgress(progress)
}

// CompleteBranch completes a branch for a mission
func (s *campaignService) CompleteBranch(playerID, missionID, branchID string) error {
	// Get the mission to find campaign ID
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return err
	}

	// Get the chapter to find campaign ID
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return err
	}

	// Get player's progress
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		return err
	}

	if progress == nil {
		return errors.New("player has not started this campaign")
	}

	// Check if this is the player's current mission and branch
	if progress.CurrentMissionID == nil || *progress.CurrentMissionID != missionID {
		return errors.New("this is not the player's current mission")
	}

	if progress.CurrentBranchID == nil || *progress.CurrentBranchID != branchID {
		return errors.New("this is not the player's current branch")
	}

	// Verify branch exists
	branch, err := s.campaignRepo.GetBranchByID(branchID)
	if err != nil {
		return err
	}

	if branch.MissionID != missionID {
		return errors.New("branch does not belong to this mission")
	}

	// Check if branch is complete
	complete, err := s.CheckBranchCompletion(playerID, branchID)
	if err != nil {
		return err
	}

	if !complete {
		return errors.New("branch is not complete")
	}

	// Add to completed branches
	progress.CompletedBranchIDs = append(progress.CompletedBranchIDs, branchID)

	// Add to completed missions
	progress.CompletedMissionIDs = append(progress.CompletedMissionIDs, missionID)

	// Find next mission
	missions, err := s.campaignRepo.GetMissionsByChapterID(mission.ChapterID)
	if err != nil {
		return err
	}

	// Find current mission index
	currentIndex := -1
	for i, m := range missions {
		if m.ID == missionID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return errors.New("current mission not found in chapter")
	}

	// Check if there are more missions in this chapter
	if currentIndex < len(missions)-1 {
		// Move to next mission
		nextMission := missions[currentIndex+1]
		progress.CurrentMissionID = &nextMission.ID
		progress.CurrentBranchID = nil
	} else {
		// Check if there are more chapters
		chapters, err := s.campaignRepo.GetChaptersByCampaignID(chapter.CampaignID)
		if err != nil {
			return err
		}

		// Find current chapter index
		currentChapterIndex := -1
		for i, c := range chapters {
			if c.ID == chapter.ID {
				currentChapterIndex = i
				break
			}
		}

		if currentChapterIndex == -1 {
			return errors.New("current chapter not found in campaign")
		}

		// Check if there are more chapters
		if currentChapterIndex < len(chapters)-1 {
			// Move to first mission of next chapter
			nextChapter := chapters[currentChapterIndex+1]
			nextMissions, err := s.campaignRepo.GetMissionsByChapterID(nextChapter.ID)
			if err != nil {
				return err
			}

			if len(nextMissions) == 0 {
				return errors.New("next chapter has no missions")
			}

			progress.CurrentMissionID = &nextMissions[0].ID
			progress.CurrentBranchID = nil
		} else {
			// Campaign complete
			progress.CurrentMissionID = nil
			progress.CurrentBranchID = nil
		}
	}

	progress.UpdatedAt = time.Now()

	if err := s.campaignRepo.UpdatePlayerCampaignProgress(progress); err != nil {
		return err
	}

	// Send notification
	message := fmt.Sprintf("You have completed mission '%s'!", mission.Name)
	s.playerService.AddNotification(playerID, message, util.NotificationTypeCampaign)

	return nil
}

// GetPOIsByBranchID retrieves POIs by branch ID
func (s *campaignService) GetPOIsByBranchID(branchID string) ([]model.CampaignPOI, error) {
	// Get the basic POIs
	pois, err := s.campaignRepo.GetPOIsByBranchID(branchID)
	if err != nil {
		return nil, err
	}

	// Enrich each POI with location information
	enrichedPOIs := make([]model.CampaignPOI, 0, len(pois))

	for _, poi := range pois {
		// Get city information
		city, err := s.territoryRepo.GetCityByID(poi.CityID)
		if err != nil {
			s.logger.Error().Err(err).Str("cityID", poi.CityID).Msg("Failed to get city for POI")
			// Add the POI without location info if we can't get the city
			enrichedPOIs = append(enrichedPOIs, poi)
			continue
		}

		// Get district information
		district, err := s.territoryRepo.GetDistrictByID(city.DistrictID)
		if err != nil {
			s.logger.Error().Err(err).Str("districtID", city.DistrictID).Msg("Failed to get district for city")
			// Add the POI without full location info if we can't get the district
			enrichedPOIs = append(enrichedPOIs, poi)
			continue
		}

		// Get region information
		region, err := s.territoryRepo.GetRegionByID(district.RegionID)
		if err != nil {
			s.logger.Error().Err(err).Str("regionID", district.RegionID).Msg("Failed to get region for district")
			// Add the POI without full location info if we can't get the region
			enrichedPOIs = append(enrichedPOIs, poi)
			continue
		}

		// Create enriched POI with location information
		// We'll add the location info to the POI's metadata field since we can't modify the struct
		if poi.Metadata == nil {
			poi.Metadata = make(map[string]interface{})
		}

		poi.Metadata["regionName"] = region.Name
		poi.Metadata["districtName"] = district.Name
		poi.Metadata["cityName"] = city.Name
		poi.Metadata["fullLocation"] = fmt.Sprintf("%s, %s, %s", city.Name, district.Name, region.Name)

		enrichedPOIs = append(enrichedPOIs, poi)
	}

	return enrichedPOIs, nil
}

// GetPOIByID retrieves a POI by ID
func (s *campaignService) GetPOIByID(id string) (*model.CampaignPOI, error) {
	return s.campaignRepo.GetPOIByID(id)
}

// GetDialoguesByPOIID retrieves dialogues by POI ID
func (s *campaignService) GetDialoguesByPOIID(poiID string) ([]model.Dialogue, error) {
	return s.campaignRepo.GetDialoguesByPOIID(poiID)
}

// InteractWithPOI allows a player to interact with a POI
func (s *campaignService) InteractWithPOI(playerID, poiID string, interactionType model.InteractionType) (*model.Dialogue, *model.ResourceEffect, error) {
	// Get the POI
	poi, err := s.campaignRepo.GetPOIByID(poiID)
	if err != nil {
		return nil, nil, err
	}

	// Get the branch to find mission ID
	branch, err := s.campaignRepo.GetBranchByID(poi.BranchID)
	if err != nil {
		return nil, nil, err
	}

	// Get the mission to find chapter ID
	mission, err := s.campaignRepo.GetMissionByID(branch.MissionID)
	if err != nil {
		return nil, nil, err
	}

	// Get the chapter to find campaign ID
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return nil, nil, err
	}

	// Get player's progress
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		return nil, nil, err
	}

	if progress == nil {
		return nil, nil, errors.New("player has not started this campaign")
	}

	// Check if this is part of the player's current branch
	if progress.CurrentBranchID == nil || *progress.CurrentBranchID != poi.BranchID {
		return nil, nil, errors.New("this POI is not part of the player's current branch")
	}

	// Get or create POI record
	poiRecord, err := s.campaignRepo.GetPlayerPOIRecordByIDs(progress.ID, poiID)
	if err != nil {
		return nil, nil, err
	}

	if poiRecord == nil {
		// Create new record
		poiRecord = &model.PlayerPOIRecord{
			PlayerID:    playerID,
			ProgressID:  progress.ID,
			POIID:       poiID,
			IsCompleted: false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.campaignRepo.CreatePlayerPOIRecord(poiRecord); err != nil {
			return nil, nil, err
		}
	}

	if poiRecord.IsCompleted {
		return nil, nil, errors.New("this POI has already been completed")
	}

	// Get dialogues
	dialogues, err := s.campaignRepo.GetDialoguesByPOIID(poiID)
	if err != nil {
		return nil, nil, err
	}

	if len(dialogues) == 0 {
		return nil, nil, errors.New("POI has no dialogues")
	}

	// Find the next interactive dialogue that matches the interaction type
	var nextDialogue *model.Dialogue
	for _, dialogue := range dialogues {
		// Skip NPC dialogues
		if dialogue.Speaker != "Player" {
			continue
		}

		// Skip dialogues with different interaction type
		if dialogue.InteractionType == nil || *dialogue.InteractionType != interactionType {
			continue
		}

		// Check if this dialogue has already been used
		dialogueState, err := s.campaignRepo.GetDialogueStateByIDs(poiRecord.ID, dialogue.ID)
		if err != nil {
			return nil, nil, err
		}

		if dialogueState == nil || !dialogueState.IsCompleted {
			nextDialogue = &dialogue
			break
		}
	}

	if nextDialogue == nil {
		return nil, nil, errors.New("no matching dialogue found for this interaction type")
	}

	// Find response dialogue (next in sequence)
	var responseDialogue *model.Dialogue
	for _, dialogue := range dialogues {
		if dialogue.Order == nextDialogue.Order+1 && dialogue.Speaker == "NPC" {
			responseDialogue = &dialogue
			break
		}
	}

	if responseDialogue == nil {
		return nil, nil, errors.New("no response dialogue found")
	}

	// Create or update dialogue state
	dialogueState, err := s.campaignRepo.GetDialogueStateByIDs(poiRecord.ID, nextDialogue.ID)
	if err != nil {
		return nil, nil, err
	}

	if dialogueState == nil {
		// Create new state
		dialogueState = &model.DialogueState{
			RecordID:     poiRecord.ID,
			DialogueID:   nextDialogue.ID,
			IsCompleted:  true,
			PlayerChoice: nextDialogue.InteractionType,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.campaignRepo.CreateDialogueState(dialogueState); err != nil {
			return nil, nil, err
		}
	} else {
		// Update existing state
		dialogueState.IsCompleted = true
		dialogueState.PlayerChoice = nextDialogue.InteractionType
		dialogueState.UpdatedAt = time.Now()

		if err := s.campaignRepo.UpdateDialogueState(dialogueState); err != nil {
			return nil, nil, err
		}
	}

	// Create dialogue state for response as well
	responseState, err := s.campaignRepo.GetDialogueStateByIDs(poiRecord.ID, responseDialogue.ID)
	if err != nil {
		return nil, nil, err
	}

	if responseState == nil {
		// Create new state
		responseState = &model.DialogueState{
			RecordID:    poiRecord.ID,
			DialogueID:  responseDialogue.ID,
			IsCompleted: true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.campaignRepo.CreateDialogueState(responseState); err != nil {
			return nil, nil, err
		}
	} else {
		// Update existing state
		responseState.IsCompleted = true
		responseState.UpdatedAt = time.Now()

		if err := s.campaignRepo.UpdateDialogueState(responseState); err != nil {
			return nil, nil, err
		}
	}

	// Apply resource effects if this is a success
	var resourceEffect *model.ResourceEffect
	if responseDialogue.IsSuccess != nil && *responseDialogue.IsSuccess {
		resourceEffect = &responseDialogue.ResourceEffect

		// Update player resources
		resourceUpdates := make(map[string]int)

		if resourceEffect.Money != 0 {
			resourceUpdates["money"] = resourceEffect.Money
		}

		if resourceEffect.Crew != 0 {
			resourceUpdates["crew"] = resourceEffect.Crew
		}

		if resourceEffect.Weapons != 0 {
			resourceUpdates["weapons"] = resourceEffect.Weapons
		}

		if resourceEffect.Vehicles != 0 {
			resourceUpdates["vehicles"] = resourceEffect.Vehicles
		}

		if resourceEffect.Respect != 0 {
			resourceUpdates["respect"] = resourceEffect.Respect
		}

		if resourceEffect.Influence != 0 {
			resourceUpdates["influence"] = resourceEffect.Influence
		}

		if resourceEffect.Heat != 0 {
			resourceUpdates["heat"] = resourceEffect.Heat
		}

		if len(resourceUpdates) > 0 {
			if err := s.playerService.UpdatePlayerResources(playerID, resourceUpdates); err != nil {
				s.logger.Error().Err(err).Msg("Failed to update player resources after dialogue interaction")
			}
		}
	}

	// Check if all dialogues are completed
	allCompleted := true
	for _, dialogue := range dialogues {
		dialogueState, err := s.campaignRepo.GetDialogueStateByIDs(poiRecord.ID, dialogue.ID)
		if err != nil {
			return nil, nil, err
		}

		if dialogueState == nil || !dialogueState.IsCompleted {
			allCompleted = false
			break
		}
	}

	// If all dialogues are completed, mark POI as completed
	if allCompleted {
		poiRecord.IsCompleted = true
		poiRecord.CompletedAt = ptrTime(time.Now())
		poiRecord.UpdatedAt = time.Now()

		if err := s.campaignRepo.UpdatePlayerPOIRecord(poiRecord); err != nil {
			return nil, nil, err
		}

		// Add to completed POIs
		progress.CompletedPOIIDs = append(progress.CompletedPOIIDs, poiID)
		progress.UpdatedAt = time.Now()

		if err := s.campaignRepo.UpdatePlayerCampaignProgress(progress); err != nil {
			return nil, nil, err
		}

		// Send notification
		message := fmt.Sprintf("You have completed interaction with %s!", poi.Name)
		s.playerService.AddNotification(playerID, message, util.NotificationTypeCampaign)
	}

	return responseDialogue, resourceEffect, nil
}

// CompletePOI marks a POI as completed
func (s *campaignService) CompletePOI(playerID, poiID string) error {
	// Get the POI
	poi, err := s.campaignRepo.GetPOIByID(poiID)
	if err != nil {
		return err
	}

	// Get the branch to find mission ID
	branch, err := s.campaignRepo.GetBranchByID(poi.BranchID)
	if err != nil {
		return err
	}

	// Get the mission to find chapter ID
	mission, err := s.campaignRepo.GetMissionByID(branch.MissionID)
	if err != nil {
		return err
	}

	// Get the chapter to find campaign ID
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return err
	}

	// Get player's progress
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		return err
	}

	if progress == nil {
		return errors.New("player has not started this campaign")
	}

	// Check if this is part of the player's current branch
	if progress.CurrentBranchID == nil || *progress.CurrentBranchID != poi.BranchID {
		return errors.New("this POI is not part of the player's current branch")
	}

	// Get or create POI record
	poiRecord, err := s.campaignRepo.GetPlayerPOIRecordByIDs(progress.ID, poiID)
	if err != nil {
		return err
	}

	if poiRecord == nil {
		// Create new record
		poiRecord = &model.PlayerPOIRecord{
			PlayerID:    playerID,
			ProgressID:  progress.ID,
			POIID:       poiID,
			IsCompleted: true,
			CompletedAt: ptrTime(time.Now()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.campaignRepo.CreatePlayerPOIRecord(poiRecord); err != nil {
			return err
		}
	} else if !poiRecord.IsCompleted {
		// Update existing record
		poiRecord.IsCompleted = true
		poiRecord.CompletedAt = ptrTime(time.Now())
		poiRecord.UpdatedAt = time.Now()

		if err := s.campaignRepo.UpdatePlayerPOIRecord(poiRecord); err != nil {
			return err
		}
	} else {
		// Already completed
		return nil
	}

	// Add to completed POIs
	progress.CompletedPOIIDs = append(progress.CompletedPOIIDs, poiID)
	progress.UpdatedAt = time.Now()

	if err := s.campaignRepo.UpdatePlayerCampaignProgress(progress); err != nil {
		return err
	}

	// Send notification
	message := fmt.Sprintf("You have completed interaction with %s!", poi.Name)
	s.playerService.AddNotification(playerID, message, util.NotificationTypeCampaign)

	return nil
}

// GetOperationsByBranchID retrieves operations by branch ID
// GetOperationsByBranchID retrieves operations by branch ID with region names
func (s *campaignService) GetOperationsByBranchID(branchID string) ([]model.CampaignOperation, error) {
	// Get the basic operations
	operations, err := s.campaignRepo.GetOperationsByBranchID(branchID)
	if err != nil {
		return nil, err
	}

	// Enrich each operation with region information
	enrichedOperations := make([]model.CampaignOperation, 0, len(operations))

	for _, operation := range operations {
		// Initialize metadata if nil
		if operation.Metadata == nil {
			operation.Metadata = make(map[string]interface{})
		}

		// Get region names for this operation
		regionNames := make([]string, 0, len(operation.RegionIDs))
		for _, regionID := range operation.RegionIDs {
			region, err := s.territoryRepo.GetRegionByID(regionID)
			if err != nil {
				s.logger.Warn().Err(err).Str("regionID", regionID).Msg("Failed to get region for operation")
				// Use a fallback name if region not found
				regionNames = append(regionNames, fmt.Sprintf("Region-%s", regionID[:8]))
				continue
			}
			regionNames = append(regionNames, region.Name)
		}

		// Add region information to metadata
		operation.Metadata["regionNames"] = regionNames
		if len(regionNames) == 0 {
			operation.Metadata["regionsDisplay"] = "All Regions"
		} else if len(regionNames) == 1 {
			operation.Metadata["regionsDisplay"] = regionNames[0]
		} else if len(regionNames) <= 3 {
			operation.Metadata["regionsDisplay"] = strings.Join(regionNames, ", ")
		} else {
			operation.Metadata["regionsDisplay"] = fmt.Sprintf("%s + %d more", strings.Join(regionNames[:2], ", "), len(regionNames)-2)
		}

		s.logger.Debug().
			Str("operationId", operation.ID).
			Str("operationName", operation.Name).
			Strs("regionIds", operation.RegionIDs).
			Strs("regionNames", regionNames).
			Msg("Successfully enriched operation with region data")

		enrichedOperations = append(enrichedOperations, operation)
	}

	return enrichedOperations, nil
}

// GetOperationByID retrieves an operation by ID
func (s *campaignService) GetOperationByID(id string) (*model.CampaignOperation, error) {
	return s.campaignRepo.GetOperationByID(id)
}

// CompleteOperation marks an operation as completed
func (s *campaignService) CompleteOperation(playerID, operationID, attemptID string) error {
	// Get the operation
	operation, err := s.campaignRepo.GetOperationByID(operationID)
	if err != nil {
		return err
	}

	// Get the branch to find mission ID
	branch, err := s.campaignRepo.GetBranchByID(operation.BranchID)
	if err != nil {
		return err
	}

	// Get the mission to find chapter ID
	mission, err := s.campaignRepo.GetMissionByID(branch.MissionID)
	if err != nil {
		return err
	}

	// Get the chapter to find campaign ID
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return err
	}

	// Get player's progress
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		return err
	}

	if progress == nil {
		return errors.New("player has not started this campaign")
	}

	// Check if this is part of the player's current branch
	if progress.CurrentBranchID == nil || *progress.CurrentBranchID != operation.BranchID {
		return errors.New("this operation is not part of the player's current branch")
	}

	// Get or create operation record
	operationRecord, err := s.campaignRepo.GetPlayerOperationRecordByIDs(progress.ID, operationID)
	if err != nil {
		return err
	}

	if operationRecord == nil {
		// Create new record
		operationRecord = &model.PlayerOperationRecord{
			PlayerID:    playerID,
			ProgressID:  progress.ID,
			OperationID: operationID,
			AttemptID:   attemptID,
			IsCompleted: true,
			CompletedAt: ptrTime(time.Now()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.campaignRepo.CreatePlayerOperationRecord(operationRecord); err != nil {
			return err
		}
	} else if !operationRecord.IsCompleted {
		// Update existing record
		operationRecord.IsCompleted = true
		operationRecord.CompletedAt = ptrTime(time.Now())
		operationRecord.UpdatedAt = time.Now()

		if err := s.campaignRepo.UpdatePlayerOperationRecord(operationRecord); err != nil {
			return err
		}
	} else {
		// Already completed
		return nil
	}

	// Add to completed operations
	progress.CompletedOperationIDs = append(progress.CompletedOperationIDs, operationID)
	progress.UpdatedAt = time.Now()

	if err := s.campaignRepo.UpdatePlayerCampaignProgress(progress); err != nil {
		return err
	}

	// Send notification
	message := fmt.Sprintf("You have completed the operation '%s' for your campaign!", operation.Name)
	s.playerService.AddNotification(playerID, message, util.NotificationTypeCampaign)

	return nil
}

// CheckBranchCompletion checks if a branch is complete
func (s *campaignService) CheckBranchCompletion(playerID, branchID string) (bool, error) {
	// Get the branch
	branch, err := s.campaignRepo.GetBranchByID(branchID)
	if err != nil {
		return false, err
	}

	// Get the mission to find chapter ID
	mission, err := s.campaignRepo.GetMissionByID(branch.MissionID)
	if err != nil {
		return false, err
	}

	// Get the chapter to find campaign ID
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return false, err
	}

	// Get player's progress
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		return false, err
	}

	if progress == nil {
		return false, errors.New("player has not started this campaign")
	}

	// Check operations
	operations, err := s.campaignRepo.GetOperationsByBranchID(branchID)
	if err != nil {
		return false, err
	}

	for _, operation := range operations {
		if !contains(progress.CompletedOperationIDs, operation.ID) {
			return false, nil
		}
	}

	// Check POIs
	pois, err := s.campaignRepo.GetPOIsByBranchID(branchID)
	if err != nil {
		return false, err
	}

	for _, poi := range pois {
		if !contains(progress.CompletedPOIIDs, poi.ID) {
			return false, nil
		}
	}

	return true, nil
}

// CheckMissionCompletion checks if a mission is complete
func (s *campaignService) CheckMissionCompletion(playerID, missionID string) (bool, error) {
	// Get the mission to find chapter ID
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		return false, err
	}

	// Get the chapter to find campaign ID
	chapter, err := s.campaignRepo.GetChapterByID(mission.ChapterID)
	if err != nil {
		return false, err
	}

	// Get player's progress
	progress, err := s.campaignRepo.GetPlayerCampaignProgress(playerID, chapter.CampaignID)
	if err != nil {
		return false, err
	}

	if progress == nil {
		return false, errors.New("player has not started this campaign")
	}

	return contains(progress.CompletedMissionIDs, missionID), nil
}

// GetInjectedHotspots implements CustomHotspotProvider
func (s *campaignService) GetInjectedHotspots(playerID string, regionID *string) ([]model.Hotspot, error) {
	// If no region specified, return empty list
	if regionID == nil {
		return []model.Hotspot{}, nil
	}

	// Get all player campaign progress
	progresses, err := s.campaignRepo.GetAllPlayerCampaignProgress(playerID)
	if err != nil {
		return nil, err
	}

	if len(progresses) == 0 {
		return []model.Hotspot{}, nil
	}

	var result []model.Hotspot

	// For each active campaign
	for _, progress := range progresses {
		// Skip if no current mission or branch
		if progress.CurrentMissionID == nil || progress.CurrentBranchID == nil {
			continue
		}

		// Get POIs for current branch
		pois, err := s.campaignRepo.GetPOIsByBranchID(*progress.CurrentBranchID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get POIs for branch")
			continue
		}

		// Filter POIs by region and check if they're already completed
		for _, poi := range pois {
			// Get city to find region
			city, err := s.territoryRepo.GetCityByID(poi.CityID)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get city for POI")
				continue
			}

			district, err := s.territoryRepo.GetDistrictByID(city.DistrictID)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get district for city")
				continue
			}

			// Skip if not in the requested region
			if district.RegionID != *regionID {
				continue
			}

			// Skip if already completed
			if contains(progress.CompletedPOIIDs, poi.ID) {
				continue
			}

			// Convert to regular hotspot
			hotspot := model.Hotspot{
				ID:                 poi.ID,
				Name:               poi.Name,
				CityID:             poi.CityID,
				Type:               poi.Type,
				BusinessType:       poi.BusinessType,
				IsLegal:            poi.IsLegal,
				Income:             0, // No income for campaign POIs
				PendingCollection:  0,
				LastCollectionTime: nil,
				LastIncomeTime:     nil,
				Crew:               0,
				Weapons:            0,
				Vehicles:           0,
				DefenseStrength:    0,
				Metadata: map[string]interface{}{
					"isCampaignPOI": true,
					"campaignID":    progress.CampaignID,
					"missionID":     *progress.CurrentMissionID,
					"branchID":      *progress.CurrentBranchID,
				},
			}

			result = append(result, hotspot)
		}
	}

	return result, nil
}

// GetInjectedOperations implements CustomOperationsProvider
func (s *campaignService) GetInjectedOperations(playerID string, regionID *string) ([]model.Operation, error) {
	// Get all player campaign progress
	progresses, err := s.campaignRepo.GetAllPlayerCampaignProgress(playerID)
	if err != nil {
		return nil, err
	}

	if len(progresses) == 0 {
		return []model.Operation{}, nil
	}

	var result []model.Operation

	// For each active campaign
	for _, progress := range progresses {
		// Skip if no current mission or branch
		if progress.CurrentMissionID == nil || progress.CurrentBranchID == nil {
			continue
		}

		// Get operations for current branch
		operations, err := s.campaignRepo.GetOperationsByBranchID(*progress.CurrentBranchID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get operations for branch")
			continue
		}

		// Filter operations and check if they're already completed
		for _, operation := range operations {
			// Skip if already completed
			if contains(progress.CompletedOperationIDs, operation.ID) {
				continue
			}

			// If region specified, filter by region
			if regionID != nil {
				// Skip if not available in this region and not global
				if len(operation.RegionIDs) > 0 && !contains(operation.RegionIDs, *regionID) {
					continue
				}
			}

			// Convert to regular operation
			availableUntil := time.Now().Add(24 * time.Hour) // 24 hours

			op := model.Operation{
				ID:             operation.ID,
				Name:           operation.Name,
				Description:    operation.Description,
				Type:           operation.Type,
				IsSpecial:      operation.IsSpecial,
				IsActive:       true,
				RegionIDs:      operation.RegionIDs,
				Requirements:   operation.Requirements,
				Resources:      operation.Resources,
				Rewards:        operation.Rewards,
				Risks:          operation.Risks,
				Duration:       operation.Duration,
				SuccessRate:    operation.SuccessRate,
				AvailableUntil: availableUntil,
				Metadata: map[string]interface{}{
					"isCampaignOperation": true,
					"campaignID":          progress.CampaignID,
					"missionID":           *progress.CurrentMissionID,
					"branchID":            *progress.CurrentBranchID,
				},
			}

			result = append(result, op)
		}
	}

	return result, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
