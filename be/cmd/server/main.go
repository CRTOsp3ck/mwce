// cmd/server/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mwce-be/internal/app"
	"mwce-be/internal/config"
	"mwce-be/pkg/logger"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "../../configs/app.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	l := logger.NewLogger()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize the application
	application, err := app.NewApp(cfg, l)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to initialize application")
	}

	// Setup the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      application.Router,
		ReadTimeout:  time.Duration(cfg.Server.TimeoutRead) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.TimeoutWrite) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.TimeoutIdle) * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		l.Info().Msgf("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info().Msg("Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.TimeoutShutdown)*time.Second)
	defer cancel()

	// Close database connections before shutting down
	if err := application.Close(); err != nil {
		l.Error().Err(err).Msg("Error during application cleanup")
	}

	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	l.Info().Msg("Server exited gracefully")
}
