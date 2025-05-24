// be/internal/model/campaign.go

package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Campaign represents the overall story container
type Campaign struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	IsActive    bool      `json:"isActive" gorm:"not null;default:true"`
	ImageURL    string    `json:"imageUrl,omitempty"`
	Chapters    []Chapter `json:"chapters,omitempty" gorm:"foreignKey:CampaignID"`
	CreatedAt   time.Time `json:"-" gorm:"not null"`
	UpdatedAt   time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new campaign
func (c *Campaign) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// Chapter represents a major story segment
type Chapter struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key"`
	CampaignID  string    `json:"campaignId" gorm:"type:uuid;not null;references:campaigns.id"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	Order       int       `json:"order" gorm:"not null"`
	Missions    []Mission `json:"missions,omitempty" gorm:"foreignKey:ChapterID"`
	CreatedAt   time.Time `json:"-" gorm:"not null"`
	UpdatedAt   time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new chapter
func (c *Chapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// Mission represents an individual playable story segment
type Mission struct {
	ID            string         `json:"id" gorm:"type:uuid;primary_key"`
	ChapterID     string         `json:"chapterId" gorm:"type:uuid;not null;references:chapters.id"`
	Name          string         `json:"name" gorm:"not null"`
	Description   string         `json:"description" gorm:"not null"`
	Order         int            `json:"order" gorm:"not null"`
	Branches      []Branch       `json:"branches,omitempty" gorm:"foreignKey:MissionID"`
	Prerequisites pq.StringArray `json:"prerequisites" gorm:"type:text[]"` // Mission IDs that must be completed first
	CreatedAt     time.Time      `json:"-" gorm:"not null"`
	UpdatedAt     time.Time      `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new mission
func (m *Mission) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// Branch represents an available outcome branch for a mission
type Branch struct {
	ID          string              `json:"id" gorm:"type:uuid;primary_key"`
	MissionID   string              `json:"missionId" gorm:"type:uuid;not null;references:missions.id"`
	Name        string              `json:"name" gorm:"not null"`
	Description string              `json:"description" gorm:"not null"`
	Operations  []CampaignOperation `json:"operations,omitempty" gorm:"foreignKey:BranchID"`
	POIs        []CampaignPOI       `json:"pois,omitempty" gorm:"foreignKey:BranchID"`
	CreatedAt   time.Time           `json:"-" gorm:"not null"`
	UpdatedAt   time.Time           `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new branch
func (b *Branch) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// CampaignOperation represents an operation that is part of a campaign branch
type CampaignOperation struct {
	ID           string                 `json:"id" gorm:"type:uuid;primary_key"`
	BranchID     string                 `json:"branchId" gorm:"type:uuid;not null;references:branches.id"`
	Name         string                 `json:"name" gorm:"not null"`
	Description  string                 `json:"description" gorm:"not null"`
	Type         string                 `json:"type" gorm:"not null"` // carjacking, goods_smuggling, etc.
	IsSpecial    bool                   `json:"isSpecial" gorm:"not null;default:true"`
	RegionIDs    pq.StringArray         `json:"regionIds" gorm:"type:text[]"`
	Requirements OperationRequirements  `json:"requirements" gorm:"embedded"`
	Resources    OperationResources     `json:"resources" gorm:"embedded"`
	Rewards      OperationRewards       `json:"rewards" gorm:"embedded"`
	Risks        OperationRisks         `json:"risks" gorm:"embedded"`
	Duration     int                    `json:"duration" gorm:"not null"`
	SuccessRate  int                    `json:"successRate" gorm:"not null"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"-"`
	CreatedAt    time.Time              `json:"-" gorm:"not null"`
	UpdatedAt    time.Time              `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new campaign operation
func (o *CampaignOperation) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// CampaignPOI represents a point of interest (hotspot) in a campaign branch
type CampaignPOI struct {
	ID           string                 `json:"id" gorm:"type:uuid;primary_key"`
	BranchID     string                 `json:"branchId" gorm:"type:uuid;not null;references:branches.id"`
	Name         string                 `json:"name" gorm:"not null"`
	Description  string                 `json:"description" gorm:"not null"`
	Type         string                 `json:"type" gorm:"not null"`
	BusinessType string                 `json:"businessType" gorm:"not null"`
	IsLegal      bool                   `json:"isLegal" gorm:"not null"`
	CityID       string                 `json:"cityId" gorm:"type:uuid;not null"`
	Dialogues    []Dialogue             `json:"dialogues,omitempty" gorm:"foreignKey:POIID"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" gorm:"-"`
	CreatedAt    time.Time              `json:"-" gorm:"not null"`
	UpdatedAt    time.Time              `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new campaign POI
func (p *CampaignPOI) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// InteractionType represents different ways to interact with NPCs
type InteractionType string

const (
	InteractionNeutral    InteractionType = "Neutral"
	InteractionConvince   InteractionType = "Convince"
	InteractionIntimidate InteractionType = "Intimidate"
)

// Dialogue represents a conversation with a POI
type Dialogue struct {
	ID              string           `json:"id" gorm:"type:uuid;primary_key"`
	POIID           string           `json:"poiId" gorm:"type:uuid;not null;references:campaign_pois.id"`
	Speaker         string           `json:"speaker" gorm:"not null"` // "Player" or "NPC"
	InteractionType *InteractionType `json:"interactionType,omitempty"`
	Text            string           `json:"text" gorm:"not null"`
	Order           int              `json:"order" gorm:"not null"`
	IsSuccess       *bool            `json:"isSuccess,omitempty"`
	ResourceEffect  ResourceEffect   `json:"resourceEffect,omitempty" gorm:"embedded"`
	CreatedAt       time.Time        `json:"-" gorm:"not null"`
	UpdatedAt       time.Time        `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new dialogue
func (d *Dialogue) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// ResourceEffect represents the effect of a dialogue on player resources
type ResourceEffect struct {
	Money     int `json:"money,omitempty" gorm:"default:0"`
	Crew      int `json:"crew,omitempty" gorm:"default:0"`
	Weapons   int `json:"weapons,omitempty" gorm:"default:0"`
	Vehicles  int `json:"vehicles,omitempty" gorm:"default:0"`
	Respect   int `json:"respect,omitempty" gorm:"default:0"`
	Influence int `json:"influence,omitempty" gorm:"default:0"`
	Heat      int `json:"heat,omitempty" gorm:"default:0"`
}

// PlayerCampaignProgress represents a player's progress in a campaign
type PlayerCampaignProgress struct {
	ID                    string         `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID              string         `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	CampaignID            string         `json:"campaignId" gorm:"type:uuid;not null;references:campaigns.id"`
	CurrentMissionID      *string        `json:"currentMissionId,omitempty" gorm:"type:uuid;references:missions.id"`
	CurrentBranchID       *string        `json:"currentBranchId,omitempty" gorm:"type:uuid;references:branches.id"`
	CompletedMissionIDs   pq.StringArray `json:"completedMissionIds" gorm:"type:text[]"`
	CompletedBranchIDs    pq.StringArray `json:"completedBranchIds" gorm:"type:text[]"`
	CompletedPOIIDs       pq.StringArray `json:"completedPoiIds" gorm:"type:text[]"`
	CompletedOperationIDs pq.StringArray `json:"completedOperationIds" gorm:"type:text[]"`
	CreatedAt             time.Time      `json:"-" gorm:"not null"`
	UpdatedAt             time.Time      `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating new player campaign progress
func (p *PlayerCampaignProgress) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PlayerOperationRecord represents a record of a player's operation for a campaign
type PlayerOperationRecord struct {
	ID          string     `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID    string     `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	ProgressID  string     `json:"progressId" gorm:"type:uuid;not null;references:player_campaign_progresses.id"`
	OperationID string     `json:"operationId" gorm:"type:uuid;not null;references:campaign_operations.id"`
	AttemptID   string     `json:"attemptId" gorm:"type:uuid;not null;references:operation_attempts.id"`
	IsCompleted bool       `json:"isCompleted" gorm:"not null;default:false"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"-" gorm:"not null"`
	UpdatedAt   time.Time  `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new player operation record
func (r *PlayerOperationRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// PlayerPOIRecord represents a record of a player's POI interaction for a campaign
type PlayerPOIRecord struct {
	ID            string          `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID      string          `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	ProgressID    string          `json:"progressId" gorm:"type:uuid;not null;references:player_campaign_progresses.id"`
	POIID         string          `json:"poiId" gorm:"type:uuid;not null;references:campaign_pois.id"`
	DialogueState []DialogueState `json:"dialogueState,omitempty" gorm:"foreignKey:RecordID"`
	IsCompleted   bool            `json:"isCompleted" gorm:"not null;default:false"`
	CompletedAt   *time.Time      `json:"completedAt,omitempty"`
	CreatedAt     time.Time       `json:"-" gorm:"not null"`
	UpdatedAt     time.Time       `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new player POI record
func (r *PlayerPOIRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// DialogueState represents the state of a dialogue in a player's POI interaction
type DialogueState struct {
	ID           string           `json:"id" gorm:"type:uuid;primary_key"`
	RecordID     string           `json:"recordId" gorm:"type:uuid;not null;references:player_poi_records.id"`
	DialogueID   string           `json:"dialogueId" gorm:"type:uuid;not null;references:dialogues.id"`
	IsCompleted  bool             `json:"isCompleted" gorm:"not null;default:false"`
	PlayerChoice *InteractionType `json:"playerChoice,omitempty"`
	CreatedAt    time.Time        `json:"-" gorm:"not null"`
	UpdatedAt    time.Time        `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new dialogue state
func (s *DialogueState) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
