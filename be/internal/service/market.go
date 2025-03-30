// internal/service/market.go

package service

import (
	"errors"
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
	gameConfig    config.GameConfig
	logger        zerolog.Logger
}

// NewMarketService creates a new market service
func NewMarketService(
	marketRepo repository.MarketRepository,
	playerRepo repository.PlayerRepository,
	playerService PlayerService,
	gameConfig config.GameConfig,
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

// SetPlayerService sets the player service (used to avoid circular dependencies)
// func (s *marketService) SetPlayerService(playerService PlayerService) {
// 	s.playerService = playerService
// }

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
	return s.marketRepo.UpdateMarketPrices()
}

// Helper functions

// formatPurchaseNotification formats a notification for a purchase
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

// formatSaleNotification formats a notification for a sale
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
