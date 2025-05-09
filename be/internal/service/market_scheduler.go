// internal/service/market_scheduler.go

package service

import (
	"time"
)

// StartPeriodicMarketPriceUpdates starts a goroutine to periodically update market prices
func (s *marketService) StartPeriodicMarketPriceUpdates() {
	// Use market price update interval from config
	updateInterval := time.Duration(s.gameConfig.MarketPriceUpdateInterval) * time.Minute

	// If config value is missing or invalid, use default
	if updateInterval < 1*time.Minute {
		updateInterval = 60 * time.Minute // Default: 1 hour

		// If mechanics config is available, use that value
		if s.gameConfig.Mechanics != nil && s.gameConfig.Mechanics.Market.PriceUpdateInterval > 0 {
			updateInterval = time.Duration(s.gameConfig.Mechanics.Market.PriceUpdateInterval) * time.Second
		}
	}

	s.logger.Info().
		Dur("interval", updateInterval).
		Msg("Starting periodic market price updates")

	// Initial update
	if err := s.UpdateMarketPrices(); err != nil {
		s.logger.Error().Err(err).Msg("Failed to perform initial market price update")
	}

	// Start ticker for periodic updates
	ticker := time.NewTicker(updateInterval)
	go func() {
		for range ticker.C {
			if err := s.UpdateMarketPrices(); err != nil {
				s.logger.Error().Err(err).Msg("Failed to update market prices")
			} else {
				s.logger.Info().Msg("Market prices updated successfully")
			}
		}
	}()
}

// // StartPeriodicMarketPriceUpdates starts a goroutine to periodically update market prices
// func (s *marketService) StartPeriodicMarketPriceUpdates() {
// 	marketUpdateInterval := time.Duration(s.gameConfig.MarketPriceUpdateInterval) * time.Minute
// 	if marketUpdateInterval < 1*time.Minute {
// 		marketUpdateInterval = 1 * time.Minute // Minimum 1 minute update interval
// 	}

// 	s.logger.Info().
// 		Dur("interval", marketUpdateInterval).
// 		Msg("Starting periodic market price updates")

// 	// Initial update
// 	if err := s.UpdateMarketPrices(); err != nil {
// 		s.logger.Error().Err(err).Msg("Failed to perform initial market price update")
// 	}

// 	// Start ticker for periodic updates
// 	ticker := time.NewTicker(marketUpdateInterval)
// 	go func() {
// 		for range ticker.C {
// 			if err := s.UpdateMarketPrices(); err != nil {
// 				s.logger.Error().Err(err).Msg("Failed to update market prices")
// 			} else {
// 				s.logger.Info().Msg("Market prices updated successfully")
// 			}
// 		}
// 	}()
// }
