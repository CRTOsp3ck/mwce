// internal/repository/player.go

package repository

import (
	"database/sql"
	"errors"
	"time"

	"mwce-be/internal/model"
	"mwce-be/pkg/database"

	"gorm.io/gorm"
)

// PlayerRepository handles database operations for players
type PlayerRepository interface {
	CreatePlayer(player *model.Player) error
	GetPlayerByID(id string) (*model.Player, error)
	GetPlayerByEmail(email string) (*model.Player, error)
	UpdatePlayer(player *model.Player) error
	DeletePlayer(id string) error
	GetPlayerStats(playerID string) (*model.PlayerStats, error)
	UpdatePlayerStats(stats *model.PlayerStats) error
	AddNotification(notification *model.Notification) error
	GetNotifications(playerID string) ([]model.Notification, error)
	MarkAllNotificationsRead(playerID string) error
	MarkNotificationRead(notificationID string) error
	UpdatePlayerResource(playerID, resourceType string, amount int) error
	GetControlledHotspotsCount(playerID string) (int, error)
	GetTotalHotspotsCount() (int, error)
	CalculateHourlyRevenue(playerID string) (int, error)
	CalculatePendingCollections(playerID string) (int, error)
	CollectAllPending(playerID string) (int, error)
	// New travel-related methods
	CreateTravelAttempt(attempt *model.TravelAttempt) error
	GetTravelHistory(playerID string, limit int) ([]model.TravelAttempt, error)
	GetPlayerCurrentRegion(playerID string) (*string, error)
}

type playerRepository struct {
	db database.Database
}

// NewPlayerRepository creates a new player repository
func NewPlayerRepository(db database.Database) PlayerRepository {
	return &playerRepository{
		db: db,
	}
}

// CreatePlayer creates a new player in the database
func (r *playerRepository) CreatePlayer(player *model.Player) error {
	return r.db.GetDB().Create(player).Error
}

// GetPlayerByID retrieves a player by ID
func (r *playerRepository) GetPlayerByID(id string) (*model.Player, error) {
	var player model.Player
	if err := r.db.GetDB().Where("id = ?", id).First(&player).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("player not found")
		}
		return nil, err
	}

	// Get additional calculated fields
	controlledHotspots, err := r.GetControlledHotspotsCount(id)
	if err != nil {
		return nil, err
	}
	player.ControlledHotspots = controlledHotspots

	totalHotspots, err := r.GetTotalHotspotsCount()
	if err != nil {
		return nil, err
	}
	player.TotalHotspots = totalHotspots

	hourlyRevenue, err := r.CalculateHourlyRevenue(id)
	if err != nil {
		return nil, err
	}
	player.HourlyRevenue = hourlyRevenue

	pendingCollections, err := r.CalculatePendingCollections(id)
	if err != nil {
		return nil, err
	}
	player.PendingCollections = pendingCollections

	return &player, nil
}

// GetPlayerByEmail retrieves a player by email
func (r *playerRepository) GetPlayerByEmail(email string) (*model.Player, error) {
	var player model.Player
	if err := r.db.GetDB().Where("email = ?", email).First(&player).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("player not found")
		}
		return nil, err
	}
	return &player, nil
}

// UpdatePlayer updates a player in the database
func (r *playerRepository) UpdatePlayer(player *model.Player) error {
	return r.db.GetDB().Save(player).Error
}

// DeletePlayer deletes a player from the database
func (r *playerRepository) DeletePlayer(id string) error {
	return r.db.GetDB().Delete(&model.Player{}, "id = ?", id).Error
}

// GetPlayerStats retrieves player statistics
func (r *playerRepository) GetPlayerStats(playerID string) (*model.PlayerStats, error) {
	var stats model.PlayerStats
	if err := r.db.GetDB().Where("player_id = ?", playerID).First(&stats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If stats don't exist, create them
			stats = model.PlayerStats{
				PlayerID:  playerID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := r.db.GetDB().Create(&stats).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &stats, nil
}

// UpdatePlayerStats updates player statistics
func (r *playerRepository) UpdatePlayerStats(stats *model.PlayerStats) error {
	return r.db.GetDB().Save(stats).Error
}

// AddNotification adds a notification for a player
func (r *playerRepository) AddNotification(notification *model.Notification) error {
	return r.db.GetDB().Create(notification).Error
}

// GetNotifications retrieves notifications for a player
func (r *playerRepository) GetNotifications(playerID string) ([]model.Notification, error) {
	var notifications []model.Notification
	if err := r.db.GetDB().Where("player_id = ?", playerID).Order("timestamp DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkAllNotificationsRead marks all notifications as read for a player
func (r *playerRepository) MarkAllNotificationsRead(playerID string) error {
	return r.db.GetDB().Model(&model.Notification{}).Where("player_id = ?", playerID).Update("read", true).Error
}

// MarkNotificationRead marks a specific notification as read
func (r *playerRepository) MarkNotificationRead(notificationID string) error {
	return r.db.GetDB().Model(&model.Notification{}).Where("id = ?", notificationID).Update("read", true).Error
}

// UpdatePlayerResource updates a player's resource amount
func (r *playerRepository) UpdatePlayerResource(playerID, resourceType string, amount int) error {
	var updateField string

	switch resourceType {
	case "crew":
		updateField = "crew"
	case "weapons":
		updateField = "weapons"
	case "vehicles":
		updateField = "vehicles"
	case "money":
		updateField = "money"
	case "respect":
		updateField = "respect"
	case "influence":
		updateField = "influence"
	case "heat":
		updateField = "heat"
	default:
		return errors.New("invalid resource type")
	}

	// Using SQL expression to prevent negative values
	return r.db.GetDB().Model(&model.Player{}).
		Where("id = ?", playerID).
		Updates(map[string]interface{}{
			updateField:   gorm.Expr("GREATEST(0, ? + ?)", gorm.Expr(updateField), amount),
			"last_active": time.Now(),
		}).Error
}

// GetControlledHotspotsCount counts hotspots controlled by a player
func (r *playerRepository) GetControlledHotspotsCount(playerID string) (int, error) {
	var count int64
	if err := r.db.GetDB().Model(&model.Hotspot{}).Where("controller_id = ?", playerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetTotalHotspotsCount counts all legal hotspots
func (r *playerRepository) GetTotalHotspotsCount() (int, error) {
	var count int64
	if err := r.db.GetDB().Model(&model.Hotspot{}).Where("is_legal = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// CalculateHourlyRevenue calculates the total hourly revenue for a player
func (r *playerRepository) CalculateHourlyRevenue(playerID string) (int, error) {
	var total sql.NullInt64
	if err := r.db.GetDB().Model(&model.Hotspot{}).
		Where("controller_id = ?", playerID).
		Select("COALESCE(SUM(income), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return int(total.Int64), nil
}

// CalculatePendingCollections calculates the total pending collections for a player
func (r *playerRepository) CalculatePendingCollections(playerID string) (int, error) {
	var total sql.NullInt64
	if err := r.db.GetDB().Model(&model.Hotspot{}).
		Where("controller_id = ?", playerID).
		Select("COALESCE(SUM(pending_collection),0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return int(total.Int64), nil
}

// CollectAllPending collects all pending resources for a player
func (r *playerRepository) CollectAllPending(playerID string) (int, error) {
	var pendingTotal int64

	// Get the total pending amount
	if err := r.db.GetDB().Model(&model.Hotspot{}).
		Where("controller_id = ?", playerID).
		Select("SUM(pending_collection)").
		Scan(&pendingTotal).Error; err != nil {
		return 0, err
	}

	// Reset pending collections on all hotspots
	if err := r.db.GetDB().Model(&model.Hotspot{}).
		Where("controller_id = ?", playerID).
		Updates(map[string]interface{}{
			"pending_collection":   0,
			"last_collection_time": time.Now(),
		}).Error; err != nil {
		return 0, err
	}

	// Update player's money
	if err := r.UpdatePlayerResource(playerID, "money", int(pendingTotal)); err != nil {
		return 0, err
	}

	return int(pendingTotal), nil
}

// CreateTravelAttempt creates a new travel attempt record
func (r *playerRepository) CreateTravelAttempt(attempt *model.TravelAttempt) error {
	return r.db.GetDB().Create(attempt).Error
}

// GetTravelHistory retrieves travel history for a player
func (r *playerRepository) GetTravelHistory(playerID string, limit int) ([]model.TravelAttempt, error) {
	var attempts []model.TravelAttempt

	query := r.db.GetDB().Where("player_id = ?", playerID).Order("timestamp DESC")

	// Apply limit if provided
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&attempts).Error; err != nil {
		return nil, err
	}

	return attempts, nil
}

// GetPlayerCurrentRegion returns the current region ID for a player
func (r *playerRepository) GetPlayerCurrentRegion(playerID string) (*string, error) {
	var player model.Player

	if err := r.db.GetDB().Select("current_region_id").Where("id = ?", playerID).First(&player).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("player not found")
		}
		return nil, err
	}

	return player.CurrentRegionID, nil
}
