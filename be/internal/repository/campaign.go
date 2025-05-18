// internal/repository/campaign.go

package repository

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mwce-be/internal/model"
	"mwce-be/pkg/database"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// CampaignRepository handles campaign data persistence
type CampaignRepository interface {
	GetDB() *gorm.DB
	GetAllCampaigns() ([]model.Campaign, error)
	GetCampaignByID(campaignID string) (*model.Campaign, error)
	GetChapterByID(chapterID string) (*model.Chapter, error)
	GetMissionByID(missionID string) (*model.Mission, error)
	GetMissionsByChapterID(chapterID string) ([]model.Mission, error)
	GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error)
	GetPlayerCampaignProgresses(playerID string) ([]model.PlayerCampaignProgress, error)
	GetPlayerMissionProgress(playerID string, missionID string) (*model.PlayerMissionProgress, error)
	SavePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error
	SavePlayerMissionProgress(progress *model.PlayerMissionProgress) error

	// Template retrieval
	GetAllPOITemplates() ([]model.POITemplate, error)
	GetConditionTemplate(templateID string) (*model.ConditionTemplate, error)
	GetConditionTemplatesByChoice(choiceID string) ([]model.ConditionTemplate, error)
	GetPOITemplate(templateID string) (*model.POITemplate, error)
	GetPOITemplatesByChoice(choiceID string) ([]model.POITemplate, error)
	GetOperationTemplate(templateID string) (*model.OperationTemplate, error)
	GetOperationTemplatesByChoice(choiceID string) ([]model.OperationTemplate, error)

	// Player-specific instances
	CreatePlayerPOI(playerPOI *model.PlayerPOI) error
	GetPlayerPOI(playerPOIID string) (*model.PlayerPOI, error)
	GetActivePlayerPOIs(playerID string) ([]model.PlayerPOI, error)
	GetPlayerPOIsByMission(playerID string, missionID string) ([]model.PlayerPOI, error)
	ActivatePlayerPOI(playerID string, templateID string) (*model.PlayerPOI, error)
	CompletePlayerPOI(playerID string, playerPOIID string) error

	CreatePlayerMissionOperation(playerOp *model.PlayerMissionOperation) error
	GetPlayerMissionOperation(playerOpID string) (*model.PlayerMissionOperation, error)
	GetActivePlayerMissionOperations(playerID string) ([]model.PlayerMissionOperation, error)
	GetPlayerMissionOperationsByMission(playerID string, missionID string) ([]model.PlayerMissionOperation, error)
	ActivatePlayerMissionOperation(playerID string, templateID string) (*model.PlayerMissionOperation, error)
	CompletePlayerMissionOperation(playerID string, playerOpID string) error

	CreatePlayerCompletionCondition(condition *model.PlayerCompletionCondition) error
	GetPlayerCompletionCondition(conditionID string) (*model.PlayerCompletionCondition, error)
	GetPlayerCompletionConditions(playerID string, choiceID string) ([]model.PlayerCompletionCondition, error)
	CompletePlayerCompletionCondition(playerID string, conditionID string) error

	// YAML import
	LoadCampaignsFromYAML(dirPath string) error
}

// campaignRepository implements CampaignRepository
type campaignRepository struct {
	db     *gorm.DB
	logger zerolog.Logger
}

// NewCampaignRepository creates a new campaign repository
func NewCampaignRepository(db database.Database, logger zerolog.Logger) CampaignRepository {
	return &campaignRepository{
		db:     db.GetDB(),
		logger: logger,
	}
}

// GetDB returns the database connection
func (r *campaignRepository) GetDB() *gorm.DB {
	return r.db
}

// GetAllPOITemplates retrieves all POI templates in the system
func (r *campaignRepository) GetAllPOITemplates() ([]model.POITemplate, error) {
	var templates []model.POITemplate
	if err := r.db.Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve POI templates: %w", err)
	}
	return templates, nil
}

// GetAllCampaigns retrieves all available campaigns
func (r *campaignRepository) GetAllCampaigns() ([]model.Campaign, error) {
	var campaigns []model.Campaign
	if err := r.db.Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// GetCampaignByID retrieves a campaign by ID with its chapters and missions
func (r *campaignRepository) GetCampaignByID(campaignID string) (*model.Campaign, error) {
	var campaign model.Campaign

	// Get campaign with its chapters
	if err := r.db.Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("chapters.order ASC")
	}).First(&campaign, "id = ?", campaignID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("campaign not found")
		}
		return nil, err
	}

	// Get missions for each chapter
	for i, chapter := range campaign.Chapters {
		var missions []model.Mission
		if err := r.db.Preload("Choices").Where("chapter_id = ?", chapter.ID).Order("missions.order ASC").Find(&missions).Error; err != nil {
			return nil, err
		}
		campaign.Chapters[i].Missions = missions
	}

	return &campaign, nil
}

// GetChapterByID retrieves a chapter by ID with its missions
func (r *campaignRepository) GetChapterByID(chapterID string) (*model.Chapter, error) {
	var chapter model.Chapter

	// Get chapter
	if err := r.db.First(&chapter, "id = ?", chapterID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chapter not found")
		}
		return nil, err
	}

	// Get missions for the chapter
	var missions []model.Mission
	if err := r.db.Preload("Choices").Where("chapter_id = ?", chapterID).Order("missions.order ASC").Find(&missions).Error; err != nil {
		return nil, err
	}
	chapter.Missions = missions

	return &chapter, nil
}

// GetMissionByID retrieves a mission by ID with its choices and related data
func (r *campaignRepository) GetMissionByID(missionID string) (*model.Mission, error) {
	var mission model.Mission

	// Get mission with choices
	if err := r.db.Preload("Choices").First(&mission, "id = ?", missionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mission not found")
		}
		return nil, err
	}

	// Preload condition templates, POI templates, and operation templates for each choice
	for i, choice := range mission.Choices {
		var conditionTemplates []model.ConditionTemplate
		if err := r.db.Where("choice_id = ?", choice.ID).Order("order_index ASC").Find(&conditionTemplates).Error; err != nil {
			return nil, err
		}
		mission.Choices[i].Conditions = conditionTemplates

		var poiTemplates []model.POITemplate
		if err := r.db.Where("choice_id = ?", choice.ID).Find(&poiTemplates).Error; err != nil {
			return nil, err
		}
		mission.Choices[i].POITemplates = poiTemplates

		var operationTemplates []model.OperationTemplate
		if err := r.db.Where("choice_id = ?", choice.ID).Find(&operationTemplates).Error; err != nil {
			return nil, err
		}
		mission.Choices[i].OperationTemplates = operationTemplates
	}

	return &mission, nil
}

// GetMissionsByChapterID gets all missions for a chapter
func (r *campaignRepository) GetMissionsByChapterID(chapterID string) ([]model.Mission, error) {
	var missions []model.Mission
	if err := r.db.Where("chapter_id = ?", chapterID).Order("order ASC").Find(&missions).Error; err != nil {
		return nil, err
	}
	return missions, nil
}

// GetPlayerCampaignProgress retrieves a player's progress for a campaign
func (r *campaignRepository) GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error) {
	var progress model.PlayerCampaignProgress
	if err := r.db.Where("player_id = ? AND campaign_id = ?", playerID, campaignID).First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No progress found, not an error
		}
		return nil, err
	}
	return &progress, nil
}

// GetPlayerCampaignProgresses retrieves all campaign progress for a player
func (r *campaignRepository) GetPlayerCampaignProgresses(playerID string) ([]model.PlayerCampaignProgress, error) {
	var progresses []model.PlayerCampaignProgress
	if err := r.db.Where("player_id = ?", playerID).Find(&progresses).Error; err != nil {
		return nil, err
	}
	return progresses, nil
}

// GetPlayerMissionProgress retrieves a player's progress for a mission
func (r *campaignRepository) GetPlayerMissionProgress(playerID string, missionID string) (*model.PlayerMissionProgress, error) {
	var progress model.PlayerMissionProgress
	if err := r.db.Where("player_id = ? AND mission_id = ?", playerID, missionID).First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No progress found, not an error
		}
		return nil, err
	}
	return &progress, nil
}

// SavePlayerCampaignProgress saves a player's campaign progress
func (r *campaignRepository) SavePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error {
	if progress.ID == "" {
		progress.ID = uuid.New().String()
		progress.CreatedAt = time.Now()
	}
	progress.UpdatedAt = time.Now()
	return r.db.Save(progress).Error
}

// SavePlayerMissionProgress saves a player's mission progress
func (r *campaignRepository) SavePlayerMissionProgress(progress *model.PlayerMissionProgress) error {
	if progress.ID == "" {
		progress.ID = uuid.New().String()
		progress.CreatedAt = time.Now()
	}
	progress.UpdatedAt = time.Now()
	return r.db.Save(progress).Error
}

// -- Template retrieval methods --

// GetConditionTemplate retrieves a condition template by ID
func (r *campaignRepository) GetConditionTemplate(templateID string) (*model.ConditionTemplate, error) {
	var template model.ConditionTemplate
	if err := r.db.First(&template, "id = ?", templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("condition template not found")
		}
		return nil, err
	}
	return &template, nil
}

// GetConditionTemplatesByChoice retrieves all condition templates for a choice
func (r *campaignRepository) GetConditionTemplatesByChoice(choiceID string) ([]model.ConditionTemplate, error) {
	var templates []model.ConditionTemplate
	if err := r.db.Where("choice_id = ?", choiceID).Order("order_index ASC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetPOITemplate retrieves a POI template by ID
func (r *campaignRepository) GetPOITemplate(templateID string) (*model.POITemplate, error) {
	var template model.POITemplate
	if err := r.db.First(&template, "id = ?", templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("POI template not found")
		}
		return nil, err
	}
	return &template, nil
}

// GetPOITemplatesByChoice retrieves all POI templates for a choice
func (r *campaignRepository) GetPOITemplatesByChoice(choiceID string) ([]model.POITemplate, error) {
	var templates []model.POITemplate
	if err := r.db.Where("choice_id = ?", choiceID).Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// GetOperationTemplate retrieves an operation template by ID
func (r *campaignRepository) GetOperationTemplate(templateID string) (*model.OperationTemplate, error) {
	var template model.OperationTemplate
	if err := r.db.First(&template, "id = ?", templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("operation template not found")
		}
		return nil, err
	}
	return &template, nil
}

// GetOperationTemplatesByChoice retrieves all operation templates for a choice
func (r *campaignRepository) GetOperationTemplatesByChoice(choiceID string) ([]model.OperationTemplate, error) {
	var templates []model.OperationTemplate
	if err := r.db.Where("choice_id = ?", choiceID).Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// -- Player-specific instances methods --

// CreatePlayerPOI creates a new player POI instance
func (r *campaignRepository) CreatePlayerPOI(playerPOI *model.PlayerPOI) error {
	if playerPOI.ID == "" {
		playerPOI.ID = uuid.New().String()
		playerPOI.CreatedAt = time.Now()
	}
	playerPOI.UpdatedAt = time.Now()
	return r.db.Create(playerPOI).Error
}

// GetPlayerPOI retrieves a player POI by ID
func (r *campaignRepository) GetPlayerPOI(playerPOIID string) (*model.PlayerPOI, error) {
	var playerPOI model.PlayerPOI
	if err := r.db.First(&playerPOI, "id = ?", playerPOIID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("player POI not found")
		}
		return nil, err
	}
	return &playerPOI, nil
}

// GetActivePlayerPOIs retrieves all active POIs for a player
func (r *campaignRepository) GetActivePlayerPOIs(playerID string) ([]model.PlayerPOI, error) {
	var pois []model.PlayerPOI
	if err := r.db.Where("player_id = ? AND is_active = ?", playerID, true).Find(&pois).Error; err != nil {
		return nil, err
	}
	return pois, nil
}

// GetPlayerPOIsByMission retrieves all POIs for a player and mission
func (r *campaignRepository) GetPlayerPOIsByMission(playerID string, missionID string) ([]model.PlayerPOI, error) {
	var pois []model.PlayerPOI
	if err := r.db.Where("player_id = ? AND mission_id = ? AND is_active = ?", playerID, missionID, true).Find(&pois).Error; err != nil {
		return nil, err
	}
	return pois, nil
}

// ActivatePlayerPOI activates a POI for a player from a template
func (r *campaignRepository) ActivatePlayerPOI(playerID string, templateID string) (*model.PlayerPOI, error) {
	// First get the template
	var template model.POITemplate
	if err := r.db.First(&template, "id = ?", templateID).Error; err != nil {
		return nil, errors.New("POI template not found")
	}

	// Check if it's already activated for this player
	var existingPOI model.PlayerPOI
	if err := r.db.Where("template_id = ? AND player_id = ?", templateID, playerID).First(&existingPOI).Error; err == nil {
		// If found and not active, just update the active status
		if !existingPOI.IsActive {
			existingPOI.IsActive = true
			existingPOI.UpdatedAt = time.Now()
			if err := r.db.Save(&existingPOI).Error; err != nil {
				return nil, err
			}
			return &existingPOI, nil
		}
		// If already active, return it
		return &existingPOI, nil
	}

	// Create a new player POI
	playerPOI := &model.PlayerPOI{
		ID:           uuid.New().String(),
		TemplateID:   templateID,
		PlayerID:     playerID,
		MissionID:    template.MissionID,
		ChoiceID:     template.ChoiceID,
		Name:         template.Name,
		Description:  template.Description,
		LocationType: template.LocationType,
		LocationID:   template.LocationID,
		IsActive:     true,
		IsCompleted:  false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := r.db.Create(playerPOI).Error; err != nil {
		return nil, err
	}

	return playerPOI, nil
}

// CompletePlayerPOI marks a player POI as completed
func (r *campaignRepository) CompletePlayerPOI(playerID string, playerPOIID string) error {
	var playerPOI model.PlayerPOI
	if err := r.db.Where("id = ? AND player_id = ?", playerPOIID, playerID).First(&playerPOI).Error; err != nil {
		return errors.New("player POI not found or not authorized")
	}

	if playerPOI.IsCompleted {
		return errors.New("POI already completed")
	}

	now := time.Now()
	playerPOI.IsCompleted = true
	playerPOI.CompletedAt = &now
	playerPOI.UpdatedAt = now

	return r.db.Save(&playerPOI).Error
}

// CreatePlayerMissionOperation creates a new player mission operation
func (r *campaignRepository) CreatePlayerMissionOperation(playerOp *model.PlayerMissionOperation) error {
	if playerOp.ID == "" {
		playerOp.ID = uuid.New().String()
		playerOp.CreatedAt = time.Now()
	}
	playerOp.UpdatedAt = time.Now()
	return r.db.Create(playerOp).Error
}

// GetPlayerMissionOperation retrieves a player mission operation by ID
func (r *campaignRepository) GetPlayerMissionOperation(playerOpID string) (*model.PlayerMissionOperation, error) {
	var playerOp model.PlayerMissionOperation
	if err := r.db.First(&playerOp, "id = ?", playerOpID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("player mission operation not found")
		}
		return nil, err
	}
	return &playerOp, nil
}

// GetActivePlayerMissionOperations retrieves all active mission operations for a player
func (r *campaignRepository) GetActivePlayerMissionOperations(playerID string) ([]model.PlayerMissionOperation, error) {
	var operations []model.PlayerMissionOperation
	if err := r.db.Where("player_id = ? AND is_active = ?", playerID, true).Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

// GetPlayerMissionOperationsByMission retrieves all mission operations for a player and mission
func (r *campaignRepository) GetPlayerMissionOperationsByMission(playerID string, missionID string) ([]model.PlayerMissionOperation, error) {
	var operations []model.PlayerMissionOperation
	if err := r.db.Where("player_id = ? AND mission_id = ? AND is_active = ?", playerID, missionID, true).Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

// ActivatePlayerMissionOperation activates a mission operation for a player from a template
func (r *campaignRepository) ActivatePlayerMissionOperation(playerID string, templateID string) (*model.PlayerMissionOperation, error) {
	// First get the template
	var template model.OperationTemplate
	if err := r.db.First(&template, "id = ?", templateID).Error; err != nil {
		return nil, errors.New("operation template not found")
	}

	// Check if it's already activated for this player
	var existingOp model.PlayerMissionOperation
	if err := r.db.Where("template_id = ? AND player_id = ?", templateID, playerID).First(&existingOp).Error; err == nil {
		// If found and not active, just update the active status
		if !existingOp.IsActive {
			existingOp.IsActive = true
			existingOp.UpdatedAt = time.Now()
			if err := r.db.Save(&existingOp).Error; err != nil {
				return nil, err
			}
			return &existingOp, nil
		}
		// If already active, return it
		return &existingOp, nil
	}

	// Create a new player mission operation
	playerOp := &model.PlayerMissionOperation{
		ID:            uuid.New().String(),
		TemplateID:    templateID,
		PlayerID:      playerID,
		MissionID:     template.MissionID,
		ChoiceID:      template.ChoiceID,
		Name:          template.Name,
		Description:   template.Description,
		OperationType: template.OperationType,
		Resources:     template.Resources,
		Rewards:       template.Rewards,
		Risks:         template.Risks,
		Duration:      template.Duration,
		SuccessRate:   template.SuccessRate,
		IsActive:      true,
		IsCompleted:   false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := r.db.Create(playerOp).Error; err != nil {
		return nil, err
	}

	return playerOp, nil
}

// CompletePlayerMissionOperation marks a player mission operation as completed
func (r *campaignRepository) CompletePlayerMissionOperation(playerID string, playerOpID string) error {
	var playerOp model.PlayerMissionOperation
	if err := r.db.Where("id = ? AND player_id = ?", playerOpID, playerID).First(&playerOp).Error; err != nil {
		return errors.New("player mission operation not found or not authorized")
	}

	if playerOp.IsCompleted {
		return errors.New("mission operation already completed")
	}

	now := time.Now()
	playerOp.IsCompleted = true
	playerOp.CompletedAt = &now
	playerOp.UpdatedAt = now

	return r.db.Save(&playerOp).Error
}

// CreatePlayerCompletionCondition creates a new player completion condition
func (r *campaignRepository) CreatePlayerCompletionCondition(condition *model.PlayerCompletionCondition) error {
	if condition.ID == "" {
		condition.ID = uuid.New().String()
		condition.CreatedAt = time.Now()
	}
	condition.UpdatedAt = time.Now()
	return r.db.Create(condition).Error
}

// GetPlayerCompletionCondition retrieves a player completion condition by ID
func (r *campaignRepository) GetPlayerCompletionCondition(conditionID string) (*model.PlayerCompletionCondition, error) {
	var condition model.PlayerCompletionCondition
	if err := r.db.First(&condition, "id = ?", conditionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("player completion condition not found")
		}
		return nil, err
	}
	return &condition, nil
}

// GetPlayerCompletionConditions retrieves all completion conditions for a player and choice
func (r *campaignRepository) GetPlayerCompletionConditions(playerID string, choiceID string) ([]model.PlayerCompletionCondition, error) {
	var conditions []model.PlayerCompletionCondition
	if err := r.db.Where("player_id = ? AND choice_id = ?", playerID, choiceID).Order("order_index ASC").Find(&conditions).Error; err != nil {
		return nil, err
	}
	return conditions, nil
}

// CompletePlayerCompletionCondition marks a player completion condition as completed
func (r *campaignRepository) CompletePlayerCompletionCondition(playerID string, conditionID string) error {
	var condition model.PlayerCompletionCondition
	if err := r.db.Where("id = ? AND player_id = ?", conditionID, playerID).First(&condition).Error; err != nil {
		return errors.New("player completion condition not found or not authorized")
	}

	if condition.IsCompleted {
		return errors.New("completion condition already completed")
	}

	now := time.Now()
	condition.IsCompleted = true
	condition.CompletedAt = &now
	condition.UpdatedAt = now

	return r.db.Save(&condition).Error
}

// LoadCampaignsFromYAML loads campaigns from YAML files
func (r *campaignRepository) LoadCampaignsFromYAML(dirPath string) error {
	// Get list of YAML files in the directory
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Filter and sort the YAML files
	var yamlFiles []string
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml") {
			yamlFiles = append(yamlFiles, file.Name())
		}
	}

	// Sort files by name (which should have numeric prefixes like "0001_" or "1_")
	sort.Strings(yamlFiles)

	// Process each file in order
	for _, filename := range yamlFiles {
		filePath := filepath.Join(dirPath, filename)
		r.logger.Info().Str("file", filePath).Msg("Loading campaign from YAML")

		// Read the YAML file
		yamlData, err := ioutil.ReadFile(filePath)
		if err != nil {
			r.logger.Error().Err(err).Str("file", filePath).Msg("Failed to read YAML file")
			continue
		}

		// Parse YAML into a campaign structure
		var campaignData struct {
			ID               string `yaml:"id"`
			Title            string `yaml:"title"`
			Description      string `yaml:"description"`
			ImageURL         string `yaml:"image_url"`
			InitialChapterID string `yaml:"initial_chapter"`
			RequiredLevel    int    `yaml:"required_level"`
			Chapters         []struct {
				ID          string `yaml:"id"`
				Title       string `yaml:"title"`
				Description string `yaml:"description"`
				ImageURL    string `yaml:"image_url"`
				Order       int    `yaml:"order"`
				Missions    []struct {
					ID           string                 `yaml:"id"`
					Title        string                 `yaml:"title"`
					Description  string                 `yaml:"description"`
					Narrative    string                 `yaml:"narrative"`
					ImageURL     string                 `yaml:"image_url"`
					Type         string                 `yaml:"type"`
					Order        int                    `yaml:"order"`
					Requirements map[string]interface{} `yaml:"requirements"`
					Rewards      map[string]interface{} `yaml:"rewards"`
					Choices      []struct {
						ID              string                 `yaml:"id"`
						Text            string                 `yaml:"text"`
						NextMission     string                 `yaml:"next_mission"`
						Requirements    map[string]interface{} `yaml:"requirements"`
						Rewards         map[string]interface{} `yaml:"rewards"`
						SequentialOrder bool                   `yaml:"sequential_order"`
						Conditions      []struct {
							Type            string `yaml:"type"`
							RequiredValue   string `yaml:"required_value"`
							AdditionalValue string `yaml:"additional_value"`
							OrderIndex      int    `yaml:"order_index"`
						} `yaml:"conditions"`
						POIs []struct {
							Name         string `yaml:"name"`
							Description  string `yaml:"description"`
							LocationType string `yaml:"location_type"`
							LocationID   string `yaml:"location_id"`
						} `yaml:"pois"`
						Operations []struct {
							Name          string                 `yaml:"name"`
							Description   string                 `yaml:"description"`
							OperationType string                 `yaml:"operation_type"`
							Resources     map[string]interface{} `yaml:"resources"`
							Rewards       map[string]interface{} `yaml:"rewards"`
							Risks         map[string]interface{} `yaml:"risks"`
							Duration      int                    `yaml:"duration"`
							SuccessRate   int                    `yaml:"success_rate"`
						} `yaml:"operations"`
					} `yaml:"choices"`
				} `yaml:"missions"`
			} `yaml:"chapters"`
		}

		if err := yaml.Unmarshal(yamlData, &campaignData); err != nil {
			r.logger.Error().Err(err).Str("file", filePath).Msg("Failed to parse YAML")
			continue
		}

		// Start transaction for this campaign
		tx := r.db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to start transaction: %w", tx.Error)
		}

		// Check if campaign already exists
		var existingCampaign model.Campaign
		if err := tx.Where("id = ?", campaignData.ID).First(&existingCampaign).Error; err == nil {
			// Campaign exists, skip or update based on your needs
			r.logger.Info().Str("id", campaignData.ID).Msg("Campaign already exists, skipping")
			tx.Rollback()
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Database error
			r.logger.Error().Err(err).Str("id", campaignData.ID).Msg("Database error checking campaign")
			tx.Rollback()
			continue
		}

		// Map to store ID mappings for cross-references
		idMap := make(map[string]string)

		// Create the campaign
		now := time.Now()
		campaign := model.Campaign{
			ID:               campaignData.ID,
			Title:            campaignData.Title,
			Description:      campaignData.Description,
			ImageURL:         campaignData.ImageURL,
			InitialChapterID: "", // Will be updated once we have the new chapter ID
			IsActive:         true,
			RequiredLevel:    campaignData.RequiredLevel,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := tx.Create(&campaign).Error; err != nil {
			r.logger.Error().Err(err).Str("id", campaignData.ID).Msg("Failed to create campaign")
			tx.Rollback()
			continue
		}

		// Create chapters
		for _, chapterData := range campaignData.Chapters {
			// Create a unique ID for the chapter
			uniqueChapterID := chapterData.ID + "_" + shortUUID()

			// Store the mapping
			idMap[chapterData.ID] = uniqueChapterID

			// Update campaign's initial chapter ID if needed
			if chapterData.ID == campaignData.InitialChapterID {
				campaign.InitialChapterID = uniqueChapterID
			}

			chapter := model.Chapter{
				ID:          uniqueChapterID,
				CampaignID:  campaignData.ID,
				Title:       chapterData.Title,
				Description: chapterData.Description,
				ImageURL:    chapterData.ImageURL,
				IsLocked:    chapterData.ID != campaignData.InitialChapterID,
				Order:       chapterData.Order,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if err := tx.Create(&chapter).Error; err != nil {
				r.logger.Error().Err(err).Str("id", uniqueChapterID).Msg("Failed to create chapter")
				tx.Rollback()
				continue
			}

			// Create missions
			for _, missionData := range chapterData.Missions {
				// Convert requirements
				missionRequirements := model.MissionRequirements{}
				if val, ok := missionData.Requirements["money"].(int); ok {
					missionRequirements.Money = val
				}
				if val, ok := missionData.Requirements["crew"].(int); ok {
					missionRequirements.Crew = val
				}
				if val, ok := missionData.Requirements["weapons"].(int); ok {
					missionRequirements.Weapons = val
				}
				if val, ok := missionData.Requirements["vehicles"].(int); ok {
					missionRequirements.Vehicles = val
				}
				if val, ok := missionData.Requirements["respect"].(int); ok {
					missionRequirements.Respect = val
				}
				if val, ok := missionData.Requirements["influence"].(int); ok {
					missionRequirements.Influence = val
				}
				if val, ok := missionData.Requirements["max_heat"].(int); ok {
					missionRequirements.MaxHeat = val
				}
				if val, ok := missionData.Requirements["min_title"].(string); ok {
					missionRequirements.MinTitle = val
				}
				if val, ok := missionData.Requirements["region_id"].(string); ok {
					missionRequirements.RegionID = val
				}
				if val, ok := missionData.Requirements["controlled_hotspots"].(int); ok {
					missionRequirements.ControlledHotspots = val
				}

				// Convert rewards
				missionRewards := model.MissionRewards{}
				if val, ok := missionData.Rewards["money"].(int); ok {
					missionRewards.Money = val
				}
				if val, ok := missionData.Rewards["crew"].(int); ok {
					missionRewards.Crew = val
				}
				if val, ok := missionData.Rewards["weapons"].(int); ok {
					missionRewards.Weapons = val
				}
				if val, ok := missionData.Rewards["vehicles"].(int); ok {
					missionRewards.Vehicles = val
				}
				if val, ok := missionData.Rewards["respect"].(int); ok {
					missionRewards.Respect = val
				}
				if val, ok := missionData.Rewards["influence"].(int); ok {
					missionRewards.Influence = val
				}
				if val, ok := missionData.Rewards["heat_reduction"].(int); ok {
					missionRewards.HeatReduction = val
				}
				if val, ok := missionData.Rewards["unlock_hotspot_id"].(string); ok {
					missionRewards.UnlockHotspotID = val
				}
				if val, ok := missionData.Rewards["unlock_chapter_id"].(string); ok {
					missionRewards.UnlockChapterID = val
				}
				if val, ok := missionData.Rewards["unlock_mission_id"].(string); ok {
					missionRewards.UnlockMissionID = val
				}

				// Create unique mission ID
				uniqueMissionID := missionData.ID + "_" + shortUUID()

				// Store the mapping
				idMap[missionData.ID] = uniqueMissionID

				mission := model.Mission{
					ID:           uniqueMissionID,
					ChapterID:    uniqueChapterID,
					Title:        missionData.Title,
					Description:  missionData.Description,
					Narrative:    missionData.Narrative,
					ImageURL:     missionData.ImageURL,
					MissionType:  missionData.Type,
					Requirements: missionRequirements,
					Rewards:      missionRewards,
					IsLocked:     missionData.Order > 1, // First mission is unlocked
					Order:        missionData.Order,
					CreatedAt:    now,
					UpdatedAt:    now,
				}

				if err := tx.Create(&mission).Error; err != nil {
					r.logger.Error().Err(err).Str("id", uniqueMissionID).Msg("Failed to create mission")
					tx.Rollback()
					continue
				}

				// Create choices
				for _, choiceData := range missionData.Choices {
					// Convert requirements
					choiceRequirements := model.MissionRequirements{}
					if val, ok := choiceData.Requirements["money"].(int); ok {
						choiceRequirements.Money = val
					}
					if val, ok := choiceData.Requirements["crew"].(int); ok {
						choiceRequirements.Crew = val
					}
					if val, ok := choiceData.Requirements["weapons"].(int); ok {
						choiceRequirements.Weapons = val
					}
					if val, ok := choiceData.Requirements["vehicles"].(int); ok {
						choiceRequirements.Vehicles = val
					}
					if val, ok := choiceData.Requirements["respect"].(int); ok {
						choiceRequirements.Respect = val
					}
					if val, ok := choiceData.Requirements["influence"].(int); ok {
						choiceRequirements.Influence = val
					}

					// Convert rewards
					choiceRewards := model.MissionRewards{}
					if val, ok := choiceData.Rewards["money"].(int); ok {
						choiceRewards.Money = val
					}
					if val, ok := choiceData.Rewards["crew"].(int); ok {
						choiceRewards.Crew = val
					}
					if val, ok := choiceData.Rewards["weapons"].(int); ok {
						choiceRewards.Weapons = val
					}
					if val, ok := choiceData.Rewards["vehicles"].(int); ok {
						choiceRewards.Vehicles = val
					}
					if val, ok := choiceData.Rewards["respect"].(int); ok {
						choiceRewards.Respect = val
					}
					if val, ok := choiceData.Rewards["influence"].(int); ok {
						choiceRewards.Influence = val
					}
					if val, ok := choiceData.Rewards["heat_reduction"].(int); ok {
						choiceRewards.HeatReduction = val
					}
					if val, ok := choiceData.Rewards["unlock_hotspot_id"].(string); ok {
						choiceRewards.UnlockHotspotID = val
					}

					// Create unique choice ID
					uniqueChoiceID := choiceData.ID + "_" + shortUUID()

					// Store the mapping and next mission reference
					idMap[choiceData.ID] = uniqueChoiceID
					if choiceData.NextMission != "" {
						idMap[choiceData.ID+"_nextMission"] = choiceData.NextMission
					}

					choice := model.MissionChoice{
						ID:              uniqueChoiceID,
						MissionID:       uniqueMissionID,
						Text:            choiceData.Text,
						NextMissionID:   "", // Will be updated later
						Requirements:    choiceRequirements,
						Rewards:         choiceRewards,
						SequentialOrder: choiceData.SequentialOrder,
						CreatedAt:       now,
						UpdatedAt:       now,
					}

					if err := tx.Create(&choice).Error; err != nil {
						r.logger.Error().Err(err).Str("id", uniqueChoiceID).Msg("Failed to create choice")
						tx.Rollback()
						continue
					}

					// Create condition templates
					for i, conditionData := range choiceData.Conditions {
						condition := model.ConditionTemplate{
							ID:              uuid.New().String(),
							ChoiceID:        uniqueChoiceID,
							Type:            conditionData.Type,
							RequiredValue:   conditionData.RequiredValue,
							AdditionalValue: conditionData.AdditionalValue,
							OrderIndex:      conditionData.OrderIndex,
							CreatedAt:       now,
							UpdatedAt:       now,
						}

						if err := tx.Create(&condition).Error; err != nil {
							r.logger.Error().Err(err).Int("index", i).Msg("Failed to create condition")
							tx.Rollback()
							continue
						}
					}

					// Create POI templates
					for i, poiData := range choiceData.POIs {
						poi := model.POITemplate{
							ID:           uuid.New().String(),
							Name:         poiData.Name,
							Description:  poiData.Description,
							LocationType: poiData.LocationType,
							LocationID:   poiData.LocationID,
							MissionID:    uniqueMissionID,
							ChoiceID:     uniqueChoiceID,
							CreatedAt:    now,
							UpdatedAt:    now,
						}

						if err := tx.Create(&poi).Error; err != nil {
							r.logger.Error().Err(err).Int("index", i).Msg("Failed to create POI template")
							tx.Rollback()
							continue
						}
					}

					// Create operation templates
					for i, opData := range choiceData.Operations {
						// Convert resources
						resources := model.OperationResources{}
						if val, ok := opData.Resources["crew"].(int); ok {
							resources.Crew = val
						}
						if val, ok := opData.Resources["weapons"].(int); ok {
							resources.Weapons = val
						}
						if val, ok := opData.Resources["vehicles"].(int); ok {
							resources.Vehicles = val
						}
						if val, ok := opData.Resources["money"].(int); ok {
							resources.Money = val
						}

						// Convert rewards
						rewards := model.OperationRewards{}
						if val, ok := opData.Rewards["money"].(int); ok {
							rewards.Money = val
						}
						if val, ok := opData.Rewards["crew"].(int); ok {
							rewards.Crew = val
						}
						if val, ok := opData.Rewards["weapons"].(int); ok {
							rewards.Weapons = val
						}
						if val, ok := opData.Rewards["vehicles"].(int); ok {
							rewards.Vehicles = val
						}
						if val, ok := opData.Rewards["respect"].(int); ok {
							rewards.Respect = val
						}
						if val, ok := opData.Rewards["influence"].(int); ok {
							rewards.Influence = val
						}
						if val, ok := opData.Rewards["heat_reduction"].(int); ok {
							rewards.HeatReduction = val
						}

						// Convert risks
						risks := model.OperationRisks{}
						if val, ok := opData.Risks["crew_loss"].(int); ok {
							risks.CrewLoss = val
						}
						if val, ok := opData.Risks["weapons_loss"].(int); ok {
							risks.WeaponsLoss = val
						}
						if val, ok := opData.Risks["vehicles_loss"].(int); ok {
							risks.VehiclesLoss = val
						}
						if val, ok := opData.Risks["money_loss"].(int); ok {
							risks.MoneyLoss = val
						}
						if val, ok := opData.Risks["heat_increase"].(int); ok {
							risks.HeatIncrease = val
						}
						if val, ok := opData.Risks["respect_loss"].(int); ok {
							risks.RespectLoss = val
						}

						op := model.OperationTemplate{
							ID:            uuid.New().String(),
							Name:          opData.Name,
							Description:   opData.Description,
							OperationType: opData.OperationType,
							MissionID:     uniqueMissionID,
							ChoiceID:      uniqueChoiceID,
							Resources:     resources,
							Rewards:       rewards,
							Risks:         risks,
							Duration:      opData.Duration,
							SuccessRate:   opData.SuccessRate,
							CreatedAt:     now,
							UpdatedAt:     now,
						}

						if err := tx.Create(&op).Error; err != nil {
							r.logger.Error().Err(err).Int("index", i).Msg("Failed to create operation template")
							tx.Rollback()
							continue
						}
					}
				}
			}
		}

		// Update campaign with initial chapter ID
		if campaign.InitialChapterID != "" {
			if err := tx.Model(&campaign).Update("initial_chapter_id", campaign.InitialChapterID).Error; err != nil {
				r.logger.Error().Err(err).Str("id", campaign.ID).Msg("Failed to update campaign initial chapter ID")
				tx.Rollback()
				continue
			}
		}

		// Update NextMissionID references for all choices
		for choiceOriginalID, uniqueChoiceID := range idMap {
			nextMissionOriginalID, exists := idMap[choiceOriginalID+"_nextMission"]
			if !exists {
				continue
			}

			nextMissionUniqueID, exists := idMap[nextMissionOriginalID]
			if !exists {
				continue
			}

			if err := tx.Model(&model.MissionChoice{}).Where("id = ?", uniqueChoiceID).Update("next_mission_id", nextMissionUniqueID).Error; err != nil {
				r.logger.Error().Err(err).Str("id", uniqueChoiceID).Msg("Failed to update NextMissionID")
			}
		}

		// Update UnlockChapterID and UnlockMissionID references
		// TODO: Add logic to update these references from Rewards if needed

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			r.logger.Error().Err(err).Str("id", campaignData.ID).Msg("Failed to commit transaction")
			continue
		}

		r.logger.Info().Str("id", campaignData.ID).Msg("Campaign loaded successfully")
	}

	return nil
}

// shortUUID generates a short unique ID suffix
func shortUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
}
