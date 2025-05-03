// internal/model/player.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Player represents a player in the game
type Player struct {
	ID                 string     `json:"id" gorm:"type:uuid;primary_key"`
	Name               string     `json:"name" gorm:"not null"`
	Email              string     `json:"email" gorm:"unique;not null"`
	Password           string     `json:"-" gorm:"not null"` // Hashed password, not returned in JSON
	Title              string     `json:"title" gorm:"not null"`
	Money              int        `json:"money" gorm:"not null;default:0"`
	Crew               int        `json:"crew" gorm:"not null;default:0"`
	MaxCrew            int        `json:"maxCrew" gorm:"not null;default:25"`
	Weapons            int        `json:"weapons" gorm:"not null;default:0"`
	MaxWeapons         int        `json:"maxWeapons" gorm:"not null;default:30"`
	Vehicles           int        `json:"vehicles" gorm:"not null;default:0"`
	MaxVehicles        int        `json:"maxVehicles" gorm:"not null;default:12"`
	Respect            int        `json:"respect" gorm:"not null;default:0"`
	Influence          int        `json:"influence" gorm:"not null;default:0"`
	Heat               int        `json:"heat" gorm:"not null;default:0"`
	CurrentRegionID    *string    `json:"currentRegionId" gorm:"type:uuid;references:regions.id"`
	LastTravelTime     *time.Time `json:"lastTravelTime"`
	CreatedAt          time.Time  `json:"createdAt" gorm:"not null"`
	LastActive         time.Time  `json:"lastActive" gorm:"not null"`
	TotalHotspots      int        `json:"totalHotspotCount" gorm:"-"`  // Calculated field, not stored in DB
	ControlledHotspots int        `json:"controlledHotspots" gorm:"-"` // Calculated field, not stored in DB
	HourlyRevenue      int        `json:"hourlyRevenue" gorm:"-"`      // Calculated field, not stored in DB
	PendingCollections int        `json:"pendingCollections" gorm:"-"` // Calculated field, not stored in DB
}

// BeforeCreate is a GORM hook to generate UUID before creating a new player
func (p *Player) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PlayerStats represents player statistics
type PlayerStats struct {
	PlayerID                 string    `json:"-" gorm:"type:uuid;primary_key;references:players.id"`
	TotalOperationsCompleted int       `json:"totalOperationsCompleted" gorm:"not null;default:0"`
	TotalMoneyEarned         int       `json:"totalMoneyEarned" gorm:"not null;default:0"`
	TotalHotspotsControlled  int       `json:"totalHotspotsControlled" gorm:"not null;default:0"`
	MaxInfluenceAchieved     int       `json:"maxInfluenceAchieved" gorm:"not null;default:0"`
	MaxRespectAchieved       int       `json:"maxRespectAchieved" gorm:"not null;default:0"`
	SuccessfulTakeovers      int       `json:"successfulTakeovers" gorm:"not null;default:0"`
	FailedTakeovers          int       `json:"failedTakeovers" gorm:"not null;default:0"`
	RegionsVisited           int       `json:"regionsVisited" gorm:"not null;default:0"`
	TotalTravelDistance      int       `json:"totalTravelDistance" gorm:"not null;default:0"` // Could be used for achievements
	CreatedAt                time.Time `json:"-" gorm:"not null"`
	UpdatedAt                time.Time `json:"-" gorm:"not null"`
}

// Notification represents a player notification
type Notification struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID  string    `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	Message   string    `json:"message" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"` // territory, operation, collection, heat, system, travel
	Timestamp time.Time `json:"timestamp" gorm:"not null"`
	Read      bool      `json:"read" gorm:"not null;default:false"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new notification
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}

// Achievement represents a player achievement
type Achievement struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	Criteria    string    `json:"criteria" gorm:"not null"`
	Reward      string    `json:"reward" gorm:"not null"`
	CreatedAt   time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new achievement
func (a *Achievement) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// PlayerAchievement represents an achievement earned by a player
type PlayerAchievement struct {
	PlayerID      string    `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	AchievementID string    `json:"achievementId" gorm:"type:uuid;not null;references:achievements.id"`
	UnlockedAt    time.Time `json:"unlockedAt" gorm:"not null"`
}

// ProfileResponse represents the player profile response
type ProfileResponse struct {
	Player Player `json:"player"`
}

// StatsResponse represents the player stats response
type StatsResponse struct {
	Stats PlayerStats `json:"stats"`
}

// NotificationsResponse represents the player notifications response
type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}

// CollectAllResponse represents the response after collecting all pending resources
type CollectAllResponse struct {
	CollectedAmount int    `json:"collectedAmount"`
	HotspotsCount   int    `json:"hotspotsCount"`
	Message         string `json:"message"`
}

// TravelRequest represents a request to travel to a new region
type TravelRequest struct {
	RegionID string `json:"regionId" binding:"required"`
}

// TravelResponse represents the response after attempting to travel
type TravelResponse struct {
	Success        bool   `json:"success"`
	RegionID       string `json:"regionId"`
	RegionName     string `json:"regionName"`
	TravelCost     int    `json:"travelCost"`
	HeatReduction  int    `json:"heatReduction,omitempty"`
	Message        string `json:"message"`
	CaughtByPolice bool   `json:"caughtByPolice,omitempty"`
	FineAmount     int    `json:"fineAmount,omitempty"`
	HeatIncrease   int    `json:"heatIncrease,omitempty"`
}

// TravelAttempt represents a travel attempt by a player
type TravelAttempt struct {
	ID             string    `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID       string    `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	FromRegionID   *string   `json:"fromRegionId" gorm:"type:uuid;references:regions.id"`
	ToRegionID     string    `json:"toRegionId" gorm:"type:uuid;not null;references:regions.id"`
	Success        bool      `json:"success" gorm:"not null"`
	CaughtByPolice bool      `json:"caughtByPolice" gorm:"not null;default:false"`
	TravelCost     int       `json:"travelCost" gorm:"not null"`
	FineAmount     int       `json:"fineAmount" gorm:"not null;default:0"`
	HeatChange     int       `json:"heatChange" gorm:"not null;default:0"` // Negative for reduction, positive for increase
	Timestamp      time.Time `json:"timestamp" gorm:"not null"`
	CreatedAt      time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new travel attempt
func (t *TravelAttempt) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}
