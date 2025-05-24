// fe/src/types/campaign.ts

import {
  OperationRequirements,
  OperationResources,
  OperationRewards,
  OperationRisks
} from "./operations";

export interface Campaign {
  id: string;
  name: string;
  description: string;
  imageUrl?: string;
  isActive: boolean;
  chapters: Chapter[];
  is_completed?: boolean;
}

export interface Chapter {
  id: string;
  campaignId: string;
  name: string;
  description: string;
  order: number;
  missions: Mission[];
  requirement_level?: number;
}

export interface Mission {
  id: string;
  chapterId: string;
  name: string;
  description: string;
  order: number;
  prerequisites: string[];
  branches: Branch[];
  rewards?: {
    cash?: number;
    exp?: number;
    reputation?: number;
  };
  is_completed?: boolean; // Added by frontend based on progress
}

export interface Branch {
  id: string;
  missionId: string;
  name: string;
  description: string;
  operations: CampaignOperation[];
  pois: CampaignPOI[];
}

export interface CampaignOperation {
  id: string;
  branchId: string;
  name: string;
  description: string;
  type: string;
  isSpecial: boolean;
  regionIds: string[];
  requirements: OperationRequirements;
  resources: OperationResources;
  rewards: OperationRewards;
  risks: OperationRisks;
  duration: number;
  successRate: number;
  metadata?: {
    regionNames?: string[];
    regionsDisplay?: string;
    [key: string]: any;
  };
}

export interface CampaignPOI {
  id: string;
  branchId: string;
  name: string;
  description: string;
  type: string;
  businessType: string;
  isLegal: boolean;
  cityId: string;
  dialogues: Dialogue[];
  metadata?: {
    regionName?: string;
    districtName?: string;
    cityName?: string;
    fullLocation?: string;
    [key: string]: any;
  };
}

export interface Dialogue {
  id: string;
  poiId: string;
  speaker: string;
  interactionType?: InteractionType;
  text: string;
  order: number;
  isSuccess?: boolean;
  resourceEffect?: ResourceEffect;
}

export enum InteractionType {
  Neutral = "Neutral",
  Convince = "Convince",
  Intimidate = "Intimidate"
}

export interface ResourceEffect {
  money?: number;
  crew?: number;
  weapons?: number;
  vehicles?: number;
  respect?: number;
  influence?: number;
  heat?: number;
}

export interface PlayerCampaignProgress {
  id: string;
  playerId: string;
  campaignId: string;
  currentMissionId?: string;
  currentBranchId?: string;
  completedMissionIds: string[];
  completedBranchIds: string[];
  completedPoiIds: string[];
  completedOperationIds: string[];
}

export interface BranchCompletionStatus {
  complete: boolean;
}

export interface InteractionResponse {
  dialogue: Dialogue;
  resourceEffect?: ResourceEffect;
}
