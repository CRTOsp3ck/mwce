// internal/service/providers.go

package service

import "mwce-be/internal/model"

// OperationsProvider defines an interface for injecting custom operations
type CustomOperationsProvider interface {
	GetInjectedOperations(playerID string, regionID *string) ([]model.Operation, error)
}

// HotspotProvider defines an interface for injecting custom hotspots
type CustomHotspotProvider interface {
	GetInjectedHotspots(playerID string, regionID *string) ([]model.Hotspot, error)
}
