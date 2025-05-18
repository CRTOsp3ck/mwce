// internal/model/campaign.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Campaign Definition Models ---

// Campaign represents a storyline with multiple chapters
type Campaign struct {
	ID               string    `json:"id" gorm:"primary_key"`
	Title            string    `json:"title" gorm:"not null"`
	Description      string    `json:"description" gorm:"not null"`
	ImageURL         string    `json:"imageUrl,omitempty"`
	InitialChapterID string    `json:"initialChapterId" gorm:"not null"`
	IsActive         bool      `json:"isActive" gorm:"not null;default:true"`
	RequiredLevel    int       `json:"requiredLevel" gorm:"not null;default:0"` // Player's level/title requirement
	CreatedAt        time.Time `json:"-" gorm:"not null"`
	UpdatedAt        time.Time `json:"-" gorm:"not null"`
	Chapters         []Chapter `json:"chapters,omitempty" gorm:"foreignKey:CampaignID"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new campaign
func (c *Campaign) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// Chapter represents a section of a campaign containing multiple missions
type Chapter struct {
	ID          string    `json:"id" gorm:"primary_key"`
	CampaignID  string    `json:"campaignId" gorm:"not null;references:campaigns.id"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	ImageURL    string    `json:"imageUrl,omitempty"`
	IsLocked    bool      `json:"isLocked" gorm:"not null;default:true"`
	Order       int       `json:"order" gorm:"not null;default:0"` // For display order
	CreatedAt   time.Time `json:"-" gorm:"not null"`
	UpdatedAt   time.Time `json:"-" gorm:"not null"`
	Missions    []Mission `json:"missions,omitempty" gorm:"foreignKey:ChapterID"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new chapter
func (c *Chapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// Mission represents an individual task within a chapter
type Mission struct {
	ID           string              `json:"id" gorm:"primary_key"`
	ChapterID    string              `json:"chapterId" gorm:"not null;references:chapters.id"`
	Title        string              `json:"title" gorm:"not null"`
	Description  string              `json:"description" gorm:"not null"`
	Narrative    string              `json:"narrative" gorm:"type:text"` // Story text for the mission
	ImageURL     string              `json:"imageUrl,omitempty"`
	MissionType  string              `json:"missionType" gorm:"not null"` // operation, territory, resource, travel, etc.
	Requirements MissionRequirements `json:"requirements" gorm:"embedded"`
	Rewards      MissionRewards      `json:"rewards" gorm:"embedded"`
	IsLocked     bool                `json:"isLocked" gorm:"not null;default:true"`
	Order        int                 `json:"order" gorm:"not null;default:0"` // For display order
	Choices      []MissionChoice     `json:"choices,omitempty" gorm:"foreignKey:MissionID"`
	CreatedAt    time.Time           `json:"-" gorm:"not null"`
	UpdatedAt    time.Time           `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new mission
func (m *Mission) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// MissionRequirements defines what a player needs to complete a mission
type MissionRequirements struct {
	Money              int    `json:"money,omitempty" gorm:"default:0"`
	Crew               int    `json:"crew,omitempty" gorm:"default:0"`
	Weapons            int    `json:"weapons,omitempty" gorm:"default:0"`
	Vehicles           int    `json:"vehicles,omitempty" gorm:"default:0"`
	Respect            int    `json:"respect,omitempty" gorm:"default:0"`
	Influence          int    `json:"influence,omitempty" gorm:"default:0"`
	MaxHeat            int    `json:"maxHeat,omitempty" gorm:"default:0"`
	MinTitle           string `json:"minTitle,omitempty" gorm:"default:''"`
	OperationType      string `json:"operationType,omitempty" gorm:"default:''"`
	OperationID        string `json:"operationId,omitempty" gorm:"default:''"`
	HotspotID          string `json:"hotspotId,omitempty" gorm:"default:''"`
	RegionID           string `json:"regionId,omitempty" gorm:"default:''"`
	ControlledHotspots int    `json:"controlledHotspots,omitempty" gorm:"default:0"`
}

// MissionRewards defines what a player receives for completing a mission
type MissionRewards struct {
	Money           int    `json:"money,omitempty" gorm:"default:0"`
	Crew            int    `json:"crew,omitempty" gorm:"default:0"`
	Weapons         int    `json:"weapons,omitempty" gorm:"default:0"`
	Vehicles        int    `json:"vehicles,omitempty" gorm:"default:0"`
	Respect         int    `json:"respect,omitempty" gorm:"default:0"`
	Influence       int    `json:"influence,omitempty" gorm:"default:0"`
	HeatReduction   int    `json:"heatReduction,omitempty" gorm:"default:0"`
	UnlockHotspotID string `json:"unlockHotspotId,omitempty" gorm:"default:''"`
	UnlockChapterID string `json:"unlockChapterId,omitempty" gorm:"default:''"`
	UnlockMissionID string `json:"unlockMissionId,omitempty" gorm:"default:''"`
}

// MissionChoice represents a decision point in a mission
type MissionChoice struct {
	ID                 string              `json:"id" gorm:"primary_key"`
	MissionID          string              `json:"missionId" gorm:"not null;references:missions.id"`
	Text               string              `json:"text" gorm:"not null"`
	NextMissionID      string              `json:"nextMissionId"`
	Requirements       MissionRequirements `json:"requirements" gorm:"embedded"`
	Rewards            MissionRewards      `json:"rewards" gorm:"embedded"`
	SequentialOrder    bool                `json:"sequentialOrder" gorm:"not null;default:false"`
	Conditions         []ConditionTemplate `json:"conditions,omitempty" gorm:"foreignKey:ChoiceID"`
	POITemplates       []POITemplate       `json:"pois,omitempty" gorm:"foreignKey:ChoiceID"`
	OperationTemplates []OperationTemplate `json:"operations,omitempty" gorm:"foreignKey:ChoiceID"`
	CreatedAt          time.Time           `json:"-" gorm:"not null"`
	UpdatedAt          time.Time           `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new mission choice
func (mc *MissionChoice) BeforeCreate(tx *gorm.DB) error {
	if mc.ID == "" {
		mc.ID = uuid.New().String()
	}
	return nil
}

// ConditionTemplate represents a completion condition template for mission choices
type ConditionTemplate struct {
	ID              string    `json:"id" gorm:"primary_key"`
	ChoiceID        string    `json:"choiceId" gorm:"not null;references:mission_choices.id"`
	Type            string    `json:"type" gorm:"not null"`                 // travel, territory, operation
	RequiredValue   string    `json:"requiredValue" gorm:"not null"`        // Region ID, hotspot ID, operation type, etc.
	AdditionalValue string    `json:"additionalValue" gorm:"default:''"`    // Additional parameters like action type
	OrderIndex      int       `json:"orderIndex" gorm:"not null;default:0"` // For sequential ordering
	CreatedAt       time.Time `json:"-" gorm:"not null"`
	UpdatedAt       time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new condition template
func (c *ConditionTemplate) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// POITemplate represents a Point of Interest template in the game world
type POITemplate struct {
	ID           string    `json:"id" gorm:"primary_key"`
	Name         string    `json:"name" gorm:"not null"`
	Description  string    `json:"description" gorm:"not null"`
	LocationType string    `json:"locationType" gorm:"not null"` // hotspot, region, district, etc.
	LocationID   string    `json:"locationId" gorm:"not null"`   // ID of the location
	MissionID    string    `json:"missionId" gorm:"not null;references:missions.id"`
	ChoiceID     string    `json:"choiceId" gorm:"references:mission_choices.id"`
	CreatedAt    time.Time `json:"-" gorm:"not null"`
	UpdatedAt    time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new POI template
func (p *POITemplate) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// OperationTemplate represents a mission operation template
type OperationTemplate struct {
	ID            string             `json:"id" gorm:"primary_key"`
	Name          string             `json:"name" gorm:"not null"`
	Description   string             `json:"description" gorm:"not null"`
	OperationType string             `json:"operationType" gorm:"not null"` // Same as in regular operations
	MissionID     string             `json:"missionId" gorm:"not null;references:missions.id"`
	ChoiceID      string             `json:"choiceId" gorm:"references:mission_choices.id"`
	Resources     OperationResources `json:"resources" gorm:"embedded"`
	Rewards       OperationRewards   `json:"rewards" gorm:"embedded"`
	Risks         OperationRisks     `json:"risks" gorm:"embedded"`
	Duration      int                `json:"duration" gorm:"not null"`
	SuccessRate   int                `json:"successRate" gorm:"not null"`
	CreatedAt     time.Time          `json:"-" gorm:"not null"`
	UpdatedAt     time.Time          `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new operation template
func (o *OperationTemplate) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// --- Player Progress Models ---

// PlayerCampaignProgress tracks a player's progress through campaigns
type PlayerCampaignProgress struct {
	ID               string     `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID         string     `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	CampaignID       string     `json:"campaignId" gorm:"not null;references:campaigns.id"`
	CurrentChapterID string     `json:"currentChapterId" gorm:"references:chapters.id"`
	CurrentMissionID string     `json:"currentMissionId" gorm:"references:missions.id"`
	IsCompleted      bool       `json:"isCompleted" gorm:"not null;default:false"`
	CompletedAt      *time.Time `json:"completedAt"`
	LastUpdated      time.Time  `json:"lastUpdated" gorm:"not null"`
	CreatedAt        time.Time  `json:"-" gorm:"not null"`
	UpdatedAt        time.Time  `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new progress record
func (p *PlayerCampaignProgress) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// MissionObjective represents a single objective for a mission
type MissionObjective struct {
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Target      string     `json:"target"`
	IsCompleted bool       `json:"isCompleted"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// PlayerMissionProgress tracks a player's progress on individual missions
type PlayerMissionProgress struct {
	ID                  string     `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID            string     `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	MissionID           string     `json:"missionId" gorm:"not null;references:missions.id"`
	Status              string     `json:"status" gorm:"not null;default:'not_started'"` // not_started, in_progress, completed, failed
	ChoiceID            string     `json:"choiceId,omitempty" gorm:"references:mission_choices.id"`
	StartedAt           *time.Time `json:"startedAt"`
	CompletedAt         *time.Time `json:"completedAt"`
	CurrentActiveChoice string     `json:"currentActiveChoice" gorm:"default:''"`
	ActionLog           string     `json:"actionLog" gorm:"type:text"` // JSON array of tracked actions
	CreatedAt           time.Time  `json:"-" gorm:"not null"`
	UpdatedAt           time.Time  `json:"-" gorm:"not null"`

	// Non-persisted fields
	Objectives  []MissionObjective `json:"objectives,omitempty" gorm:"-"`
	CanComplete bool
}

// BeforeCreate is a GORM hook to generate UUID before creating a new mission progress record
func (p *PlayerMissionProgress) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PlayerPOI represents a player-specific Point of Interest instance
type PlayerPOI struct {
	ID           string     `json:"id" gorm:"primary_key"`
	TemplateID   string     `json:"templateId" gorm:"not null;references:poi_templates.id"`
	PlayerID     string     `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	MissionID    string     `json:"missionId" gorm:"not null;references:missions.id"`
	ChoiceID     string     `json:"choiceId" gorm:"references:mission_choices.id"`
	Name         string     `json:"name" gorm:"not null"`
	Description  string     `json:"description" gorm:"not null"`
	LocationType string     `json:"locationType" gorm:"not null"`
	LocationID   string     `json:"locationId" gorm:"not null"`
	IsActive     bool       `json:"isActive" gorm:"not null;default:false"`
	IsCompleted  bool       `json:"isCompleted" gorm:"not null;default:false"`
	CompletedAt  *time.Time `json:"completedAt"`
	CreatedAt    time.Time  `json:"-" gorm:"not null"`
	UpdatedAt    time.Time  `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new player POI
func (p *PlayerPOI) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PlayerMissionOperation represents a player-specific mission operation instance
type PlayerMissionOperation struct {
	ID            string             `json:"id" gorm:"primary_key"`
	TemplateID    string             `json:"templateId" gorm:"not null;references:operation_templates.id"`
	PlayerID      string             `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	MissionID     string             `json:"missionId" gorm:"not null;references:missions.id"`
	ChoiceID      string             `json:"choiceId" gorm:"references:mission_choices.id"`
	Name          string             `json:"name" gorm:"not null"`
	Description   string             `json:"description" gorm:"not null"`
	OperationType string             `json:"operationType" gorm:"not null"`
	Resources     OperationResources `json:"resources" gorm:"embedded"`
	Rewards       OperationRewards   `json:"rewards" gorm:"embedded"`
	Risks         OperationRisks     `json:"risks" gorm:"embedded"`
	Duration      int                `json:"duration" gorm:"not null"`
	SuccessRate   int                `json:"successRate" gorm:"not null"`
	IsActive      bool               `json:"isActive" gorm:"not null;default:false"`
	IsCompleted   bool               `json:"isCompleted" gorm:"not null;default:false"`
	CompletedAt   *time.Time         `json:"completedAt"`
	CreatedAt     time.Time          `json:"-" gorm:"not null"`
	UpdatedAt     time.Time          `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new player mission operation
func (p *PlayerMissionOperation) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PlayerCompletionCondition represents a player-specific completion condition
type PlayerCompletionCondition struct {
	ID              string     `json:"id" gorm:"primary_key"`
	TemplateID      string     `json:"templateId" gorm:"not null;references:condition_templates.id"`
	PlayerID        string     `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	ChoiceID        string     `json:"choiceId" gorm:"not null;references:mission_choices.id"`
	Type            string     `json:"type" gorm:"not null"`
	RequiredValue   string     `json:"requiredValue" gorm:"not null"`
	AdditionalValue string     `json:"additionalValue" gorm:"default:''"`
	OrderIndex      int        `json:"orderIndex" gorm:"not null;default:0"`
	IsCompleted     bool       `json:"isCompleted" gorm:"not null;default:false"`
	CompletedAt     *time.Time `json:"completedAt"`
	CreatedAt       time.Time  `json:"-" gorm:"not null"`
	UpdatedAt       time.Time  `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new player completion condition
func (p *PlayerCompletionCondition) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// MissionCompleteResult contains the results of completing a mission
type MissionCompleteResult struct {
	Success     bool                   `json:"success"`
	Rewards     MissionRewards         `json:"rewards"`
	NextMission *Mission               `json:"nextMission,omitempty"`
	Progress    *PlayerMissionProgress `json:"progress"`
	Message     string                 `json:"message"`
}
