package service

import "time"

// StartPeriodicOperationsRefresh starts a goroutine to periodically refresh operations
func (s *operationsService) StartPeriodicOperationsRefresh() {
	refreshInterval := time.Duration(s.gameConfig.OperationsRefreshInterval) * time.Minute
	if refreshInterval < 1*time.Minute {
		refreshInterval = 1 * time.Minute // Minimum 1 minute refresh interval
	}

	s.logger.Info().
		Dur("interval", refreshInterval).
		Int("basic_count", s.gameConfig.DailyOperationsCount).
		Int("special_count", s.gameConfig.SpecialOperationsCount).
		Msg("Starting periodic operations refresh")

	// Initial refresh
	if err := s.RefreshDailyOperations(); err != nil {
		s.logger.Error().Err(err).Msg("Failed to perform initial operations refresh")
	}

	// Start ticker for periodic refresh
	ticker := time.NewTicker(refreshInterval)
	go func() {
		for range ticker.C {
			if err := s.RefreshDailyOperations(); err != nil {
				s.logger.Error().Err(err).Msg("Failed to refresh operations")
			} else {
				s.logger.Info().Msg("Operations refreshed successfully")
			}
		}
	}()
}
