// src/services/territoryService.ts

import api, { ApiResponse } from './api';
import { 
  Region, 
  District, 
  City, 
  Hotspot, 
  TerritoryAction, 
  TerritoryActionType, 
  ActionResources,
  ActionResult,
  CollectResponse,
  CollectAllResponse
} from '@/types/territory';

// Define interfaces that match backend models
export interface PerformActionRequest {
  hotspotId: string;
  resources: ActionResources;
}

export interface GameMessageResponse<T> {
  result: T;
  gameMessage: {
    type: string;
    message: string;
  };
}

// Endpoints
const ENDPOINTS = {
  REGIONS: '/territory/regions',
  DISTRICTS: '/territory/districts',
  CITIES: '/territory/cities',
  HOTSPOTS: '/territory/hotspots',
  CONTROLLED_HOTSPOTS: '/territory/hotspots/controlled',
  ACTIONS: '/territory/actions',
  COLLECT_HOTSPOT: '/territory/hotspots', // + /:id/collect
  COLLECT_ALL: '/territory/hotspots/collect-all'
};

export default {
  /**
   * Get all regions
   */
  getRegions() {
    return api.get<ApiResponse<Region[]>>(ENDPOINTS.REGIONS);
  },
  
  /**
   * Get a specific region
   */
  getRegion(regionId: string) {
    return api.get<ApiResponse<Region>>(`${ENDPOINTS.REGIONS}/${regionId}`);
  },
  
  /**
   * Get all districts
   */
  getDistricts() {
    return api.get<ApiResponse<District[]>>(ENDPOINTS.DISTRICTS);
  },
  
  /**
   * Get districts in a specific region
   */
  getDistrictsInRegion(regionId: string) {
    return api.get<ApiResponse<District[]>>(`${ENDPOINTS.DISTRICTS}?regionId=${regionId}`);
  },
  
  /**
   * Get a specific district
   */
  getDistrict(districtId: string) {
    return api.get<ApiResponse<District>>(`${ENDPOINTS.DISTRICTS}/${districtId}`);
  },
  
  /**
   * Get all cities
   */
  getCities() {
    return api.get<ApiResponse<City[]>>(ENDPOINTS.CITIES);
  },
  
  /**
   * Get cities in a specific district
   */
  getCitiesInDistrict(districtId: string) {
    return api.get<ApiResponse<City[]>>(`${ENDPOINTS.CITIES}?districtId=${districtId}`);
  },
  
  /**
   * Get a specific city
   */
  getCity(cityId: string) {
    return api.get<ApiResponse<City>>(`${ENDPOINTS.CITIES}/${cityId}`);
  },
  
  /**
   * Get all hotspots
   */
  getHotspots() {
    return api.get<ApiResponse<Hotspot[]>>(ENDPOINTS.HOTSPOTS);
  },
  
  /**
   * Get hotspots in a specific city
   */
  getHotspotsInCity(cityId: string) {
    return api.get<ApiResponse<Hotspot[]>>(`${ENDPOINTS.HOTSPOTS}?cityId=${cityId}`);
  },
  
  /**
   * Get a specific hotspot
   */
  getHotspot(hotspotId: string) {
    return api.get<ApiResponse<Hotspot>>(`${ENDPOINTS.HOTSPOTS}/${hotspotId}`);
  },
  
  /**
   * Get hotspots controlled by the player
   */
  getControlledHotspots() {
    return api.get<ApiResponse<Hotspot[]>>(ENDPOINTS.CONTROLLED_HOTSPOTS);
  },
  
  /**
   * Get recent territory actions
   */
  getRecentActions() {
    return api.get<ApiResponse<TerritoryAction[]>>(ENDPOINTS.ACTIONS);
  },
  
  /**
   * Perform a territory action (extortion, takeover, collection, defend)
   */
  performAction(actionType: TerritoryActionType, hotspotId: string, resources: ActionResources) {
    const request: PerformActionRequest = {
      hotspotId,
      resources
    };
    return api.post<ApiResponse<GameMessageResponse<ActionResult>>>(`${ENDPOINTS.ACTIONS}/${actionType}`, request);
  },
  
  /**
   * Extort an illegal business
   */
  extort(hotspotId: string, resources: ActionResources) {
    return this.performAction(TerritoryActionType.EXTORTION, hotspotId, resources);
  },
  
  /**
   * Attempt to take over a legal business
   */
  takeover(hotspotId: string, resources: ActionResources) {
    return this.performAction(TerritoryActionType.TAKEOVER, hotspotId, resources);
  },
  
  /**
   * Collect resources from a controlled hotspot
   */
  collect(hotspotId: string, resources: ActionResources) {
    return this.performAction(TerritoryActionType.COLLECTION, hotspotId, resources);
  },
  
  /**
   * Allocate resources to defend a hotspot
   */
  defend(hotspotId: string, resources: ActionResources) {
    return this.performAction(TerritoryActionType.DEFEND, hotspotId, resources);
  },

  /**
   * Collect income from a specific hotspot
   */
  collectHotspotIncome(hotspotId: string) {
    return api.post<ApiResponse<GameMessageResponse<CollectResponse>>>(`${ENDPOINTS.COLLECT_HOTSPOT}/${hotspotId}/collect`);
  },
  
  /**
   * Collect income from all controlled hotspots
   */
  collectAllHotspotIncome() {
    return api.post<ApiResponse<GameMessageResponse<CollectAllResponse>>>(ENDPOINTS.COLLECT_ALL);
  }
};