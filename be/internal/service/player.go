// internal/service/player.go

package service

import (
	"fmt"
	"time"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
)

// PlayerService handles player-related business logic
type PlayerService interface {
	GetProfile(playerID string) (*model.Player, error)
	GetStats(playerID string) (*model.PlayerStats, error)
	GetNotifications(playerID string) ([]model.Notification, error)
	MarkAllNotificationsRead(playerID string) error
	MarkNotificationRead(notificationID, playerID string) error
	CollectAllPending(playerID string) (*model.CollectAllResponse, error)
	UpdatePlayerResources(playerID string, resourceUpdates map[string]int) error
	AddNotification(playerID, message, notificationType string) error
	UpdateTitle(playerID string) error
	CreateNewPlayer(name, email, password string) (*model.Player, error)
}

type playerService struct {
	playerRepo repository.PlayerRepository
	gameConfig config.GameConfig
	logger     zerolog.Logger
}

// NewPlayerService creates a new player service
func NewPlayerService(playerRepo repository.PlayerRepository, gameConfig config.GameConfig, logger zerolog.Logger) PlayerService {
	return &playerService{
		playerRepo: playerRepo,
		gameConfig: gameConfig,
		logger:     logger,
	}
}

// CreateNewPlayer creates a new player with initial resources from config
func (s *playerService) CreateNewPlayer(name, email, password string) (*model.Player, error) {
	// Create player with initial resource values from config
	now := time.Now()
	player := &model.Player{
		Name:        name,
		Email:       email,
		Password:    password,                  // Should be hashed by caller
		Title:       util.PlayerTitleAssociate, // Starting title
		Money:       s.gameConfig.ResourceLimit.InitialMoney,
		Crew:        s.gameConfig.ResourceLimit.InitialCrew,
		MaxCrew:     s.gameConfig.ResourceLimit.MaxCrew,
		Weapons:     s.gameConfig.ResourceLimit.InitialWeapons,
		MaxWeapons:  s.gameConfig.ResourceLimit.MaxWeapons,
		Vehicles:    s.gameConfig.ResourceLimit.InitialVehicles,
		MaxVehicles: s.gameConfig.ResourceLimit.MaxVehicles,
		Respect:     s.gameConfig.ResourceLimit.InitialRespect,
		Influence:   s.gameConfig.ResourceLimit.InitialInfluence,
		Heat:        s.gameConfig.ResourceLimit.InitialHeat,
		CreatedAt:   now,
		LastActive:  now,
	}

	return player, nil
}

// GetProfile retrieves a player's profile
func (s *playerService) GetProfile(playerID string) (*model.Player, error) {
	return s.playerRepo.GetPlayerByID(playerID)
}

// GetStats retrieves a player's statistics
func (s *playerService) GetStats(playerID string) (*model.PlayerStats, error) {
	return s.playerRepo.GetPlayerStats(playerID)
}

// GetNotifications retrieves a player's notifications
func (s *playerService) GetNotifications(playerID string) ([]model.Notification, error) {
	return s.playerRepo.GetNotifications(playerID)
}

// MarkAllNotificationsRead marks all notifications as read for a player
func (s *playerService) MarkAllNotificationsRead(playerID string) error {
	return s.playerRepo.MarkAllNotificationsRead(playerID)
}

// MarkNotificationRead marks a specific notification as read
func (s *playerService) MarkNotificationRead(notificationID, playerID string) error {
	// In a real implementation, we'd check if the notification belongs to the player
	return s.playerRepo.MarkNotificationRead(notificationID)
}

// CollectAllPending collects all pending resources for a player
func (s *playerService) CollectAllPending(playerID string) (*model.CollectAllResponse, error) {
	// Collect pending resources
	collectedAmount, err := s.playerRepo.CollectAllPending(playerID)
	if err != nil {
		return nil, err
	}

	// Generate response message
	var message string
	if collectedAmount > 0 {
		message = fmt.Sprintf("Successfully collected $%s from your controlled businesses.", formatMoney(collectedAmount))
	} else {
		message = "No resources available to collect at this time."
	}

	return &model.CollectAllResponse{
		CollectedAmount: collectedAmount,
		Message:         message,
	}, nil
}

// UpdatePlayerResources updates multiple resources for a player
func (s *playerService) UpdatePlayerResources(playerID string, resourceUpdates map[string]int) error {
	for resourceType, amount := range resourceUpdates {
		if err := s.playerRepo.UpdatePlayerResource(playerID, resourceType, amount); err != nil {
			s.logger.Error().Err(err).
				Str("playerID", playerID).
				Str("resourceType", resourceType).
				Int("amount", amount).
				Msg("Failed to update player resource")
			return err
		}
	}

	// Update the player's title based on their new stats
	if err := s.UpdateTitle(playerID); err != nil {
		s.logger.Error().Err(err).Str("playerID", playerID).Msg("Failed to update player title")
	}

	return nil
}

// AddNotification adds a notification for a player
func (s *playerService) AddNotification(playerID, message, notificationType string) error {
	// Check the notification limits from config
	if s.gameConfig.Mechanics != nil {
		notificationsForPlayer, err := s.playerRepo.GetNotifications(playerID)
		if err == nil {
			maxUnread := s.gameConfig.Mechanics.Notifications.MaxUnread
			maxTotal := s.gameConfig.Mechanics.Notifications.MaxTotal

			// Count unread notifications
			unreadCount := 0
			for _, n := range notificationsForPlayer {
				if !n.Read {
					unreadCount++
				}
			}

			// If over limits, mark oldest as read or prune
			if unreadCount >= maxUnread || len(notificationsForPlayer) >= maxTotal {
				// Implementation for pruning would go here
				s.logger.Warn().
					Str("playerID", playerID).
					Int("unreadCount", unreadCount).
					Int("totalCount", len(notificationsForPlayer)).
					Msg("Notification limits reached")
			}
		}
	}

	notification := &model.Notification{
		PlayerID:  playerID,
		Message:   message,
		Type:      notificationType,
		Timestamp: time.Now(),
		Read:      false,
	}
	return s.playerRepo.AddNotification(notification)
}

// UpdateTitle updates a player's title based on their respect and influence
func (s *playerService) UpdateTitle(playerID string) error {
	player, err := s.playerRepo.GetPlayerByID(playerID)
	if err != nil {
		return err
	}

	// Using title requirements from mechanics config
	newTitle := util.PlayerTitleAssociate
	respectInfluence := player.Respect + player.Influence

	// Use title requirements from config if available
	if s.gameConfig.Mechanics != nil && len(s.gameConfig.Mechanics.Progression.TitleRequirements) > 0 {
		for title, requirements := range s.gameConfig.Mechanics.Progression.TitleRequirements {
			if respectValue, exists := requirements["respect_influence"]; exists {
				if respectInfluence >= respectValue {
					newTitle = title
				}
			}
		}
	} else {
		// Fallback to hardcoded values if config not available
		if respectInfluence >= 20 && respectInfluence < 40 {
			newTitle = util.PlayerTitleSoldier
		} else if respectInfluence >= 40 && respectInfluence < 60 {
			newTitle = util.PlayerTitleCapo
		} else if respectInfluence >= 60 && respectInfluence < 80 {
			newTitle = util.PlayerTitleUnderboss
		} else if respectInfluence >= 80 && respectInfluence < 100 {
			newTitle = util.PlayerTitleConsigliere
		} else if respectInfluence >= 100 && respectInfluence < 150 {
			newTitle = util.PlayerTitleBoss
		} else if respectInfluence >= 150 {
			newTitle = util.PlayerTitleGodfather
		}
	}

	// Only update if the title has changed
	if player.Title != newTitle {
		player.Title = newTitle
		if err := s.playerRepo.UpdatePlayer(player); err != nil {
			return err
		}

		// Add a notification about the title change
		message := fmt.Sprintf("Congratulations! Your criminal influence has earned you the title of %s.", newTitle)
		if err := s.AddNotification(playerID, message, util.NotificationTypeSystem); err != nil {
			s.logger.Error().Err(err).Msg("Failed to add title change notification")
		}
	}

	return nil
}
