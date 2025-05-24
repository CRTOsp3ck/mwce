// internal/repository/territory.go

package repository

import (
	"errors"
	"time"

	"mwce-be/internal/model"
	"mwce-be/pkg/database"

	"gorm.io/gorm"
)

// TerritoryRepository handles database operations for territories
type TerritoryRepository interface {
	GetDB() *gorm.DB
	GetAllRegions() ([]model.Region, error)
	GetRegionByID(id string) (*model.Region, error)
	GetAllDistricts() ([]model.District, error)
	GetDistrictsByRegionID(regionID string) ([]model.District, error)
	GetDistrictByID(id string) (*model.District, error)
	GetAllCities() ([]model.City, error)
	GetCitiesByDistrictID(districtID string) ([]model.City, error)
	GetCityByID(id string) (*model.City, error)
	GetAllHotspots() ([]model.Hotspot, error)
	GetHotspotsByCity(cityID string) ([]model.Hotspot, error)
	GetHotspotByID(id string) (*model.Hotspot, error)
	GetHotspotsByRegion(regionID string) ([]model.Hotspot, error)
	GetControlledHotspots(playerID string) ([]model.Hotspot, error)
	GetControlledHotspotsByRegion(playerID string, regionID string) ([]model.Hotspot, error)
	UpdateHotspot(hotspot *model.Hotspot) error
	AddTerritoryAction(action *model.TerritoryAction) error
	GetRecentActions(limit int) ([]model.TerritoryAction, error)
	GetRecentActionsByPlayer(playerID string, limit int) ([]model.TerritoryAction, error)
	GetRecentActionsByPlayerAndRegion(playerID string, regionID string, limit int) ([]model.TerritoryAction, error)
	UpdateHotspotPendingCollection(hotspotID string, amount int) error
	RefreshIllegalBusinesses() error
	GetAllControlledLegalHotspots() ([]model.Hotspot, error)
	GetAllControlledLegalHotspotsByRegion(regionID string) ([]model.Hotspot, error)
	UpdateHotspotLastIncomeTime(hotspotID string, lastIncomeTime time.Time) error
}

type territoryRepository struct {
	db database.Database
}

// NewTerritoryRepository creates a new territory repository
func NewTerritoryRepository(db database.Database) TerritoryRepository {
	return &territoryRepository{
		db: db,
	}
}

// GetDB returns the database connection instance
func (r *territoryRepository) GetDB() *gorm.DB {
	return r.db.GetDB()
}

// GetAllRegions retrieves all regions
func (r *territoryRepository) GetAllRegions() ([]model.Region, error) {
	var regions []model.Region
	if err := r.db.GetDB().Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

// GetRegionByID retrieves a region by ID
func (r *territoryRepository) GetRegionByID(id string) (*model.Region, error) {
	var region model.Region
	if err := r.db.GetDB().
		Preload("Districts").
		Where("id = ?", id).
		First(&region).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("region not found")
		}
		return nil, err
	}
	return &region, nil
}

// GetAllDistricts retrieves all districts
func (r *territoryRepository) GetAllDistricts() ([]model.District, error) {
	var districts []model.District
	if err := r.db.GetDB().Find(&districts).Error; err != nil {
		return nil, err
	}
	return districts, nil
}

// GetDistrictsByRegionID retrieves districts by region ID
func (r *territoryRepository) GetDistrictsByRegionID(regionID string) ([]model.District, error) {
	var districts []model.District
	if err := r.db.GetDB().Where("region_id = ?", regionID).Find(&districts).Error; err != nil {
		return nil, err
	}
	return districts, nil
}

// GetDistrictByID retrieves a district by ID
func (r *territoryRepository) GetDistrictByID(id string) (*model.District, error) {
	var district model.District
	if err := r.db.GetDB().
		Preload("Cities").
		Where("id = ?", id).
		First(&district).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("district not found")
		}
		return nil, err
	}
	return &district, nil
}

// GetAllCities retrieves all cities
func (r *territoryRepository) GetAllCities() ([]model.City, error) {
	var cities []model.City
	if err := r.db.GetDB().Find(&cities).Error; err != nil {
		return nil, err
	}
	return cities, nil
}

// GetCitiesByDistrictID retrieves cities by district ID
func (r *territoryRepository) GetCitiesByDistrictID(districtID string) ([]model.City, error) {
	var cities []model.City
	if err := r.db.GetDB().Where("district_id = ?", districtID).Find(&cities).Error; err != nil {
		return nil, err
	}
	return cities, nil
}

// GetCityByID retrieves a city by ID
func (r *territoryRepository) GetCityByID(id string) (*model.City, error) {
	var city model.City
	if err := r.db.GetDB().
		Preload("Hotspots").
		Where("id = ?", id).
		First(&city).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("city not found")
		}
		return nil, err
	}
	return &city, nil
}

// GetAllHotspots retrieves all hotspots
func (r *territoryRepository) GetAllHotspots() ([]model.Hotspot, error) {
	var hotspots []model.Hotspot
	if err := r.db.GetDB().Find(&hotspots).Error; err != nil {
		return nil, err
	}

	// Get controller names for each hotspot
	for i, hotspot := range hotspots {
		if hotspot.ControllerID != nil {
			var player model.Player
			if err := r.db.GetDB().
				Select("name").
				Where("id = ?", *hotspot.ControllerID).
				First(&player).Error; err == nil {
				controllerName := player.Name
				hotspots[i].ControllerName = &controllerName
			}
		}
	}

	return hotspots, nil
}

// GetHotspotsByCity retrieves hotspots by city ID
func (r *territoryRepository) GetHotspotsByCity(cityID string) ([]model.Hotspot, error) {
	var hotspots []model.Hotspot
	if err := r.db.GetDB().Where("city_id = ?", cityID).Find(&hotspots).Error; err != nil {
		return nil, err
	}

	// Get controller names for each hotspot
	for i, hotspot := range hotspots {
		if hotspot.ControllerID != nil {
			var player model.Player
			if err := r.db.GetDB().
				Select("name").
				Where("id = ?", *hotspot.ControllerID).
				First(&player).Error; err == nil {
				controllerName := player.Name
				hotspots[i].ControllerName = &controllerName
			}
		}
	}

	return hotspots, nil
}

// GetHotspotsByRegion retrieves all hotspots in a region
func (r *territoryRepository) GetHotspotsByRegion(regionID string) ([]model.Hotspot, error) {
	// Get districts in the region
	var districts []model.District
	if err := r.db.GetDB().Where("region_id = ?", regionID).Find(&districts).Error; err != nil {
		return nil, err
	}

	if len(districts) == 0 {
		return []model.Hotspot{}, nil
	}

	// Get district IDs
	districtIDs := make([]string, len(districts))
	for i, district := range districts {
		districtIDs[i] = district.ID
	}

	// Get cities in these districts
	var cities []model.City
	if err := r.db.GetDB().Where("district_id IN ?", districtIDs).Find(&cities).Error; err != nil {
		return nil, err
	}

	if len(cities) == 0 {
		return []model.Hotspot{}, nil
	}

	// Get city IDs
	cityIDs := make([]string, len(cities))
	for i, city := range cities {
		cityIDs[i] = city.ID
	}

	// Get hotspots in these cities
	var hotspots []model.Hotspot
	if err := r.db.GetDB().Where("city_id IN ?", cityIDs).Find(&hotspots).Error; err != nil {
		return nil, err
	}

	// Get controller names for each hotspot
	for i, hotspot := range hotspots {
		if hotspot.ControllerID != nil {
			var player model.Player
			if err := r.db.GetDB().
				Select("name").
				Where("id = ?", *hotspot.ControllerID).
				First(&player).Error; err == nil {
				controllerName := player.Name
				hotspots[i].ControllerName = &controllerName
			}
		}
	}

	return hotspots, nil
}

// GetHotspotByID retrieves a hotspot by ID
func (r *territoryRepository) GetHotspotByID(id string) (*model.Hotspot, error) {
	var hotspot model.Hotspot
	if err := r.db.GetDB().Where("id = ?", id).First(&hotspot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("hotspot not found")
		}
		return nil, err
	}

	// Get controller name if applicable
	if hotspot.ControllerID != nil {
		var player model.Player
		if err := r.db.GetDB().
			Select("name").
			Where("id = ?", *hotspot.ControllerID).
			First(&player).Error; err == nil {
			controllerName := player.Name
			hotspot.ControllerName = &controllerName
		}
	}

	return &hotspot, nil
}

// GetControlledHotspots retrieves hotspots controlled by a player
func (r *territoryRepository) GetControlledHotspots(playerID string) ([]model.Hotspot, error) {
	var hotspots []model.Hotspot
	if err := r.db.GetDB().Where("controller_id = ?", playerID).Find(&hotspots).Error; err != nil {
		return nil, err
	}

	// Set controller name for all hotspots
	var player model.Player
	if err := r.db.GetDB().
		Select("name").
		Where("id = ?", playerID).
		First(&player).Error; err == nil {
		for i := range hotspots {
			controllerName := player.Name
			hotspots[i].ControllerName = &controllerName
		}
	}

	return hotspots, nil
}

// GetControlledHotspotsByRegion retrieves hotspots controlled by a player in a specific region
func (r *territoryRepository) GetControlledHotspotsByRegion(playerID string, regionID string) ([]model.Hotspot, error) {
	// First get all hotspots in the region
	regionalHotspots, err := r.GetHotspotsByRegion(regionID)
	if err != nil {
		return nil, err
	}

	// Filter for those controlled by the player
	controlledHotspots := make([]model.Hotspot, 0)
	for _, hotspot := range regionalHotspots {
		if hotspot.ControllerID != nil && *hotspot.ControllerID == playerID {
			controlledHotspots = append(controlledHotspots, hotspot)
		}
	}

	// Set controller name for all hotspots
	var player model.Player
	if err := r.db.GetDB().
		Select("name").
		Where("id = ?", playerID).
		First(&player).Error; err == nil {
		for i := range controlledHotspots {
			controllerName := player.Name
			controlledHotspots[i].ControllerName = &controllerName
		}
	}

	return controlledHotspots, nil
}

// UpdateHotspot updates a hotspot
func (r *territoryRepository) UpdateHotspot(hotspot *model.Hotspot) error {
	// Calculate defense strength based on allocated resources
	hotspot.DefenseStrength = (hotspot.Crew * 10) + (hotspot.Weapons * 15) + (hotspot.Vehicles * 20)
	return r.db.GetDB().Save(hotspot).Error
}

// AddTerritoryAction records a territory action
func (r *territoryRepository) AddTerritoryAction(action *model.TerritoryAction) error {
	return r.db.GetDB().Create(action).Error
}

// GetRecentActions retrieves recent territory actions
func (r *territoryRepository) GetRecentActions(limit int) ([]model.TerritoryAction, error) {
	var actions []model.TerritoryAction
	if err := r.db.GetDB().
		Order("timestamp DESC").
		Limit(limit).
		Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// GetRecentActionsByPlayer retrieves recent territory actions by a player
func (r *territoryRepository) GetRecentActionsByPlayer(playerID string, limit int) ([]model.TerritoryAction, error) {
	var actions []model.TerritoryAction
	if err := r.db.GetDB().
		Where("player_id = ?", playerID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// GetRecentActionsByPlayerAndRegion retrieves recent territory actions by a player in a specific region
func (r *territoryRepository) GetRecentActionsByPlayerAndRegion(playerID string, regionID string, limit int) ([]model.TerritoryAction, error) {
	// Get hotspots in the region
	regionalHotspots, err := r.GetHotspotsByRegion(regionID)
	if err != nil {
		return nil, err
	}

	if len(regionalHotspots) == 0 {
		return []model.TerritoryAction{}, nil
	}

	// Get hotspot IDs
	hotspotIDs := make([]string, len(regionalHotspots))
	for i, hotspot := range regionalHotspots {
		hotspotIDs[i] = hotspot.ID
	}

	// Get actions for these hotspots by the player
	var actions []model.TerritoryAction
	if err := r.db.GetDB().
		Where("player_id = ? AND hotspot_id IN ?", playerID, hotspotIDs).
		Order("timestamp DESC").
		Limit(limit).
		Find(&actions).Error; err != nil {
		return nil, err
	}

	return actions, nil
}

// UpdateHotspotPendingCollection updates the pending collection amount for a hotspot
func (r *territoryRepository) UpdateHotspotPendingCollection(hotspotID string, amount int) error {
	return r.db.GetDB().Model(&model.Hotspot{}).
		Where("id = ?", hotspotID).
		Updates(map[string]interface{}{
			"pending_collection": gorm.Expr("pending_collection + ?", amount),
		}).Error
}

// RefreshIllegalBusinesses randomly refreshes illegal businesses
func (r *territoryRepository) RefreshIllegalBusinesses() error {
	// In a real implementation, this would randomly change some illegal businesses
	// For now, we just ensure all illegal businesses have no controller
	return r.db.GetDB().Model(&model.Hotspot{}).
		Where("is_legal = ?", false).
		Updates(map[string]interface{}{
			"controller_id":    nil,
			"crew":             0,
			"weapons":          0,
			"vehicles":         0,
			"defense_strength": 0,
		}).Error
}

// GetAllControlledLegalHotspots retrieves all legal hotspots with controllers
func (r *territoryRepository) GetAllControlledLegalHotspots() ([]model.Hotspot, error) {
	var hotspots []model.Hotspot
	if err := r.db.GetDB().
		Where("is_legal = ? AND controller_id IS NOT NULL", true).
		Find(&hotspots).Error; err != nil {
		return nil, err
	}

	return hotspots, nil
}

// GetAllControlledLegalHotspotsByRegion retrieves all legal hotspots with controllers in a specific region
func (r *territoryRepository) GetAllControlledLegalHotspotsByRegion(regionID string) ([]model.Hotspot, error) {
	// Get all hotspots in the region
	regionalHotspots, err := r.GetHotspotsByRegion(regionID)
	if err != nil {
		return nil, err
	}

	// Filter for legal hotspots with controllers
	controlledLegalHotspots := make([]model.Hotspot, 0)
	for _, hotspot := range regionalHotspots {
		if hotspot.IsLegal && hotspot.ControllerID != nil {
			controlledLegalHotspots = append(controlledLegalHotspots, hotspot)
		}
	}

	return controlledLegalHotspots, nil
}

// UpdateHotspotLastIncomeTime updates the last income time for a hotspot
func (r *territoryRepository) UpdateHotspotLastIncomeTime(hotspotID string, lastIncomeTime time.Time) error {
	return r.db.GetDB().Model(&model.Hotspot{}).
		Where("id = ?", hotspotID).
		Updates(map[string]interface{}{
			"last_income_time": lastIncomeTime,
		}).Error
}

// GetHotspotsByRegionAndController retrieves all hotspots in a region controlled by a specific player
func (r *territoryRepository) GetHotspotsByRegionAndController(regionID, controllerID string) ([]model.Hotspot, error) {
	var hotspots []model.Hotspot

	// Complex query to join through the location hierarchy
	err := r.db.GetDB().
		Select("hotspots.*").
		Table("hotspots").
		Joins("JOIN cities ON cities.id = hotspots.city_id").
		Joins("JOIN districts ON districts.id = cities.district_id").
		Joins("JOIN regions ON regions.id = districts.region_id").
		Where("regions.id = ?", regionID).
		Where("hotspots.controller_id = ?", controllerID).
		Find(&hotspots).Error

	if err != nil {
		return nil, err
	}

	return hotspots, nil
}
