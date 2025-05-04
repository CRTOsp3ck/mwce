// src/services/campaignService.ts

import api from './api';
import {
  Campaign,
  Chapter,
  Mission,
  PlayerCampaignProgress,
  PlayerMissionProgress,
  MissionCompleteResult,
  POI,
  MissionOperation
} from '@/types/campaign';

// Endpoints
const ENDPOINTS = {
  CAMPAIGNS: '/campaigns',
  CAMPAIGN_DETAIL: '/campaigns',  // + /:id
  START_CAMPAIGN: '/campaigns',   // + /:id/start
  CHAPTERS: '/campaigns/chapters', // + /:id
  MISSIONS: '/campaigns/missions', // + /:id
  START_MISSION: '/campaigns/missions', // + /:id/start
  COMPLETE_MISSION: '/campaigns/missions', // + /:id/complete
  ACTIVE_POIS: '/campaigns/pois',
  COMPLETE_POI: '/campaigns/pois', // + /:id/complete
  ACTIVE_OPERATIONS: '/campaigns/operations',
  START_OPERATION: '/campaigns/operations', // + /:id/start
  COMPLETE_OPERATION: '/campaigns/operations', // + /:id/complete
  TRACK_ACTION: '/campaigns/actions/track',
};

export default {
  /**
   * Get all available campaigns
   */
  getCampaigns() {
    return api.get<{
      campaigns: Campaign[],
      progress: { [key: string]: PlayerCampaignProgress }
    }>(ENDPOINTS.CAMPAIGNS);
  },

  /**
   * Get a campaign with chapters and missions
   */
  getCampaignDetail(campaignId: string) {
    return api.get<{
      campaign: Campaign,
      progress: PlayerCampaignProgress,
      missionProgress: { [key: string]: PlayerMissionProgress }
    }>(`${ENDPOINTS.CAMPAIGN_DETAIL}/${campaignId}`);
  },

  /**
   * Start a campaign
   */
  startCampaign(campaignId: string) {
    return api.post<PlayerCampaignProgress>(`${ENDPOINTS.START_CAMPAIGN}/${campaignId}/start`);
  },

  /**
   * Get a chapter with missions
   */
  getChapter(chapterId: string) {
    return api.get<{
      chapter: Chapter,
      missionProgress: { [key: string]: PlayerMissionProgress }
    }>(`${ENDPOINTS.CHAPTERS}/${chapterId}`);
  },

  /**
   * Get a mission with details
   */
  getMission(missionId: string) {
    return api.get<{
      mission: Mission,
      progress: PlayerMissionProgress,
      meetsRequirements: boolean,
      failedRequirements: string[]
    }>(`${ENDPOINTS.MISSIONS}/${missionId}`);
  },

  /**
   * Start a mission
   */
  startMission(missionId: string) {
    return api.post<PlayerMissionProgress>(`${ENDPOINTS.START_MISSION}/${missionId}/start`);
  },

  /**
   * Complete a mission
   */
  completeMission(missionId: string, choiceId?: string) {
    return api.post<MissionCompleteResult>(`${ENDPOINTS.COMPLETE_MISSION}/${missionId}/complete`, {
      choiceId: choiceId || ""
    });
  },

  /**
   * Get active POIs for the player
   */
  getActivePOIs() {
    return api.get<POI[]>(ENDPOINTS.ACTIVE_POIS);
  },

  /**
   * Complete a POI
   */
  completePOI(poiId: string) {
    return api.post(`${ENDPOINTS.COMPLETE_POI}/${poiId}/complete`);
  },

  /**
   * Get active mission operations for the player
   */
  getActiveMissionOperations() {
    return api.get<MissionOperation[]>(ENDPOINTS.ACTIVE_OPERATIONS);
  },

  /**
   * Start a mission operation
   */
  startMissionOperation(operationId: string) {
    return api.post(`${ENDPOINTS.START_OPERATION}/${operationId}/start`);
  },

  /**
   * Complete a mission operation
   */
  completeMissionOperation(operationId: string) {
    return api.post(`${ENDPOINTS.COMPLETE_OPERATION}/${operationId}/complete`);
  },

  /**
   * Track a player action
   */
  trackPlayerAction(actionType: string, actionValue: string) {
    return api.post(ENDPOINTS.TRACK_ACTION, {
      actionType,
      actionValue
    });
  }
};
