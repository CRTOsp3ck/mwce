// src/types/territory.ts

export interface Region {
  id: string;
  name: string;
  districts: District[];
}

export interface District {
  id: string;
  name: string;
  regionId: string;
  cities: City[];
}

export interface City {
  id: string;
  name: string;
  districtId: string;
  hotspots: Hotspot[];
}

export interface Hotspot {
  id: string;
  name: string;
  cityId: string;
  type: HotspotType;
  businessType: BusinessType;
  isLegal: boolean;
  controller?: string; // Player ID who controls this hotspot (null if uncontrolled)
  controllerName?: string; // Name of the controller (for display purposes)
  income: number; // Income per hour
  pendingCollection: number;
  lastCollectionTime: string;
  crew: number;
  weapons: number;
  vehicles: number;
  defenseStrength: number;
}

export enum HotspotType {
  BAR = "Bar",
  RESTAURANT = "Restaurant",
  CLUB = "Club",
  CASINO = "Casino",
  HOTEL = "Hotel",
  WAREHOUSE = "Warehouse",
  DOCK = "Dock",
  FACTORY = "Factory",
  SHOP = "Shop",
  CONSTRUCTION = "Construction Site",
}

export enum BusinessType {
  GAMBLING = "Gambling",
  ENTERTAINMENT = "Entertainment",
  PROTECTION = "Protection",
  SMUGGLING = "Smuggling",
  BLACKMARKET = "Black Market",
  LOAN_SHARKING = "Loan Sharking",
  COUNTERFEITING = "Counterfeiting",
  RACKETEERING = "Racketeering",
}

export interface TerritoryAction {
  id: string;
  type: TerritoryActionType;
  playerId: string;
  hotspotId: string;
  resources: ActionResources;
  result?: ActionResult;
  timestamp: string;
}

export enum TerritoryActionType {
  EXTORTION = "extortion",
  TAKEOVER = "takeover",
  COLLECTION = "collection",
  DEFEND = "defend",
}

export interface ActionResources {
  crew: number;
  weapons: number;
  vehicles: number;
}

export interface ActionResult {
  success: boolean;
  moneyGained?: number;
  moneyLost?: number;
  crewGained?: number;
  crewLost?: number;
  weaponsGained?: number;
  weaponsLost?: number;
  vehiclesGained?: number;
  vehiclesLost?: number;
  heatGenerated?: number;
  respectGained?: number;
  respectLost?: number;
  influenceGained?: number;
  influenceLost?: number;
  message: string;
}
