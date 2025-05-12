// internal/service/operations.go

package service

import (
	"errors"
	"fmt"
	"math/rand"
	"mwce-be/internal/util"
	"strconv"
	"sync"
	"time"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// OperationsService handles operations-related business logic
type OperationsService interface {
	GetAvailableOperations(playerID string, validOnly bool) ([]model.Operation, error)
	GetOperationByID(id string) (*model.Operation, error)
	GetCurrentOperations(playerID string) ([]model.OperationAttempt, error)
	GetCompletedOperations(playerID string) ([]model.OperationAttempt, error)
	StartOperation(playerID, operationID string, resources model.OperationResources) (*model.OperationAttempt, error)
	CancelOperation(playerID, attemptID string) error
	CollectOperation(playerID, attemptID string) (*model.OperationResult, error)
	CollectOperationReward(playerID, attemptID string) (*model.OperationResult, error)
	RefreshDailyOperations() error
	CheckAndCompleteOperations() error

	// Scheduled jobs
	StartPeriodicOperationsRefresh()

	GetOperationsRefreshInfo() (*model.OperationsRefreshInfo, error)
}

type operationsService struct {
	operationsRepo repository.OperationsRepository
	playerRepo     repository.PlayerRepository
	playerService  PlayerService
	sseService     SSEService
	gameConfig     config.GameConfig
	logger         zerolog.Logger

	lastRefreshTime time.Time
	refreshMutex    sync.RWMutex
}

// NewOperationsService creates a new operations service
func NewOperationsService(
	operationsRepo repository.OperationsRepository,
	playerRepo repository.PlayerRepository,
	playerService PlayerService,
	sseService SSEService,
	gameConfig config.GameConfig,
	logger zerolog.Logger,
) OperationsService {
	return &operationsService{
		operationsRepo:  operationsRepo,
		playerRepo:      playerRepo,
		playerService:   playerService,
		sseService:      sseService,
		gameConfig:      gameConfig,
		logger:          logger,
		lastRefreshTime: time.Now(),
		refreshMutex:    sync.RWMutex{},
	}
}

func (s *operationsService) GetOperationsRefreshInfo() (*model.OperationsRefreshInfo, error) {
	// Get refresh information safely with lock
	s.refreshMutex.RLock()
	lastRefreshTime := s.lastRefreshTime
	s.refreshMutex.RUnlock()

	// Calculate next refresh time
	refreshInterval := s.gameConfig.OperationsRefreshInterval
	nextRefreshTime := lastRefreshTime.Add(time.Duration(refreshInterval) * time.Minute)

	return &model.OperationsRefreshInfo{
		RefreshInterval: refreshInterval,
		LastRefreshTime: lastRefreshTime.Format(time.RFC3339),
		NextRefreshTime: nextRefreshTime.Format(time.RFC3339),
	}, nil
}

// GetAvailableOperations retrieves available operations for a player - NOW REGION-AWARE
func (s *operationsService) GetAvailableOperations(playerID string, validOnly bool) ([]model.Operation, error) {
	// Get the player to find their current region
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return nil, err
	}

	// If player is not in any region, return global operations only (no region ID)
	if player.CurrentRegionID == nil {
		var globalOps []model.Operation
		query := s.operationsRepo.GetDB().Where("is_active = ?", true).Where("region_id IS NULL")

		if validOnly {
			query = query.Where("available_until > ?", time.Now())
		}

		if err := query.Find(&globalOps).Error; err != nil {
			return nil, err
		}

		return s.filterOperationsByPlayerAttempts(globalOps, playerID)
	}

	// Get operations for the player's current region and global operations
	var operations []model.Operation
	query := s.operationsRepo.GetDB().Where("is_active = ?", true)

	if validOnly {
		query = query.Where("available_until > ?", time.Now())
	}

	// Filter by current region - either global (region_id IS NULL), specific to current region,
	// or multi-region operations that include the current region
	query = query.Where("region_id IS NULL OR region_id = ? OR ? = ANY(region_ids)",
		*player.CurrentRegionID, *player.CurrentRegionID)

	if err := query.Find(&operations).Error; err != nil {
		return nil, err
	}

	// Filter operations based on player's in-progress attempts and requirements
	return s.filterOperationsByPlayerAttempts(s.filterOperationsByRequirements(operations, player), playerID)
}

// New helper function to filter operations based on player requirements
func (s *operationsService) filterOperationsByRequirements(operations []model.Operation, player *model.Player) []model.Operation {
	var filteredOps []model.Operation

	for _, op := range operations {
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
				if !meetsMinimumTitle(player.Title, op.Requirements.MinTitle) {
					continue
				}
			}
		}

		filteredOps = append(filteredOps, op)
	}

	return filteredOps
}

// New helper function to filter out operations that the player already has in progress
func (s *operationsService) filterOperationsByPlayerAttempts(operations []model.Operation, playerID string) ([]model.Operation, error) {
	// Get player's in-progress operations
	inProgressOps, err := s.operationsRepo.GetCurrentOperations(playerID)
	if err != nil {
		return nil, err
	}

	// Create a map of operation IDs that are already in progress
	inProgressOpMap := make(map[string]model.OperationAttempt)
	for _, op := range inProgressOps {
		inProgressOpMap[op.OperationID] = op
	}

	// Filter out operations that the player already has in progress
	var filteredOps []model.Operation
	for _, op := range operations {
		if _, exists := inProgressOpMap[op.ID]; !exists {
			filteredOps = append(filteredOps, op)
		} else {
			// If operation is in progress, add it to PlayerAttempts field but don't include in available list
			op.PlayerAttempts = append(op.PlayerAttempts, inProgressOpMap[op.ID])
		}
	}

	return filteredOps, nil
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
	now := time.Now()
	if operation.AvailableUntil.Before(now) {
		return nil, errors.New("operation is no longer available")
	}

	// Check if the player already has this operation in progress
	inProgressOps, err := s.operationsRepo.GetCurrentOperations(playerID)
	if err != nil {
		return nil, errors.New("failed to check in-progress operations")
	}

	for _, op := range inProgressOps {
		if op.OperationID == operationID {
			return nil, errors.New("you already have this operation in progress")
		}
	}

	// Check if there's enough time remaining to complete the operation
	timeRemaining := operation.AvailableUntil.Sub(now).Seconds()
	if timeRemaining < float64(operation.Duration) {
		return nil, errors.New("insufficient time remaining to complete this operation")
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

// CollectOperation completes an operation and moves it to completed status without applying rewards
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
	successChance := s.calculateSuccessChance(operation, attempt.Resources, playerID)
	success := rand.Float64()*100 < float64(successChance)

	// Generate result without applying rewards (this will be done in CollectOperationReward)
	result := &model.OperationResult{
		Success:          success,
		Message:          "",
		RewardsCollected: false,
	}

	// Process potential outcomes based on success without applying them
	if success {
		// Calculate rewards that would be given
		if operation.Rewards.Money > 0 {
			result.MoneyGained = operation.Rewards.Money
		}

		if operation.Rewards.Crew > 0 {
			result.CrewGained = operation.Rewards.Crew
		}

		if operation.Rewards.Weapons > 0 {
			result.WeaponsGained = operation.Rewards.Weapons
		}

		if operation.Rewards.Vehicles > 0 {
			result.VehiclesGained = operation.Rewards.Vehicles
		}

		if operation.Rewards.Respect > 0 {
			result.RespectGained = operation.Rewards.Respect
		}

		if operation.Rewards.Influence > 0 {
			result.InfluenceGained = operation.Rewards.Influence
		}

		if operation.Rewards.HeatReduction > 0 {
			result.HeatReduced = operation.Rewards.HeatReduction
		}

		// Success message
		result.Message = fmt.Sprintf("Operation successful! %s", s.getSuccessMessage(operation.Type))
	} else {
		// Calculate losses that would be applied
		if operation.Risks.CrewLoss > 0 {
			result.CrewLost = rand.Intn(operation.Risks.CrewLoss) + 1
		}

		if operation.Risks.WeaponsLoss > 0 {
			result.WeaponsLost = rand.Intn(operation.Risks.WeaponsLoss) + 1
		}

		if operation.Risks.VehiclesLoss > 0 {
			result.VehiclesLost = rand.Intn(operation.Risks.VehiclesLoss) + 1
		}

		if operation.Risks.MoneyLoss > 0 {
			result.MoneyLost = operation.Risks.MoneyLoss
		}

		if operation.Risks.HeatIncrease > 0 {
			result.HeatGenerated = operation.Risks.HeatIncrease
		}

		// Failure message
		result.Message = fmt.Sprintf("Operation failed! %s", s.getFailureMessage(operation.Type))
	}

	// Update operation attempt
	attempt.Result = result
	attempt.Status = func() string {
		if success {
			return util.OperationStatusCompleted
		}
		return util.OperationStatusFailed
	}()
	attempt.CompletionTime = ptrTime(time.Now())
	attempt.Notified = true // Mark as notified since the player is actively collecting this operation
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

// CollectOperationReward collects rewards from a completed operation
func (s *operationsService) CollectOperationReward(playerID, attemptID string) (*model.OperationResult, error) {
	// Get the operation attempt
	attempt, err := s.operationsRepo.GetOperationAttemptByID(attemptID)
	if err != nil {
		return nil, errors.New("operation attempt not found")
	}

	// Check if the attempt belongs to the player
	if attempt.PlayerID != playerID {
		return nil, errors.New("not authorized to collect this operation reward")
	}

	// Check if the attempt is completed
	if attempt.Status != util.OperationStatusCompleted && attempt.Status != util.OperationStatusFailed {
		return nil, errors.New("can only collect rewards from completed or failed operations")
	}

	// Check if the operation was successful
	if attempt.Result == nil {
		return nil, errors.New("operation result not found")
	}

	// Check if rewards have already been collected
	if attempt.Result.RewardsCollected {
		return nil, errors.New("rewards for this operation have already been collected")
	}

	// Update player resources based on result
	resourceUpdates := make(map[string]int)

	// Apply rewards
	if attempt.Result.Success {
		if attempt.Result.MoneyGained > 0 {
			resourceUpdates["money"] = attempt.Result.MoneyGained
		}

		if attempt.Result.CrewGained > 0 {
			resourceUpdates["crew"] = attempt.Result.CrewGained
		}

		if attempt.Result.WeaponsGained > 0 {
			resourceUpdates["weapons"] = attempt.Result.WeaponsGained
		}

		if attempt.Result.VehiclesGained > 0 {
			resourceUpdates["vehicles"] = attempt.Result.VehiclesGained
		}

		if attempt.Result.RespectGained > 0 {
			resourceUpdates["respect"] = attempt.Result.RespectGained
		}

		if attempt.Result.InfluenceGained > 0 {
			resourceUpdates["influence"] = attempt.Result.InfluenceGained
		}

		if attempt.Result.HeatReduced > 0 {
			resourceUpdates["heat"] = -attempt.Result.HeatReduced
		}
	} else {
		// Apply losses for failed operations
		if attempt.Result.CrewLost > 0 {
			resourceUpdates["crew"] = -attempt.Result.CrewLost
		}

		if attempt.Result.WeaponsLost > 0 {
			resourceUpdates["weapons"] = -attempt.Result.WeaponsLost
		}

		if attempt.Result.VehiclesLost > 0 {
			resourceUpdates["vehicles"] = -attempt.Result.VehiclesLost
		}

		if attempt.Result.MoneyLost > 0 {
			resourceUpdates["money"] = -attempt.Result.MoneyLost
		}

		if attempt.Result.HeatGenerated > 0 {
			resourceUpdates["heat"] = attempt.Result.HeatGenerated
		}
	}

	// Update player resources
	if err := s.playerService.UpdatePlayerResources(playerID, resourceUpdates); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update player resources after collecting operation rewards")
		return nil, errors.New("failed to update player resources")
	}

	// Mark rewards as collected
	attempt.Result.RewardsCollected = true
	attempt.UpdatedAt = time.Now()

	if err := s.operationsRepo.UpdateOperationAttempt(attempt); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update operation attempt after collecting rewards")
		return nil, errors.New("failed to update operation attempt")
	}

	// Create a success message based on the rewards
	message := "Rewards collected: "
	hasRewards := false

	if attempt.Result.MoneyGained > 0 {
		message += fmt.Sprintf("$%d, ", attempt.Result.MoneyGained)
		hasRewards = true
	}

	if attempt.Result.CrewGained > 0 {
		message += fmt.Sprintf("%d crew, ", attempt.Result.CrewGained)
		hasRewards = true
	}

	if attempt.Result.WeaponsGained > 0 {
		message += fmt.Sprintf("%d weapons, ", attempt.Result.WeaponsGained)
		hasRewards = true
	}

	if attempt.Result.VehiclesGained > 0 {
		message += fmt.Sprintf("%d vehicles, ", attempt.Result.VehiclesGained)
		hasRewards = true
	}

	if attempt.Result.RespectGained > 0 {
		message += fmt.Sprintf("%d respect, ", attempt.Result.RespectGained)
		hasRewards = true
	}

	if attempt.Result.InfluenceGained > 0 {
		message += fmt.Sprintf("%d influence, ", attempt.Result.InfluenceGained)
		hasRewards = true
	}

	if attempt.Result.HeatReduced > 0 {
		message += fmt.Sprintf("%d heat reduction, ", attempt.Result.HeatReduced)
		hasRewards = true
	}

	// For failed operations
	if !attempt.Result.Success {
		message = "Operation losses applied: "
		hasRewards = false

		if attempt.Result.MoneyLost > 0 {
			message += fmt.Sprintf("$%d, ", attempt.Result.MoneyLost)
			hasRewards = true
		}

		if attempt.Result.CrewLost > 0 {
			message += fmt.Sprintf("%d crew, ", attempt.Result.CrewLost)
			hasRewards = true
		}

		if attempt.Result.WeaponsLost > 0 {
			message += fmt.Sprintf("%d weapons, ", attempt.Result.WeaponsLost)
			hasRewards = true
		}

		if attempt.Result.VehiclesLost > 0 {
			message += fmt.Sprintf("%d vehicles, ", attempt.Result.VehiclesLost)
			hasRewards = true
		}

		if attempt.Result.HeatGenerated > 0 {
			message += fmt.Sprintf("%d heat increase, ", attempt.Result.HeatGenerated)
			hasRewards = true
		}
	}

	// Remove trailing comma and space
	if hasRewards && len(message) > 18 {
		message = message[:len(message)-2]
	} else if !hasRewards {
		message = ""
		// attempt.Result.Success ? "Rewards collected successfully!" : "Operation losses processed.";
		if attempt.Result.Success {
			message = "Rewards collected successfully!"
		} else {
			message = "Operation losses processed."
		}
	}

	// Add notification
	s.playerService.AddNotification(playerID, message, util.NotificationTypeOperation)

	return attempt.Result, nil
}

// CheckAndCompleteOperations checks for operations that should be completed and processes them
func (s *operationsService) CheckAndCompleteOperations() error {
	// Get all in-progress operations
	var attempts []model.OperationAttempt
	if err := s.operationsRepo.GetDB().
		Where("status = ?", util.OperationStatusInProgress).
		Find(&attempts).Error; err != nil {
		return err
	}

	now := time.Now()
	completed := 0

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
			// Determine success or failure
			successChance := s.calculateSuccessChance(operation, attempt.Resources, attempt.PlayerID)
			success := rand.Float64()*100 < float64(successChance)

			// Generate result without applying rewards (this will be done in CollectOperationReward)
			result := &model.OperationResult{
				Success:          success,
				Message:          "",
				RewardsCollected: false,
			}

			// Calculate potential rewards or losses without applying them
			if success {
				// Rewards
				if operation.Rewards.Money > 0 {
					result.MoneyGained = operation.Rewards.Money
				}

				if operation.Rewards.Crew > 0 {
					result.CrewGained = operation.Rewards.Crew
				}

				if operation.Rewards.Weapons > 0 {
					result.WeaponsGained = operation.Rewards.Weapons
				}

				if operation.Rewards.Vehicles > 0 {
					result.VehiclesGained = operation.Rewards.Vehicles
				}

				if operation.Rewards.Respect > 0 {
					result.RespectGained = operation.Rewards.Respect
				}

				if operation.Rewards.Influence > 0 {
					result.InfluenceGained = operation.Rewards.Influence
				}

				if operation.Rewards.HeatReduction > 0 {
					result.HeatReduced = operation.Rewards.HeatReduction
				}

				// Success message
				result.Message = fmt.Sprintf("Operation successful! %s", s.getSuccessMessage(operation.Type))
			} else {
				// Losses
				if operation.Risks.CrewLoss > 0 {
					result.CrewLost = rand.Intn(operation.Risks.CrewLoss) + 1
				}

				if operation.Risks.WeaponsLoss > 0 {
					result.WeaponsLost = rand.Intn(operation.Risks.WeaponsLoss) + 1
				}

				if operation.Risks.VehiclesLoss > 0 {
					result.VehiclesLost = rand.Intn(operation.Risks.VehiclesLoss) + 1
				}

				if operation.Risks.MoneyLoss > 0 {
					result.MoneyLost = operation.Risks.MoneyLoss
				}

				if operation.Risks.HeatIncrease > 0 {
					result.HeatGenerated = operation.Risks.HeatIncrease
				}

				// Failure message
				result.Message = fmt.Sprintf("Operation failed! %s", s.getFailureMessage(operation.Type))
			}

			// Update operation attempt
			attempt.Result = result
			attempt.Status = func() string {
				if success {
					return util.OperationStatusCompleted
				}
				return util.OperationStatusFailed
			}()
			attempt.CompletionTime = ptrTime(now)
			attempt.Notified = false // Mark as not notified yet, the notification service will pick this up
			attempt.UpdatedAt = now

			if err := s.operationsRepo.UpdateOperationAttempt(&attempt); err != nil {
				s.logger.Error().Err(err).Msg("Failed to update operation attempt")
				continue
			}

			// Update player stats
			stats, err := s.playerRepo.GetPlayerStats(attempt.PlayerID)
			if err == nil {
				stats.TotalOperationsCompleted++
				stats.UpdatedAt = now
				s.playerRepo.UpdatePlayerStats(stats)
			}

			// Add notification
			notificationMsg := fmt.Sprintf("Operation '%s' is ready to collect!", operation.Name)
			s.playerService.AddNotification(attempt.PlayerID, notificationMsg, util.NotificationTypeOperation)

			completed++
		}
	}

	if completed > 0 {
		s.logger.Info().Int("count", completed).Msg("Completed operations")
	}

	return nil
}

// RefreshDailyOperations refreshes operations from YAML pool with region support
func (s *operationsService) RefreshDailyOperations() error {
	// Load operations from YAML
	operationsData, err := loadOperationsFromYAML()
	if err != nil {
		return fmt.Errorf("failed to load operations data: %w", err)
	}

	// Get all regions for multi-region support
	var regions []model.Region
	if err := s.operationsRepo.GetDB().Find(&regions).Error; err != nil {
		return err
	}

	// Create a map of region IDs to region objects for easy lookup
	regionMap := make(map[string]model.Region)
	for _, region := range regions {
		regionMap[region.ID] = region
	}

	// Mark expired operations as inactive
	if err := s.operationsRepo.GetDB().
		Model(&model.Operation{}).
		Where("available_until < ?", time.Now()).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to mark expired operations as inactive: %w", err)
	}

	// For existing operations that are still valid, ensure they remain active
	if err := s.operationsRepo.GetDB().
		Model(&model.Operation{}).
		Where("available_until >= ?", time.Now()).
		Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to ensure valid operations remain active: %w", err)
	}

	// Create global and region-specific operations from the YAML pool
	now := time.Now()

	// Refresh basic operations
	for _, template := range operationsData.BasicOperations {
		// Skip if we've already reached the daily operations count
		var currentCount int64
		if err := s.operationsRepo.GetDB().
			Model(&model.Operation{}).
			Where("is_special = ? AND is_active = ?", false, true).
			Count(&currentCount).Error; err != nil {
			return fmt.Errorf("failed to count active basic operations: %w", err)
		}

		if int(currentCount) >= s.gameConfig.DailyOperationsCount {
			break
		}

		// Determine availability duration (in minutes)
		availabilityDuration := template.AvailabilityDuration
		if availabilityDuration <= 0 {
			availabilityDuration = s.gameConfig.OperationsRefreshInterval // Default to refresh interval
		}

		availableUntil := now.Add(time.Duration(availabilityDuration) * time.Minute)

		// Create operation
		operation := model.Operation{
			ID:                   uuid.New().String(),
			Name:                 template.Name,
			Description:          template.Description,
			Type:                 template.Type,
			IsSpecial:            template.IsSpecial,
			IsActive:             true,
			RegionIDs:            template.Regions, // Use regions from YAML
			Requirements:         template.Requirements,
			Resources:            template.Resources,
			Rewards:              template.Rewards,
			Risks:                template.Risks,
			Duration:             template.Duration,
			AvailabilityDuration: availabilityDuration,
			SuccessRate:          template.SuccessRate,
			AvailableUntil:       availableUntil,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		if err := s.operationsRepo.CreateOperation(&operation); err != nil {
			return err
		}
	}

	// Refresh special operations
	for _, template := range operationsData.SpecialOperations {
		// Skip if we've already reached the special operations count
		var currentCount int64
		if err := s.operationsRepo.GetDB().
			Model(&model.Operation{}).
			Where("is_special = ? AND is_active = ?", true, true).
			Count(&currentCount).Error; err != nil {
			return fmt.Errorf("failed to count active special operations: %w", err)
		}

		if int(currentCount) >= s.gameConfig.SpecialOperationsCount {
			break
		}

		// Determine availability duration (in minutes)
		availabilityDuration := template.AvailabilityDuration
		if availabilityDuration <= 0 {
			availabilityDuration = s.gameConfig.OperationsRefreshInterval // Default to refresh interval
		}

		availableUntil := now.Add(time.Duration(availabilityDuration) * time.Minute)

		// Create operation
		operation := model.Operation{
			ID:                   uuid.New().String(),
			Name:                 template.Name,
			Description:          template.Description,
			Type:                 template.Type,
			IsSpecial:            template.IsSpecial,
			IsActive:             true,
			RegionIDs:            template.Regions, // Use regions from YAML
			Requirements:         template.Requirements,
			Resources:            template.Resources,
			Rewards:              template.Rewards,
			Risks:                template.Risks,
			Duration:             template.Duration,
			AvailabilityDuration: availabilityDuration,
			SuccessRate:          template.SuccessRate,
			AvailableUntil:       availableUntil,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		if err := s.operationsRepo.CreateOperation(&operation); err != nil {
			return err
		}
	}

	// Update last refresh time
	s.refreshMutex.Lock()
	s.lastRefreshTime = now
	s.refreshMutex.Unlock()

	// Get all active operations for SSE notification
	var allOperations []model.Operation
	if err := s.operationsRepo.GetDB().
		Where("is_active = ? AND available_until > ?", true, time.Now()).
		Find(&allOperations).Error; err != nil {
		s.logger.Error().Err(err).Msg("Failed to fetch operations for SSE notification")
		return nil // Don't fail the refresh just because we couldn't send notification
	}

	nextRefreshTime := now.Add(time.Duration(s.gameConfig.OperationsRefreshInterval) * time.Minute)

	s.sseService.SendEventToAll("operations_refreshed", map[string]interface{}{
		"operations": allOperations,
		"timestamp":  now.Format(time.RFC3339),
		"refreshInfo": map[string]interface{}{
			"refreshInterval": s.gameConfig.OperationsRefreshInterval,
			"lastRefreshTime": now.Format(time.RFC3339),
			"nextRefreshTime": nextRefreshTime.Format(time.RFC3339),
		},
	})

	return nil
}

// calculateSuccessChance calculates the success chance for an operation
func (s *operationsService) calculateSuccessChance(operation *model.Operation, resources model.OperationResources, playerID string) int {
	// If mechanics config is available, use it for calculations
	if s.gameConfig.Mechanics != nil {
		// Try to get operation-specific success chance from config
		if successChanceConfig, exists := s.gameConfig.Mechanics.SuccessChances[operation.Type]; exists {
			// Base success chance from config
			successChance := successChanceConfig.BaseChance

			// Apply resource multipliers from config
			resourceCommitmentBonus := 0.0
			if len(successChanceConfig.ResourceMultiplier) > 0 {
				// Calculate resource commitment level
				if multiplier, exists := successChanceConfig.ResourceMultiplier["crew"]; exists && operation.Resources.Crew > 0 {
					crewCommitment := float64(resources.Crew) / float64(operation.Resources.Crew) * multiplier
					resourceCommitmentBonus += crewCommitment
				}

				if multiplier, exists := successChanceConfig.ResourceMultiplier["weapons"]; exists && operation.Resources.Weapons > 0 {
					weaponsCommitment := float64(resources.Weapons) / float64(operation.Resources.Weapons) * multiplier
					resourceCommitmentBonus += weaponsCommitment
				}

				if multiplier, exists := successChanceConfig.ResourceMultiplier["vehicles"]; exists && operation.Resources.Vehicles > 0 {
					vehiclesCommitment := float64(resources.Vehicles) / float64(operation.Resources.Vehicles) * multiplier
					resourceCommitmentBonus += vehiclesCommitment
				}

				if multiplier, exists := successChanceConfig.ResourceMultiplier["money"]; exists && operation.Resources.Money > 0 {
					moneyCommitment := float64(resources.Money) / float64(operation.Resources.Money) * multiplier
					resourceCommitmentBonus += moneyCommitment
				}
			}

			// Apply resource commitment bonus
			successChance += int(resourceCommitmentBonus)

			// Apply heat penalty if applicable
			if s.gameConfig.Mechanics.Heat.Effects != nil {
				// Check for operation success penalties in heat effects
				operationPenalties, exists := s.gameConfig.Mechanics.Heat.Effects["operation_success_penalty"]
				if exists && playerID != "" {
					// Get player's current heat
					player, err := s.playerRepo.GetPlayerByID(playerID)
					if err == nil {
						// Find the appropriate penalty tier based on heat level
						// We need to process the nested map structure correctly
						var appliedPenalty int
						highestApplicableThreshold := 0

						// Find the highest threshold that's less than or equal to player's heat
						for heatThresholdStr, penaltyVal := range operationPenalties {
							// Convert string key to int for comparison
							heatThreshold, err := strconv.Atoi(heatThresholdStr)
							if err != nil {
								s.logger.Warn().
									Str("threshold", heatThresholdStr).
									Msg("Invalid heat threshold format in config")
								continue
							}

							// If player heat meets or exceeds this threshold and it's higher than previous matches
							if player.Heat >= heatThreshold && heatThreshold > highestApplicableThreshold {
								highestApplicableThreshold = heatThreshold

								// Extract the penalty value
								penalty, exists := penaltyVal[highestApplicableThreshold]
								if !exists {
									s.logger.Warn().
										Int("threshold", highestApplicableThreshold).
										Msg("Penalty not found for threshold in config")
									continue
								}
								appliedPenalty = penalty
							}
						}

						// Apply the highest applicable penalty
						if highestApplicableThreshold > 0 {
							successChance -= appliedPenalty
							s.logger.Debug().
								Int("playerHeat", player.Heat).
								Int("thresholdApplied", highestApplicableThreshold).
								Int("penaltyApplied", appliedPenalty).
								Int("finalSuccessChance", successChance).
								Msg("Applied heat penalty to operation")
						}
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
	}

	// Fallback to original calculation if config not available
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
func (s *operationsService) getSuccessMessage(operationType string) string {
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
func (s *operationsService) getFailureMessage(operationType string) string {
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
