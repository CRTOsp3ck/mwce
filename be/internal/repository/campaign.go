// be/internal/repository/campaign.go

package repository

import (
	"errors"
	"mwce-be/internal/model"
	"mwce-be/pkg/database"

	"gorm.io/gorm"
)

// CampaignRepository handles database operations for campaigns
type CampaignRepository interface {
	GetDB() *gorm.DB
	GetAllCampaigns() ([]model.Campaign, error)
	GetCampaignByID(id string) (*model.Campaign, error)
	GetChaptersByCampaignID(campaignID string) ([]model.Chapter, error)
	GetChapterByID(id string) (*model.Chapter, error)
	GetMissionsByChapterID(chapterID string) ([]model.Mission, error)
	GetMissionByID(id string) (*model.Mission, error)
	GetBranchesByMissionID(missionID string) ([]model.Branch, error)
	GetBranchByID(id string) (*model.Branch, error)
	GetOperationsByBranchID(branchID string) ([]model.CampaignOperation, error)
	GetOperationByID(id string) (*model.CampaignOperation, error)
	GetPOIsByBranchID(branchID string) ([]model.CampaignPOI, error)
	GetPOIByID(id string) (*model.CampaignPOI, error)
	GetDialoguesByPOIID(poiID string) ([]model.Dialogue, error)

	// Player Progress
	GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error)
	GetAllPlayerCampaignProgress(playerID string) ([]model.PlayerCampaignProgress, error)
	CreatePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error
	UpdatePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error

	// Player Records
	GetPlayerOperationRecords(progressID string) ([]model.PlayerOperationRecord, error)
	GetPlayerOperationRecordByIDs(progressID, operationID string) (*model.PlayerOperationRecord, error)
	CreatePlayerOperationRecord(record *model.PlayerOperationRecord) error
	UpdatePlayerOperationRecord(record *model.PlayerOperationRecord) error

	GetPlayerPOIRecords(progressID string) ([]model.PlayerPOIRecord, error)
	GetPlayerPOIRecordByIDs(progressID, poiID string) (*model.PlayerPOIRecord, error)
	CreatePlayerPOIRecord(record *model.PlayerPOIRecord) error
	UpdatePlayerPOIRecord(record *model.PlayerPOIRecord) error

	GetDialogueStateByIDs(recordID, dialogueID string) (*model.DialogueState, error)
	CreateDialogueState(state *model.DialogueState) error
	UpdateDialogueState(state *model.DialogueState) error

	// Campaign Data Management
	CreateCampaign(campaign *model.Campaign) error
	CreateChapter(chapter *model.Chapter) error
	CreateMission(mission *model.Mission) error
	CreateBranch(branch *model.Branch) error
	CreateCampaignOperation(operation *model.CampaignOperation) error
	CreateCampaignPOI(poi *model.CampaignPOI) error
	CreateDialogue(dialogue *model.Dialogue) error
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

// GetDB returns the database connection instance
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

// GetCampaignByID retrieves a campaign by ID
func (r *campaignRepository) GetCampaignByID(id string) (*model.Campaign, error) {
	var campaign model.Campaign
	if err := r.db.GetDB().
		Preload("Chapters", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("id = ?", id).
		First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("campaign not found")
		}
		return nil, err
	}
	return &campaign, nil
}

// GetChaptersByCampaignID retrieves chapters by campaign ID
func (r *campaignRepository) GetChaptersByCampaignID(campaignID string) ([]model.Chapter, error) {
	var chapters []model.Chapter
	if err := r.db.GetDB().
		Where("campaign_id = ?", campaignID).
		Order("\"order\" ASC").
		Find(&chapters).Error; err != nil {
		return nil, err
	}
	return chapters, nil
}

// GetChapterByID retrieves a chapter by ID
func (r *campaignRepository) GetChapterByID(id string) (*model.Chapter, error) {
	var chapter model.Chapter
	if err := r.db.GetDB().
		Preload("Missions", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("id = ?", id).
		First(&chapter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chapter not found")
		}
		return nil, err
	}
	return &chapter, nil
}

// GetMissionsByChapterID retrieves missions by chapter ID
func (r *campaignRepository) GetMissionsByChapterID(chapterID string) ([]model.Mission, error) {
	var missions []model.Mission
	if err := r.db.GetDB().
		Where("chapter_id = ?", chapterID).
		Order("\"order\" ASC").
		Find(&missions).Error; err != nil {
		return nil, err
	}
	return missions, nil
}

// GetMissionByID retrieves a mission by ID
func (r *campaignRepository) GetMissionByID(id string) (*model.Mission, error) {
	var mission model.Mission
	if err := r.db.GetDB().
		Preload("Branches").
		Where("id = ?", id).
		First(&mission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mission not found")
		}
		return nil, err
	}
	return &mission, nil
}

// GetBranchesByMissionID retrieves branches by mission ID
func (r *campaignRepository) GetBranchesByMissionID(missionID string) ([]model.Branch, error) {
	var branches []model.Branch
	if err := r.db.GetDB().
		Where("mission_id = ?", missionID).
		Find(&branches).Error; err != nil {
		return nil, err
	}
	return branches, nil
}

// GetBranchByID retrieves a branch by ID
func (r *campaignRepository) GetBranchByID(id string) (*model.Branch, error) {
	var branch model.Branch
	if err := r.db.GetDB().
		Preload("Operations").
		Preload("POIs").
		Where("id = ?", id).
		First(&branch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("branch not found")
		}
		return nil, err
	}
	return &branch, nil
}

// GetOperationsByBranchID retrieves operations by branch ID
func (r *campaignRepository) GetOperationsByBranchID(branchID string) ([]model.CampaignOperation, error) {
	var operations []model.CampaignOperation
	if err := r.db.GetDB().
		Where("branch_id = ?", branchID).
		Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

// GetOperationByID retrieves an operation by ID
func (r *campaignRepository) GetOperationByID(id string) (*model.CampaignOperation, error) {
	var operation model.CampaignOperation
	if err := r.db.GetDB().
		Where("id = ?", id).
		First(&operation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("operation not found")
		}
		return nil, err
	}
	return &operation, nil
}

// GetPOIsByBranchID retrieves POIs by branch ID
func (r *campaignRepository) GetPOIsByBranchID(branchID string) ([]model.CampaignPOI, error) {
	var pois []model.CampaignPOI
	if err := r.db.GetDB().
		Where("branch_id = ?", branchID).
		Find(&pois).Error; err != nil {
		return nil, err
	}
	return pois, nil
}

// GetPOIByID retrieves a POI by ID
func (r *campaignRepository) GetPOIByID(id string) (*model.CampaignPOI, error) {
	var poi model.CampaignPOI
	if err := r.db.GetDB().
		Preload("Dialogues", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("id = ?", id).
		First(&poi).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("POI not found")
		}
		return nil, err
	}
	return &poi, nil
}

// GetDialoguesByPOIID retrieves dialogues by POI ID
func (r *campaignRepository) GetDialoguesByPOIID(poiID string) ([]model.Dialogue, error) {
	var dialogues []model.Dialogue
	if err := r.db.GetDB().
		Where("poi_id = ?", poiID).
		Order("\"order\" ASC").
		Find(&dialogues).Error; err != nil {
		return nil, err
	}
	return dialogues, nil
}

// GetPlayerCampaignProgress retrieves a player's campaign progress
func (r *campaignRepository) GetPlayerCampaignProgress(playerID string, campaignID string) (*model.PlayerCampaignProgress, error) {
	var progress model.PlayerCampaignProgress
	if err := r.db.GetDB().
		Where("player_id = ? AND campaign_id = ?", playerID, campaignID).
		First(&progress).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error to indicate no progress found
		}
		return nil, err
	}
	return &progress, nil
}

// GetAllPlayerCampaignProgress retrieves all campaign progress for a player
func (r *campaignRepository) GetAllPlayerCampaignProgress(playerID string) ([]model.PlayerCampaignProgress, error) {
	var progresses []model.PlayerCampaignProgress
	if err := r.db.GetDB().
		Where("player_id = ?", playerID).
		Find(&progresses).Error; err != nil {
		return nil, err
	}
	return progresses, nil
}

// CreatePlayerCampaignProgress creates a new player campaign progress
func (r *campaignRepository) CreatePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error {
	return r.db.GetDB().Create(progress).Error
}

// UpdatePlayerCampaignProgress updates a player's campaign progress
func (r *campaignRepository) UpdatePlayerCampaignProgress(progress *model.PlayerCampaignProgress) error {
	return r.db.GetDB().Save(progress).Error
}

// GetPlayerOperationRecords retrieves a player's operation records for a campaign
func (r *campaignRepository) GetPlayerOperationRecords(progressID string) ([]model.PlayerOperationRecord, error) {
	var records []model.PlayerOperationRecord
	if err := r.db.GetDB().
		Where("progress_id = ?", progressID).
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// GetPlayerOperationRecordByIDs retrieves a player's operation record by progress and operation IDs
func (r *campaignRepository) GetPlayerOperationRecordByIDs(progressID, operationID string) (*model.PlayerOperationRecord, error) {
	var record model.PlayerOperationRecord
	if err := r.db.GetDB().
		Where("progress_id = ? AND operation_id = ?", progressID, operationID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error to indicate no record found
		}
		return nil, err
	}
	return &record, nil
}

// CreatePlayerOperationRecord creates a new player operation record
func (r *campaignRepository) CreatePlayerOperationRecord(record *model.PlayerOperationRecord) error {
	return r.db.GetDB().Create(record).Error
}

// UpdatePlayerOperationRecord updates a player's operation record
func (r *campaignRepository) UpdatePlayerOperationRecord(record *model.PlayerOperationRecord) error {
	return r.db.GetDB().Save(record).Error
}

// GetPlayerPOIRecords retrieves a player's POI records for a campaign
func (r *campaignRepository) GetPlayerPOIRecords(progressID string) ([]model.PlayerPOIRecord, error) {
	var records []model.PlayerPOIRecord
	if err := r.db.GetDB().
		Preload("DialogueState").
		Where("progress_id = ?", progressID).
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// GetPlayerPOIRecordByIDs retrieves a player's POI record by progress and POI IDs
func (r *campaignRepository) GetPlayerPOIRecordByIDs(progressID, poiID string) (*model.PlayerPOIRecord, error) {
	var record model.PlayerPOIRecord
	if err := r.db.GetDB().
		Preload("DialogueState").
		Where("progress_id = ? AND poi_id = ?", progressID, poiID).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error to indicate no record found
		}
		return nil, err
	}
	return &record, nil
}

// CreatePlayerPOIRecord creates a new player POI record
func (r *campaignRepository) CreatePlayerPOIRecord(record *model.PlayerPOIRecord) error {
	return r.db.GetDB().Create(record).Error
}

// UpdatePlayerPOIRecord updates a player's POI record
func (r *campaignRepository) UpdatePlayerPOIRecord(record *model.PlayerPOIRecord) error {
	return r.db.GetDB().Save(record).Error
}

// GetDialogueStateByIDs retrieves a dialogue state by record and dialogue IDs
func (r *campaignRepository) GetDialogueStateByIDs(recordID, dialogueID string) (*model.DialogueState, error) {
	var state model.DialogueState
	if err := r.db.GetDB().
		Where("record_id = ? AND dialogue_id = ?", recordID, dialogueID).
		First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error to indicate no state found
		}
		return nil, err
	}
	return &state, nil
}

// CreateDialogueState creates a new dialogue state
func (r *campaignRepository) CreateDialogueState(state *model.DialogueState) error {
	return r.db.GetDB().Create(state).Error
}

// UpdateDialogueState updates a dialogue state
func (r *campaignRepository) UpdateDialogueState(state *model.DialogueState) error {
	return r.db.GetDB().Save(state).Error
}

// CreateCampaign creates a new campaign
func (r *campaignRepository) CreateCampaign(campaign *model.Campaign) error {
	return r.db.GetDB().Create(campaign).Error
}

// CreateChapter creates a new chapter
func (r *campaignRepository) CreateChapter(chapter *model.Chapter) error {
	return r.db.GetDB().Create(chapter).Error
}

// CreateMission creates a new mission
func (r *campaignRepository) CreateMission(mission *model.Mission) error {
	return r.db.GetDB().Create(mission).Error
}

// CreateBranch creates a new branch
func (r *campaignRepository) CreateBranch(branch *model.Branch) error {
	return r.db.GetDB().Create(branch).Error
}

// CreateCampaignOperation creates a new campaign operation
func (r *campaignRepository) CreateCampaignOperation(operation *model.CampaignOperation) error {
	return r.db.GetDB().Create(operation).Error
}

// CreateCampaignPOI creates a new campaign POI
func (r *campaignRepository) CreateCampaignPOI(poi *model.CampaignPOI) error {
	return r.db.GetDB().Create(poi).Error
}

// CreateDialogue creates a new dialogue
func (r *campaignRepository) CreateDialogue(dialogue *model.Dialogue) error {
	return r.db.GetDB().Create(dialogue).Error
}
