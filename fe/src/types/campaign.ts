// src/types/campaign.ts

export interface Campaign {
  id: string;
  title: string;
  description: string;
  imageUrl?: string;
  initialChapterId: string;
  isActive: boolean;
  requiredLevel: number;
  chapters?: Chapter[];
}

export interface Chapter {
  id: string;
  campaignId: string;
  title: string;
  description: string;
  imageUrl?: string;
  isLocked: boolean;
  order: number;
  missions?: Mission[];
}

export interface Mission {
  id: string;
  chapterId: string;
  title: string;
  description: string;
  narrative: string;
  imageUrl?: string;
  missionType: MissionType;
  requirements: MissionRequirements;
  rewards: MissionRewards;
  isLocked: boolean;
  order: number;
  choices?: MissionChoice[];
}

export enum MissionType {
  OPERATION = "operation",
  TERRITORY = "territory",
  RESOURCE = "resource",
  TRAVEL = "travel",
  MARKET = "market",
}

export interface MissionRequirements {
  money?: number;
  crew?: number;
  weapons?: number;
  vehicles?: number;
  respect?: number;
  influence?: number;
  maxHeat?: number;
  minTitle?: string;
  operationType?: string;
  operationId?: string;
  hotspotId?: string;
  regionId?: string;
  controlledHotspots?: number;
}

export interface MissionRewards {
  money?: number;
  crew?: number;
  weapons?: number;
  vehicles?: number;
  respect?: number;
  influence?: number;
  heatReduction?: number;
  unlockHotspotId?: string;
  unlockChapterId?: string;
  unlockMissionId?: string;
}

export interface MissionChoice {
  id: string;
  missionId: string;
  text: string;
  nextMissionId: string;
  requirements: MissionRequirements;
  rewards: MissionRewards;
}

export interface PlayerCampaignProgress {
  id: string;
  playerId: string;
  campaignId: string;
  currentChapterId: string;
  currentMissionId: string;
  isCompleted: boolean;
  completedAt?: string;
  lastUpdated: string;
}

export interface PlayerMissionProgress {
  id: string;
  playerId: string;
  missionId: string;
  status: MissionStatus;
  choiceId?: string;
  startedAt?: string;
  completedAt?: string;
}

export enum MissionStatus {
  NOT_STARTED = "not_started",
  IN_PROGRESS = "in_progress",
  COMPLETED = "completed",
  FAILED = "failed",
}

export interface MissionCompleteResult {
  success: boolean;
  rewards: MissionRewards;
  nextMission?: Mission;
  progress: PlayerMissionProgress;
  message: string;
}

export interface CompletionCondition {
  id: string;
  choiceId: string;
  type: string;
  requiredValue: string;
  additionalValue: string;
  orderIndex: number;
  isCompleted: boolean;
  completedAt?: string;
}

export interface POI {
  id: string;
  name: string;
  description: string;
  locationType: string;
  locationId: string;
  missionId: string;
  choiceId: string;
  isActive: boolean;
  isCompleted: boolean;
  completedAt?: string;
}

export interface MissionOperation {
  id: string;
  name: string;
  description: string;
  operationType: string;
  missionId: string;
  choiceId: string;
  resources: OperationResources;
  rewards: OperationRewards;
  risks: OperationRisks;
  duration: number;
  successRate: number;
  isActive: boolean;
  isCompleted: boolean;
  completedAt?: string;
}

export interface OperationResources {
  crew: number;
  weapons: number;
  vehicles: number;
  money: number;
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

// Update MissionChoice interface
export interface MissionChoice {
  id: string;
  missionId: string;
  text: string;
  nextMissionId: string;
  requirements: MissionRequirements;
  rewards: MissionRewards;
  // Add new properties
  sequentialOrder: boolean;
  conditions?: CompletionCondition[];
  pois?: POI[];
  operations?: MissionOperation[];
}

// Update PlayerMissionProgress
export interface PlayerMissionProgress {
  id: string;
  playerId: string;
  missionId: string;
  status: MissionStatus;
  choiceId?: string;
  startedAt?: string;
  completedAt?: string;
  // Add new fields
  currentActiveChoice?: string;
  actionLog?: string;
}
