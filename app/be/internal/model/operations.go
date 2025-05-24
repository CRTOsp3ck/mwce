// internal/model/operations.go

package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Operation represents an operation that can be performed by the player
type Operation struct {
	ID                   string                 `json:"id" gorm:"type:uuid;primary_key"`
	Name                 string                 `json:"name" gorm:"not null"`
	Description          string                 `json:"description" gorm:"not null"`
	Type                 string                 `json:"type" gorm:"not null"` // carjacking, goods_smuggling, etc.
	IsSpecial            bool                   `json:"isSpecial" gorm:"not null;default:false"`
	IsActive             bool                   `json:"isActive" gorm:"not null;default:true"`
	IsLocked             bool                   `json:"isLocked" gorm:"-"`             // Not stored in DB, calculated on retrieval
	LockReason           string                 `json:"lockReason,omitempty" gorm:"-"` // Reason why it's locked
	RegionIDs            pq.StringArray         `json:"regionIds" gorm:"type:text[]"`  // For multi-region operations
	Requirements         OperationRequirements  `json:"requirements" gorm:"embedded"`
	Resources            OperationResources     `json:"resources" gorm:"embedded"`
	Rewards              OperationRewards       `json:"rewards" gorm:"embedded"`
	Risks                OperationRisks         `json:"risks" gorm:"embedded"`
	Duration             int                    `json:"duration" gorm:"not null"`              // in seconds
	AvailabilityDuration int                    `json:"availabilityDuration" gorm:"default:0"` // in minutes, 0 means no expiration
	SuccessRate          int                    `json:"successRate" gorm:"not null"`           // percentage
	AvailableUntil       time.Time              `json:"availableUntil" gorm:"not null"`
	CreatedAt            time.Time              `json:"-" gorm:"not null"`
	UpdatedAt            time.Time              `json:"-" gorm:"not null"`
	PlayerAttempts       []OperationAttempt     `json:"playerAttempts,omitempty" gorm:"-"`
	Metadata             map[string]interface{} `json:"metadata,omitempty" gorm:"-"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new operation
func (o *Operation) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// OperationRequirements represents requirements for an operation
type OperationRequirements struct {
	MinInfluence         int    `json:"minInfluence,omitempty" gorm:"default:0"`
	MaxHeat              int    `json:"maxHeat,omitempty" gorm:"default:0"`
	MinTitle             string `json:"minTitle,omitempty" gorm:"default:''"`
	RequiredHotspotTypes string `json:"requiredHotspotTypes,omitempty" gorm:"default:''"` // Comma-separated list
}

// OperationResources represents resources required for an operation
type OperationResources struct {
	Crew     int `json:"crew" gorm:"not null;default:0"`
	Weapons  int `json:"weapons" gorm:"not null;default:0"`
	Vehicles int `json:"vehicles" gorm:"not null;default:0"`
	Money    int `json:"money,omitempty" gorm:"default:0"`
}

// OperationRewards represents rewards for an operation
type OperationRewards struct {
	Money         int `json:"money,omitempty" gorm:"default:0"`
	Crew          int `json:"crew,omitempty" gorm:"default:0"`
	Weapons       int `json:"weapons,omitempty" gorm:"default:0"`
	Vehicles      int `json:"vehicles,omitempty" gorm:"default:0"`
	Respect       int `json:"respect,omitempty" gorm:"default:0"`
	Influence     int `json:"influence,omitempty" gorm:"default:0"`
	HeatReduction int `json:"heatReduction,omitempty" gorm:"default:0"`
}

// OperationRisks represents risks for an operation
type OperationRisks struct {
	CrewLoss     int `json:"crewLoss,omitempty" gorm:"default:0"`
	WeaponsLoss  int `json:"weaponsLoss,omitempty" gorm:"default:0"`
	VehiclesLoss int `json:"vehiclesLoss,omitempty" gorm:"default:0"`
	MoneyLoss    int `json:"moneyLoss,omitempty" gorm:"default:0"`
	HeatIncrease int `json:"heatIncrease,omitempty" gorm:"default:0"`
	RespectLoss  int `json:"respectLoss,omitempty" gorm:"default:0"`
}

// OperationAttempt represents a player's attempt at an operation
type OperationAttempt struct {
	ID              string             `json:"id" gorm:"type:uuid;primary_key"`
	OperationID     string             `json:"operationId" gorm:"type:uuid;not null;references:operations.id"`
	PlayerID        string             `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	Timestamp       time.Time          `json:"timestamp" gorm:"not null"`
	Resources       OperationResources `json:"resources" gorm:"embedded"`
	Result          *OperationResult   `json:"result,omitempty" gorm:"embedded"`
	CompletionTime  *time.Time         `json:"completionTime,omitempty"`
	Status          string             `json:"status" gorm:"not null"` // in_progress, completed, failed, cancelled
	Notified        bool               `json:"notified" gorm:"default:false"`
	CreatedAt       time.Time          `json:"-" gorm:"not null"`
	UpdatedAt       time.Time          `json:"-" gorm:"not null"`
	OperationDetail *Operation         `json:"operationDetail,omitempty" gorm:"-"` // Not stored in DB, populated when needed
}

// BeforeCreate is a GORM hook to generate UUID before creating a new operation attempt
func (o *OperationAttempt) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = uuid.New().String()
	}
	return nil
}

// OperationResult represents the result of an operation attempt
type OperationResult struct {
	Success          bool   `json:"success" gorm:"default:false"`
	MoneyGained      int    `json:"moneyGained,omitempty" gorm:"default:0"`
	MoneyLost        int    `json:"moneyLost,omitempty" gorm:"default:0"`
	CrewGained       int    `json:"crewGained,omitempty" gorm:"default:0"`
	CrewLost         int    `json:"crewLost,omitempty" gorm:"default:0"`
	WeaponsGained    int    `json:"weaponsGained,omitempty" gorm:"default:0"`
	WeaponsLost      int    `json:"weaponsLost,omitempty" gorm:"default:0"`
	VehiclesGained   int    `json:"vehiclesGained,omitempty" gorm:"default:0"`
	VehiclesLost     int    `json:"vehiclesLost,omitempty" gorm:"default:0"`
	RespectGained    int    `json:"respectGained,omitempty" gorm:"default:0"`
	InfluenceGained  int    `json:"influenceGained,omitempty" gorm:"default:0"`
	HeatGenerated    int    `json:"heatGenerated,omitempty" gorm:"default:0"`
	HeatReduced      int    `json:"heatReduced,omitempty" gorm:"default:0"`
	RewardsCollected bool   `json:"rewardsCollected" gorm:"default:false"`
	Message          string `json:"message" gorm:"not null"`
}

// StartOperationRequest represents a request to start an operation
type StartOperationRequest struct {
	Resources OperationResources `json:"resources" binding:"required"`
}

// OperationsRefreshInfo represents information about operations refresh timing
type OperationsRefreshInfo struct {
	RefreshInterval int    `json:"refreshInterval"` // in minutes
	LastRefreshTime string `json:"lastRefreshTime"` // ISO8601 format
	NextRefreshTime string `json:"nextRefreshTime"` // ISO8601 format
}
