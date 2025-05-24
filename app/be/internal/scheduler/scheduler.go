package scheduler

import (
	"mwce-be/internal/config"
	"mwce-be/internal/service"

	"github.com/rs/zerolog"
)

// 1. OperationsListingPeriodicRefresh
// 2. MarketPricesPeriodicUpdate

type Scheduler struct {
	operationsService service.OperationsService
	marketService     service.MarketService
	gameConfig        config.GameConfig
	logger            zerolog.Logger
}

func NewScheduler(
	operationsSvc service.OperationsService,
	marketSvc service.MarketService,
	gameConfig config.GameConfig,
	logger zerolog.Logger,
) *Scheduler {
	return &Scheduler{
		operationsService: operationsSvc,
		marketService:     marketSvc,
	}
}
