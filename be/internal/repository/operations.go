// internal/repository/operations.go

package repository

import (
	"errors"
	"time"

	"mwce-be/internal/model"
	"mwce-be/internal/util"
	"mwce-be/pkg/database"

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
