// internal/service/operations.go

package service

import (
	"errors"
	"fmt"
	"math/rand"
	"mwce-be/internal/util"
	"time"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// OperationsService handles operations-related business logic
type OperationsService interface {
	GetAvailableOperations(playerID string) ([]model.Operation, error)
	GetOperationByID(id string) (*model.Operation, error)
	GetCurrentOperations(playerID string) ([]model.OperationAttempt, error)
	GetCompletedOperations(playerID string) ([]model.OperationAttempt, error)
	StartOperation(playerID, operationID string, resources model.OperationResources) (*model.OperationAttempt, error)
	CancelOperation(playerID, attemptID string) error
	CollectOperation(playerID, attemptID string) (*model.OperationResult, error)
	RefreshDailyOperations() error
	CheckAndCompleteOperations() error

	// Scheduled jobs
	StartPeriodicOperationsRefresh()
}

type operationsService struct {
	operationsRepo repository.OperationsRepository
	playerRepo     repository.PlayerRepository
	playerService  PlayerService
	gameConfig     config.GameConfig
	logger         zerolog.Logger
}

// NewOperationsService creates a new operations service
func NewOperationsService(
	operationsRepo repository.OperationsRepository,
	playerRepo repository.PlayerRepository,
	playerService PlayerService,
	gameConfig config.GameConfig,
	logger zerolog.Logger,
) OperationsService {
	return &operationsService{
		operationsRepo: operationsRepo,
		playerRepo:     playerRepo,
		playerService:  playerService,
		gameConfig:     gameConfig,
		logger:         logger,
	}
}

// GetAvailableOperations retrieves available operations for a player
func (s *operationsService) GetAvailableOperations(playerID string) ([]model.Operation, error) {
	// Get all operations
	operations, err := s.operationsRepo.GetAllOperations()
	if err != nil {
		return nil, err
	}

	// Get the player
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, err
	}

	// Filter operations based on requirements
	var availableOperations []model.Operation
	for _, op := range operations {
		// Check if it's available
		if op.AvailableUntil.Before(time.Now()) {
			continue
		}

		// For special operations, check requirements
		if op.IsSpecial {
			// Check influence
			if op.Requirements.MinInfluence > 0 && player.Influence < op.Requirements.MinInfluence {
				continue
			}

			// Check heat
			if op.Requirements.MaxHeat > 0 && player.Heat > op.Requirements.MaxHeat {
				continue
			}

			// Check title
			if op.Requirements.MinTitle != "" {
				// Skip if player title doesn't meet the requirement
				if !meetsMinimumTitle(player.Title, op.Requirements.MinTitle) {
					continue
				}
			}
		}

		availableOperations = append(availableOperations, op)
	}

	return availableOperations, nil
}

// GetOperationByID retrieves an operation by ID
func (s *operationsService) GetOperationByID(id string) (*model.Operation, error) {
	return s.operationsRepo.GetOperationByID(id)
}

// GetCurrentOperations retrieves in-progress operations for a player
func (s *operationsService) GetCurrentOperations(playerID string) ([]model.OperationAttempt, error) {
	return s.operationsRepo.GetCurrentOperations(playerID)
}

// GetCompletedOperations retrieves completed operations for a player
func (s *operationsService) GetCompletedOperations(playerID string) ([]model.OperationAttempt, error) {
	return s.operationsRepo.GetCompletedOperations(playerID)
}

// StartOperation starts a new operation
func (s *operationsService) StartOperation(playerID, operationID string, resources model.OperationResources) (*model.OperationAttempt, error) {
	// Get the operation
	operation, err := s.operationsRepo.GetOperationByID(operationID)
	if err != nil {
		return nil, errors.New("operation not found")
	}

	// Check if operation is still available
	if operation.AvailableUntil.Before(time.Now()) {
		return nil, errors.New("operation is no longer available")
	}

	// Get the player
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, errors.New("player not found")
	}

	// Check if the player meets the requirements for special operations
	if operation.IsSpecial {
		// Check influence
		if operation.Requirements.MinInfluence > 0 && player.Influence < operation.Requirements.MinInfluence {
			return nil, errors.New("insufficient influence for this operation")
		}

		// Check heat
		if operation.Requirements.MaxHeat > 0 && player.Heat > operation.Requirements.MaxHeat {
			return nil, errors.New("heat level too high for this operation")
		}

		// Check title
		if operation.Requirements.MinTitle != "" {
			if !meetsMinimumTitle(player.Title, operation.Requirements.MinTitle) {
				return nil, errors.New("your title rank is too low for this operation")
			}
		}
	}

	// Check if the player has enough resources
	if player.Crew < resources.Crew {
		return nil, errors.New("not enough crew members")
	}
	if player.Weapons < resources.Weapons {
		return nil, errors.New("not enough weapons")
	}
	if player.Vehicles < resources.Vehicles {
		return nil, errors.New("not enough vehicles")
	}
	if resources.Money > 0 && player.Money < resources.Money {
		return nil, errors.New("not enough money")
	}

	// Deduct resources from player
	resourceUpdates := map[string]int{
		"crew":     -resources.Crew,
		"weapons":  -resources.Weapons,
		"vehicles": -resources.Vehicles,
	}

	if resources.Money > 0 {
		resourceUpdates["money"] = -resources.Money
	}

	if err := s.playerService.UpdatePlayerResources(playerID, resourceUpdates); err != nil {
		return nil, errors.New("failed to update player resources")
	}

	// Create operation attempt
	attempt := &model.OperationAttempt{
		ID:          uuid.New().String(),
		OperationID: operationID,
		PlayerID:    playerID,
		Timestamp:   time.Now(),
		Resources:   resources,
		Status:      util.OperationStatusInProgress,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save the attempt
	if err := s.operationsRepo.CreateOperationAttempt(attempt); err != nil {
		// If we fail to save, try to refund resources
		refundUpdates := map[string]int{
			"crew":     resources.Crew,
			"weapons":  resources.Weapons,
			"vehicles": resources.Vehicles,
		}

		if resources.Money > 0 {
			refundUpdates["money"] = resources.Money
		}
 
		s.playerService.UpdatePlayerResources(playerID, refundUpdates)

		return nil, errors.New("failed to start operation")
	}

	// Add notification
	message := fmt.Sprintf("Operation '%s' started. Check back in %s for results.",
		operation.Name, formatDuration(operation.Duration))
	s.playerService.AddNotification(playerID, message, util.NotificationTypeOperation)

	return attempt, nil
}

// CancelOperation cancels an in-progress operation
func (s *operationsService) CancelOperation(playerID, attemptID string) error {
	// Get the operation attempt
	attempt, err := s.operationsRepo.GetOperationAttemptByID(attemptID)
	if err != nil {
		return errors.New("operation attempt not found")
	}

	// Check if the attempt belongs to the player
	if attempt.PlayerID != playerID {
		return errors.New("not authorized to cancel this operation")
	}

	// Check if the attempt is in progress
	if attempt.Status != util.OperationStatusInProgress {
		return errors.New("can only cancel in-progress operations")
	}

	// Update attempt status
	attempt.Status = util.OperationStatusCancelled
	attempt.CompletionTime = ptrTime(time.Now())
	attempt.UpdatedAt = time.Now()

	// Save the changes
	if err := s.operationsRepo.UpdateOperationAttempt(attempt); err != nil {
		return errors.New("failed to cancel operation")
	}

	// Refund a portion of the resources (50%)
	refundUpdates := map[string]int{
		"crew":     attempt.Resources.Crew / 2,
		"weapons":  attempt.Resources.Weapons / 2,
		"vehicles": attempt.Resources.Vehicles / 2,
	}

	if attempt.Resources.Money > 0 {
		refundUpdates["money"] = attempt.Resources.Money / 2
	}

	if err := s.playerService.UpdatePlayerResources(playerID, refundUpdates); err != nil {
		s.logger.Error().Err(err).Msg("Failed to refund resources after cancellation")
	}

	// Add notification
	operation, _ := s.operationsRepo.GetOperationByID(attempt.OperationID)
	operationName := "Unknown operation"
	if operation != nil {
		operationName = operation.Name
	}

	message := fmt.Sprintf("Operation '%s' cancelled. 50%% of committed resources have been returned.", operationName)
	s.playerService.AddNotification(playerID, message, util.NotificationTypeOperation)

	return nil
}

// CollectOperation completes an operation and collects the results
func (s *operationsService) CollectOperation(playerID, attemptID string) (*model.OperationResult, error) {
	// Get the operation attempt
	attempt, err := s.operationsRepo.GetOperationAttemptByID(attemptID)
	if err != nil {
		return nil, errors.New("operation attempt not found")
	}

	// Check if the attempt belongs to the player
	if attempt.PlayerID != playerID {
		return nil, errors.New("not authorized to collect this operation")
	}

	// Check if the attempt is in progress
	if attempt.Status != util.OperationStatusInProgress {
		return nil, errors.New("can only collect in-progress operations")
	}

	// Get the operation
	operation, err := s.operationsRepo.GetOperationByID(attempt.OperationID)
	if err != nil {
		return nil, errors.New("operation not found")
	}

	// Check if the operation has been running long enough
	timeSinceStart := time.Since(attempt.Timestamp)
	if timeSinceStart.Seconds() < float64(operation.Duration) {
		return nil, errors.New("operation is still in progress")
	}

	// Determine success or failure
	successChance := s.calculateSuccessChance(operation, attempt.Resources)
	success := rand.Float64()*100 < float64(successChance)

	// Generate result
	result := &model.OperationResult{
		Success: success,
		Message: "",
	}

	// Process rewards or losses based on success
	resourceUpdates := make(map[string]int)

	if success {
		// Rewards
		if operation.Rewards.Money > 0 {
			moneyGained := operation.Rewards.Money
			result.MoneyGained = moneyGained
			resourceUpdates["money"] = moneyGained
		}

		if operation.Rewards.Crew > 0 {
			crewGained := operation.Rewards.Crew
			result.CrewGained = crewGained
			resourceUpdates["crew"] = crewGained
		}

		if operation.Rewards.Weapons > 0 {
			weaponsGained := operation.Rewards.Weapons
			result.WeaponsGained = weaponsGained
			resourceUpdates["weapons"] = weaponsGained
		}

		if operation.Rewards.Vehicles > 0 {
			vehiclesGained := operation.Rewards.Vehicles
			result.VehiclesGained = vehiclesGained
			resourceUpdates["vehicles"] = vehiclesGained
		}

		if operation.Rewards.Respect > 0 {
			respectGained := operation.Rewards.Respect
			result.RespectGained = respectGained
			resourceUpdates["respect"] = respectGained
		}

		if operation.Rewards.Influence > 0 {
			influenceGained := operation.Rewards.Influence
			result.InfluenceGained = influenceGained
			resourceUpdates["influence"] = influenceGained
		}

		if operation.Rewards.HeatReduction > 0 {
			heatReduced := operation.Rewards.HeatReduction
			result.HeatReduced = heatReduced
			resourceUpdates["heat"] = -heatReduced
		}

		// Success message
		result.Message = fmt.Sprintf("Operation successful! %s", getSuccessMessage(operation.Type))
	} else {
		// Losses
		if operation.Risks.CrewLoss > 0 {
			crewLost := rand.Intn(operation.Risks.CrewLoss) + 1
			result.CrewLost = crewLost
			resourceUpdates["crew"] = -crewLost
		}

		if operation.Risks.WeaponsLoss > 0 {
			weaponsLost := rand.Intn(operation.Risks.WeaponsLoss) + 1
			result.WeaponsLost = weaponsLost
			resourceUpdates["weapons"] = -weaponsLost
		}

		if operation.Risks.VehiclesLoss > 0 {
			vehiclesLost := rand.Intn(operation.Risks.VehiclesLoss) + 1
			result.VehiclesLost = vehiclesLost
			resourceUpdates["vehicles"] = -vehiclesLost
		}

		if operation.Risks.MoneyLoss > 0 {
			moneyLost := operation.Risks.MoneyLoss
			result.MoneyLost = moneyLost
			resourceUpdates["money"] = -moneyLost
		}

		if operation.Risks.HeatIncrease > 0 {
			heatGenerated := operation.Risks.HeatIncrease
			result.HeatGenerated = heatGenerated
			resourceUpdates["heat"] = heatGenerated
		}

		// Failure message
		result.Message = fmt.Sprintf("Operation failed! %s", getFailureMessage(operation.Type))
	}

	// Update player resources
	if err := s.playerService.UpdatePlayerResources(playerID, resourceUpdates); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update player resources after operation")
	}

	// Update operation attempt
	attempt.Result = result
	attempt.Status = func() string {
		// success ? util.OperationStatusCompleted : util.OperationStatusFailed
		if success {
			return util.OperationStatusCompleted
		}
		return util.OperationStatusFailed
	}()
	attempt.CompletionTime = ptrTime(time.Now())
	attempt.UpdatedAt = time.Now()

	if err := s.operationsRepo.UpdateOperationAttempt(attempt); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update operation attempt")
		return nil, errors.New("failed to complete operation")
	}

	// Update player stats
	stats, err := s.playerRepo.GetPlayerStats(playerID)
	if err == nil {
		stats.TotalOperationsCompleted++
		stats.UpdatedAt = time.Now()
		s.playerRepo.UpdatePlayerStats(stats)
	}

	// Add notification
	s.playerService.AddNotification(playerID, result.Message, util.NotificationTypeOperation)

	return result, nil
}

// RefreshDailyOperations refreshes the daily operations
func (s *operationsService) RefreshDailyOperations() error {
	// Get current operations to check if we need to refresh
	basicOperations, err := s.operationsRepo.GetBasicOperations()
	if err != nil {
		return err
	}

	specialOperations, err := s.operationsRepo.GetSpecialOperations()
	if err != nil {
		return err
	}

	// Only generate new operations if we're below the desired count
	if len(basicOperations) < s.gameConfig.DailyOperationsCount ||
		len(specialOperations) < s.gameConfig.SpecialOperationsCount {

		s.logger.Info().
			Int("current_basic", len(basicOperations)).
			Int("target_basic", s.gameConfig.DailyOperationsCount).
			Int("current_special", len(specialOperations)).
			Int("target_special", s.gameConfig.SpecialOperationsCount).
			Msg("Generating new operations to meet target counts")

		return s.operationsRepo.GenerateDailyOperations(
			s.gameConfig.DailyOperationsCount,
			s.gameConfig.SpecialOperationsCount,
		)
	}

	return nil
}

// CheckAndCompleteOperations checks for completed operations and processes them
func (s *operationsService) CheckAndCompleteOperations() error {
	// Get all in-progress operations
	var attempts []model.OperationAttempt
	if err := s.operationsRepo.GetDB().
		Where("status = ?", util.OperationStatusInProgress).
		Find(&attempts).Error; err != nil {
		return err
	}

	now := time.Now()

	for _, attempt := range attempts {
		// Get the operation
		operation, err := s.operationsRepo.GetOperationByID(attempt.OperationID)
		if err != nil {
			s.logger.Error().Err(err).Str("operationID", attempt.OperationID).Msg("Failed to get operation")
			continue
		}

		// Check if the operation has been running long enough
		timeSinceStart := now.Sub(attempt.Timestamp)
		if timeSinceStart.Seconds() >= float64(operation.Duration) {
			// Send notification that the operation is ready to collect
			message := fmt.Sprintf("Operation '%s' is ready to collect!", operation.Name)
			s.playerService.AddNotification(attempt.PlayerID, message, util.NotificationTypeOperation)
		}
	}

	return nil
}

// calculateSuccessChance calculates the success chance for an operation
func (s *operationsService) calculateSuccessChance(operation *model.Operation, resources model.OperationResources) int {
	// Base success chance from the operation
	successChance := operation.SuccessRate

	// Calculate resource commitment level (0-100%)
	requiredCrew := operation.Resources.Crew
	requiredWeapons := operation.Resources.Weapons
	requiredVehicles := operation.Resources.Vehicles

	// Avoid division by zero
	crewCommitment := 100.0
	if requiredCrew > 0 {
		crewCommitment = float64(resources.Crew) / float64(requiredCrew) * 100.0
		if crewCommitment > 200.0 {
			crewCommitment = 200.0 // Cap at 200%
		}
	}

	weaponsCommitment := 100.0
	if requiredWeapons > 0 {
		weaponsCommitment = float64(resources.Weapons) / float64(requiredWeapons) * 100.0
		if weaponsCommitment > 200.0 {
			weaponsCommitment = 200.0 // Cap at 200%
		}
	}

	vehiclesCommitment := 100.0
	if requiredVehicles > 0 {
		vehiclesCommitment = float64(resources.Vehicles) / float64(requiredVehicles) * 100.0
		if vehiclesCommitment > 200.0 {
			vehiclesCommitment = 200.0 // Cap at 200%
		}
	}

	// Weight the commitment levels
	averageCommitment := (crewCommitment + weaponsCommitment + vehiclesCommitment) / 3.0

	// Adjust success chance based on commitment
	if averageCommitment > 100.0 {
		// Bonus for over-committing resources
		successChance += int((averageCommitment - 100.0) / 10.0)
	} else if averageCommitment < 100.0 {
		// Penalty for under-committing resources
		successChance -= int((100.0 - averageCommitment) / 5.0)
	}

	// Cap success chance between 5% and 95%
	if successChance < 5 {
		successChance = 5
	} else if successChance > 95 {
		successChance = 95
	}

	return successChance
}

// getSuccessMessage returns a success message for the operation type
func getSuccessMessage(operationType string) string {
	switch operationType {
	case util.OperationTypeCarjacking:
		return "You successfully stole the vehicles and made a clean getaway."
	case util.OperationTypeGoodsSmuggling:
		return "The contraband was delivered safely to its destination."
	case util.OperationTypeDrugTrafficking:
		return "The product was successfully moved and sold for a substantial profit."
	case util.OperationTypeOfficialBribing:
		return "The officials have accepted your bribe and will turn a blind eye."
	case util.OperationTypeIntelligence:
		return "Valuable intelligence gathered on rival operations."
	case util.OperationTypeCrewRecruitment:
		return "New members have joined your crew."
	default:
		return "The operation was completed successfully."
	}
}

// getFailureMessage returns a failure message for the operation type
func getFailureMessage(operationType string) string {
	switch operationType {
	case util.OperationTypeCarjacking:
		return "The police spotted your crew during the theft and intervened."
	case util.OperationTypeGoodsSmuggling:
		return "The contraband was intercepted at a checkpoint."
	case util.OperationTypeDrugTrafficking:
		return "An undercover agent was among the buyers. Your crew barely escaped."
	case util.OperationTypeOfficialBribing:
		return "The official reported your bribe attempt to the authorities."
	case util.OperationTypeIntelligence:
		return "Your informant provided misleading information."
	case util.OperationTypeCrewRecruitment:
		return "The potential recruits were scared off by police presence."
	default:
		return "The operation failed due to unforeseen complications."
	}
}

// formatDuration formats duration in seconds to a human-readable string
func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

// meetsMinimumTitle checks if the player's title meets the minimum required title
func meetsMinimumTitle(playerTitle, requiredTitle string) bool {
	titleRanks := map[string]int{
		util.PlayerTitleAssociate:   1,
		util.PlayerTitleSoldier:     2,
		util.PlayerTitleCapo:        3,
		util.PlayerTitleUnderboss:   4,
		util.PlayerTitleConsigliere: 5,
		util.PlayerTitleBoss:        6,
		util.PlayerTitleGodfather:   7,
	}

	playerRank, ok := titleRanks[playerTitle]
	if !ok {
		return false
	}

	requiredRank, ok := titleRanks[requiredTitle]
	if !ok {
		return false
	}

	return playerRank >= requiredRank
}

// ptrTime creates a pointer to a time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}
