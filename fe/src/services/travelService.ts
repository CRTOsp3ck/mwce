// src/services/travelService.ts

import api, { ApiResponse } from './api';
import { Region } from '@/types/territory';

// Define interfaces that match backend models
export interface TravelAttempt {
  id: string;
  playerId: string;
  fromRegionId: string | null;
  toRegionId: string;
  success: boolean;
  caughtByPolice: boolean;
  travelCost: number;
  fineAmount: number;
  heatChange: number; // Negative for reduction, positive for increase
  timestamp: string;
}

export interface TravelRequest {
  regionId: string;
}

export interface TravelResponse {
  success: boolean;
  regionId: string;
  regionName: string;
  travelCost: number;
  heatReduction?: number;
  message: string;
  caughtByPolice?: boolean;
  fineAmount?: number;
  heatIncrease?: number;
}

// Endpoints
const ENDPOINTS = {
  AVAILABLE_REGIONS: '/travel/available',
  CURRENT_REGION: '/travel/current',
  TRAVEL: '/travel',
  TRAVEL_HISTORY: '/travel/history'
};

export default {
  /**
   * Get all regions available for travel
   */
  getAvailableRegions() {
    return api.get<Region[]>(ENDPOINTS.AVAILABLE_REGIONS);
  },

  /**
   * Get the player's current region
   */
  getCurrentRegion() {
    return api.get<Region | null>(ENDPOINTS.CURRENT_REGION);
  },

  /**
   * Travel to a specific region
   */
  travel(regionId: string) {
    const request: TravelRequest = {
      regionId
    };
    return api.post<TravelResponse>(ENDPOINTS.TRAVEL, request);
  },

  /**
   * Get the player's travel history
   */
  getTravelHistory(limit?: number) {
    const queryParams = limit ? `?limit=${limit}` : '';
    return api.get<TravelAttempt[]>(`${ENDPOINTS.TRAVEL_HISTORY}${queryParams}`);
  }
};
