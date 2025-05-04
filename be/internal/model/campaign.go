// internal/model/campaign.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Campaign represents a storyline with multiple chapters
type Campaign struct {
	ID               string    `json:"id" gorm:"primary_key"` // Changed from type:uuid to allow string IDs
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
	ID          string    `json:"id" gorm:"primary_key"`                              // Changed from type:uuid
	CampaignID  string    `json:"campaignId" gorm:"not null;references:campaigns.id"` // Changed from type:uuid
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
	ID           string              `json:"id" gorm:"primary_key"`                            // Changed from type:uuid
	ChapterID    string              `json:"chapterId" gorm:"not null;references:chapters.id"` // Changed from type:uuid
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
	ID            string              `json:"id" gorm:"primary_key"`                            // Changed from type:uuid
	MissionID     string              `json:"missionId" gorm:"not null;references:missions.id"` // Changed from type:uuid
	Text          string              `json:"text" gorm:"not null"`                             // Choice text shown to player
	NextMissionID string              `json:"nextMissionId"`                                    // Changed from type:uuid
	Requirements  MissionRequirements `json:"requirements" gorm:"embedded"`                     // Requirements to select this choice
	Rewards       MissionRewards      `json:"rewards" gorm:"embedded"`                          // Additional rewards for this choice
	CreatedAt     time.Time           `json:"-" gorm:"not null"`
	UpdatedAt     time.Time           `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new mission choice
func (mc *MissionChoice) BeforeCreate(tx *gorm.DB) error {
	if mc.ID == "" {
		mc.ID = uuid.New().String()
	}
	return nil
}

// PlayerCampaignProgress tracks a player's progress through campaigns
type PlayerCampaignProgress struct {
	ID               string     `json:"id" gorm:"type:uuid;primary_key"` // Keep as UUID for player progress
	PlayerID         string     `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	CampaignID       string     `json:"campaignId" gorm:"not null;references:campaigns.id"` // Changed from type:uuid
	CurrentChapterID string     `json:"currentChapterId" gorm:"references:chapters.id"`     // Changed from type:uuid
	CurrentMissionID string     `json:"currentMissionId" gorm:"references:missions.id"`     // Changed from type:uuid
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

// PlayerMissionProgress tracks a player's progress on individual missions
type PlayerMissionProgress struct {
	ID          string     `json:"id" gorm:"type:uuid;primary_key"` // Keep as UUID for player progress
	PlayerID    string     `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	MissionID   string     `json:"missionId" gorm:"not null;references:missions.id"`        // Changed from type:uuid
	Status      string     `json:"status" gorm:"not null;default:'not_started'"`            // not_started, in_progress, completed, failed
	ChoiceID    string     `json:"choiceId,omitempty" gorm:"references:mission_choices.id"` // Changed from type:uuid
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   time.Time  `json:"-" gorm:"not null"`
	UpdatedAt   time.Time  `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new mission progress record
func (p *PlayerMissionProgress) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
