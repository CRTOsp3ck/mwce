// src/stores/modules/territory.ts

import { defineStore } from "pinia";
import territoryService from "@/services/territoryService";
import {
  Region,
  District,
  City,
  Hotspot,
  TerritoryAction,
  TerritoryActionType,
  ActionResources,
  ActionResult,
  HotspotType,
  BusinessType,
} from "@/types/territory";
import { usePlayerStore } from "./player";

interface TerritoryState {
  regions: Region[];
  districts: District[];
  cities: City[];
  hotspots: Hotspot[];
  selectedRegionId: string | null;
  selectedDistrictId: string | null;
  selectedCityId: string | null;
  selectedHotspotId: string | null;
  filteredHotspots: Hotspot[];
  recentActions: TerritoryAction[];
  isLoading: boolean;
  error: string | null;
}

export const useTerritoryStore = defineStore("territory", {
  state: (): TerritoryState => ({
    regions: [],
    districts: [],
    cities: [],
    hotspots: [],
    selectedRegionId: null,
    selectedDistrictId: null,
    selectedCityId: null,
    selectedHotspotId: null,
    filteredHotspots: [],
    recentActions: [],
    isLoading: false,
    error: null,
  }),

  getters: {
    selectedRegion: (state) => {
      return state.selectedRegionId
        ? state.regions.find((r) => r.id === state.selectedRegionId)
        : null;
    },

    selectedDistrict: (state) => {
      return state.selectedDistrictId
        ? state.districts.find((d) => d.id === state.selectedDistrictId)
        : null;
    },

    selectedCity: (state) => {
      return state.selectedCityId
        ? state.cities.find((c) => c.id === state.selectedCityId)
        : null;
    },

    selectedHotspot: (state) => {
      return state.selectedHotspotId
        ? state.hotspots.find((h) => h.id === state.selectedHotspotId)
        : null;
    },

    controlledHotspots: (state) => {
      return state.hotspots.filter((h) => h.controller);
    },

    legalBusinesses: (state) => {
      return state.hotspots.filter((h) => h.isLegal);
    },

    illegalBusinesses: (state) => {
      return state.hotspots.filter((h) => !h.isLegal);
    },

    districtsInRegion: (state) => (regionId: string) => {
      return state.districts.filter((d) => d.regionId === regionId);
    },

    citiesInDistrict: (state) => (districtId: string) => {
      return state.cities.filter((c) => c.districtId === districtId);
    },

    hotspotsInCity: (state) => (cityId: string) => {
      return state.hotspots.filter((h) => h.cityId === cityId);
    },
  },

  actions: {
    async fetchTerritoryData() {
      this.isLoading = true;
      this.error = null;

      try {
        const regionsResponse = await territoryService.getRegions();
        this.regions = regionsResponse.data;

        const districtsResponse = await territoryService.getDistricts();
        this.districts = districtsResponse.data;

        const citiesResponse = await territoryService.getCities();
        this.cities = citiesResponse.data;

        const hotspotsResponse = await territoryService.getHotspots();
        this.hotspots = hotspotsResponse.data;

        // Set initial filtered hotspots
        this.updateFilteredHotspots();
      } catch (error) {
        this.error = "Failed to load territory data";
        console.error("Error fetching territory data:", error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchRecentActions() {
      this.isLoading = true;

      try {
        const response = await territoryService.getRecentActions();
        this.recentActions = response.data;
      } catch (error) {
        console.error("Error fetching recent actions:", error);
      } finally {
        this.isLoading = false;
      }
    },

    selectRegion(regionId: string | null) {
      this.selectedRegionId = regionId;
      this.selectedDistrictId = null;
      this.selectedCityId = null;
      this.selectedHotspotId = null;
      this.updateFilteredHotspots();
    },

    selectDistrict(districtId: string | null) {
      this.selectedDistrictId = districtId;
      this.selectedCityId = null;
      this.selectedHotspotId = null;
      this.updateFilteredHotspots();
    },

    selectCity(cityId: string | null) {
      this.selectedCityId = cityId;
      this.selectedHotspotId = null;
      this.updateFilteredHotspots();
    },

    selectHotspot(hotspotId: string | null) {
      this.selectedHotspotId = hotspotId;
    },

    updateFilteredHotspots() {
      if (this.selectedCityId) {
        this.filteredHotspots = this.hotspots.filter(
          (h) => h.cityId === this.selectedCityId
        );
      } else if (this.selectedDistrictId) {
        const citiesInDistrict = this.cities.filter(
          (c) => c.districtId === this.selectedDistrictId
        );
        const cityIds = citiesInDistrict.map((c) => c.id);
        this.filteredHotspots = this.hotspots.filter((h) =>
          cityIds.includes(h.cityId)
        );
      } else if (this.selectedRegionId) {
        const districtsInRegion = this.districts.filter(
          (d) => d.regionId === this.selectedRegionId
        );
        const districtIds = districtsInRegion.map((d) => d.id);
        const citiesInRegion = this.cities.filter((c) =>
          districtIds.includes(c.districtId)
        );
        const cityIds = citiesInRegion.map((c) => c.id);
        this.filteredHotspots = this.hotspots.filter((h) =>
          cityIds.includes(h.cityId)
        );
      } else {
        this.filteredHotspots = [...this.hotspots];
      }
    },

    async performTerritoryAction(
      actionType: TerritoryActionType,
      hotspotId: string,
      resources: ActionResources
    ) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await territoryService.performAction(
          actionType,
          hotspotId,
          resources
        );
        const result = response.data as ActionResult;

        // Update player resources based on action result
        const playerStore = usePlayerStore();
        if (playerStore.profile) {
          // Update money
          if (result.moneyGained) {
            playerStore.profile.money += result.moneyGained;
          }
          if (result.moneyLost) {
            playerStore.profile.money -= result.moneyLost;
          }

          // Update crew
          if (result.crewGained) {
            playerStore.profile.crew += result.crewGained;
          }
          if (result.crewLost) {
            playerStore.profile.crew -= result.crewLost;
          }

          // Update weapons
          if (result.weaponsGained) {
            playerStore.profile.weapons += result.weaponsGained;
          }
          if (result.weaponsLost) {
            playerStore.profile.weapons -= result.weaponsLost;
          }

          // Update vehicles
          if (result.vehiclesGained) {
            playerStore.profile.vehicles += result.vehiclesGained;
          }
          if (result.vehiclesLost) {
            playerStore.profile.vehicles -= result.vehiclesLost;
          }

          // Update player stats
          if (result.respectGained) {
            playerStore.profile.respect += result.respectGained;
          }
          if (result.respectLost) {
            playerStore.profile.respect -= result.respectLost;
          }

          if (result.influenceGained) {
            playerStore.profile.influence += result.influenceGained;
          }
          if (result.influenceLost) {
            playerStore.profile.influence -= result.influenceLost;
          }

          if (result.heatGenerated) {
            playerStore.profile.heat += result.heatGenerated;
          }
        }

        // Update hotspot data if action was successful
        if (result.success) {
          const updatedHotspot = await territoryService.getHotspot(hotspotId);
          const index = this.hotspots.findIndex((h) => h.id === hotspotId);

          if (index !== -1) {
            this.hotspots[index] = updatedHotspot.data;

            // Also update in filtered hotspots
            const filteredIndex = this.filteredHotspots.findIndex(
              (h) => h.id === hotspotId
            );
            if (filteredIndex !== -1) {
              this.filteredHotspots[filteredIndex] = updatedHotspot.data;
            }
          }

          // If takeover was successful, update controlled hotspots count
          if (actionType === TerritoryActionType.TAKEOVER) {
            playerStore.profile.controlledHotspots += 1;
          }
        }

        // Add the action to recent actions
        const newAction: TerritoryAction = {
          id: Date.now().toString(), // Temporary ID
          type: actionType,
          playerId: playerStore.profile?.id || "",
          hotspotId: hotspotId,
          resources,
          result,
          timestamp: new Date().toISOString(),
        };

        this.recentActions.unshift(newAction);

        return result;
      } catch (error) {
        this.error = "Failed to perform territory action";
        console.error("Error performing territory action:", error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },
  },
});
