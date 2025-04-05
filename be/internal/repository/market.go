// internal/repository/market.go

package repository

import (
	"errors"
	"math/rand"
	"time"

	"mwce-be/internal/model"
	"mwce-be/internal/util"
	"mwce-be/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarketRepository handles database operations for the market
type MarketRepository interface {
	GetAllListings() ([]model.MarketListing, error)
	GetListingByType(resourceType string) (*model.MarketListing, error)
	CreateListing(listing *model.MarketListing) error
	UpdateListing(listing *model.MarketListing) error
	CreateTransaction(transaction *model.MarketTransaction) error
	GetTransactionsByPlayer(playerID string) ([]model.MarketTransaction, error)
	UpdateMarketPrices() error
	CreatePriceHistory(history *model.MarketPriceHistory) error
	GetResourcePriceHistory(resourceType string, days int) (*model.MarketHistoryResponse, error)
	GetAllPriceHistory(days int) ([]model.MarketHistoryResponse, error)
}

type marketRepository struct {
	db database.Database
}

// NewMarketRepository creates a new market repository
func NewMarketRepository(db database.Database) MarketRepository {
	return &marketRepository{
		db: db,
	}
}

// GetAllListings retrieves all market listings
func (r *marketRepository) GetAllListings() ([]model.MarketListing, error) {
	var listings []model.MarketListing
	if err := r.db.GetDB().Find(&listings).Error; err != nil {
		return nil, err
	}
	return listings, nil
}

// GetListingByType retrieves a market listing by resource type
func (r *marketRepository) GetListingByType(resourceType string) (*model.MarketListing, error) {
	var listing model.MarketListing
	if err := r.db.GetDB().Where("type = ?", resourceType).First(&listing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("listing not found")
		}
		return nil, err
	}
	return &listing, nil
}

// CreateListing creates a new market listing
func (r *marketRepository) CreateListing(listing *model.MarketListing) error {
	if listing.ID == "" {
		listing.ID = uuid.New().String()
	}
	return r.db.GetDB().Create(listing).Error
}

// UpdateListing updates a market listing
func (r *marketRepository) UpdateListing(listing *model.MarketListing) error {
	return r.db.GetDB().Save(listing).Error
}

// CreateTransaction records a market transaction
func (r *marketRepository) CreateTransaction(transaction *model.MarketTransaction) error {
	return r.db.GetDB().Create(transaction).Error
}

// GetTransactionsByPlayer retrieves market transactions for a player
func (r *marketRepository) GetTransactionsByPlayer(playerID string) ([]model.MarketTransaction, error) {
	var transactions []model.MarketTransaction
	if err := r.db.GetDB().
		Where("player_id = ?", playerID).
		Order("timestamp DESC").
		Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// CreatePriceHistory creates a new price history record
func (r *marketRepository) CreatePriceHistory(history *model.MarketPriceHistory) error {
	if history.ID == "" {
		history.ID = uuid.New().String()
	}
	return r.db.GetDB().Create(history).Error
}

// UpdateMarketPrices updates market prices based on supply and demand and creates initial listings
func (r *marketRepository) UpdateMarketPrices() error {
	// Get all listings
	listings, err := r.GetAllListings()
	if err != nil {
		return err
	}

	// If no listings exist, create initial ones
	if len(listings) == 0 {
		if err := r.createInitialListings(); err != nil {
			return err
		}
		listings, err = r.GetAllListings()
		if err != nil {
			return err
		}
	}

	// Update each listing with new prices
	for _, listing := range listings {
		// Calculate price change (in a real implementation, this would be based on recent transactions)
		priceChange := rand.Intn(5) - 2 // Random change between -2% and +2%

		// Update trend
		if priceChange > 0 {
			listing.Trend = util.PriceTrendUp
		} else if priceChange < 0 {
			listing.Trend = util.PriceTrendDown
		} else {
			listing.Trend = util.PriceTrendStable
		}

		listing.TrendPercentage = abs(priceChange)

		// Apply price change
		newPrice := listing.Price + (listing.Price * priceChange / 100)
		// Ensure minimum price
		if newPrice < 100 {
			newPrice = 100
		}
		listing.Price = newPrice

		// Record price history
		history := model.MarketPriceHistory{
			ID:           uuid.New().String(),
			ResourceType: listing.Type,
			Price:        listing.Price,
			Timestamp:    time.Now(),
			CreatedAt:    time.Now(),
		}

		// Save changes
		if err := r.CreatePriceHistory(&history); err != nil {
			return err
		}

		if err := r.UpdateListing(&listing); err != nil {
			return err
		}
	}

	return nil
}

// GetResourcePriceHistory gets price history for a specific resource
func (r *marketRepository) GetResourcePriceHistory(resourceType string, days int) (*model.MarketHistoryResponse, error) {
	var historyRecords []model.MarketPriceHistory

	// Get records from the last 'days' days
	startDate := time.Now().AddDate(0, 0, -days)

	if err := r.db.GetDB().
		Where("resource_type = ?", resourceType).
		Where("timestamp > ?", startDate).
		Order("timestamp ASC").
		Find(&historyRecords).Error; err != nil {
		return nil, err
	}

	// Convert to response format
	response := &model.MarketHistoryResponse{
		ResourceType: resourceType,
		TimePoints:   make([]string, len(historyRecords)),
		Prices:       make([]int, len(historyRecords)),
	}

	for i, record := range historyRecords {
		response.TimePoints[i] = record.Timestamp.Format(time.RFC3339)
		response.Prices[i] = record.Price
	}

	return response, nil
}

// GetAllPriceHistory gets price history for all resources
func (r *marketRepository) GetAllPriceHistory(days int) ([]model.MarketHistoryResponse, error) {
	// Get all resource types
	resourceTypes := []string{
		util.ResourceTypeCrew,
		util.ResourceTypeWeapons,
		util.ResourceTypeVehicles,
	}

	var responses []model.MarketHistoryResponse

	for _, resourceType := range resourceTypes {
		response, err := r.GetResourcePriceHistory(resourceType, days)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

// createInitialListings creates initial market listings if none exist
func (r *marketRepository) createInitialListings() error {
	listings := []model.MarketListing{
		{
			ID:              uuid.New().String(),
			Type:            util.ResourceTypeCrew,
			Price:           1000,
			Quantity:        999,
			Trend:           util.PriceTrendStable,
			TrendPercentage: 0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New().String(),
			Type:            util.ResourceTypeWeapons,
			Price:           2000,
			Quantity:        999,
			Trend:           util.PriceTrendStable,
			TrendPercentage: 0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              uuid.New().String(),
			Type:            util.ResourceTypeVehicles,
			Price:           5000,
			Quantity:        999,
			Trend:           util.PriceTrendStable,
			TrendPercentage: 0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, listing := range listings {
		if err := r.CreateListing(&listing); err != nil {
			return err
		}

		// Create initial price history
		history := model.MarketPriceHistory{
			ID:           uuid.New().String(),
			ResourceType: listing.Type,
			Price:        listing.Price,
			Timestamp:    time.Now(),
			CreatedAt:    time.Now(),
		}

		if err := r.CreatePriceHistory(&history); err != nil {
			return err
		}
	}

	return nil
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
