// internal/util/constants.go

package util

// Player titles
const (
	PlayerTitleAssociate   = "Associate"
	PlayerTitleSoldier     = "Soldier"
	PlayerTitleCapo        = "Capo"
	PlayerTitleUnderboss   = "Underboss"
	PlayerTitleConsigliere = "Consigliere"
	PlayerTitleBoss        = "Boss"
	PlayerTitleGodfather   = "Godfather"
)

// Resource types
const (
	ResourceTypeCrew      = "crew"
	ResourceTypeWeapons   = "weapons"
	ResourceTypeVehicles  = "vehicles"
	ResourceTypeMoney     = "money"
	ResourceTypeRespect   = "respect"
	ResourceTypeInfluence = "influence"
	ResourceTypeHeat      = "heat"
)

// Territory action types
const (
	TerritoryActionTypeExtortion  = "extortion"
	TerritoryActionTypeTakeover   = "takeover"
	TerritoryActionTypeCollection = "collection"
	TerritoryActionTypeDefend     = "defend"
)

// Operation types
const (
	OperationTypeCarjacking      = "carjacking"
	OperationTypeGoodsSmuggling  = "goods_smuggling"
	OperationTypeDrugTrafficking = "drug_trafficking"
	OperationTypeOfficialBribing = "official_bribing"
	OperationTypeIntelligence    = "intelligence_gathering"
	OperationTypeCrewRecruitment = "crew_recruitment"
)

// Operation statuses
const (
	OperationStatusInProgress = "in_progress"
	OperationStatusCompleted  = "completed"
	OperationStatusFailed     = "failed"
	OperationStatusCancelled  = "cancelled"
)

// Market price trends
const (
	PriceTrendUp     = "up"
	PriceTrendDown   = "down"
	PriceTrendStable = "stable"
)

// Transaction types
const (
	TransactionTypeBuy  = "buy"
	TransactionTypeSell = "sell"
)

// Notification types
const (
	NotificationTypeTerritory  = "territory"
	NotificationTypeOperation  = "operation"
	NotificationTypeCollection = "collection"
	NotificationTypeHeat       = "heat"
	NotificationTypeSystem     = "system"
	NotificationTypeTravel     = "travel"
	NotificationTypeCampaign   = "campaign"
)

// Game message types
const (
	GameMessageTypeSuccess = "success"
	GameMessageTypeError   = "error"
	GameMessageTypeInfo    = "info"
	GameMessageTypeWarning = "warning"
)
