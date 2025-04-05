// internal/service/market.go

package service

import (
	"errors"
	"math/rand"
	"time"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
)

// MarketService handles market-related business logic
type MarketService interface {
	GetListings() ([]model.MarketListing, error)
	GetListingByType(resourceType string) (*model.MarketListing, error)
	GetTransactions(playerID string) ([]model.MarketTransaction, error)
	GetPriceHistory(days int) ([]model.MarketHistoryResponse, error)
	GetResourcePriceHistory(resourceType string, days int) (*model.MarketHistoryResponse, error)
	BuyResource(playerID string, request model.ResourceTransaction) (*model.MarketTransaction, error)
	SellResource(playerID string, request model.ResourceTransaction) (*model.MarketTransaction, error)
	UpdateMarketPrices() error

	// Scheduled jobs
	StartPeriodicMarketPriceUpdates()
}

type marketService struct {
	marketRepo    repository.MarketRepository
	playerRepo    repository.PlayerRepository
	playerService PlayerService
	gameConfig    *config.GameConfig
	logger        zerolog.Logger
}

// NewMarketService creates a new market service
func NewMarketService(
	marketRepo repository.MarketRepository,
	playerRepo repository.PlayerRepository,
	playerService PlayerService,
	gameConfig *config.GameConfig,
	logger zerolog.Logger,
) MarketService {
	return &marketService{
		marketRepo:    marketRepo,
		playerRepo:    playerRepo,
		playerService: playerService,
		gameConfig:    gameConfig,
		logger:        logger,
	}
}

// GetListings retrieves all market listings
func (s *marketService) GetListings() ([]model.MarketListing, error) {
	return s.marketRepo.GetAllListings()
}

// GetListingByType retrieves a market listing by resource type
func (s *marketService) GetListingByType(resourceType string) (*model.MarketListing, error) {
	return s.marketRepo.GetListingByType(resourceType)
}

// GetTransactions retrieves market transactions for a player
func (s *marketService) GetTransactions(playerID string) ([]model.MarketTransaction, error) {
	return s.marketRepo.GetTransactionsByPlayer(playerID)
}

// GetPriceHistory retrieves price history for all resources
func (s *marketService) GetPriceHistory(days int) ([]model.MarketHistoryResponse, error) {
	return s.marketRepo.GetAllPriceHistory(days)
}

// GetResourcePriceHistory retrieves price history for a specific resource
func (s *marketService) GetResourcePriceHistory(resourceType string, days int) (*model.MarketHistoryResponse, error) {
	return s.marketRepo.GetResourcePriceHistory(resourceType, days)
}

// BuyResource handles a resource purchase
func (s *marketService) BuyResource(playerID string, request model.ResourceTransaction) (*model.MarketTransaction, error) {
	// Get the listing
	listing, err := s.marketRepo.GetListingByType(request.ResourceType)
	if err != nil {
		return nil, errors.New("resource type not available in the market")
	}

	// Calculate total cost
	totalCost := listing.Price * request.Quantity

	// Get the player
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	// Check if player has enough money
	if player.Money < totalCost {
		return nil, errors.New("not enough money to complete this purchase")
	}

	// Check max resource limits
	var currentResource, maxResource int
	switch request.ResourceType {
	case util.ResourceTypeCrew:
		currentResource = player.Crew
		maxResource = player.MaxCrew
	case util.ResourceTypeWeapons:
		currentResource = player.Weapons
		maxResource = player.MaxWeapons
	case util.ResourceTypeVehicles:
		currentResource = player.Vehicles
		maxResource = player.MaxVehicles
	default:
		return nil, errors.New("invalid resource type")
	}

	// Check if purchase would exceed maximum
	if currentResource+request.Quantity > maxResource {
		return nil, errors.New("this purchase would exceed your maximum capacity")
	}

	// Create the transaction
	transaction := &model.MarketTransaction{
		PlayerID:        playerID,
		ResourceType:    request.ResourceType,
		Quantity:        request.Quantity,
		Price:           listing.Price,
		TotalCost:       totalCost,
		Timestamp:       time.Now(),
		TransactionType: util.TransactionTypeBuy,
		CreatedAt:       time.Now(),
	}

	// Update player resources
	resourceUpdates := map[string]int{
		"money": -totalCost,
	}

	switch request.ResourceType {
	case util.ResourceTypeCrew:
		resourceUpdates["crew"] = request.Quantity
	case util.ResourceTypeWeapons:
		resourceUpdates["weapons"] = request.Quantity
	case util.ResourceTypeVehicles:
		resourceUpdates["vehicles"] = request.Quantity
	}

	if err := s.playerService.UpdatePlayerResources(playerID, resourceUpdates); err != nil {
		return nil, errors.New("failed to update player resources")
	}

	// Record the transaction
	if err := s.marketRepo.CreateTransaction(transaction); err != nil {
		// Try to revert the transaction if recording fails
		revertUpdates := map[string]int{
			"money": totalCost,
		}

		switch request.ResourceType {
		case util.ResourceTypeCrew:
			revertUpdates["crew"] = -request.Quantity
		case util.ResourceTypeWeapons:
			revertUpdates["weapons"] = -request.Quantity
		case util.ResourceTypeVehicles:
			revertUpdates["vehicles"] = -request.Quantity
		}

		s.playerService.UpdatePlayerResources(playerID, revertUpdates)

		return nil, errors.New("failed to record transaction")
	}

	// Add notification
	message := formatPurchaseNotification(request.Quantity, request.ResourceType, totalCost)
	s.playerService.AddNotification(playerID, message, util.NotificationTypeSystem)

	return transaction, nil
}

// SellResource handles a resource sale
func (s *marketService) SellResource(playerID string, request model.ResourceTransaction) (*model.MarketTransaction, error) {
	// Get the listing
	listing, err := s.marketRepo.GetListingByType(request.ResourceType)
	if err != nil {
		return nil, errors.New("resource type not available in the market")
	}

	// Calculate total value
	totalValue := listing.Price * request.Quantity

	// Get the player
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	// Check if player has enough resources
	var currentResource int
	switch request.ResourceType {
	case util.ResourceTypeCrew:
		currentResource = player.Crew
	case util.ResourceTypeWeapons:
		currentResource = player.Weapons
	case util.ResourceTypeVehicles:
		currentResource = player.Vehicles
	default:
		return nil, errors.New("invalid resource type")
	}

	if currentResource < request.Quantity {
		return nil, errors.New("not enough resources to complete this sale")
	}

	// Create the transaction
	transaction := &model.MarketTransaction{
		PlayerID:        playerID,
		ResourceType:    request.ResourceType,
		Quantity:        request.Quantity,
		Price:           listing.Price,
		TotalCost:       totalValue, // For sales, total cost is the value received
		Timestamp:       time.Now(),
		TransactionType: util.TransactionTypeSell,
		CreatedAt:       time.Now(),
	}

	// Update player resources
	resourceUpdates := map[string]int{
		"money": totalValue,
	}

	switch request.ResourceType {
	case util.ResourceTypeCrew:
		resourceUpdates["crew"] = -request.Quantity
	case util.ResourceTypeWeapons:
		resourceUpdates["weapons"] = -request.Quantity
	case util.ResourceTypeVehicles:
		resourceUpdates["vehicles"] = -request.Quantity
	}

	if err := s.playerService.UpdatePlayerResources(playerID, resourceUpdates); err != nil {
		return nil, errors.New("failed to update player resources")
	}

	// Record the transaction
	if err := s.marketRepo.CreateTransaction(transaction); err != nil {
		// Try to revert the transaction if recording fails
		revertUpdates := map[string]int{
			"money": -totalValue,
		}

		switch request.ResourceType {
		case util.ResourceTypeCrew:
			revertUpdates["crew"] = request.Quantity
		case util.ResourceTypeWeapons:
			revertUpdates["weapons"] = request.Quantity
		case util.ResourceTypeVehicles:
			revertUpdates["vehicles"] = request.Quantity
		}

		s.playerService.UpdatePlayerResources(playerID, revertUpdates)

		return nil, errors.New("failed to record transaction")
	}

	// Add notification
	message := formatSaleNotification(request.Quantity, request.ResourceType, totalValue)
	s.playerService.AddNotification(playerID, message, util.NotificationTypeSystem)

	return transaction, nil
}

// UpdateMarketPrices updates market prices based on supply and demand
func (s *marketService) UpdateMarketPrices() error {
	// First check if we have any mechanics config at all
	if s.gameConfig == nil {
		s.logger.Error().Msg("Game configuration is nil")
		return errors.New("game configuration is nil")
	}

	// Then check if we have mechanics config
	if s.gameConfig.Mechanics == nil {
		s.logger.Warn().Msg("Mechanics configuration not available, using default values")
		return s.marketRepo.UpdateMarketPrices()
	}

	// Finally, check specifically for market config
	marketConfig := s.gameConfig.Mechanics.Market
	if marketConfig.PriceFluctuationRange == 0 && len(marketConfig.BasePrices) == 0 {
		s.logger.Warn().Msg("Market mechanics configuration not available, using default values")
		return s.marketRepo.UpdateMarketPrices()
	}

	// If we get here, we have valid market config to use
	s.logger.Info().
		Int("fluctRange", marketConfig.PriceFluctuationRange).
		Int("updateInterval", marketConfig.PriceUpdateInterval).
		Int("numBasePrices", len(marketConfig.BasePrices)).
		Msg("Using market mechanics configuration")

	// Use the market configuration from mechanics.yaml
	// marketConfig := s.gameConfig.Mechanics.Market

	// Get current listings
	listings, err := s.marketRepo.GetAllListings()
	if err != nil {
		return err
	}

	// If no listings exist, create them using config values
	if len(listings) == 0 {
		// Create initial listings using base prices from config
		for resourceType, basePrice := range marketConfig.BasePrices {
			listing := model.MarketListing{
				Type:            resourceType,
				Price:           basePrice,
				Quantity:        999, // Default quantity
				Trend:           util.PriceTrendStable,
				TrendPercentage: 0,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			if err := s.marketRepo.CreateListing(&listing); err != nil {
				s.logger.Error().Err(err).
					Str("resourceType", resourceType).
					Msg("Failed to create initial market listing")
				continue
			}

			// Create initial price history entry
			history := model.MarketPriceHistory{
				ResourceType: resourceType,
				Price:        basePrice,
				Timestamp:    time.Now(),
				CreatedAt:    time.Now(),
			}

			if err := s.marketRepo.CreatePriceHistory(&history); err != nil {
				s.logger.Error().Err(err).
					Str("resourceType", resourceType).
					Msg("Failed to create initial price history")
			}
		}

		return nil
	}

	// For existing listings, update prices based on config
	for _, listing := range listings {
		// Get min and max prices from config
		minPrice, hasMin := marketConfig.MinPrices[listing.Type]
		maxPrice, hasMax := marketConfig.MaxPrices[listing.Type]

		if !hasMin || !hasMax {
			s.logger.Warn().
				Str("resourceType", listing.Type).
				Msg("Min or max price not defined in config, using defaults")
			minPrice = 100               // Default minimum
			maxPrice = listing.Price * 2 // Default maximum
		}

		// Calculate price fluctuation using config values
		fluctuationRange := marketConfig.PriceFluctuationRange
		if fluctuationRange <= 0 {
			fluctuationRange = 5 // Default 5%
		}

		// Calculate price change (-fluctuationRange to +fluctuationRange percent)
		priceChange := (rand.Float64()*float64(fluctuationRange*2) - float64(fluctuationRange)) / 100.0

		// Apply price change
		newPrice := float64(listing.Price) * (1.0 + priceChange)

		// Ensure price stays within configured min/max
		if newPrice < float64(minPrice) {
			newPrice = float64(minPrice)
		} else if newPrice > float64(maxPrice) {
			newPrice = float64(maxPrice)
		}

		// Update trend
		if priceChange > 0 {
			listing.Trend = util.PriceTrendUp
		} else if priceChange < 0 {
			listing.Trend = util.PriceTrendDown
		} else {
			listing.Trend = util.PriceTrendStable
		}

		listing.TrendPercentage = int(priceChange * 100.0)
		if listing.TrendPercentage < 0 {
			listing.TrendPercentage = -listing.TrendPercentage
		}

		listing.Price = int(newPrice)
		listing.UpdatedAt = time.Now()

		// Save the updated listing
		if err := s.marketRepo.UpdateListing(&listing); err != nil {
			s.logger.Error().Err(err).
				Str("resourceType", listing.Type).
				Msg("Failed to update market listing")
			continue
		}

		// Record price history
		history := model.MarketPriceHistory{
			ResourceType: listing.Type,
			Price:        listing.Price,
			Timestamp:    time.Now(),
			CreatedAt:    time.Now(),
		}

		if err := s.marketRepo.CreatePriceHistory(&history); err != nil {
			s.logger.Error().Err(err).
				Str("resourceType", listing.Type).
				Msg("Failed to create price history entry")
		}
	}

	return nil
}

// Helper functions for formatting notifications
func formatPurchaseNotification(quantity int, resourceType string, totalCost int) string {
	var resourceName string

	switch resourceType {
	case util.ResourceTypeCrew:
		if quantity == 1 {
			resourceName = "crew member"
		} else {
			resourceName = "crew members"
		}
	case util.ResourceTypeWeapons:
		if quantity == 1 {
			resourceName = "weapon"
		} else {
			resourceName = "weapons"
		}
	case util.ResourceTypeVehicles:
		if quantity == 1 {
			resourceName = "vehicle"
		} else {
			resourceName = "vehicles"
		}
	default:
		resourceName = "resources"
	}

	return formatMessage("Purchase completed", "You bought %d %s for $%s.", quantity, resourceName, formatMoney(totalCost))
}

func formatSaleNotification(quantity int, resourceType string, totalValue int) string {
	var resourceName string

	switch resourceType {
	case util.ResourceTypeCrew:
		if quantity == 1 {
			resourceName = "crew member"
		} else {
			resourceName = "crew members"
		}
	case util.ResourceTypeWeapons:
		if quantity == 1 {
			resourceName = "weapon"
		} else {
			resourceName = "weapons"
		}
	case util.ResourceTypeVehicles:
		if quantity == 1 {
			resourceName = "vehicle"
		} else {
			resourceName = "vehicles"
		}
	default:
		resourceName = "resources"
	}

	return formatMessage("Sale completed", "You sold %d %s for $%s.", quantity, resourceName, formatMoney(totalValue))
}
