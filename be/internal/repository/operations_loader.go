package repository

import (
	"fmt"
	"mwce-be/internal/model"
	"os"

	"gopkg.in/yaml.v3"
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

// loadOperationsFromYAML reads operations data from the configured YAML file
func loadOperationsFromYAML() (*OperationsData, error) {
	// Get the operations YAML file path from environment or use default
	operationsFile := os.Getenv("OPERATIONS_FILE")
	if operationsFile == "" {
		operationsFile = "../../configs/operations.yaml"
	}

	data, err := os.ReadFile(operationsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read operations data file: %w", err)
	}

	var operationsData OperationsData
	if err := yaml.Unmarshal(data, &operationsData); err != nil {
		return nil, fmt.Errorf("failed to parse operations data: %w", err)
	}

	return &operationsData, nil
}
