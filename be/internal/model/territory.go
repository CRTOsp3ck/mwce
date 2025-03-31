// internal/model/territory.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Region represents a region on the map
type Region struct {
	ID        string     `json:"id" gorm:"type:uuid;primary_key"`
	Name      string     `json:"name" gorm:"not null"`
	Districts []District `json:"districts,omitempty" gorm:"foreignKey:RegionID"`
	CreatedAt time.Time  `json:"-" gorm:"not null"`
	UpdatedAt time.Time  `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new region
func (r *Region) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// District represents a district within a region
type District struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key"`
	Name      string    `json:"name" gorm:"not null"`
	RegionID  string    `json:"regionId" gorm:"type:uuid;not null;references:regions.id"`
	Cities    []City    `json:"cities,omitempty" gorm:"foreignKey:DistrictID"`
	CreatedAt time.Time `json:"-" gorm:"not null"`
	UpdatedAt time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new district
func (d *District) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// City represents a city within a district
type City struct {
	ID         string    `json:"id" gorm:"type:uuid;primary_key"`
	Name       string    `json:"name" gorm:"not null"`
	DistrictID string    `json:"districtId" gorm:"type:uuid;not null;references:districts.id"`
	Hotspots   []Hotspot `json:"hotspots,omitempty" gorm:"foreignKey:CityID"`
	CreatedAt  time.Time `json:"-" gorm:"not null"`
	UpdatedAt  time.Time `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new city
func (c *City) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

type Hotspot struct {
	ID                 string     `json:"id" gorm:"type:uuid;primary_key"`
	Name               string     `json:"name" gorm:"not null"`
	CityID             string     `json:"cityId" gorm:"type:uuid;not null;references:cities.id"`
	Type               string     `json:"type" gorm:"not null"`         // Bar, Restaurant, Club, Casino, etc.
	BusinessType       string     `json:"businessType" gorm:"not null"` // Gambling, Entertainment, Protection, etc.
	IsLegal            bool       `json:"isLegal" gorm:"not null"`
	ControllerID       *string    `json:"controller,omitempty" gorm:"type:uuid;references:players.id"`
	ControllerName     *string    `json:"controllerName,omitempty" gorm:"-"`
	Income             int        `json:"income" gorm:"not null;default:0"` // Income per hour
	PendingCollection  int        `json:"pendingCollection" gorm:"not null;default:0"`
	LastCollectionTime *time.Time `json:"lastCollectionTime"`
	LastIncomeTime     *time.Time `json:"lastIncomeTime"`          // Time of last income generation
	NextIncomeTime     time.Time  `json:"nextIncomeTime" gorm:"-"` // Calculated field for next income time
	Crew               int        `json:"crew" gorm:"not null;default:0"`
	Weapons            int        `json:"weapons" gorm:"not null;default:0"`
	Vehicles           int        `json:"vehicles" gorm:"not null;default:0"`
	DefenseStrength    int        `json:"defenseStrength" gorm:"not null;default:0"` // Calculated from resources
	CreatedAt          time.Time  `json:"-" gorm:"not null"`
	UpdatedAt          time.Time  `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new hotspot
func (h *Hotspot) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	return nil
}

// TerritoryAction represents an action taken on a territory
type TerritoryAction struct {
	ID        string          `json:"id" gorm:"type:uuid;primary_key"`
	Type      string          `json:"type" gorm:"not null"` // extortion, takeover, collection, defend
	PlayerID  string          `json:"playerId" gorm:"type:uuid;not null;references:players.id"`
	HotspotID string          `json:"hotspotId" gorm:"type:uuid;not null;references:hotspots.id"`
	Resources ActionResources `json:"resources" gorm:"embedded"`
	Result    *ActionResult   `json:"result" gorm:"embedded"`
	Timestamp time.Time       `json:"timestamp" gorm:"not null"`
	CreatedAt time.Time       `json:"-" gorm:"not null"`
}

// BeforeCreate is a GORM hook to generate UUID before creating a new territory action
func (t *TerritoryAction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// ActionResources represents resources allocated to an action
type ActionResources struct {
	Crew     int `json:"crew" gorm:"not null;default:0"`
	Weapons  int `json:"weapons" gorm:"not null;default:0"`
	Vehicles int `json:"vehicles" gorm:"not null;default:0"`
}

// ActionResult represents the result of an action
type ActionResult struct {
	Success         bool   `json:"success" gorm:"not null"`
	MoneyGained     int    `json:"moneyGained,omitempty" gorm:"default:0"`
	MoneyLost       int    `json:"moneyLost,omitempty" gorm:"default:0"`
	CrewGained      int    `json:"crewGained,omitempty" gorm:"default:0"`
	CrewLost        int    `json:"crewLost,omitempty" gorm:"default:0"`
	WeaponsGained   int    `json:"weaponsGained,omitempty" gorm:"default:0"`
	WeaponsLost     int    `json:"weaponsLost,omitempty" gorm:"default:0"`
	VehiclesGained  int    `json:"vehiclesGained,omitempty" gorm:"default:0"`
	VehiclesLost    int    `json:"vehiclesLost,omitempty" gorm:"default:0"`
	RespectGained   int    `json:"respectGained,omitempty" gorm:"default:0"`
	RespectLost     int    `json:"respectLost,omitempty" gorm:"default:0"`
	InfluenceGained int    `json:"influenceGained,omitempty" gorm:"default:0"`
	InfluenceLost   int    `json:"influenceLost,omitempty" gorm:"default:0"`
	HeatGenerated   int    `json:"heatGenerated,omitempty" gorm:"default:0"`
	Message         string `json:"message" gorm:"not null"`
}

// PerformActionRequest represents a request to perform a territory action
type PerformActionRequest struct {
	HotspotID string          `json:"hotspotId"`
	Resources ActionResources `json:"resources"`
}

// CollectResponse represents the response after collecting income from a hotspot
type CollectResponse struct {
	HotspotID       string `json:"hotspotId"`
	HotspotName     string `json:"hotspotName"`
	CollectedAmount int    `json:"collectedAmount"`
	Message         string `json:"message"`
}
