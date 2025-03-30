// internal/util/constants.go

package util

// Constants for the game
const (
	// Resource types
	ResourceTypeCrew     = "crew"
	ResourceTypeWeapons  = "weapons"
	ResourceTypeVehicles = "vehicles"
	ResourceTypeMoney    = "money"

	// Player titles
	PlayerTitleAssociate   = "Associate"
	PlayerTitleSoldier     = "Soldier"
	PlayerTitleCapo        = "Capo"
	PlayerTitleUnderboss   = "Underboss"
	PlayerTitleConsigliere = "Consigliere"
	PlayerTitleBoss        = "Boss"
	PlayerTitleGodfather   = "Godfather"

	// Notification types
	NotificationTypeTerritory  = "territory"
	NotificationTypeOperation  = "operation"
	NotificationTypeCollection = "collection"
	NotificationTypeHeat       = "heat"
	NotificationTypeSystem     = "system"

	// Operation types
	OperationTypeCarjacking      = "carjacking"
	OperationTypeGoodsSmuggling  = "goods_smuggling"
	OperationTypeDrugTrafficking = "drug_trafficking"
	OperationTypeOfficialBribing = "official_bribing"
	OperationTypeIntelligence    = "intelligence_gathering"
	OperationTypeCrewRecruitment = "crew_recruitment"

	// Operation statuses
	OperationStatusInProgress = "in_progress"
	OperationStatusCompleted  = "completed"
	OperationStatusFailed     = "failed"
	OperationStatusCancelled  = "cancelled"

	// Territory action types
	TerritoryActionTypeExtortion  = "extortion"
	TerritoryActionTypeTakeover   = "takeover"
	TerritoryActionTypeCollection = "collection"
	TerritoryActionTypeDefend     = "defend"

	// Market transaction types
	TransactionTypeBuy  = "buy"
	TransactionTypeSell = "sell"

	// Market price trends
	PriceTrendUp     = "up"
	PriceTrendDown   = "down"
	PriceTrendStable = "stable"

	// Game message types
	GameMessageTypeSuccess = "success"
	GameMessageTypeWarning = "warning"
	GameMessageTypeError   = "error"
	GameMessageTypeInfo    = "info"
)
