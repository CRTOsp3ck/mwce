package service

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/pkg/database"
	"mwce-be/pkg/logger"
)

// TerritoryData represents the structure of the territory.yaml file
type TerritoryData struct {
	Regions []RegionData `yaml:"regions"`
}

// RegionData represents a region in the territory structure
type RegionData struct {
	ID        string         `yaml:"id"`
	Name      string         `yaml:"name"`
	Districts []DistrictData `yaml:"districts"`
}

// DistrictData represents a district in the territory structure
type DistrictData struct {
	ID     string     `yaml:"id"`
	Name   string     `yaml:"name"`
	Cities []CityData `yaml:"cities"`
}

// CityData represents a city in the territory structure
type CityData struct {
	ID       string        `yaml:"id"`
	Name     string        `yaml:"name"`
	Hotspots []HotspotData `yaml:"hotspots"`
}

// HotspotData represents a hotspot in the territory structure
type HotspotData struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	Type         string `yaml:"type"`
	BusinessType string `yaml:"business_type"`
	IsLegal      bool   `yaml:"is_legal"`
	Income       int    `yaml:"income"`
}

func RunTerritorySeeder(configPath, territoryPath string) {
	// Initialize logger
	l := logger.NewLogger()

	// Load configuration
	cfg, err := config.LoadAllConfigs(configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Read territory data
	territoryData, err := loadTerritoryData(territoryPath)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load territory data")
	}

	// Connect to database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Clear existing territory data
	if err := clearExistingTerritoryData(db.GetDB(), l); err != nil {
		l.Fatal().Err(err).Msg("Failed to clear existing territory data")
	}

	// Seed territory data
	if err := seedTerritoryData(db.GetDB(), territoryData, l); err != nil {
		l.Fatal().Err(err).Msg("Failed to seed territory data")
	}

	l.Info().Msg("Territory data seeded successfully")
}

// loadTerritoryData reads the territory data from a YAML file
func loadTerritoryData(path string) (*TerritoryData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read territory data file: %w", err)
	}

	var territoryData TerritoryData
	if err := yaml.Unmarshal(data, &territoryData); err != nil {
		return nil, fmt.Errorf("failed to parse territory data: %w", err)
	}

	return &territoryData, nil
}

// clearExistingTerritoryData clears existing territory data from the database
func clearExistingTerritoryData(db *gorm.DB, l zerolog.Logger) error {
	l.Info().Msg("Clearing existing territory data...")

	// Delete in reverse order of dependencies
	if err := db.Exec("DELETE FROM territory_actions").Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM hotspots").Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM cities").Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM districts").Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM regions").Error; err != nil {
		return err
	}

	return nil
}

// seedTerritoryData seeds the territory data into the database
func seedTerritoryData(db *gorm.DB, data *TerritoryData, l zerolog.Logger) error {
	l.Info().Msg("Seeding territory data...")

	regionCount := 0
	districtCount := 0
	cityCount := 0
	hotspotCount := 0

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

	// Create regions
	for _, regionData := range data.Regions {
		region := model.Region{
			ID:        regionData.ID, // Use ID from YAML file
			Name:      regionData.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := tx.Create(&region).Error; err != nil {
			return handleError(err, "failed to create region")
		}

		regionCount++

		// Create districts for this region
		for _, districtData := range regionData.Districts {
			district := model.District{
				ID:        districtData.ID, // Use ID from YAML file
				Name:      districtData.Name,
				RegionID:  region.ID,
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := tx.Create(&district).Error; err != nil {
				return handleError(err, "failed to create district")
			}

			districtCount++

			// Create cities for this district
			for _, cityData := range districtData.Cities {
				city := model.City{
					ID:         cityData.ID, // Use ID from YAML file
					Name:       cityData.Name,
					DistrictID: district.ID,
					CreatedAt:  now,
					UpdatedAt:  now,
				}

				if err := tx.Create(&city).Error; err != nil {
					return handleError(err, "failed to create city")
				}

				cityCount++

				// Create hotspots for this city
				for _, hotspotData := range cityData.Hotspots {
					hotspot := model.Hotspot{
						ID:                 hotspotData.ID, // Use ID from YAML file
						Name:               hotspotData.Name,
						CityID:             city.ID,
						Type:               hotspotData.Type,
						BusinessType:       hotspotData.BusinessType,
						IsLegal:            hotspotData.IsLegal,
						Income:             hotspotData.Income,
						PendingCollection:  0,
						LastCollectionTime: &now,
						CreatedAt:          now,
						UpdatedAt:          now,
					}

					if err := tx.Create(&hotspot).Error; err != nil {
						return handleError(err, "failed to create hotspot")
					}

					hotspotCount++
				}
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	l.Info().
		Int("regions", regionCount).
		Int("districts", districtCount).
		Int("cities", cityCount).
		Int("hotspots", hotspotCount).
		Msg("Territory data seeded successfully")

	return nil
}
