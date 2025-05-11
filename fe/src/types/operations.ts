// src/types/operations.ts

export interface Operation {
  id: string;
  name: string;
  description: string;
  type: OperationType;
  isSpecial: boolean;
  regionId: string;
  requirements: OperationRequirements;
  resources: OperationResources;
  rewards: OperationRewards;
  risks: OperationRisks;
  duration: number; // in seconds
  successRate: number; // base success rate percentage
  availableUntil: string;
  playerAttempts?: OperationAttempt[];
}

export enum OperationType {
  CARJACKING = "carjacking",
  GOODS_SMUGGLING = "goods_smuggling",
  DRUG_TRAFFICKING = "drug_trafficking",
  OFFICIAL_BRIBING = "official_bribing",
  INTELLIGENCE_GATHERING = "intelligence_gathering",
  CREW_RECRUITMENT = "crew_recruitment",
}

export interface OperationRequirements {
  minInfluence?: number;
  maxHeat?: number;
  minTitle?: string;
  requiredHotspotTypes?: string[];
}

export interface OperationResources {
  crew: number;
  weapons: number;
  vehicles: number;
  money?: number;
}

export interface OperationRewards {
  money?: number;
  crew?: number;
  weapons?: number;
  vehicles?: number;
  respect?: number;
  influence?: number;
  heatReduction?: number;
}

export interface OperationRisks {
  crewLoss?: number;
  weaponsLoss?: number;
  vehiclesLoss?: number;
  moneyLoss?: number;
  heatIncrease?: number;
  respectLoss?: number;
}

export interface OperationAttempt {
  id: string;
  operationId: string;
  playerId: string;
  timestamp: string;
  resources: OperationResources;
  result?: OperationResult;
  completionTime?: string;
  status: OperationStatus;
}

export interface OperationResult {
  success: boolean;
  moneyGained?: number;
  moneyLost?: number;
  crewGained?: number;
  crewLost?: number;
  weaponsGained?: number;
  weaponsLost?: number;
  vehiclesGained?: number;
  vehiclesLost?: number;
  respectGained?: number;
  respectLost?: number;
  influenceGained?: number;
  heatGenerated?: number;
  heatReduced?: number;
  message: string;
}

export enum OperationStatus {
  IN_PROGRESS = "in_progress",
  COMPLETED = "completed",
  FAILED = "failed",
  CANCELLED = "cancelled",
}

export interface OperationsRefreshInfo {
  refreshInterval: number;  // in minutes
  lastRefreshTime: string;  // ISO8601 format
  nextRefreshTime: string;  // ISO8601 format
}
