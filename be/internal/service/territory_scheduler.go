package service

import "time"

// StartPeriodicIncomeGeneration starts a goroutine that periodically generates income for all hotspots
func (s *territoryService) StartPeriodicIncomeGeneration() {
	ticker := time.NewTicker(time.Second) // Check every second

	go func() {
		for {
			select {
			case <-ticker.C:
				// s.logger.Info().Msg("Running periodic income generation for hotspots")
				if err := s.UpdateHotspotIncome(); err != nil {
					s.logger.Error().Err(err).Msg("Failed to update hotspot income")
				}
			}
		}
	}()

	s.logger.Info().Msg("Started periodic income generation scheduler")
}
