// internal/model/market.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarketListing represents a resource listing on the market
type MarketListing struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key"`
	Type            string    `json:"type" gorm:"not null"` // crew, weapons, vehicles
	Price           int       `json:"price" gorm:"not null"`
	Quantity        int       `json:"quantity" gorm:"not null"`
	Trend           string    `json:"trend" gorm:"not null"` // up, down, stable
	TrendPercentage int       `json:"trendPercentage" gorm:"not null"`
	CreatedAt       time.Time `json:"-" gorm:"not null"`
	UpdatedAt       time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new market listing
func (m *MarketListing) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// MarketTransaction represents a transaction on the market
type MarketTransaction struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key"`
	PlayerID        string    `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	ResourceType    string    `json:"resourceType" gorm:"not null"` // crew, weapons, vehicles
	Quantity        int       `json:"quantity" gorm:"not null"`
	Price           int       `json:"price" gorm:"not null"`      // Price per unit at time of transaction
	TotalCost       int       `json:"totalCost" gorm:"not null"` // Total cost of the transaction
	Timestamp       time.Time `json:"timestamp" gorm:"not null"`
	TransactionType string    `json:"transaction_type" gorm:"not null"` // buy, sell
	CreatedAt       time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new market transaction
func (m *MarketTransaction) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// MarketPriceHistory represents the price history for a resource type
type MarketPriceHistory struct {
	ID           string    `json:"id" gorm:"type:uuid;primary_key"`
	ResourceType string    `json:"resourceType" gorm:"not null"` // crew, weapons, vehicles
	Price        int       `json:"price" gorm:"not null"`
	Timestamp    time.Time `json:"timestamp" gorm:"not null"`
	CreatedAt    time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new market price history record
func (m *MarketPriceHistory) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// MarketHistoryResponse represents the market price history response structure
type MarketHistoryResponse struct {
	ResourceType string   `json:"resourceType"`
	TimePoints   []string `json:"timePoints"` // Timestamps
	Prices       []int    `json:"prices"`      // Corresponding prices
}

// ResourceTransaction represents a request to buy or sell resources
type ResourceTransaction struct {
	ResourceType string `json:"resourceType" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required,gt=0"`
}
