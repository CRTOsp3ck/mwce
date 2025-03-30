// internal/service/player.go

package service

import (
	"fmt"
	"time"

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
}

type playerService struct {
	playerRepo repository.PlayerRepository
	logger     zerolog.Logger
}

// NewPlayerService creates a new player service
func NewPlayerService(playerRepo repository.PlayerRepository, logger zerolog.Logger) PlayerService {
	return &playerService{
		playerRepo: playerRepo,
		logger:     logger,
	}
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

	// Determine the new title based on respect and influence
	newTitle := util.PlayerTitleAssociate

	respectInfluence := player.Respect + player.Influence

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
