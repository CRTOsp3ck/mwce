package service

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
	Name                 string                    `yaml:"name"`
	Description          string                    `yaml:"description"`
	Type                 string                    `yaml:"type"`
	IsSpecial            bool                      `yaml:"is_special"`
	Regions              []string                  `yaml:"regions,omitempty"`               // New field for region support
	AvailabilityDuration int                       `yaml:"availability_duration,omitempty"` // New field for expiration
	Requirements         OperationRequirementsYAML `yaml:"requirements"`
	Resources            OperationResourcesYAML    `yaml:"resources"`
	Rewards              OperationRewardsYAML      `yaml:"rewards"`
	Risks                OperationRisksYAML        `yaml:"risks"`
	Duration             int                       `yaml:"duration"`
	SuccessRate          int                       `yaml:"success_rate"`
}

// YAML-specific structs for proper snake_case mapping
type OperationRequirementsYAML struct {
	MinInfluence         int    `yaml:"min_influence"`
	MaxHeat              int    `yaml:"max_heat"`
	MinTitle             string `yaml:"min_title"`
	RequiredHotspotTypes string `yaml:"required_hotspot_types"`
}

type OperationResourcesYAML struct {
	Crew     int `yaml:"crew"`
	Weapons  int `yaml:"weapons"`
	Vehicles int `yaml:"vehicles"`
	Money    int `yaml:"money"`
}

type OperationRewardsYAML struct {
	Money         int `yaml:"money"`
	Crew          int `yaml:"crew"`
	Weapons       int `yaml:"weapons"`
	Vehicles      int `yaml:"vehicles"`
	Respect       int `yaml:"respect"`
	Influence     int `yaml:"influence"`
	HeatReduction int `yaml:"heat_reduction"`
}

type OperationRisksYAML struct {
	CrewLoss     int `yaml:"crew_loss"`
	WeaponsLoss  int `yaml:"weapons_loss"`
	VehiclesLoss int `yaml:"vehicles_loss"`
	MoneyLoss    int `yaml:"money_loss"`
	HeatIncrease int `yaml:"heat_increase"`
	RespectLoss  int `yaml:"respect_loss"`
}

// Convert YAML structs to model structs
func (y OperationRequirementsYAML) ToModel() model.OperationRequirements {
	return model.OperationRequirements{
		MinInfluence:         y.MinInfluence,
		MaxHeat:              y.MaxHeat,
		MinTitle:             y.MinTitle,
		RequiredHotspotTypes: y.RequiredHotspotTypes,
	}
}

func (y OperationResourcesYAML) ToModel() model.OperationResources {
	return model.OperationResources{
		Crew:     y.Crew,
		Weapons:  y.Weapons,
		Vehicles: y.Vehicles,
		Money:    y.Money,
	}
}

func (y OperationRewardsYAML) ToModel() model.OperationRewards {
	return model.OperationRewards{
		Money:         y.Money,
		Crew:          y.Crew,
		Weapons:       y.Weapons,
		Vehicles:      y.Vehicles,
		Respect:       y.Respect,
		Influence:     y.Influence,
		HeatReduction: y.HeatReduction,
	}
}

func (y OperationRisksYAML) ToModel() model.OperationRisks {
	return model.OperationRisks{
		CrewLoss:     y.CrewLoss,
		WeaponsLoss:  y.WeaponsLoss,
		VehiclesLoss: y.VehiclesLoss,
		MoneyLoss:    y.MoneyLoss,
		HeatIncrease: y.HeatIncrease,
		RespectLoss:  y.RespectLoss,
	}
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
