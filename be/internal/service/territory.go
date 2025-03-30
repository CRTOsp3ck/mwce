// internal/service/territory.go

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

// TerritoryService handles territory-related business logic
type TerritoryService interface {
	GetAllRegions() ([]model.Region, error)
	GetRegionByID(id string) (*model.Region, error)
	GetAllDistricts() ([]model.District, error)
	GetDistrictsByRegionID(regionID string) ([]model.District, error)
	GetDistrictByID(id string) (*model.District, error)
	GetAllCities() ([]model.City, error)
	GetCitiesByDistrictID(districtID string) ([]model.City, error)
	GetCityByID(id string) (*model.City, error)
	GetAllHotspots() ([]model.Hotspot, error)
	GetHotspotsByCity(cityID string) ([]model.Hotspot, error)
	GetHotspotByID(id string) (*model.Hotspot, error)
	GetControlledHotspots(playerID string) ([]model.Hotspot, error)
	GetRecentActions(playerID string) ([]model.TerritoryAction, error)
	PerformAction(playerID, actionType string, request model.PerformActionRequest) (*model.ActionResult, error)
	UpdateHotspotIncome() error
	CollectHotspotIncome(playerID, hotspotID string) (*model.CollectResponse, error)
	CollectAllHotspotIncome(playerID string) (*model.CollectAllResponse, error)

	// Scheduled jobs
	StartPeriodicIncomeGeneration()

	// TEMP!!!
	GetSSEService() SSEService
}

type territoryService struct {
	territoryRepo repository.TerritoryRepository
	playerRepo    repository.PlayerRepository
	sseService    SSEService
	gameConfig    config.GameConfig
	logger        zerolog.Logger
}

// NewTerritoryService creates a new territory service
func NewTerritoryService(
	territoryRepo repository.TerritoryRepository,
	playerRepo repository.PlayerRepository,
	sseService SSEService,
	gameConfig config.GameConfig,
	logger zerolog.Logger,
) TerritoryService {
	return &territoryService{
		territoryRepo: territoryRepo,
		playerRepo:    playerRepo,
		sseService:    sseService,
		gameConfig:    gameConfig,
		logger:        logger,
	}
}

// TEMP!!!
func (s *territoryService) GetSSEService() SSEService {
	return s.sseService
}

// GetAllRegions retrieves all regions
func (s *territoryService) GetAllRegions() ([]model.Region, error) {
	return s.territoryRepo.GetAllRegions()
}

// GetRegionByID retrieves a region by ID
func (s *territoryService) GetRegionByID(id string) (*model.Region, error) {
	return s.territoryRepo.GetRegionByID(id)
}

// GetAllDistricts retrieves all districts
func (s *territoryService) GetAllDistricts() ([]model.District, error) {
	return s.territoryRepo.GetAllDistricts()
}

// GetDistrictsByRegionID retrieves districts by region ID
func (s *territoryService) GetDistrictsByRegionID(regionID string) ([]model.District, error) {
	return s.territoryRepo.GetDistrictsByRegionID(regionID)
}

// GetDistrictByID retrieves a district by ID
func (s *territoryService) GetDistrictByID(id string) (*model.District, error) {
	return s.territoryRepo.GetDistrictByID(id)
}

// GetAllCities retrieves all cities
func (s *territoryService) GetAllCities() ([]model.City, error) {
	return s.territoryRepo.GetAllCities()
}

// GetCitiesByDistrictID retrieves cities by district ID
func (s *territoryService) GetCitiesByDistrictID(districtID string) ([]model.City, error) {
	return s.territoryRepo.GetCitiesByDistrictID(districtID)
}

// GetCityByID retrieves a city by ID
func (s *territoryService) GetCityByID(id string) (*model.City, error) {
	return s.territoryRepo.GetCityByID(id)
}

// GetAllHotspots retrieves all hotspots
func (s *territoryService) GetAllHotspots() ([]model.Hotspot, error) {
	return s.territoryRepo.GetAllHotspots()
}

// GetHotspotsByCity retrieves hotspots by city ID
func (s *territoryService) GetHotspotsByCity(cityID string) ([]model.Hotspot, error) {
	return s.territoryRepo.GetHotspotsByCity(cityID)
}

// GetHotspotByID retrieves a hotspot by ID
func (s *territoryService) GetHotspotByID(id string) (*model.Hotspot, error) {
	return s.territoryRepo.GetHotspotByID(id)
}

// GetControlledHotspots retrieves hotspots controlled by a player
func (s *territoryService) GetControlledHotspots(playerID string) ([]model.Hotspot, error) {
	return s.territoryRepo.GetControlledHotspots(playerID)
}

// GetRecentActions retrieves recent territory actions
func (s *territoryService) GetRecentActions(playerID string) ([]model.TerritoryAction, error) {
	return s.territoryRepo.GetRecentActionsByPlayer(playerID, 20)
}

// PerformAction performs a territory action
func (s *territoryService) PerformAction(playerID, actionType string, request model.PerformActionRequest) (*model.ActionResult, error) {
	// Get the player
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("failed to get player")
	}

	// Check if player has enough resources
	if player.Crew < request.Resources.Crew {
		return nil, errors.New("not enough crew members")
	}
	if player.Weapons < request.Resources.Weapons {
		return nil, errors.New("not enough weapons")
	}
	if player.Vehicles < request.Resources.Vehicles {
		return nil, errors.New("not enough vehicles")
	}

	// Get the hotspot
	hotspot, err := s.territoryRepo.GetHotspotByID(request.HotspotID)
	if err != nil {
		return nil, errors.New("hotspot not found")
	}

	// Initialize action and result
	action := &model.TerritoryAction{
		Type:      actionType,
		PlayerID:  playerID,
		HotspotID: request.HotspotID,
		Resources: request.Resources,
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}

	// Process the action based on type
	var result *model.ActionResult
	switch actionType {
	case util.TerritoryActionTypeExtortion:
		result, err = s.handleExtortion(player, hotspot, request.Resources)
	case util.TerritoryActionTypeTakeover:
		result, err = s.handleTakeover(player, hotspot, request.Resources)
	case util.TerritoryActionTypeCollection:
		result, err = s.handleCollection(player, hotspot, request.Resources)
	case util.TerritoryActionTypeDefend:
		result, err = s.handleDefend(player, hotspot, request.Resources)
	default:
		return nil, errors.New("invalid action type")
	}

	if err != nil {
		return nil, err
	}

	// Set the result on the action
	action.Result = result

	// Record the action
	if err := s.territoryRepo.AddTerritoryAction(action); err != nil {
		s.logger.Error().Err(err).Msg("Failed to record territory action")
	}

	return result, nil
}

// handleExtortion processes an extortion action
func (s *territoryService) handleExtortion(player *model.Player, hotspot *model.Hotspot, resources model.ActionResources) (*model.ActionResult, error) {
	// Validate the action
	if hotspot.IsLegal {
		return nil, errors.New("cannot extort legal businesses")
	}

	// Calculate success chance
	successChance := s.calculateSuccessChance(resources, 70, hotspot.DefenseStrength)

	// Roll for success
	success := rand.Float64()*100 < float64(successChance)

	// Initialize result
	result := &model.ActionResult{
		Success: success,
		Message: "",
	}

	// Deduct player resources used for the action
	resourceUpdates := map[string]int{
		"crew":     -resources.Crew,
		"weapons":  -resources.Weapons,
		"vehicles": -resources.Vehicles,
	}

	// Process the result based on success/failure
	if success {
		// Calculate money gained (based on business type and resources committed)
		baseGain := 500 + (rand.Intn(11) * 100) // $500-$1500 base
		resourceMultiplier := 1.0 + (float64(resources.Crew+resources.Weapons*2+resources.Vehicles*3) / 20.0)
		moneyGained := int(float64(baseGain) * resourceMultiplier)

		// Small chance to gain additional resources
		if rand.Intn(100) < 20 {
			// Potentially gain crew
			if rand.Intn(100) < 30 {
				crewGained := rand.Intn(2) + 1
				result.CrewGained = crewGained
				resourceUpdates["crew"] += crewGained
			}

			// Potentially gain weapons
			if rand.Intn(100) < 20 {
				weaponsGained := rand.Intn(2) + 1
				result.WeaponsGained = weaponsGained
				resourceUpdates["weapons"] += weaponsGained
			}

			// Very small chance to gain a vehicle
			if rand.Intn(100) < 5 {
				result.VehiclesGained = 1
				resourceUpdates["vehicles"] += 1
			}
		}

		// Add money gained
		result.MoneyGained = moneyGained
		resourceUpdates["money"] = moneyGained

		// Generate heat
		heatGenerated := 5 + rand.Intn(6) // 5-10 heat
		result.HeatGenerated = heatGenerated
		resourceUpdates["heat"] = heatGenerated

		// Generate respect
		respectGained := 1 + rand.Intn(3) // 1-3 respect
		result.RespectGained = respectGained
		resourceUpdates["respect"] = respectGained

		// Set success message
		result.Message = fmt.Sprintf("Extortion successful. You collected $%s from %s.", formatMoney(moneyGained), hotspot.Name)
	} else {
		// On failure, potential resource loss and higher heat

		// Chance to lose crew
		if rand.Intn(100) < 30 {
			crewLost := rand.Intn(resources.Crew) + 1
			if crewLost > resources.Crew {
				crewLost = resources.Crew
			}
			result.CrewLost = crewLost
			resourceUpdates["crew"] -= crewLost
		}

		// Chance to lose weapons
		if rand.Intn(100) < 20 && resources.Weapons > 0 {
			weaponsLost := rand.Intn(resources.Weapons) + 1
			if weaponsLost > resources.Weapons {
				weaponsLost = resources.Weapons
			}
			result.WeaponsLost = weaponsLost
			resourceUpdates["weapons"] -= weaponsLost
		}

		// Smaller chance to lose a vehicle
		if rand.Intn(100) < 10 && resources.Vehicles > 0 {
			vehiclesLost := 1
			result.VehiclesLost = vehiclesLost
			resourceUpdates["vehicles"] -= vehiclesLost
		}

		// Generate higher heat on failure
		heatGenerated := 8 + rand.Intn(8) // 8-15 heat
		result.HeatGenerated = heatGenerated
		resourceUpdates["heat"] = heatGenerated

		// Set failure message
		result.Message = fmt.Sprintf("Extortion failed. The owners of %s called the police.", hotspot.Name)
	}

	// Update player resources
	if err := s.updatePlayerResources(player.ID, resourceUpdates); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update player resources after extortion")
		return nil, errors.New("failed to update player resources")
	}

	// Add notification
	if err := s.addNotification(player.ID, result.Message, util.NotificationTypeTerritory); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add notification after extortion")
	}

	return result, nil
}

// handleTakeover processes a takeover action
func (s *territoryService) handleTakeover(player *model.Player, hotspot *model.Hotspot, resources model.ActionResources) (*model.ActionResult, error) {
	// Validate the action
	if !hotspot.IsLegal {
		return nil, errors.New("cannot take over illegal businesses")
	}

	// Calculate base success chance
	var baseSuccessChance int
	var defenseStrength int

	// If the hotspot is already controlled, takeover is harder
	if hotspot.ControllerID != nil {
		// Cannot take over your own hotspot
		if *hotspot.ControllerID == player.ID {
			return nil, errors.New("you already control this business")
		}

		baseSuccessChance = 50 // Harder to take from another player
		defenseStrength = hotspot.DefenseStrength
	} else {
		baseSuccessChance = 75 // Easier to take an uncontrolled business
		defenseStrength = 0    // No defense
	}

	// Calculate final success chance
	successChance := s.calculateSuccessChance(resources, baseSuccessChance, defenseStrength)

	// Roll for success
	success := rand.Float64()*100 < float64(successChance)

	// Initialize result
	result := &model.ActionResult{
		Success: success,
		Message: "",
	}

	// Deduct player resources used for the action
	resourceUpdates := map[string]int{
		"crew":     -resources.Crew,
		"weapons":  -resources.Weapons,
		"vehicles": -resources.Vehicles,
	}

	// Process the result based on success/failure
	if success {
		// Get previous controller ID (if any)
		previousControllerID := hotspot.ControllerID

		// Update the hotspot control
		hotspot.ControllerID = &player.ID
		hotspot.Crew = resources.Crew
		hotspot.Weapons = resources.Weapons
		hotspot.Vehicles = resources.Vehicles
		hotspot.DefenseStrength = (resources.Crew * 10) + (resources.Weapons * 15) + (resources.Vehicles * 20)

		// If there was a previous controller, they lose the resources allocated
		if previousControllerID != nil && *previousControllerID != player.ID {
			// Add notification to previous controller
			previousControllerMessage := fmt.Sprintf("Your business %s has been taken over by %s!", hotspot.Name, player.Name)
			if err := s.addNotification(*previousControllerID, previousControllerMessage, util.NotificationTypeTerritory); err != nil {
				s.logger.Error().Err(err).Msg("Failed to add notification to previous controller")
			}

			// Update previous controller's stats
			// In a complete implementation, we would update the previous controller's stats here
		}

		// Update the hotspot
		if err := s.territoryRepo.UpdateHotspot(hotspot); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update hotspot after takeover")
			return nil, errors.New("failed to update hotspot")
		}

		// Generate respect and influence
		respectGained := 3 + rand.Intn(3)   // 3-5 respect
		influenceGained := 2 + rand.Intn(3) // 2-4 influence
		result.RespectGained = respectGained
		result.InfluenceGained = influenceGained
		resourceUpdates["respect"] = respectGained
		resourceUpdates["influence"] = influenceGained

		// Generate heat
		heatGenerated := 3 + rand.Intn(5) // 3-7 heat
		result.HeatGenerated = heatGenerated
		resourceUpdates["heat"] = heatGenerated

		// Set success message
		if previousControllerID != nil {
			result.Message = fmt.Sprintf("Takeover successful! You now control %s.", hotspot.Name)
		} else {
			result.Message = fmt.Sprintf("You successfully took control of %s.", hotspot.Name)
		}

		// Update player stats
		stats, err := s.playerRepo.GetPlayerStats(player.ID)
		if err == nil {
			stats.SuccessfulTakeovers++
			stats.TotalHotspotsControlled++
			s.playerRepo.UpdatePlayerStats(stats)
		}
	} else {
		// On failure, lose resources and generate heat

		// Chance to lose additional crew
		if rand.Intn(100) < 40 {
			crewLost := rand.Intn(resources.Crew) + 1
			if crewLost > resources.Crew {
				crewLost = resources.Crew
			}
			result.CrewLost = crewLost
			resourceUpdates["crew"] -= crewLost
		}

		// Chance to lose additional weapons
		if rand.Intn(100) < 30 && resources.Weapons > 0 {
			weaponsLost := rand.Intn(resources.Weapons) + 1
			if weaponsLost > resources.Weapons {
				weaponsLost = resources.Weapons
			}
			result.WeaponsLost = weaponsLost
			resourceUpdates["weapons"] -= weaponsLost
		}

		// Chance to lose additional vehicles
		if rand.Intn(100) < 20 && resources.Vehicles > 0 {
			vehiclesLost := 1
			result.VehiclesLost = vehiclesLost
			resourceUpdates["vehicles"] -= vehiclesLost
		}

		// Generate heat
		heatGenerated := 5 + rand.Intn(6) // 5-10 heat
		result.HeatGenerated = heatGenerated
		resourceUpdates["heat"] = heatGenerated

		// Lose respect on failure
		respectLost := 1 + rand.Intn(2) // 1-2 respect
		result.RespectLost = respectLost
		resourceUpdates["respect"] = -respectLost

		// Set failure message
		if hotspot.ControllerID != nil {
			result.Message = fmt.Sprintf("Takeover failed. The defenders of %s fought back successfully.", hotspot.Name)

			// Notify the defender
			defenderMessage := fmt.Sprintf("You successfully defended %s from a takeover attempt by %s!", hotspot.Name, player.Name)
			if err := s.addNotification(*hotspot.ControllerID, defenderMessage, util.NotificationTypeTerritory); err != nil {
				s.logger.Error().Err(err).Msg("Failed to add notification to defender")
			}
		} else {
			result.Message = fmt.Sprintf("Takeover failed. The police intervened before you could secure %s.", hotspot.Name)
		}

		// Update player stats
		stats, err := s.playerRepo.GetPlayerStats(player.ID)
		if err == nil {
			stats.FailedTakeovers++
			s.playerRepo.UpdatePlayerStats(stats)
		}
	}

	// Update player resources
	if err := s.updatePlayerResources(player.ID, resourceUpdates); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update player resources after takeover")
		return nil, errors.New("failed to update player resources")
	}

	// Add notification
	if err := s.addNotification(player.ID, result.Message, util.NotificationTypeTerritory); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add notification after takeover")
	}

	return result, nil
}

// handleCollection processes a collection action
func (s *territoryService) handleCollection(player *model.Player, hotspot *model.Hotspot, resources model.ActionResources) (*model.ActionResult, error) {
	// Validate the action
	if !hotspot.IsLegal {
		return nil, errors.New("cannot collect from illegal businesses")
	}

	if hotspot.ControllerID == nil || *hotspot.ControllerID != player.ID {
		return nil, errors.New("you do not control this business")
	}

	if hotspot.PendingCollection <= 0 {
		return nil, errors.New("no pending collections available")
	}

	// Calculate success chance (higher amounts are riskier)
	baseSuccessChance := 95 - (hotspot.PendingCollection / 1000)
	if baseSuccessChance < 60 {
		baseSuccessChance = 60 // Minimum 60% chance
	}

	successChance := s.calculateSuccessChance(resources, baseSuccessChance, 0)

	// Roll for success
	success := rand.Float64()*100 < float64(successChance)

	// Initialize result
	result := &model.ActionResult{
		Success: success,
		Message: "",
	}

	// Deduct player resources used for the action
	resourceUpdates := map[string]int{
		"crew":     -resources.Crew,
		"weapons":  -resources.Weapons,
		"vehicles": -resources.Vehicles,
	}

	// Process the result based on success/failure
	if success {
		// Calculate money collected
		moneyGained := hotspot.PendingCollection

		// Reset pending collection
		hotspot.PendingCollection = 0
		hotspot.LastCollectionTime = func() *time.Time {
			now := time.Now()
			return &now
		}()

		// Update the hotspot
		if err := s.territoryRepo.UpdateHotspot(hotspot); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update hotspot after collection")
			return nil, errors.New("failed to update hotspot")
		}

		// Add money to player
		result.MoneyGained = moneyGained
		resourceUpdates["money"] = moneyGained

		// Small amount of heat generation
		heatGenerated := 1 + rand.Intn(3) // 1-3 heat
		result.HeatGenerated = heatGenerated
		resourceUpdates["heat"] = heatGenerated

		// Set success message
		result.Message = fmt.Sprintf("Collection successful. $%s added to your account.", formatMoney(moneyGained))
	} else {
		// On failure, lose money and generate heat

		// Calculate money lost (portion of pending collection)
		percentLost := 30 + rand.Intn(41) // 30-70%
		moneyLost := (hotspot.PendingCollection * percentLost) / 100

		// Reduce pending collection
		hotspot.PendingCollection -= moneyLost

		// Update the hotspot
		if err := s.territoryRepo.UpdateHotspot(hotspot); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update hotspot after failed collection")
			return nil, errors.New("failed to update hotspot")
		}

		result.MoneyLost = moneyLost

		// Chance to lose crew
		if rand.Intn(100) < 30 && resources.Crew > 0 {
			crewLost := rand.Intn(resources.Crew) + 1
			if crewLost > resources.Crew {
				crewLost = resources.Crew
			}
			result.CrewLost = crewLost
			resourceUpdates["crew"] -= crewLost
		}

		// Generate higher heat on failure
		heatGenerated := 5 + rand.Intn(6) // 5-10 heat
		result.HeatGenerated = heatGenerated
		resourceUpdates["heat"] = heatGenerated

		// Set failure message
		result.Message = fmt.Sprintf("Collection interrupted by police. $%s was lost.", formatMoney(moneyLost))
	}

	// Update player resources
	if err := s.updatePlayerResources(player.ID, resourceUpdates); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update player resources after collection")
		return nil, errors.New("failed to update player resources")
	}

	// Add notification
	if err := s.addNotification(player.ID, result.Message, util.NotificationTypeCollection); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add notification after collection")
	}

	return result, nil
}

// handleDefend processes a defend action
func (s *territoryService) handleDefend(player *model.Player, hotspot *model.Hotspot, resources model.ActionResources) (*model.ActionResult, error) {
	// Validate the action
	if !hotspot.IsLegal {
		return nil, errors.New("cannot defend illegal businesses")
	}

	if hotspot.ControllerID == nil || *hotspot.ControllerID != player.ID {
		return nil, errors.New("you do not control this business")
	}

	// Defense is always successful
	result := &model.ActionResult{
		Success: true,
		Message: "",
	}

	// Deduct player resources used for the action
	resourceUpdates := map[string]int{
		"crew":     -resources.Crew,
		"weapons":  -resources.Weapons,
		"vehicles": -resources.Vehicles,
	}

	// Add resources to hotspot defense
	hotspot.Crew += resources.Crew
	hotspot.Weapons += resources.Weapons
	hotspot.Vehicles += resources.Vehicles

	// Calculate new defense strength
	hotspot.DefenseStrength = (hotspot.Crew * 10) + (hotspot.Weapons * 15) + (hotspot.Vehicles * 20)

	// Update the hotspot
	if err := s.territoryRepo.UpdateHotspot(hotspot); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update hotspot defense")
		return nil, errors.New("failed to update hotspot")
	}

	// Update player resources
	if err := s.updatePlayerResources(player.ID, resourceUpdates); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update player resources after defense allocation")
		return nil, errors.New("failed to update player resources")
	}

	// Set success message
	result.Message = fmt.Sprintf("Defense reinforced for %s. Current defense strength: %d", hotspot.Name, hotspot.DefenseStrength)

	// Add notification
	if err := s.addNotification(player.ID, result.Message, util.NotificationTypeTerritory); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add notification after defend action")
	}

	return result, nil
}

// CollectHotspotIncome collects pending income from a specific hotspot
func (s *territoryService) CollectHotspotIncome(playerID, hotspotID string) (*model.CollectResponse, error) {
	// Verify player owns the hotspot
	hotspot, err := s.territoryRepo.GetHotspotByID(hotspotID)
	if err != nil {
		return nil, err
	}

	if hotspot.ControllerID == nil || *hotspot.ControllerID != playerID {
		return nil, errors.New("you do not control this hotspot")
	}

	if hotspot.PendingCollection <= 0 {
		return nil, errors.New("no resources available to collect")
	}

	// Get the collected amount
	collectedAmount := hotspot.PendingCollection

	// Reset pending collection
	hotspot.PendingCollection = 0
	hotspot.LastCollectionTime = func() *time.Time {
		now := time.Now()
		return &now
	}()

	// Update the hotspot
	if err := s.territoryRepo.UpdateHotspot(hotspot); err != nil {
		s.logger.Error().Err(err).
			Str("hotspotID", hotspotID).
			Msg("Failed to update hotspot after collection")
		return nil, errors.New("failed to update hotspot")
	}

	// Update player's money
	if err := s.playerRepo.UpdatePlayerResource(playerID, "money", collectedAmount); err != nil {
		s.logger.Error().Err(err).
			Str("playerID", playerID).
			Str("resourceType", "money").
			Int("amount", collectedAmount).
			Msg("Failed to update player money after collection")
		return nil, errors.New("failed to update player resources")
	}

	// Generate message
	message := fmt.Sprintf("Successfully collected $%s from %s.", formatMoney(collectedAmount), hotspot.Name)

	// Add notification to player
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   message,
		Type:      util.NotificationTypeCollection,
		Timestamp: time.Now(),
		Read:      false,
	}
	if err := s.playerRepo.AddNotification(notification); err != nil {
		s.logger.Error().Err(err).Msg("Failed to add collection notification")
	}

	return &model.CollectResponse{
		HotspotID:       hotspotID,
		HotspotName:     hotspot.Name,
		CollectedAmount: collectedAmount,
		Message:         message,
	}, nil
}

// CollectAllHotspotIncome collects pending income from all hotspots controlled by a player
func (s *territoryService) CollectAllHotspotIncome(playerID string) (*model.CollectAllResponse, error) {
	// Get all controlled hotspots
	hotspots, err := s.territoryRepo.GetControlledHotspots(playerID)
	if err != nil {
		return nil, err
	}

	totalCollected := 0
	collectedHotspots := 0

	// Collect from each hotspot
	for _, hotspot := range hotspots {
		if hotspot.PendingCollection > 0 {
			// Reset pending collection
			collectedAmount := hotspot.PendingCollection
			totalCollected += collectedAmount
			collectedHotspots++

			hotspot.PendingCollection = 0
			hotspot.LastCollectionTime = func() *time.Time {
				now := time.Now()
				return &now
			}()

			// Update the hotspot
			if err := s.territoryRepo.UpdateHotspot(&hotspot); err != nil {
				s.logger.Error().Err(err).
					Str("hotspotID", hotspot.ID).
					Msg("Failed to update hotspot after collection")
				// Continue with others even if one fails
				continue
			}
		}
	}

	// Update player's money
	if totalCollected > 0 {
		if err := s.playerRepo.UpdatePlayerResource(playerID, "money", totalCollected); err != nil {
			s.logger.Error().Err(err).
				Str("playerID", playerID).
				Str("resourceType", "money").
				Int("amount", totalCollected).
				Msg("Failed to update player money after collection")
			return nil, errors.New("failed to update player resources")
		}
	}

	// Generate response message
	var message string
	if totalCollected > 0 {
		message = fmt.Sprintf("Successfully collected $%s from %d businesses.", formatMoney(totalCollected), collectedHotspots)
	} else {
		message = "No resources available to collect at this time."
	}

	// Add notification to player if resources were collected
	if totalCollected > 0 {
		notification := &model.Notification{
			PlayerID:  playerID,
			Message:   message,
			Type:      util.NotificationTypeCollection,
			Timestamp: time.Now(),
			Read:      false,
		}
		if err := s.playerRepo.AddNotification(notification); err != nil {
			s.logger.Error().Err(err).Msg("Failed to add collection notification")
		}
	}

	return &model.CollectAllResponse{
		CollectedAmount: totalCollected,
		HotspotsCount:   collectedHotspots,
		Message:         message,
	}, nil
}

// Update the UpdateHotspotIncome method to send SSE events
func (s *territoryService) UpdateHotspotIncome() error {
	// Get all legal hotspots with controllers
	hotspots, err := s.territoryRepo.GetAllControlledLegalHotspots()
	if err != nil {
		return err
	}

	currentTime := time.Now()

	// Map to collect income updates by player
	playerIncomeUpdates := make(map[string][]map[string]interface{})

	// Update each hotspot's pending collection
	for _, hotspot := range hotspots {
		if hotspot.ControllerID == nil {
			continue
		}

		playerID := *hotspot.ControllerID

		// Calculate time since last income generation
		timeSinceLastIncome := func() time.Duration {
			if hotspot.LastIncomeTime != nil {
				return currentTime.Sub(*hotspot.LastIncomeTime)
			}

			// NOTE: This should be time since last takeover
			return currentTime.Sub(time.Now())
		}()
		hoursSinceLastIncome := timeSinceLastIncome.Hours()

		// Only add income if at least one hour has passed
		if hoursSinceLastIncome >= 1.0 {
			// Calculate number of full hours that have passed
			fullHoursPassed := int(hoursSinceLastIncome)

			// Calculate new income (hourly rate * hours elapsed)
			newIncome := hotspot.Income * fullHoursPassed

			// Add new income to pending collection
			if err := s.territoryRepo.UpdateHotspotPendingCollection(hotspot.ID, newIncome); err != nil {
				s.logger.Error().Err(err).
					Str("hotspotID", hotspot.ID).
					Int("income", newIncome).
					Msg("Failed to update hotspot pending collection")
				continue
			}

			// Update last income time to the current time minus any partial hour
			partialHour := timeSinceLastIncome - time.Duration(fullHoursPassed)*time.Hour
			newLastIncomeTime := currentTime.Add(-partialHour)

			if err := s.territoryRepo.UpdateHotspotLastIncomeTime(hotspot.ID, newLastIncomeTime); err != nil {
				s.logger.Error().Err(err).
					Str("hotspotID", hotspot.ID).
					Msg("Failed to update hotspot last income time")
			}

			// Collect income update for SSE event
			// Get the updated hotspot
			updatedHotspot, err := s.territoryRepo.GetHotspotByID(hotspot.ID)
			if err != nil {
				s.logger.Error().Err(err).
					Str("hotspotID", hotspot.ID).
					Msg("Failed to get updated hotspot")
				continue
			}

			// Add to player updates
			if _, ok := playerIncomeUpdates[playerID]; !ok {
				playerIncomeUpdates[playerID] = make([]map[string]interface{}, 0)
			}

			playerIncomeUpdates[playerID] = append(playerIncomeUpdates[playerID], map[string]interface{}{
				"hotspotId":         updatedHotspot.ID,
				"hotspotName":       updatedHotspot.Name,
				"newIncome":         newIncome,
				"pendingCollection": updatedHotspot.PendingCollection,
				"lastIncomeTime":    updatedHotspot.LastIncomeTime,
				"nextIncomeTime":    updatedHotspot.LastIncomeTime.Add(time.Hour),
			})

			// If significant amount accumulated, send notification to player
			if newIncome > 1000 {
				message := fmt.Sprintf("$%s is ready for collection at %s.", formatMoney(newIncome), hotspot.Name)
				if err := s.addNotification(playerID, message, util.NotificationTypeCollection); err != nil {
					s.logger.Error().Err(err).Msg("Failed to add collection notification")
				}
			}
		}
	}

	// Send SSE events to players
	for playerID, updates := range playerIncomeUpdates {
		// Calculate total pending collections
		totalPending, err := s.playerRepo.CalculatePendingCollections(playerID)
		if err != nil {
			s.logger.Error().Err(err).
				Str("playerID", playerID).
				Msg("Failed to calculate pending collections")
			continue
		}

		// Send the updates via SSE
		s.sseService.SendEventToPlayer(playerID, "income_generated", map[string]interface{}{
			"updates":      updates,
			"totalPending": totalPending,
			"timestamp":    currentTime,
		})
	}

	return nil
}

// calculateSuccessChance calculates the success chance for an action
func (s *territoryService) calculateSuccessChance(resources model.ActionResources, baseChance, opponentStrength int) int {
	// Calculate player strength
	playerStrength := (resources.Crew * 10) + (resources.Weapons * 15) + (resources.Vehicles * 20)

	// Base success rate
	successChance := baseChance

	// Adjust for resources committed
	if playerStrength > 0 {
		if opponentStrength > 0 {
			// For actions against opponents, compare strengths
			strengthRatio := float64(playerStrength) / float64(opponentStrength)
			if strengthRatio >= 2.0 {
				successChance += 20 // Major advantage
			} else if strengthRatio >= 1.5 {
				successChance += 15 // Significant advantage
			} else if strengthRatio >= 1.0 {
				successChance += 10 // Slight advantage
			} else if strengthRatio >= 0.75 {
				successChance += 5 // Nearly even
			} else if strengthRatio >= 0.5 {
				successChance -= 5 // Disadvantage
			} else if strengthRatio >= 0.25 {
				successChance -= 10 // Major disadvantage
			} else {
				successChance -= 20 // Severe disadvantage
			}
		} else {
			// For actions without opposition, just add based on strength
			if playerStrength >= 100 {
				successChance += 20
			} else if playerStrength >= 75 {
				successChance += 15
			} else if playerStrength >= 50 {
				successChance += 10
			} else if playerStrength >= 25 {
				successChance += 5
			}
		}
	}

	// Cap success chance between 5% and 95%
	if successChance < 5 {
		successChance = 5
	} else if successChance > 95 {
		successChance = 95
	}

	return successChance
}

// updatePlayerResources updates multiple resources for a player
func (s *territoryService) updatePlayerResources(playerID string, resourceUpdates map[string]int) error {
	for resourceType, amount := range resourceUpdates {
		if err := s.playerRepo.UpdatePlayerResource(playerID, resourceType, amount); err != nil {
			s.logger.Error().Err(err).
				Str("playerID", playerID).
				Str("resourceType", resourceType).
				Int("amount", amount).
				Msg("Failed to update player resource")
			return err
		}
	}

	// After updating resources, we should update the player's title based on new stats
	if err := s.updatePlayerTitle(playerID); err != nil {
		s.logger.Error().Err(err).Str("playerID", playerID).Msg("Failed to update player title")
	}

	return nil
}

// addNotification adds a notification for a player
func (s *territoryService) addNotification(playerID, message, notificationType string) error {
	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   message,
		Type:      notificationType,
		Timestamp: time.Now(),
		Read:      false,
	}
	return s.playerRepo.AddNotification(notification)
}

// updatePlayerTitle updates a player's title based on their respect and influence
func (s *territoryService) updatePlayerTitle(playerID string) error {
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return err
	}

	// Determine the new title based on respect and influence
	newTitle := util.PlayerTitleAssociate

	respectInfluence := player.Respect + player.Influence

	if respectInfluence >= 20 && respectInfluence < 40 {
		newTitle = util.PlayerTitleSoldier
	} else if respectInfluence >= 40 && respectInfluence < 60 {
		newTitle = util.PlayerTitleCapo
	} else if respectInfluence >= 60 && respectInfluence < 80 {
		newTitle = util.PlayerTitleUnderboss
	} else if respectInfluence >= 80 && respectInfluence < 100 {
		newTitle = util.PlayerTitleConsigliere
	} else if respectInfluence >= 100 && respectInfluence < 150 {
		newTitle = util.PlayerTitleBoss
	} else if respectInfluence >= 150 {
		newTitle = util.PlayerTitleGodfather
	}

	// Only update if the title has changed
	if player.Title != newTitle {
		player.Title = newTitle
		if err := s.playerRepo.UpdatePlayer(player); err != nil {
			return err
		}

		// Add a notification about the title change
		message := fmt.Sprintf("Congratulations! Your criminal influence has earned you the title of %s.", newTitle)
		if err := s.addNotification(playerID, message, util.NotificationTypeSystem); err != nil {
			s.logger.Error().Err(err).Msg("Failed to add title change notification")
		}
	}

	return nil
}
