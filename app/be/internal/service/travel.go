// internal/service/travel.go

package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
)

// TravelService handles travel-related business logic
type TravelService interface {
	// Travel to a specific region
	Travel(playerID string, regionID string) (*model.TravelResponse, error)

	// Get available regions for travel
	GetAvailableRegions(playerID string) ([]model.Region, error)

	// Get travel history for a player
	GetTravelHistory(playerID string, limit int) ([]model.TravelAttempt, error)

	// Get current region for a player
	GetCurrentRegion(playerID string) (*model.Region, error)
}

type travelService struct {
	playerRepo    repository.PlayerRepository
	territoryRepo repository.TerritoryRepository
	sseService    SSEService
	gameConfig    config.GameConfig
	logger        zerolog.Logger
}

// NewTravelService creates a new travel service
func NewTravelService(
	playerRepo repository.PlayerRepository,
	territoryRepo repository.TerritoryRepository,
	sseService SSEService,
	gameConfig config.GameConfig,
	logger zerolog.Logger,
) TravelService {
	return &travelService{
		playerRepo:    playerRepo,
		territoryRepo: territoryRepo,
		sseService:    sseService,
		gameConfig:    gameConfig,
		logger:        logger,
	}
}

// Travel handles a player's attempt to travel to a new region
func (s *travelService) Travel(playerID string, regionID string) (*model.TravelResponse, error) {
	// Get the player
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	// Get the destination region
	destRegion, err := s.territoryRepo.GetRegionByID(regionID)
	if err != nil {
		return nil, errors.New("destination region not found")
	}

	// Get the player's current region if any
	var fromRegionID *string
	var fromRegion *model.Region
	if player.CurrentRegionID != nil {
		fromRegionID = player.CurrentRegionID
		fromRegion, _ = s.territoryRepo.GetRegionByID(*player.CurrentRegionID)

		// Check if player is already in the requested region
		if *player.CurrentRegionID == regionID {
			return &model.TravelResponse{
				Success:    true,
				RegionID:   regionID,
				RegionName: destRegion.Name,
				TravelCost: 0,
				Message:    fmt.Sprintf("You are already in %s.", destRegion.Name),
			}, nil
		}
	}

	// Calculate travel cost - could be based on distance, but we'll use a simple fixed cost for now
	// In a more advanced implementation, this could be based on a distance matrix between regions
	baseTravelCost := s.gameConfig.Mechanics.Travel.BaseCost
	travelCost := baseTravelCost

	// Check if player has enough money
	if player.Money < travelCost {
		return nil, errors.New("not enough money to travel")
	}

	// Calculate chance of getting caught based on player's heat level
	// Higher heat = higher chance of getting caught
	baseCatchChance := s.gameConfig.Mechanics.Travel.BaseCatchChance
	heatMultiplier := s.gameConfig.Mechanics.Travel.HeatMultiplier

	catchChance := baseCatchChance + (float64(player.Heat) * heatMultiplier)

	// Cap the catch chance at a maximum value (e.g., 75%)
	maxCatchChance := s.gameConfig.Mechanics.Travel.MaxCatchChance
	if catchChance > maxCatchChance {
		catchChance = maxCatchChance
	}

	// Determine if player gets caught
	caughtByPolice := rand.Float64()*100 < catchChance

	// Initialize travel attempt
	travelAttempt := &model.TravelAttempt{
		PlayerID:       playerID,
		FromRegionID:   fromRegionID,
		ToRegionID:     regionID,
		Success:        !caughtByPolice,
		CaughtByPolice: caughtByPolice,
		TravelCost:     travelCost,
		Timestamp:      time.Now(),
		CreatedAt:      time.Now(),
	}

	// Build response
	response := &model.TravelResponse{
		Success:        !caughtByPolice,
		RegionID:       regionID,
		RegionName:     destRegion.Name,
		TravelCost:     travelCost,
		CaughtByPolice: caughtByPolice,
	}

	// Handle the travel outcome
	if caughtByPolice {
		// Player got caught - apply penalties
		baseFineFactor := s.gameConfig.Mechanics.Travel.BaseFineFactor
		heatIncrease := s.gameConfig.Mechanics.Travel.CaughtHeatIncrease

		// Calculate fine - a percentage of the player's money, with a minimum amount
		finePercent := baseFineFactor
		fineAmount := int(float64(player.Money) * finePercent)
		minFine := s.gameConfig.Mechanics.Travel.MinimumFine

		if fineAmount < minFine {
			fineAmount = minFine
		}

		// Cap the fine to prevent wiping out the player
		maxFinePercent := s.gameConfig.Mechanics.Travel.MaxFinePercent
		maxFine := int(float64(player.Money) * maxFinePercent)
		if fineAmount > maxFine {
			fineAmount = maxFine
		}

		// Ensure we don't take more money than the player has
		if fineAmount > player.Money {
			fineAmount = player.Money
		}

		// Apply the penalties
		player.Money -= fineAmount
		player.Heat += heatIncrease

		// Update the travel attempt
		travelAttempt.FineAmount = fineAmount
		travelAttempt.HeatChange = heatIncrease

		// Update the response
		response.FineAmount = fineAmount
		response.HeatIncrease = heatIncrease
		response.Message = fmt.Sprintf("You were caught by the police while trying to travel to %s. You've been fined $%d and your heat has increased by %d.", destRegion.Name, fineAmount, heatIncrease)

		// Player stays in current region
	} else {
		// Successful travel - apply costs and benefits
		player.Money -= travelCost

		// Heat reduction for successful travel
		heatReduction := s.gameConfig.Mechanics.Travel.SuccessHeatReduction

		// Ensure heat doesn't go negative
		if heatReduction > player.Heat {
			heatReduction = player.Heat
		}

		player.Heat -= heatReduction
		player.CurrentRegionID = &regionID
		player.LastTravelTime = ptrTime(time.Now())

		// Update the travel attempt
		travelAttempt.HeatChange = -heatReduction

		// Update the response
		response.HeatReduction = heatReduction

		// Create a nice message
		fromText := "headquarters"
		if fromRegion != nil {
			fromText = fromRegion.Name
		}

		response.Message = fmt.Sprintf("You have successfully traveled from %s to %s for $%d. Your heat has decreased by %d.", fromText, destRegion.Name, travelCost, heatReduction)
	}

	// Save the travel attempt
	err = s.playerRepo.CreateTravelAttempt(travelAttempt)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to save travel attempt")
	}

	// Update player
	err = s.playerRepo.UpdatePlayer(player)
	if err != nil {
		return nil, errors.New("failed to update player after travel")
	}

	// Create notification
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   response.Message,
		Type:      util.NotificationTypeTravel,
		Timestamp: time.Now(),
		Read:      false,
	}

	// Save notification
	_ = s.playerRepo.AddNotification(notification)

	// Send SSE notification only if travel was successful
	if response.Success {
		s.logger.Info().
			Str("playerID", playerID).
			Str("regionID", regionID).
			Msg("Sending travel success SSE notification")

		// Send the notification
		s.sseService.SendEventToPlayer(playerID, "player_region_changed", map[string]interface{}{
			"event":      "player_region_changed",
			"playerId":   playerID,
			"regionId":   regionID,
			"regionName": destRegion.Name,
			"timestamp":  time.Now().Format(time.RFC3339),
		})
	}

	return response, nil
}

// GetAvailableRegions returns the list of regions available for travel
func (s *travelService) GetAvailableRegions(playerID string) ([]model.Region, error) {
	// In a simple implementation, all regions are available for travel
	// In a more complex implementation, regions might be unlocked based on player progress

	return s.territoryRepo.GetAllRegions()
}

// GetTravelHistory retrieves the travel history for a player
func (s *travelService) GetTravelHistory(playerID string, limit int) ([]model.TravelAttempt, error) {
	return s.playerRepo.GetTravelHistory(playerID, limit)
}

// GetCurrentRegion returns the current region of a player
func (s *travelService) GetCurrentRegion(playerID string) (*model.Region, error) {
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	if player.CurrentRegionID == nil {
		return nil, errors.New("player has no current region")
	}

	return s.territoryRepo.GetRegionByID(*player.CurrentRegionID)
}
