// internal/service/providers.go

package service

import (
	"mwce-be/internal/model"
)

// CustomHotspotProvider defines an interface for services that can inject custom hotspots
type CustomHotspotProvider interface {
	// GetInjectedHotspots returns custom hotspots for a player in a specific region
	GetInjectedHotspots(playerID string, regionID *string) ([]model.Hotspot, error)
}

// CustomOperationsProvider defines an interface for services that can inject custom operations
type CustomOperationsProvider interface {
	// GetInjectedOperations returns custom operations for a player in a specific region
	GetInjectedOperations(playerID string, regionID *string) ([]model.Operation, error)
}
