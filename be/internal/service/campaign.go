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
	GetMissionBranchesProgress(playerID, missionID string) (map[string]bool, error)

	// Provider implementations
	GetInjectedHotspots(playerID string, regionID *string) ([]model.Hotspot, error)
	GetInjectedOperations(playerID string, regionID *string) ([]model.Operation, error)
	HandlePOITakeover(playerID, hotspotID string) error
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

// SelectBranch is deprecated - branches are automatically determined by completion
// This method is kept for backward compatibility but does nothing
func (s *campaignService) SelectBranch(playerID, missionID, branchID string) error {
	// No longer needed - branches are selected based on completion
	return nil
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

	// Set the completed branch as current (for tracking which path was taken)
	progress.CurrentBranchID = &branchID

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
	message := fmt.Sprintf("You have completed mission '%s' using the '%s' approach!", mission.Name, branch.Name)
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

// InteractWithPOI is deprecated - POIs are now simple completion markers
func (s *campaignService) InteractWithPOI(playerID, poiID string, interactionType model.InteractionType) (*model.Dialogue, *model.ResourceEffect, error) {
	// This method is kept for backward compatibility but simply completes the POI
	err := s.CompletePOI(playerID, poiID)
	if err != nil {
		return nil, nil, err
	}
	
	// Return empty dialogue and effect
	return &model.Dialogue{
		Text: "POI completed successfully",
	}, nil, nil
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

	// Check if this POI is part of the current mission
	if progress.CurrentMissionID == nil || *progress.CurrentMissionID != mission.ID {
		return errors.New("this POI is not part of the player's current mission")
	}

	// Check if already completed
	if contains(progress.CompletedPOIIDs, poiID) {
		return nil // Already completed
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

	// Check if this operation is part of the current mission
	if progress.CurrentMissionID == nil || *progress.CurrentMissionID != mission.ID {
		return errors.New("this operation is not part of the player's current mission")
	}

	// Check if already completed
	if contains(progress.CompletedOperationIDs, operationID) {
		return nil // Already completed
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

// GetMissionBranchesProgress gets the completion status of all branches for a mission
func (s *campaignService) GetMissionBranchesProgress(playerID, missionID string) (map[string]bool, error) {
	// Get all branches for the mission
	branches, err := s.campaignRepo.GetBranchesByMissionID(missionID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)

	// Check completion status for each branch
	for _, branch := range branches {
		complete, err := s.CheckBranchCompletion(playerID, branch.ID)
		if err != nil {
			// Log error but continue checking other branches
			s.logger.Error().Err(err).Str("branchID", branch.ID).Msg("Failed to check branch completion")
			result[branch.ID] = false
			continue
		}
		result[branch.ID] = complete
	}

	return result, nil
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
		// Skip if no current mission
		if progress.CurrentMissionID == nil {
			continue
		}

		// Get all branches for the current mission
		branches, err := s.campaignRepo.GetBranchesByMissionID(*progress.CurrentMissionID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get branches for mission")
			continue
		}

		// Get POIs for all branches of the current mission
		for _, branch := range branches {
			pois, err := s.campaignRepo.GetPOIsByBranchID(branch.ID)
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
						"branchID":      branch.ID,
						"branchName":    branch.Name,
					},
				}

				result = append(result, hotspot)
			}
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
		// Skip if no current mission
		if progress.CurrentMissionID == nil {
			continue
		}

		// Get all branches for the current mission
		branches, err := s.campaignRepo.GetBranchesByMissionID(*progress.CurrentMissionID)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get branches for mission")
			continue
		}

		// Get operations for all branches of the current mission
		for _, branch := range branches {
			operations, err := s.campaignRepo.GetOperationsByBranchID(branch.ID)
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
						"branchID":            branch.ID,
						"branchName":          branch.Name,
					},
				}

				result = append(result, op)
			}
		}
	}

	return result, nil
}

// HandlePOITakeover handles when a campaign POI is taken over in the territory system
func (s *campaignService) HandlePOITakeover(playerID, hotspotID string) error {
	// The hotspotID is actually the POI ID
	return s.CompletePOI(playerID, hotspotID)
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
