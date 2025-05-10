// internal/repository/operations.go

package repository

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"mwce-be/internal/model"
	"mwce-be/internal/util"
	"mwce-be/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OperationsRepository handles database operations for operations
type OperationsRepository interface {
	GetDB() *gorm.DB
	GetAllOperations() ([]model.Operation, error)
	GetOperationByID(id string) (*model.Operation, error)
	GetSpecialOperations() ([]model.Operation, error)
	GetBasicOperations() ([]model.Operation, error)
	CreateOperation(operation *model.Operation) error
	UpdateOperation(operation *model.Operation) error
	DeleteOperation(id string) error
	GenerateDailyOperations(basicCount, specialCount int) error
	GetOperationAttemptByID(id string) (*model.OperationAttempt, error)
	GetCurrentOperations(playerID string) ([]model.OperationAttempt, error)
	GetCompletedOperations(playerID string) ([]model.OperationAttempt, error)
	CreateOperationAttempt(attempt *model.OperationAttempt) error
	UpdateOperationAttempt(attempt *model.OperationAttempt) error
}

type operationsRepository struct {
	db database.Database
}

// NewOperationsRepository creates a new operations repository
func NewOperationsRepository(db database.Database) OperationsRepository {
	return &operationsRepository{
		db: db,
	}
}

// GetDB returns the database connection instance
func (r *operationsRepository) GetDB() *gorm.DB {
	return r.db.GetDB()
}

// GetAllOperations retrieves all operations
func (r *operationsRepository) GetAllOperations() ([]model.Operation, error) {
	var operations []model.Operation
	if err := r.db.GetDB().Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

// GetOperationByID retrieves an operation by ID
func (r *operationsRepository) GetOperationByID(id string) (*model.Operation, error) {
	var operation model.Operation
	if err := r.db.GetDB().Where("id = ?", id).First(&operation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("operation not found")
		}
		return nil, err
	}
	return &operation, nil
}

// GetSpecialOperations retrieves all special operations
func (r *operationsRepository) GetSpecialOperations() ([]model.Operation, error) {
	var operations []model.Operation
	if err := r.db.GetDB().
		Where("is_special = ?", true).
		Where("is_active = ?", true).
		Where("available_until > ?", time.Now()).
		Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

// GetBasicOperations retrieves all basic operations
func (r *operationsRepository) GetBasicOperations() ([]model.Operation, error) {
	var operations []model.Operation
	if err := r.db.GetDB().
		Where("is_special = ?", false).
		Where("is_active = ?", true).
		Where("available_until > ?", time.Now()).
		Find(&operations).Error; err != nil {
		return nil, err
	}
	return operations, nil
}

// CreateOperation creates a new operation
func (r *operationsRepository) CreateOperation(operation *model.Operation) error {
	return r.db.GetDB().Create(operation).Error
}

// UpdateOperation updates an operation
func (r *operationsRepository) UpdateOperation(operation *model.Operation) error {
	return r.db.GetDB().Save(operation).Error
}

// DeleteOperation deletes an operation
func (r *operationsRepository) DeleteOperation(id string) error {
	return r.db.GetDB().Delete(&model.Operation{}, "id = ?", id).Error
}

// GenerateDailyOperations generates a new set of daily operations
func (r *operationsRepository) GenerateDailyOperations(basicCount, specialCount int) error {
	// First, mark all operations with expired availability as inactive
	if err := r.db.GetDB().
		Model(&model.Operation{}).
		Where("available_until < ?", time.Now()).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to mark expired operations as inactive: %w", err)
	}

	// For existing operations that are still valid, ensure they remain active
	// This helps maintain continuity for operations players have already started
	/* Something not right here. Gotta think about this more. */
	if err := r.db.GetDB().
		Model(&model.Operation{}).
		Where("available_until >= ?", time.Now()).
		Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to ensure valid operations remain active: %w", err)
	}

	// Count existing active operations by type
	var basicOperationsCount, specialOperationsCount int64
	if err := r.db.GetDB().Model(&model.Operation{}).
		Where("is_special = ? AND is_active = ?", false, true).
		Count(&basicOperationsCount).Error; err != nil {
		return fmt.Errorf("failed to count active basic operations: %w", err)
	}

	if err := r.db.GetDB().Model(&model.Operation{}).
		Where("is_special = ? AND is_active = ?", true, true).
		Count(&specialOperationsCount).Error; err != nil {
		return fmt.Errorf("failed to count active special operations: %w", err)
	}

	// Get the operations pool from the yaml file
	operationsData, err := loadOperationsFromYAML()
	if err != nil {
		return fmt.Errorf("failed to load operations data: %w", err)
	}

	now := time.Now()

	// Temporary adding the available until here as 2 minutes.
	// This should be as part of the operations yaml file
	availableUntil := now.Add(45 * time.Minute)

	// Generate basic operations from the pool if needed
	remainingBasic := int(basicCount) - int(basicOperationsCount)
	if remainingBasic > 0 {
		basicPool := operationsData.BasicOperations

		// Check if pool is exhausted
		if len(basicPool) == 0 {
			log.Printf("WARNING: Basic operations pool is empty")
		} else if len(basicPool) < remainingBasic {
			log.Printf("WARNING: Basic operations pool has only %d operations, but %d are needed",
				len(basicPool), remainingBasic)
		}

		// Add operations from pool
		addedBasicCount := 0
		for _, opTemplate := range basicPool {
			if addedBasicCount >= remainingBasic {
				break
			}

			operation := model.Operation{
				ID:             uuid.New().String(),
				Name:           opTemplate.Name,
				Description:    opTemplate.Description,
				Type:           opTemplate.Type,
				IsSpecial:      false,
				IsActive:       true,
				Requirements:   model.OperationRequirements{}, // No requirements for basic operations
				Resources:      opTemplate.Resources,
				Rewards:        opTemplate.Rewards,
				Risks:          opTemplate.Risks,
				Duration:       opTemplate.Duration,
				SuccessRate:    opTemplate.SuccessRate,
				AvailableUntil: availableUntil,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := r.CreateOperation(&operation); err != nil {
				return err
			}

			addedBasicCount++
		}

		// If we still need more basic operations, generate random ones
		if addedBasicCount < remainingBasic {
			remainingToAdd := remainingBasic - addedBasicCount
			for i := 0; i < remainingToAdd; i++ {
				// Create random basic operation
				operation := model.Operation{
					ID:          uuid.New().String(),
					Name:        "Basic Operation " + uuid.New().String()[:8],
					Description: "A basic operation that can be completed by any player",
					Type:        getRandomOperationType(),
					IsSpecial:   false,
					IsActive:    true,
					Requirements: model.OperationRequirements{
						MinInfluence: 0,
						MaxHeat:      100,
					},
					Resources: model.OperationResources{
						Crew:     rand.Intn(3) + 1,
						Weapons:  rand.Intn(2) + 1,
						Vehicles: rand.Intn(2) + 1,
					},
					Rewards: model.OperationRewards{
						Money:   1000 + rand.Intn(3000),
						Respect: 3 + rand.Intn(7),
					},
					Risks: model.OperationRisks{
						CrewLoss:     rand.Intn(2) + 1,
						WeaponsLoss:  rand.Intn(2) + 1,
						HeatIncrease: 5 + rand.Intn(15),
					},
					Duration:       1800 + rand.Intn(3600), // 30 minutes to 90 minutes
					SuccessRate:    60 + rand.Intn(20),
					AvailableUntil: availableUntil,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := r.CreateOperation(&operation); err != nil {
					return err
				}
			}
		}
	}

	// Generate special operations from the pool if needed
	remainingSpecial := int(specialCount) - int(specialOperationsCount)
	if remainingSpecial > 0 {
		specialPool := operationsData.SpecialOperations

		// Check if pool is exhausted
		if len(specialPool) == 0 {
			log.Printf("WARNING: Special operations pool is empty")
		} else if len(specialPool) < remainingSpecial {
			log.Printf("WARNING: Special operations pool has only %d operations, but %d are needed",
				len(specialPool), remainingSpecial)
		}

		// Add operations from pool
		addedSpecialCount := 0
		for _, opTemplate := range specialPool {
			if addedSpecialCount >= remainingSpecial {
				break
			}

			operation := model.Operation{
				ID:             uuid.New().String(),
				Name:           opTemplate.Name,
				Description:    opTemplate.Description,
				Type:           opTemplate.Type,
				IsSpecial:      true,
				IsActive:       true,
				Requirements:   opTemplate.Requirements,
				Resources:      opTemplate.Resources,
				Rewards:        opTemplate.Rewards,
				Risks:          opTemplate.Risks,
				Duration:       opTemplate.Duration,
				SuccessRate:    opTemplate.SuccessRate,
				AvailableUntil: availableUntil,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := r.CreateOperation(&operation); err != nil {
				return err
			}

			addedSpecialCount++
		}

		// If we still need more special operations, generate random ones
		if addedSpecialCount < remainingSpecial {
			remainingToAdd := remainingSpecial - addedSpecialCount
			for i := 0; i < remainingToAdd; i++ {
				// Create random special operation
				operation := model.Operation{
					ID:          uuid.New().String(),
					Name:        "Special Operation " + uuid.New().String()[:8],
					Description: "A special operation that requires high influence",
					Type:        getRandomOperationType(),
					IsSpecial:   true,
					IsActive:    true,
					Requirements: model.OperationRequirements{
						MinInfluence: 20 + rand.Intn(30),
						MaxHeat:      40 + rand.Intn(30),
						MinTitle:     getRandomMinimumTitle(),
					},
					Resources: model.OperationResources{
						Crew:     3 + rand.Intn(4),
						Weapons:  2 + rand.Intn(4),
						Vehicles: 1 + rand.Intn(3),
					},
					Rewards: model.OperationRewards{
						Money:     5000 + rand.Intn(15000),
						Respect:   10 + rand.Intn(20),
						Influence: 5 + rand.Intn(15),
					},
					Risks: model.OperationRisks{
						CrewLoss:     1 + rand.Intn(3),
						WeaponsLoss:  1 + rand.Intn(3),
						VehiclesLoss: rand.Intn(2),
						HeatIncrease: 15 + rand.Intn(35),
					},
					Duration:       3600 + rand.Intn(14400), // 1 hour to 5 hours
					SuccessRate:    45 + rand.Intn(20),
					AvailableUntil: availableUntil,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := r.CreateOperation(&operation); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Helper functions for random operation generation
func getRandomOperationType() string {
	operationTypes := []string{
		util.OperationTypeCarjacking,
		util.OperationTypeGoodsSmuggling,
		util.OperationTypeDrugTrafficking,
		util.OperationTypeOfficialBribing,
		util.OperationTypeIntelligence,
		util.OperationTypeCrewRecruitment,
	}
	return operationTypes[rand.Intn(len(operationTypes))]
}

func getRandomMinimumTitle() string {
	titles := []string{
		util.PlayerTitleSoldier,
		util.PlayerTitleCapo,
		util.PlayerTitleUnderboss,
		util.PlayerTitleConsigliere,
	}
	return titles[rand.Intn(len(titles))]
}

// GetOperationAttemptByID retrieves an operation attempt by ID
func (r *operationsRepository) GetOperationAttemptByID(id string) (*model.OperationAttempt, error) {
	var attempt model.OperationAttempt
	if err := r.db.GetDB().Where("id = ?", id).First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("operation attempt not found")
		}
		return nil, err
	}
	return &attempt, nil
}

// GetCurrentOperations retrieves in-progress operations for a player
func (r *operationsRepository) GetCurrentOperations(playerID string) ([]model.OperationAttempt, error) {
	var attempts []model.OperationAttempt
	if err := r.db.GetDB().
		Where("player_id = ?", playerID).
		Where("status = ?", util.OperationStatusInProgress).
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

// GetCompletedOperations retrieves completed operations for a player
func (r *operationsRepository) GetCompletedOperations(playerID string) ([]model.OperationAttempt, error) {
	var attempts []model.OperationAttempt
	if err := r.db.GetDB().
		Where("player_id = ?", playerID).
		Where("status IN ?", []string{
			util.OperationStatusCompleted,
			util.OperationStatusFailed,
			util.OperationStatusCancelled,
		}).
		Order("completion_time DESC").
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

// CreateOperationAttempt creates a new operation attempt
func (r *operationsRepository) CreateOperationAttempt(attempt *model.OperationAttempt) error {
	return r.db.GetDB().Create(attempt).Error
}

// UpdateOperationAttempt updates an operation attempt
func (r *operationsRepository) UpdateOperationAttempt(attempt *model.OperationAttempt) error {
	return r.db.GetDB().Save(attempt).Error
}
