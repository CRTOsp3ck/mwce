// cmd/seed/market.go
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/internal/util"
	"mwce-be/pkg/database"
	"mwce-be/pkg/logger"
)

func startMarketSeeder() {
	// Parse command line flags
	configPath := flag.String("config", "../../configs/app.yaml", "Path to app configuration file")
	flag.Parse()

	// Initialize logger
	l := logger.NewLogger()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Clear existing market data
	if err := clearExistingMarketData(db.GetDB(), l); err != nil {
		l.Fatal().Err(err).Msg("Failed to clear existing market data")
	}

	// Seed market data
	if err := seedMarketData(db.GetDB(), l); err != nil {
		l.Fatal().Err(err).Msg("Failed to seed market data")
	}

	l.Info().Msg("Market data seeded successfully")
}

// clearExistingMarketData clears existing market data from the database
func clearExistingMarketData(db *gorm.DB, l zerolog.Logger) error {
	l.Info().Msg("Clearing existing market data...")

	// Delete in reverse order of dependencies
	if err := db.Exec("DELETE FROM market_price_histories").Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM market_transactions").Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM market_listings").Error; err != nil {
		return err
	}

	return nil
}

// seedMarketData seeds the market data into the database
func seedMarketData(db *gorm.DB, l zerolog.Logger) error {
	l.Info().Msg("Seeding market data...")

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

	// Create initial market listings
	listings := []model.MarketListing{
		{
			ID:              uuid.New().String(),
			Type:            util.ResourceTypeCrew,
			Price:           1000,
			Quantity:        999,
			Trend:           util.PriceTrendStable,
			TrendPercentage: 0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.New().String(),
			Type:            util.ResourceTypeWeapons,
			Price:           2000,
			Quantity:        999,
			Trend:           util.PriceTrendStable,
			TrendPercentage: 0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.New().String(),
			Type:            util.ResourceTypeVehicles,
			Price:           5000,
			Quantity:        999,
			Trend:           util.PriceTrendStable,
			TrendPercentage: 0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	for _, listing := range listings {
		if err := tx.Create(&listing).Error; err != nil {
			return handleError(err, "failed to create market listing")
		}

		// Create initial price history for each listing
		history := model.MarketPriceHistory{
			ID:           uuid.New().String(),
			ResourceType: listing.Type,
			Price:        listing.Price,
			Timestamp:    now,
			CreatedAt:    now,
		}

		if err := tx.Create(&history).Error; err != nil {
			return handleError(err, "failed to create market price history")
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	l.Info().
		Int("listings", len(listings)).
		Msg("Market data seeded successfully")

	return nil
}
