// fe/src/services/campaignService.ts

import api, { ApiResponse } from './api';
import {
  Campaign,
  Chapter,
  Mission,
  Branch,
  CampaignOperation,
  CampaignPOI,
  Dialogue,
  PlayerCampaignProgress,
  BranchCompletionStatus,
  InteractionType,
  InteractionResponse
} from '@/types/campaign';

// Endpoints
const ENDPOINTS = {
  CAMPAIGNS: '/campaigns',
  CHAPTERS: '/campaigns/chapters',
  MISSIONS: '/campaigns/missions',
  BRANCHES: '/campaigns/branches',
  OPERATIONS: '/campaigns/operations',
  POIS: '/campaigns/pois',
};

export default {
  /**
   * Get all campaigns
   */
  getCampaigns() {
    return api.get<Campaign[]>(ENDPOINTS.CAMPAIGNS);
  },

  /**
   * Get a specific campaign
   */
  getCampaign(campaignId: string) {
    return api.get<Campaign>(`${ENDPOINTS.CAMPAIGNS}/${campaignId}`);
  },

  /**
   * Get chapters for a campaign
   */
  getChaptersByCampaignId(campaignId: string) {
    return api.get<Chapter[]>(`${ENDPOINTS.CAMPAIGNS}/${campaignId}/chapters`);
  },

  /**
   * Get a specific chapter
   */
  getChapter(chapterId: string) {
    return api.get<Chapter>(`${ENDPOINTS.CHAPTERS}/${chapterId}`);
  },

  /**
   * Get missions for a chapter
   */
  getMissionsByChapterId(chapterId: string) {
    return api.get<Mission[]>(`${ENDPOINTS.CHAPTERS}/${chapterId}/missions`);
  },

  /**
   * Get a specific mission
   */
  getMission(missionId: string) {
    return api.get<Mission>(`${ENDPOINTS.MISSIONS}/${missionId}`);
  },

  /**
   * Get branches for a mission
   */
  getBranchesByMissionId(missionId: string) {
    return api.get<Branch[]>(`${ENDPOINTS.MISSIONS}/${missionId}/branches`);
  },

  /**
   * Get a specific branch
   */
  getBranch(branchId: string) {
    return api.get<Branch>(`${ENDPOINTS.BRANCHES}/${branchId}`);
  },

  /**
   * Get operations for a branch
   */
  getOperationsByBranchId(branchId: string) {
    return api.get<CampaignOperation[]>(`${ENDPOINTS.BRANCHES}/${branchId}/operations`);
  },

  /**
   * Get POIs for a branch
   */
  getPOIsByBranchId(branchId: string) {
    return api.get<CampaignPOI[]>(`${ENDPOINTS.BRANCHES}/${branchId}/pois`);
  },

  /**
   * Get a specific POI
   */
  getPOI(poiId: string) {
    return api.get<CampaignPOI>(`${ENDPOINTS.POIS}/${poiId}`);
  },

  /**
   * Get dialogues for a POI
   */
  getDialoguesByPOIId(poiId: string) {
    return api.get<Dialogue[]>(`${ENDPOINTS.POIS}/${poiId}/dialogues`);
  },

  /**
   * Get a player's campaign progress
   */
  getPlayerProgress(campaignId: string) {
    return api.get<PlayerCampaignProgress>(`${ENDPOINTS.CAMPAIGNS}/${campaignId}/progress`);
  },

  /**
   * Start a campaign for a player
   */
  startCampaign(campaignId: string) {
    return api.post<PlayerCampaignProgress>(`${ENDPOINTS.CAMPAIGNS}/${campaignId}/start`);
  },

  /**
   * Get a player's current mission in a campaign
   */
  getCurrentMission(campaignId: string) {
    return api.get<Mission>(`${ENDPOINTS.CAMPAIGNS}/${campaignId}/current-mission`);
  },

  /**
   * Select a branch for a mission
   */
  selectBranch(missionId: string, branchId: string) {
    return api.post<Branch>(`${ENDPOINTS.MISSIONS}/${missionId}/select-branch`, { branchId });
  },

  /**
   * Complete a branch
   */
  completeBranch(missionId: string, branchId: string) {
    return api.post<PlayerCampaignProgress>(`${ENDPOINTS.MISSIONS}/${missionId}/branches/${branchId}/complete`);
  },

  /**
   * Check if a branch is complete
   */
  checkBranchCompletion(branchId: string) {
    return api.get<BranchCompletionStatus>(`${ENDPOINTS.BRANCHES}/${branchId}/check-completion`);
  },

  /**
   * Interact with a POI
   */
  interactWithPOI(poiId: string, interactionType: InteractionType) {
    return api.post<InteractionResponse>(`${ENDPOINTS.POIS}/${poiId}/interact`, { interactionType });
  },

  /**
   * Complete a POI
   */
  completePOI(poiId: string) {
    return api.post<{ success: boolean }>(`${ENDPOINTS.POIS}/${poiId}/complete`);
  },

  /**
   * Complete an operation
   */
  completeOperation(operationId: string, attemptId: string) {
    return api.post<{ success: boolean }>(`${ENDPOINTS.OPERATIONS}/${operationId}/complete`, { attemptId });
  }
};
