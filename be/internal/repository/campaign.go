// internal/repository/campaign.go

package repository

import (
	"errors"
	"fmt"
	"io/ioutil"
	"mwce-be/internal/model"
	"mwce-be/pkg/database"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// CampaignDefinition represents the YAML structure for campaign data
type CampaignDefinition struct {
	ID             string              `yaml:"id"`
	Title          string              `yaml:"title"`
	Description    string              `yaml:"description"`
	ImageURL       string              `yaml:"image_url"`
	InitialChapter string              `yaml:"initial_chapter"`
	RequiredLevel  int                 `yaml:"required_level"`
	Chapters       []ChapterDefinition `yaml:"chapters"`
}

// ChapterDefinition represents the YAML structure for chapter data
type ChapterDefinition struct {
	ID          string              `yaml:"id"`
	Title       string              `yaml:"title"`
	Description string              `yaml:"description"`
	ImageURL    string              `yaml:"image_url"`
	Order       int                 `yaml:"order"`
	Missions    []MissionDefinition `yaml:"missions"`
}

// MissionDefinition represents the YAML structure for mission data
type MissionDefinition struct {
	ID           string                    `yaml:"id"`
	Title        string                    `yaml:"title"`
	Description  string                    `yaml:"description"`
	Narrative    string                    `yaml:"narrative"`
	ImageURL     string                    `yaml:"image_url"`
	Type         string                    `yaml:"type"`
	Order        int                       `yaml:"order"`
	Requirements map[string]interface{}    `yaml:"requirements"`
	Rewards      map[string]interface{}    `yaml:"rewards"`
	Choices      []MissionChoiceDefinition `yaml:"choices"`
	Next         string                    `yaml:"next"`
}

// MissionChoiceDefinition represents the YAML structure for mission choice data
type MissionChoiceDefinition struct {
	ID              string                          `yaml:"id"`
	Text            string                          `yaml:"text"`
	NextMission     string                          `yaml:"next_mission"`
	Requirements    map[string]interface{}          `yaml:"requirements"`
	Rewards         map[string]interface{}          `yaml:"rewards"`
	SequentialOrder bool                            `yaml:"sequential_order"`
	Conditions      []CompletionConditionDefinition `yaml:"conditions"`
	POIs            []POIDefinition                 `yaml:"pois"`
	Operations      []MissionOperationDefinition    `yaml:"operations"`
}

type CompletionConditionDefinition struct {
	Type            string `yaml:"type"`
	RequiredValue   string `yaml:"required_value"`
	AdditionalValue string `yaml:"additional_value"`
	OrderIndex      int    `yaml:"order_index"`
}

type POIDefinition struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	LocationType string `yaml:"location_type"`
	LocationID   string `yaml:"location_id"`
}

type MissionOperationDefinition struct {
	ID            string                 `yaml:"id"`
	Name          string                 `yaml:"name"`
	Description   string                 `yaml:"description"`
	OperationType string                 `yaml:"operation_type"`
	Resources     map[string]interface{} `yaml:"resources"`
	Rewards       map[string]interface{} `yaml:"rewards"`
	Risks         map[string]interface{} `yaml:"risks"`
	Duration      int                    `yaml:"duration"`
	SuccessRate   int                    `yaml:"success_rate"`
}

// CampaignRepository handles database operations for campaigns
type CampaignRepository interface {
	GetDB() *gorm.DB
	GetAllCampaigns() ([]model.Campaign, error)
	GetCampaignByID(id string) (*model.Campaign, error)
	GetChaptersByID(campaignID string) ([]model.Chapter, error)
	GetChapterByID(id string) (*model.Chapter, error)
	GetMissionsByChapterID(chapterID string) ([]model.Mission, error)
	GetMissionByID(id string) (*model.Mission, error)
	GetChoicesByMissionID(missionID string) ([]model.MissionChoice, error)
	GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error)
	GetPlayerCampaignProgresses(playerID string) ([]model.PlayerCampaignProgress, error)
	GetPlayerMissionProgress(playerID string, missionID string) (*model.PlayerMissionProgress, error)
	SavePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error
	SavePlayerMissionProgress(progress *model.PlayerMissionProgress) error
	LoadCampaignsFromYAML(dirPath string) error

	// POI methods
	GetPOIs() ([]model.POI, error)
	GetPOIByID(id string) (*model.POI, error)
	GetPOIsByMission(missionID string) ([]model.POI, error)
	GetPOIsByChoice(choiceID string) ([]model.POI, error)
	GetActivePlayerPOIs(playerID string) ([]model.POI, error)
	ActivatePOI(poiID string, playerID string) error
	CompletePOI(poiID string, playerID string) error
	SavePOI(poi *model.POI) error

	// MissionOperation methods
	GetMissionOperations() ([]model.MissionOperation, error)
	GetMissionOperationByID(id string) (*model.MissionOperation, error)
	GetMissionOperationsByMission(missionID string) ([]model.MissionOperation, error)
	GetMissionOperationsByChoice(choiceID string) ([]model.MissionOperation, error)
	GetActivePlayerMissionOperations(playerID string) ([]model.MissionOperation, error)
	ActivateMissionOperation(operationID string, playerID string) error
	CompleteMissionOperation(operationID string, playerID string) error
	SaveMissionOperation(operation *model.MissionOperation) error

	// CompletionCondition methods
	GetCompletionConditions(choiceID string) ([]model.CompletionCondition, error)
	GetPlayerCompletionConditions(playerID string, choiceID string) ([]model.CompletionCondition, error)
	CompleteCondition(conditionID string, playerID string) error
	SaveCompletionCondition(condition *model.CompletionCondition) error
}

type campaignRepository struct {
	db database.Database
}

// NewCampaignRepository creates a new campaign repository
func NewCampaignRepository(db database.Database) CampaignRepository {
	return &campaignRepository{
		db: db,
	}
}

// GetDB returns the database connection
func (r *campaignRepository) GetDB() *gorm.DB {
	return r.db.GetDB()
}

// GetAllCampaigns retrieves all campaigns
func (r *campaignRepository) GetAllCampaigns() ([]model.Campaign, error) {
	var campaigns []model.Campaign
	if err := r.db.GetDB().Find(&campaigns).Error; err != nil {
		return nil, err
	}
	return campaigns, nil
}

// GetCampaignByID retrieves a campaign by its ID
func (r *campaignRepository) GetCampaignByID(id string) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := r.db.GetDB().Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("chapters.order ASC")
	}).Where("id = ?", id).First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("campaign not found")
		}
		return nil, err
	}
	return &campaign, nil
}

// GetChaptersByID retrieves chapters for a campaign
func (r *campaignRepository) GetChaptersByID(campaignID string) ([]model.Chapter, error) {
	var chapters []model.Chapter
	if err := r.db.GetDB().Where("campaign_id = ?", campaignID).Order("order ASC").Find(&chapters).Error; err != nil {
		return nil, err
	}
	return chapters, nil
}

// GetChapterByID retrieves a chapter by its ID
func (r *campaignRepository) GetChapterByID(id string) (*model.Chapter, error) {
	var chapter model.Chapter
	if err := r.db.GetDB().Preload("Missions", func(db *gorm.DB) *gorm.DB {
		return db.Order("missions.order ASC")
	}).Where("id = ?", id).First(&chapter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chapter not found")
		}
		return nil, err
	}
	return &chapter, nil
}

// GetMissionsByChapterID retrieves missions for a chapter
func (r *campaignRepository) GetMissionsByChapterID(chapterID string) ([]model.Mission, error) {
	var missions []model.Mission
	if err := r.db.GetDB().Where("chapter_id = ?", chapterID).Order("order ASC").Find(&missions).Error; err != nil {
		return nil, err
	}
	return missions, nil
}

// GetMissionByID retrieves a mission by its ID
func (r *campaignRepository) GetMissionByID(id string) (*model.Mission, error) {
	var mission model.Mission
	if err := r.db.GetDB().Preload("Choices").Where("id = ?", id).First(&mission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mission not found")
		}
		return nil, err
	}
	return &mission, nil
}

// GetChoicesByMissionID retrieves choices for a mission
func (r *campaignRepository) GetChoicesByMissionID(missionID string) ([]model.MissionChoice, error) {
	var choices []model.MissionChoice
	if err := r.db.GetDB().Where("mission_id = ?", missionID).Find(&choices).Error; err != nil {
		return nil, err
	}
	return choices, nil
}

// GetPlayerCampaignProgress retrieves a player's progress for a campaign
func (r *campaignRepository) GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error) {
	var progress model.PlayerCampaignProgress
	if err := r.db.GetDB().Where("player_id = ? AND campaign_id = ?", playerID, campaignID).First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No progress found, but not an error
		}
		return nil, err
	}
	return &progress, nil
}

// GetPlayerCampaignProgresses retrieves all campaign progress for a player
func (r *campaignRepository) GetPlayerCampaignProgresses(playerID string) ([]model.PlayerCampaignProgress, error) {
	var progresses []model.PlayerCampaignProgress
	if err := r.db.GetDB().Where("player_id = ?", playerID).Find(&progresses).Error; err != nil {
		return nil, err
	}
	return progresses, nil
}

// GetPlayerMissionProgress retrieves a player's progress for a mission
func (r *campaignRepository) GetPlayerMissionProgress(playerID string, missionID string) (*model.PlayerMissionProgress, error) {
	var progress model.PlayerMissionProgress
	if err := r.db.GetDB().Where("player_id = ? AND mission_id = ?", playerID, missionID).First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No progress found, but not an error
		}
		return nil, err
	}
	return &progress, nil
}

// SavePlayerCampaignProgress saves a player's campaign progress
func (r *campaignRepository) SavePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error {
	return r.db.GetDB().Save(progress).Error
}

// SavePlayerMissionProgress saves a player's mission progress
func (r *campaignRepository) SavePlayerMissionProgress(progress *model.PlayerMissionProgress) error {
	return r.db.GetDB().Save(progress).Error
}

// LoadCampaignsFromYAML loads campaign data from YAML files in a directory
func (r *campaignRepository) LoadCampaignsFromYAML(dirPath string) error {
	// Get list of YAML files in directory
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read campaign directory: %w", err)
	}

	// Process each YAML file
	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml") {
			// Read the file
			filePath := filepath.Join(dirPath, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read campaign file %s: %w", filePath, err)
			}

			// Unmarshal the YAML data
			var campaignDef CampaignDefinition
			if err := yaml.Unmarshal(data, &campaignDef); err != nil {
				return fmt.Errorf("failed to parse campaign YAML %s: %w", filePath, err)
			}

			// Convert to model and save
			if err := r.saveCampaignFromDefinition(campaignDef); err != nil {
				return fmt.Errorf("failed to save campaign %s: %w", campaignDef.Title, err)
			}
		}
	}

	return nil
}

// Helper function to convert YAML definition to model
func (r *campaignRepository) saveCampaignFromDefinition(def CampaignDefinition) error {
	// Start a transaction
	tx := r.db.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create campaign
	campaign := model.Campaign{
		ID:               def.ID, // Use the ID from YAML
		Title:            def.Title,
		Description:      def.Description,
		ImageURL:         def.ImageURL,
		InitialChapterID: def.InitialChapter, // Use the ID from YAML
		IsActive:         true,
		RequiredLevel:    def.RequiredLevel,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save campaign
	if err := tx.Save(&campaign).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create chapters
	for _, chapterDef := range def.Chapters {
		chapter := model.Chapter{
			ID:          chapterDef.ID, // Use the ID from YAML
			CampaignID:  campaign.ID,
			Title:       chapterDef.Title,
			Description: chapterDef.Description,
			ImageURL:    chapterDef.ImageURL,
			IsLocked:    chapterDef.ID != def.InitialChapter, // Lock all chapters except initial
			Order:       chapterDef.Order,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save chapter
		if err := tx.Save(&chapter).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Create missions for this chapter
		for _, missionDef := range chapterDef.Missions {
			// Convert requirements and rewards from maps to structs
			reqStruct := parseRequirements(missionDef.Requirements)
			rewardStruct := parseRewards(missionDef.Rewards)

			mission := model.Mission{
				ID:           missionDef.ID, // Use the ID from YAML
				ChapterID:    chapter.ID,
				Title:        missionDef.Title,
				Description:  missionDef.Description,
				Narrative:    missionDef.Narrative,
				ImageURL:     missionDef.ImageURL,
				MissionType:  missionDef.Type,
				Requirements: reqStruct,
				Rewards:      rewardStruct,
				IsLocked:     true, // All missions start locked except first in chapter
				Order:        missionDef.Order,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// Save mission
			if err := tx.Save(&mission).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Create choices for this mission
			for _, choiceDef := range missionDef.Choices {
				choiceReqStruct := parseRequirements(choiceDef.Requirements)
				choiceRewardStruct := parseRewards(choiceDef.Rewards)

				choice := model.MissionChoice{
					ID:            choiceDef.ID, // Use the ID from YAML
					MissionID:     mission.ID,
					Text:          choiceDef.Text,
					NextMissionID: choiceDef.NextMission,
					Requirements:  choiceReqStruct,
					Rewards:       choiceRewardStruct,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}

				// Save choice's sequential order property
				choice.SequentialOrder = choiceDef.SequentialOrder

				// Process conditions
				for _, conditionDef := range choiceDef.Conditions {
					condition := model.CompletionCondition{
						ChoiceID:        choice.ID,
						Type:            conditionDef.Type,
						RequiredValue:   conditionDef.RequiredValue,
						AdditionalValue: conditionDef.AdditionalValue,
						OrderIndex:      conditionDef.OrderIndex,
						CreatedAt:       time.Now(),
						UpdatedAt:       time.Now(),
					}

					if err := tx.Save(&condition).Error; err != nil {
						tx.Rollback()
						return err
					}
				}

				// Process POIs
				for _, poiDef := range choiceDef.POIs {
					poi := model.POI{
						ID:           poiDef.ID,
						Name:         poiDef.Name,
						Description:  poiDef.Description,
						LocationType: poiDef.LocationType,
						LocationID:   poiDef.LocationID,
						MissionID:    mission.ID,
						ChoiceID:     choice.ID,
						IsActive:     false,
						IsCompleted:  false,
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}

					if err := tx.Save(&poi).Error; err != nil {
						tx.Rollback()
						return err
					}
				}

				// Process mission operations
				for _, opDef := range choiceDef.Operations {
					// Convert resources, rewards, risks from maps to structs
					resources := parseOperationResources(opDef.Resources)
					rewards := parseOperationRewards(opDef.Rewards)
					risks := parseOperationRisks(opDef.Risks)

					operation := model.MissionOperation{
						ID:            opDef.ID,
						Name:          opDef.Name,
						Description:   opDef.Description,
						OperationType: opDef.OperationType,
						MissionID:     mission.ID,
						ChoiceID:      choice.ID,
						Resources:     resources,
						Rewards:       rewards,
						Risks:         risks,
						Duration:      opDef.Duration,
						SuccessRate:   opDef.SuccessRate,
						IsActive:      false,
						IsCompleted:   false,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					}

					if err := tx.Save(&operation).Error; err != nil {
						tx.Rollback()
						return err
					}
				}

				// Save choice
				if err := tx.Save(&choice).Error; err != nil {
					tx.Rollback()
					return err
				}
			}

		}
	}

	// Commit transaction
	return tx.Commit().Error
}

func parseOperationResources(resourceMap map[string]interface{}) model.OperationResources {
	resources := model.OperationResources{}

	for key, value := range resourceMap {
		switch key {
		case "crew":
			if v, ok := value.(int); ok {
				resources.Crew = v
			}
		case "weapons":
			if v, ok := value.(int); ok {
				resources.Weapons = v
			}
		case "vehicles":
			if v, ok := value.(int); ok {
				resources.Vehicles = v
			}
		case "money":
			if v, ok := value.(int); ok {
				resources.Money = v
			}
		}
	}

	return resources
}

func parseOperationRewards(rewardMap map[string]interface{}) model.OperationRewards {
	rewards := model.OperationRewards{}

	for key, value := range rewardMap {
		switch key {
		case "money":
			if v, ok := value.(int); ok {
				rewards.Money = v
			}
		case "crew":
			if v, ok := value.(int); ok {
				rewards.Crew = v
			}
		case "weapons":
			if v, ok := value.(int); ok {
				rewards.Weapons = v
			}
		case "vehicles":
			if v, ok := value.(int); ok {
				rewards.Vehicles = v
			}
		case "respect":
			if v, ok := value.(int); ok {
				rewards.Respect = v
			}
		case "influence":
			if v, ok := value.(int); ok {
				rewards.Influence = v
			}
		case "heat_reduction":
			if v, ok := value.(int); ok {
				rewards.HeatReduction = v
			}
		}
	}

	return rewards
}

func parseOperationRisks(riskMap map[string]interface{}) model.OperationRisks {
	risks := model.OperationRisks{}

	for key, value := range riskMap {
		switch key {
		case "crew_loss":
			if v, ok := value.(int); ok {
				risks.CrewLoss = v
			}
		case "weapons_loss":
			if v, ok := value.(int); ok {
				risks.WeaponsLoss = v
			}
		case "vehicles_loss":
			if v, ok := value.(int); ok {
				risks.VehiclesLoss = v
			}
		case "money_loss":
			if v, ok := value.(int); ok {
				risks.MoneyLoss = v
			}
		case "heat_increase":
			if v, ok := value.(int); ok {
				risks.HeatIncrease = v
			}
		case "respect_loss":
			if v, ok := value.(int); ok {
				risks.RespectLoss = v
			}
		}
	}

	return risks
}

// POI Repository Methods

func (r *campaignRepository) GetPOIs() ([]model.POI, error) {
	var pois []model.POI
	if err := r.db.GetDB().Find(&pois).Error; err != nil {
		return nil, err
	}
	return pois, nil
}

func (r *campaignRepository) GetPOIByID(id string) (*model.POI, error) {
	var poi model.POI
	if err := r.db.GetDB().Where("id = ?", id).First(&poi).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("POI not found")
		}
		return nil, err
	}
	return &poi, nil
}

func (r *campaignRepository) GetPOIsByMission(missionID string) ([]model.POI, error) {
	var pois []model.POI
	if err := r.db.GetDB().Where("mission_id = ?", missionID).Find(&pois).Error; err != nil {
		return nil, err
	}
	return pois, nil
}

func (r *campaignRepository) GetPOIsByChoice(choiceID string) ([]model.POI, error) {
	var pois []model.POI
	if err := r.db.GetDB().Where("choice_id = ?", choiceID).Find(&pois).Error; err != nil {
		return nil, err
	}
	return pois, nil
}

func (r *campaignRepository) GetActivePlayerPOIs(playerID string) ([]model.POI, error) {
	var pois []model.POI
	if err := r.db.GetDB().Where("player_id = ? AND is_active = ? AND is_completed = ?",
		playerID, true, false).Find(&pois).Error; err != nil {
		return nil, err
	}
	return pois, nil
}

func (r *campaignRepository) ActivatePOI(poiID string, playerID string) error {
	return r.db.GetDB().Model(&model.POI{}).
		Where("id = ?", poiID).
		Updates(map[string]interface{}{
			"is_active": true,
			"player_id": playerID,
		}).Error
}

func (r *campaignRepository) CompletePOI(poiID string, playerID string) error {
	now := time.Now()
	return r.db.GetDB().Model(&model.POI{}).
		Where("id = ? AND player_id = ?", poiID, playerID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": now,
		}).Error
}

func (r *campaignRepository) SavePOI(poi *model.POI) error {
	return r.db.GetDB().Save(poi).Error
}

// MissionOperation Repository Methods
func (r *campaignRepository) GetMissionOperations() ([]model.MissionOperation, error) {
	var operations []model.MissionOperation
	if err := r.db.GetDB().Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

func (r *campaignRepository) GetMissionOperationByID(id string) (*model.MissionOperation, error) {
	var operation model.MissionOperation
	if err := r.db.GetDB().Where("id = ?", id).First(&operation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mission operation not found")
		}
		return nil, err
	}
	return &operation, nil
}

func (r *campaignRepository) GetMissionOperationsByMission(missionID string) ([]model.MissionOperation, error) {
	var operations []model.MissionOperation
	if err := r.db.GetDB().Where("mission_id = ?", missionID).Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

func (r *campaignRepository) GetMissionOperationsByChoice(choiceID string) ([]model.MissionOperation, error) {
	var operations []model.MissionOperation
	if err := r.db.GetDB().Where("choice_id = ?", choiceID).Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

func (r *campaignRepository) GetActivePlayerMissionOperations(playerID string) ([]model.MissionOperation, error) {
	var operations []model.MissionOperation
	if err := r.db.GetDB().Where("player_id = ? AND is_active = ? AND is_completed = ?",
		playerID, true, false).Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

func (r *campaignRepository) ActivateMissionOperation(operationID string, playerID string) error {
	return r.db.GetDB().Model(&model.MissionOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]interface{}{
			"is_active": true,
			"player_id": playerID,
		}).Error
}

func (r *campaignRepository) CompleteMissionOperation(operationID string, playerID string) error {
	now := time.Now()
	return r.db.GetDB().Model(&model.MissionOperation{}).
		Where("id = ? AND player_id = ?", operationID, playerID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": now,
		}).Error
}

func (r *campaignRepository) SaveMissionOperation(operation *model.MissionOperation) error {
	return r.db.GetDB().Save(operation).Error
}

// CompletionCondition Repository Methods
func (r *campaignRepository) GetCompletionConditions(choiceID string) ([]model.CompletionCondition, error) {
	var conditions []model.CompletionCondition
	if err := r.db.GetDB().Where("choice_id = ?", choiceID).Order("order_index ASC").Find(&conditions).Error; err != nil {
		return nil, err
	}
	return conditions, nil
}

func (r *campaignRepository) GetPlayerCompletionConditions(playerID string, choiceID string) ([]model.CompletionCondition, error) {
	var conditions []model.CompletionCondition
	if err := r.db.GetDB().Where("player_id = ? AND choice_id = ?", playerID, choiceID).
		Order("order_index ASC").Find(&conditions).Error; err != nil {
		return nil, err
	}
	return conditions, nil
}

func (r *campaignRepository) CompleteCondition(conditionID string, playerID string) error {
	now := time.Now()
	return r.db.GetDB().Model(&model.CompletionCondition{}).
		Where("id = ? AND player_id = ?", conditionID, playerID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": now,
		}).Error
}

func (r *campaignRepository) SaveCompletionCondition(condition *model.CompletionCondition) error {
	return r.db.GetDB().Save(condition).Error
}

// Helper functions to parse YAML maps into structured types
func parseRequirements(reqMap map[string]interface{}) model.MissionRequirements {
	req := model.MissionRequirements{}

	for key, value := range reqMap {
		switch key {
		case "money":
			if v, ok := value.(int); ok {
				req.Money = v
			}
		case "crew":
			if v, ok := value.(int); ok {
				req.Crew = v
			}
		case "weapons":
			if v, ok := value.(int); ok {
				req.Weapons = v
			}
		case "vehicles":
			if v, ok := value.(int); ok {
				req.Vehicles = v
			}
		case "respect":
			if v, ok := value.(int); ok {
				req.Respect = v
			}
		case "influence":
			if v, ok := value.(int); ok {
				req.Influence = v
			}
		case "max_heat":
			if v, ok := value.(int); ok {
				req.MaxHeat = v
			}
		case "min_title":
			if v, ok := value.(string); ok {
				req.MinTitle = v
			}
		case "operation_type":
			if v, ok := value.(string); ok {
				req.OperationType = v
			}
		case "operation_id":
			if v, ok := value.(string); ok {
				req.OperationID = v
			}
		case "hotspot_id":
			if v, ok := value.(string); ok {
				req.HotspotID = v
			}
		case "region_id":
			if v, ok := value.(string); ok {
				req.RegionID = v
			}
		case "controlled_hotspots":
			if v, ok := value.(int); ok {
				req.ControlledHotspots = v
			}
		}
	}

	return req
}

func parseRewards(rewardMap map[string]interface{}) model.MissionRewards {
	reward := model.MissionRewards{}

	for key, value := range rewardMap {
		switch key {
		case "money":
			if v, ok := value.(int); ok {
				reward.Money = v
			}
		case "crew":
			if v, ok := value.(int); ok {
				reward.Crew = v
			}
		case "weapons":
			if v, ok := value.(int); ok {
				reward.Weapons = v
			}
		case "vehicles":
			if v, ok := value.(int); ok {
				reward.Vehicles = v
			}
		case "respect":
			if v, ok := value.(int); ok {
				reward.Respect = v
			}
		case "influence":
			if v, ok := value.(int); ok {
				reward.Influence = v
			}
		case "heat_reduction":
			if v, ok := value.(int); ok {
				reward.HeatReduction = v
			}
		case "unlock_hotspot_id":
			if v, ok := value.(string); ok {
				reward.UnlockHotspotID = v
			}
		case "unlock_chapter_id":
			if v, ok := value.(string); ok {
				reward.UnlockChapterID = v
			}
		case "unlock_mission_id":
			if v, ok := value.(string); ok {
				reward.UnlockMissionID = v
			}
		}
	}

	return reward
}
