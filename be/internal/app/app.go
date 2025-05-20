// internal/app/app.go

package app

import (
	"fmt"

	"mwce-be/internal/config"
	"mwce-be/internal/controller"
	appMiddleware "mwce-be/internal/middleware"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/service"
	"mwce-be/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
)

// App represents the application
type App struct {
	Router *chi.Mux
	DB     database.Database
	logger zerolog.Logger
}

// NewApp initializes the application
func NewApp(cfg *config.Config, logger zerolog.Logger) (*App, error) {
	// Initialize database connection
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Migrate the models
	err = db.GetDB().AutoMigrate(
		&model.Player{},
		&model.PlayerStats{},
		&model.Notification{},
		&model.Achievement{},
		&model.PlayerAchievement{},
		&model.Region{},
		&model.District{},
		&model.City{},
		&model.Hotspot{},
		&model.TerritoryAction{},
		&model.Operation{},
		&model.OperationAttempt{},
		&model.MarketListing{},
		&model.MarketTransaction{},
		&model.MarketPriceHistory{},
		&model.TravelAttempt{},

		// Campaign-related models
		&model.Campaign{},
		&model.Chapter{},
		&model.Mission{},
		&model.Branch{},
		&model.CampaignOperation{},
		&model.CampaignPOI{},
		&model.Dialogue{},
		&model.PlayerCampaignProgress{},
		&model.PlayerOperationRecord{},
		&model.PlayerPOIRecord{},
		&model.DialogueState{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(appMiddleware.NewLoggerMiddleware(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Heartbeat("/health"))

	// CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize repositories
	playerRepo := repository.NewPlayerRepository(db)
	territoryRepo := repository.NewTerritoryRepository(db)
	operationsRepo := repository.NewOperationsRepository(db)
	marketRepo := repository.NewMarketRepository(db)
	campaignRepo := repository.NewCampaignRepository(db)

	// Initialize services
	playerService := service.NewPlayerService(playerRepo, *cfg.Game, logger)
	authService := service.NewAuthService(playerRepo, playerService, cfg.JWT, logger)
	sseService := service.NewSSEService(logger)

	// Initialize territory and operations services with empty slices for providers
	territoryService := service.NewTerritoryService(territoryRepo, playerRepo, sseService, *cfg.Game, logger, []service.CustomHotspotProvider{})
	operationsService := service.NewOperationsService(operationsRepo, territoryRepo, playerRepo, playerService, sseService, *cfg.Game, logger, []service.CustomOperationsProvider{})
	campaignService := service.NewCampaignService(campaignRepo, playerRepo, territoryRepo, playerService, sseService, logger)

	// Add the campaign service as a provider to territory and operations services
	territoryService.AddCustomHotspotProvider(campaignService)
	operationsService.AddCustomOperationsProvider(campaignService)

	// -- If no regions, seed territory data --
	regions, err := territoryService.GetAllRegions()
	if err != nil {
		logger.Warn().Err(err)
		return nil, err
	}
	if regions == nil || len(regions) <= 0 {
		service.RunTerritorySeeder("../../configs/app.yaml", "../../configs/territory.yaml")
	}

	// -- If no campaigns, seed campaign data --
	campaigns, err := campaignService.GetCampaigns()
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to check existing campaigns")
	} else if len(campaigns) <= 0 {
		logger.Info().Msg("No campaigns found, seeding campaign data")
		service.RunCampaignSeeder(campaignRepo, territoryRepo, logger)
	}

	marketService := service.NewMarketService(marketRepo, playerRepo, playerService, cfg.Game, logger)
	travelService := service.NewTravelService(playerRepo, territoryRepo, sseService, *cfg.Game, logger)

	// Start scheduled jobs
	operationsService.StartPeriodicOperationsRefresh()
	marketService.StartPeriodicMarketPriceUpdates()
	territoryService.StartPeriodicIncomeGeneration()

	// Initialize controllers
	authController := controller.NewAuthController(authService, logger)
	sseController := controller.NewSSEController(authService, sseService, logger)
	playerController := controller.NewPlayerController(playerService, logger)
	territoryController := controller.NewTerritoryController(territoryService, logger)
	operationsController := controller.NewOperationsController(operationsService, logger)
	marketController := controller.NewMarketController(marketService, logger)
	travelController := controller.NewTravelController(travelService, logger)
	campaignController := controller.NewCampaignController(campaignService, logger)

	// Auth middleware
	authMiddleware := appMiddleware.NewAuthMiddleware(authService)

	// API routes
	router.Route("/api", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/register", authController.Register)
			r.Post("/auth/login", authController.Login)
			r.Get("/auth/validate", authController.Validate)

			// SSE route for real-time updates
			r.Get("/sse", sseController.HandleConnection)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// Player routes
			r.Route("/player", func(r chi.Router) {
				r.Get("/profile", playerController.GetProfile)
				r.Get("/stats", playerController.GetStats)
				r.Get("/notifications", playerController.GetNotifications)
				r.Post("/notifications/read", playerController.MarkAllNotificationsRead)
				r.Post("/notifications/{id}/read", playerController.MarkNotificationRead)
				r.Post("/collect-all", playerController.CollectAllPending)
			})

			// Travel routes
			r.Route("/travel", func(r chi.Router) {
				r.Get("/available", travelController.GetAvailableRegions)
				r.Get("/current", travelController.GetCurrentRegion)
				r.Post("/", travelController.Travel)
				r.Get("/history", travelController.GetTravelHistory)
			})

			// Territory routes
			r.Route("/territory", func(r chi.Router) {
				r.Get("/regions", territoryController.GetRegions)
				r.Get("/regions/{id}", territoryController.GetRegion)
				r.Get("/districts", territoryController.GetDistricts)
				r.Get("/districts/{id}", territoryController.GetDistrict)
				r.Get("/cities", territoryController.GetCities)
				r.Get("/cities/{id}", territoryController.GetCity)
				r.Get("/hotspots", territoryController.GetHotspots)
				r.Get("/hotspots/{id}", territoryController.GetHotspot)
				r.Get("/hotspots/controlled", territoryController.GetControlledHotspots)
				r.Get("/actions", territoryController.GetRecentActions)
				r.Post("/actions/{action}", territoryController.PerformAction)
				r.Post("/hotspots/{id}/collect", territoryController.CollectHotspotIncome)
				r.Post("/hotspots/collect-all", territoryController.CollectAllHotspotIncome)
				r.Post("/hotspots/collect-all-regional", territoryController.CollectAllRegionalHotspotIncome)
			})

			// Operations routes
			r.Route("/operations", func(r chi.Router) {
				r.Get("/", operationsController.GetAvailableOperations)
				r.Get("/{id}", operationsController.GetOperation)
				r.Get("/current", operationsController.GetCurrentOperations)
				r.Get("/completed", operationsController.GetCompletedOperations)
				r.Post("/{id}/start", operationsController.StartOperation)
				r.Post("/{id}/cancel", operationsController.CancelOperation)
				r.Post("/{id}/collect", operationsController.CollectOperation)
				r.Post("/{id}/collect-reward", operationsController.CollectOperationReward)
				r.Get("/refresh-info", operationsController.GetOperationsRefreshInfo)
			})

			// Market routes
			r.Route("/market", func(r chi.Router) {
				r.Get("/listings", marketController.GetListings)
				r.Get("/listings/{type}", marketController.GetListing)
				r.Get("/transactions", marketController.GetTransactions)
				r.Get("/history", marketController.GetPriceHistory)
				r.Get("/history/{type}", marketController.GetResourcePriceHistory)
				r.Post("/buy", marketController.BuyResource)
				r.Post("/sell", marketController.SellResource)
			})

			// Campaign routes
			r.Route("/campaigns", func(r chi.Router) {
				r.Get("/", campaignController.GetCampaigns)
				r.Get("/{id}", campaignController.GetCampaign)

				r.Get("/{id}/chapters", campaignController.GetChaptersByCampaign)
				r.Get("/chapters/{id}", campaignController.GetChapter)
				r.Get("/chapters/{id}/missions", campaignController.GetMissionsByChapter)
				r.Get("/missions/{id}", campaignController.GetMission)
				r.Get("/missions/{id}/branches", campaignController.GetBranchesByMission)
				r.Get("/branches/{id}", campaignController.GetBranch)
				r.Get("/branches/{id}/operations", campaignController.GetOperationsByBranch)
				r.Get("/branches/{id}/pois", campaignController.GetPOIsByBranch)
				r.Get("/pois/{id}", campaignController.GetPOI)
				r.Get("/pois/{id}/dialogues", campaignController.GetPOIDialogues)

				r.Get("/{id}/progress", campaignController.GetPlayerProgress)
				r.Post("/{id}/start", campaignController.StartCampaign)
				r.Get("/{id}/current-mission", campaignController.GetCurrentMission)

				r.Post("/missions/{missionId}/select-branch", campaignController.SelectBranch)
				r.Post("/missions/{missionId}/branches/{branchId}/complete", campaignController.CompleteBranch)
				r.Get("/branches/{id}/check-completion", campaignController.CheckBranchCompletion)

				r.Post("/pois/{id}/interact", campaignController.InteractWithPOI)
				r.Post("/pois/{id}/complete", campaignController.CompletePOI)

				r.Post("/operations/{id}/complete", campaignController.CompleteOperation)
			})
		})
	})

	// Create app
	app := &App{
		Router: router,
		DB:     db,
		logger: logger,
	}

	return app, nil
}

// Close cleans up application resources
func (a *App) Close() error {
	return a.DB.Close()
}
