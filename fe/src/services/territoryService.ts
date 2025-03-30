// src/services/territoryService.ts

import api from './api';
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

// Endpoints
const ENDPOINTS = {
  REGIONS: '/territory/regions',
  DISTRICTS: '/territory/districts',
  CITIES: '/territory/cities',
  HOTSPOTS: '/territory/hotspots',
  ACTIONS: '/territory/actions'
};

export default {
  /**
   * Get all regions
   */
  getRegions() {
    return api.get<Region[]>(ENDPOINTS.REGIONS);
  },
  
  /**
   * Get a specific region
   */
  getRegion(regionId: string) {
    return api.get<Region>(`${ENDPOINTS.REGIONS}/${regionId}`);
  },
  
  /**
   * Get all districts
   */
  getDistricts() {
    return api.get<District[]>(ENDPOINTS.DISTRICTS);
  },
  
  /**
   * Get districts in a specific region
   */
  getDistrictsInRegion(regionId: string) {
    return api.get<District[]>(`${ENDPOINTS.DISTRICTS}?regionId=${regionId}`);
  },
  
  /**
   * Get a specific district
   */
  getDistrict(districtId: string) {
    return api.get<District>(`${ENDPOINTS.DISTRICTS}/${districtId}`);
  },
  
  /**
   * Get all cities
   */
  getCities() {
    return api.get<City[]>(ENDPOINTS.CITIES);
  },
  
  /**
   * Get cities in a specific district
   */
  getCitiesInDistrict(districtId: string) {
    return api.get<City[]>(`${ENDPOINTS.CITIES}?districtId=${districtId}`);
  },
  
  /**
   * Get a specific city
   */
  getCity(cityId: string) {
    return api.get<City>(`${ENDPOINTS.CITIES}/${cityId}`);
  },
  
  /**
   * Get all hotspots
   */
  getHotspots() {
    return api.get<Hotspot[]>(ENDPOINTS.HOTSPOTS);
  },
  
  /**
   * Get hotspots in a specific city
   */
  getHotspotsInCity(cityId: string) {
    return api.get<Hotspot[]>(`${ENDPOINTS.HOTSPOTS}?cityId=${cityId}`);
  },
  
  /**
   * Get a specific hotspot
   */
  getHotspot(hotspotId: string) {
    return api.get<Hotspot>(`${ENDPOINTS.HOTSPOTS}/${hotspotId}`);
  },
  
  /**
   * Get hotspots controlled by the player
   */
  getControlledHotspots() {
    return api.get<Hotspot[]>(`${ENDPOINTS.HOTSPOTS}/controlled`);
  },
  
  /**
   * Get recent territory actions
   */
  getRecentActions() {
    return api.get<TerritoryAction[]>(ENDPOINTS.ACTIONS);
  },
  
  /**
   * Perform a territory action (extortion, takeover, collection, defend)
   */
  performAction(actionType: TerritoryActionType, hotspotId: string, resources: ActionResources) {
    return api.post<ActionResult>(`${ENDPOINTS.ACTIONS}/${actionType}`, {
      hotspotId,
      resources
    });
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
    return api.post<CollectResponse>(`${ENDPOINTS.HOTSPOTS}/${hotspotId}/collect`);
  },
  
  /**
   * Collect income from all controlled hotspots
   */
  collectAllHotspotIncome() {
    return api.post<CollectAllResponse>(`${ENDPOINTS.HOTSPOTS}/collect-all`);
  }
};