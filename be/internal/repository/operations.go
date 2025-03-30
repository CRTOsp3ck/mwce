// internal/repository/operations.go

package repository

import (
	"errors"
	"fmt"
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
	// First, delete all expired operations
	if err := r.db.GetDB().
		Where("available_until < ?", time.Now()).
		Delete(&model.Operation{}).Error; err != nil {
		return err
	}

	// Get the operations pool from the yaml file
	operationsData, err := loadOperationsFromYAML()
	if err != nil {
		return fmt.Errorf("failed to load operations data: %w", err)
	}

	// Maintain a record of operations already in the DB to avoid duplicates
	var existingOperations []model.Operation
	if err := r.db.GetDB().Find(&existingOperations).Error; err != nil {
		return fmt.Errorf("failed to fetch existing operations: %w", err)
	}

	existingNames := make(map[string]bool)
	for _, op := range existingOperations {
		existingNames[op.Name] = true
	}

	now := time.Now()
	availableUntil := now.Add(24 * time.Hour)

	// Generate basic operations from the pool
	addedBasicCount := 0
	for _, opTemplate := range operationsData.BasicOperations {
		if addedBasicCount >= basicCount {
			break
		}

		// Skip if this operation is already active in the DB
		if existingNames[opTemplate.Name] {
			continue
		}

		operation := model.Operation{
			ID:             uuid.New().String(),
			Name:           opTemplate.Name,
			Description:    opTemplate.Description,
			Type:           opTemplate.Type,
			IsSpecial:      false,
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

		existingNames[opTemplate.Name] = true
		addedBasicCount++
	}

	// Generate special operations from the pool
	addedSpecialCount := 0
	for _, opTemplate := range operationsData.SpecialOperations {
		if addedSpecialCount >= specialCount {
			break
		}

		// Skip if this operation is already active in the DB
		if existingNames[opTemplate.Name] {
			continue
		}

		operation := model.Operation{
			ID:             uuid.New().String(),
			Name:           opTemplate.Name,
			Description:    opTemplate.Description,
			Type:           opTemplate.Type,
			IsSpecial:      true,
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

		existingNames[opTemplate.Name] = true
		addedSpecialCount++
	}

	// If we couldn't add enough operations from the pool, generate random ones to fill the quota
	if addedBasicCount < basicCount {
		remainingBasic := basicCount - addedBasicCount
		for i := 0; i < remainingBasic; i++ {
			// Create random basic operation
			operation := model.Operation{
				ID:          uuid.New().String(),
				Name:        "Basic Operation " + uuid.New().String()[:8],
				Description: "A basic operation that can be completed by any player",
				Type:        getRandomOperationType(),
				IsSpecial:   false,
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

	if addedSpecialCount < specialCount {
		remainingSpecial := specialCount - addedSpecialCount
		for i := 0; i < remainingSpecial; i++ {
			// Create random special operation
			operation := model.Operation{
				ID:          uuid.New().String(),
				Name:        "Special Operation " + uuid.New().String()[:8],
				Description: "A special operation that requires high influence",
				Type:        getRandomOperationType(),
				IsSpecial:   true,
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
