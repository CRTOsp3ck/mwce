// cmd/seed/operations.go
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/pkg/database"
	"mwce-be/pkg/logger"
)

// OperationsData represents the structure of the operations.yaml file
type OperationsData struct {
	BasicOperations   []OperationTemplate `yaml:"basic_operations"`
	SpecialOperations []OperationTemplate `yaml:"special_operations"`
}

// OperationTemplate represents an operation template in the operations structure
type OperationTemplate struct {
	Name         string                      `yaml:"name"`
	Description  string                      `yaml:"description"`
	Type         string                      `yaml:"type"`
	IsSpecial    bool                        `yaml:"is_special"`
	Requirements model.OperationRequirements `yaml:"requirements"`
	Resources    model.OperationResources    `yaml:"resources"`
	Rewards      model.OperationRewards      `yaml:"rewards"`
	Risks        model.OperationRisks        `yaml:"risks"`
	Duration     int                         `yaml:"duration"`
	SuccessRate  int                         `yaml:"success_rate"`
}

func startOperationsSeeder() {
	// Parse command line flags
	configPath := flag.String("config", "../../configs/app.yaml", "Path to app configuration file")
	operationsPath := flag.String("operations", "../../configs/operations.yaml", "Path to operations configuration file")
	flag.Parse()

	// Initialize logger
	l := logger.NewLogger()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Read operations data
	operationsData, err := loadOperationsData(*operationsPath)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load operations data")
	}

	// Connect to database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Clear existing operations data
	if err := clearExistingOperationsData(db.GetDB(), l); err != nil {
		l.Fatal().Err(err).Msg("Failed to clear existing operations data")
	}

	// Seed operations data
	if err := seedOperationsData(db.GetDB(), operationsData, l); err != nil {
		l.Fatal().Err(err).Msg("Failed to seed operations data")
	}

	l.Info().Msg("Operations data seeded successfully")
}

// loadOperationsData reads the operations data from a YAML file
func loadOperationsData(path string) (*OperationsData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read operations data file: %w", err)
	}

	var operationsData OperationsData
	if err := yaml.Unmarshal(data, &operationsData); err != nil {
		return nil, fmt.Errorf("failed to parse operations data: %w", err)
	}

	return &operationsData, nil
}

// clearExistingOperationsData clears existing operations data from the database
func clearExistingOperationsData(db *gorm.DB, l zerolog.Logger) error {
	l.Info().Msg("Clearing existing operations data...")

	// First delete operation attempts
	if err := db.Exec("DELETE FROM operation_attempts").Error; err != nil {
		return err
	}

	// Then delete operations
	if err := db.Exec("DELETE FROM operations").Error; err != nil {
		return err
	}

	return nil
}

// seedOperationsData seeds the operations data into the database
func seedOperationsData(db *gorm.DB, data *OperationsData, l zerolog.Logger) error {
	l.Info().Msg("Seeding operations data...")

	basicCount := 0
	specialCount := 0

	// Start a transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Helper function to handle error and rollback
	handleError := func(err error, msg string) error {
		tx.Rollback()
		return fmt.Errorf("%s: %w", msg, err)
	}

	now := time.Now()
	availableUntil := now.Add(24 * time.Hour) // Operations available for next 24 hours

	// Create basic operations
	for _, opTemplate := range data.BasicOperations {
		operation := model.Operation{
			ID:             uuid.NewString(),
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

		if err := tx.Create(&operation).Error; err != nil {
			return handleError(err, "failed to create basic operation")
		}

		basicCount++
	}

	// Create special operations
	for _, opTemplate := range data.SpecialOperations {
		operation := model.Operation{
			ID:             uuid.NewString(),
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

		if err := tx.Create(&operation).Error; err != nil {
			return handleError(err, "failed to create special operation")
		}

		specialCount++
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	l.Info().
		Int("basic_operations", basicCount).
		Int("special_operations", specialCount).
		Msg("Operations data seeded successfully")

	return nil
}
