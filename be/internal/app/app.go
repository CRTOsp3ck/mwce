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
		&model.Campaign{},
		&model.Chapter{},
		&model.Mission{},
		&model.MissionChoice{},
		&model.PlayerCampaignProgress{},
		&model.PlayerMissionProgress{},
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
	territoryService := service.NewTerritoryService(territoryRepo, playerRepo, sseService, *cfg.Game, logger)
	operationsService := service.NewOperationsService(operationsRepo, playerRepo, playerService, *cfg.Game, logger)
	marketService := service.NewMarketService(marketRepo, playerRepo, playerService, cfg.Game, logger)
	travelService := service.NewTravelService(playerRepo, territoryRepo, *cfg.Game, logger)
	campaignService := service.NewCampaignService(campaignRepo, playerRepo, playerService, operationsService, territoryService, logger)
	if err := campaignService.LoadCampaigns("../../configs/campaigns"); err != nil {
		logger.Warn().Err(err).Msg("Failed to load campaigns data")
	}

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
				r.Get("/{id}", campaignController.GetCampaignDetail)
				r.Post("/{id}/start", campaignController.StartCampaign)
				r.Get("/chapters/{id}", campaignController.GetChapter)
				r.Get("/missions/{id}", campaignController.GetMission)
				r.Post("/missions/{id}/start", campaignController.StartMission)
				r.Post("/missions/{id}/complete", campaignController.CompleteMission)
				r.Get("/pois", campaignController.GetPlayerActivePOIs)
				r.Post("/pois/{id}/complete", campaignController.CompletePOI)
				r.Get("/operations", campaignController.GetPlayerActiveMissionOperations)
				r.Post("/operations/{id}/start", campaignController.StartMissionOperation)
				r.Post("/operations/{id}/complete", campaignController.CompleteMissionOperation)
				r.Post("/actions/track", campaignController.TrackPlayerAction)
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
