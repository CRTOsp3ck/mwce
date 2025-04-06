// src/stores/modules/territory.ts

import { defineStore } from 'pinia';
import territoryService from '@/services/territoryService';
import {
  Region,
  District,
  City,
  Hotspot,
  TerritoryAction,
  TerritoryActionType,
  ActionResources,
  ActionResult
} from '@/types/territory';
import { usePlayerStore } from './player';

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
  // Timer-related properties
  incomeTimerInterval: number | null;
  timerRefreshCounter: number; // Used to force updates of timers
}

export const useTerritoryStore = defineStore('territory', {
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
    incomeTimerInterval: null,
    timerRefreshCounter: 0
  }),

  getters: {
    // Basic getters remain the same...
    selectedRegion: state => {
      return state.selectedRegionId ? state.regions.find(r => r.id === state.selectedRegionId) : null;
    },

    selectedDistrict: state => {
      return state.selectedDistrictId ? state.districts.find(d => d.id === state.selectedDistrictId) : null;
    },

    selectedCity: state => {
      return state.selectedCityId ? state.cities.find(c => c.id === state.selectedCityId) : null;
    },

    selectedHotspot: state => {
      return state.selectedHotspotId ? state.hotspots.find(h => h.id === state.selectedHotspotId) : null;
    },

    controlledHotspots: state => {
      const playerStore = usePlayerStore();
      const playerId = playerStore.profile?.id;
      return state.hotspots.filter(h => h.controller === playerId);
    },

    legalBusinesses: state => {
      return state.hotspots.filter(h => h.isLegal);
    },

    illegalBusinesses: state => {
      return state.hotspots.filter(h => !h.isLegal);
    },

    // Time-related getters with reactivity through timerRefreshCounter
    getTimeRemaining:
      state =>
      (hotspotId: string): string => {
        // Use the refresh counter to make this getter reactive to timer changes
        const _ = state.timerRefreshCounter;

        const hotspot = state.hotspots.find(h => h.id === hotspotId);

        if (!hotspot) {
          return 'Unknown';
        }

        if (!hotspot.nextIncomeTime) {
          if (hotspot.lastIncomeTime) {
            // Calculate next income time if we have lastIncomeTime but not nextIncomeTime
            const lastIncomeTime = new Date(hotspot.lastIncomeTime);
            const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
            return formatTimeRemaining(nextIncomeTime.toISOString());
          }
          return 'Initializing...';
        }

        return formatTimeRemaining(hotspot.nextIncomeTime);
      },

    isIncomeSoon:
      state =>
      (hotspotId: string): boolean => {
        // Use the refresh counter to make this getter reactive to timer changes
        const _ = state.timerRefreshCounter;

        const hotspot = state.hotspots.find(h => h.id === hotspotId);
        if (!hotspot || !hotspot.nextIncomeTime) return false;

        const now = new Date();
        const nextIncomeTime = new Date(hotspot.nextIncomeTime);
        const diffMs = nextIncomeTime.getTime() - now.getTime();

        // If already passed or less than 5 minutes remaining
        return diffMs <= 5 * 60 * 1000 && diffMs >= 0;
      },

    // Other getters remain the same...
    districtsInRegion: state => (regionId: string) => {
      return state.districts.filter(d => d.regionId === regionId);
    },

    citiesInDistrict: state => (districtId: string) => {
      return state.cities.filter(c => c.districtId === districtId);
    },

    hotspotsInCity: state => (cityId: string) => {
      return state.hotspots.filter(h => h.cityId === cityId);
    }
  },

  actions: {
    async fetchTerritoryData() {
      this.isLoading = true;
      this.error = null;

      try {
        // Get regions
        const regionsResponse = await territoryService.getRegions();
        if (regionsResponse.success && regionsResponse.data) {
          this.regions = regionsResponse.data;
        }

        // Get districts
        const districtsResponse = await territoryService.getDistricts();
        if (districtsResponse.success && districtsResponse.data) {
          this.districts = districtsResponse.data;
        }

        // Get cities
        const citiesResponse = await territoryService.getCities();
        if (citiesResponse.success && citiesResponse.data) {
          this.cities = citiesResponse.data;
        }

        // Get hotspots
        const hotspotsResponse = await territoryService.getHotspots();
        if (hotspotsResponse.success && hotspotsResponse.data) {
          this.hotspots = hotspotsResponse.data;

          // Process hotspot timing data
          this.ensureAllIncomeTimes();
        }

        // Set initial filtered hotspots
        this.updateFilteredHotspots();

        // Start the income timer
        this.startIncomeTimer();
      } catch (error) {
        this.error = 'Failed to load territory data';
        console.error('Error fetching territory data:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchRecentActions() {
      this.isLoading = true;

      try {
        const response = await territoryService.getRecentActions();
        if (response.success && response.data) {
          this.recentActions = response.data;
        }
      } catch (error) {
        console.error('Error fetching recent actions:', error);
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
        this.filteredHotspots = this.hotspots.filter(h => h.cityId === this.selectedCityId);
      } else if (this.selectedDistrictId) {
        const citiesInDistrict = this.cities.filter(c => c.districtId === this.selectedDistrictId);
        const cityIds = citiesInDistrict.map(c => c.id);
        this.filteredHotspots = this.hotspots.filter(h => cityIds.includes(h.cityId));
      } else if (this.selectedRegionId) {
        const districtsInRegion = this.districts.filter(d => d.regionId === this.selectedRegionId);
        const districtIds = districtsInRegion.map(d => d.id);
        const citiesInRegion = this.cities.filter(c => districtIds.includes(c.districtId));
        const cityIds = citiesInRegion.map(c => c.id);
        this.filteredHotspots = this.hotspots.filter(h => cityIds.includes(h.cityId));
      } else {
        this.filteredHotspots = [...this.hotspots];
      }
    },

    async performTerritoryAction(actionType: TerritoryActionType, hotspotId: string, resources: ActionResources) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await territoryService.performAction(actionType, hotspotId, resources);

        if (!response.success || !response.data) {
          throw new Error('Failed to perform territory action');
        }

        // Extract the result from the response
        let result: ActionResult;

        if ('result' in response.data) {
          result = response.data.result as ActionResult;
        } else {
          result = response.data as unknown as ActionResult;
        }

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
          const updatedHotspotResponse = await territoryService.getHotspot(hotspotId);
          if (updatedHotspotResponse.success && updatedHotspotResponse.data) {
            const updatedHotspot = updatedHotspotResponse.data;

            // Critical: Ensure nextIncomeTime is properly set for newly taken over hotspots
            if (actionType === TerritoryActionType.TAKEOVER && result.success) {
              if (!updatedHotspot.nextIncomeTime && updatedHotspot.lastIncomeTime) {
                const lastIncomeTime = new Date(updatedHotspot.lastIncomeTime);
                const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
                updatedHotspot.nextIncomeTime = nextIncomeTime.toISOString();

                console.log(
                  `Set nextIncomeTime for newly taken over hotspot: ${updatedHotspot.name}`,
                  updatedHotspot.nextIncomeTime
                );
              }
            }

            // Update in both hotspots arrays
            const index = this.hotspots.findIndex(h => h.id === hotspotId);
            if (index !== -1) {
              this.hotspots[index] = updatedHotspot;
            }

            const filteredIndex = this.filteredHotspots.findIndex(h => h.id === hotspotId);
            if (filteredIndex !== -1) {
              this.filteredHotspots[filteredIndex] = updatedHotspot;
            }

            // If takeover was successful, update controlled hotspots count
            if (actionType === TerritoryActionType.TAKEOVER) {
              if (playerStore.profile) {
                playerStore.profile.controlledHotspots += 1;
              }
            }
          }
        }

        // Add the action to recent actions
        const newAction: TerritoryAction = {
          id: Date.now().toString(), // Temporary ID
          type: actionType,
          playerId: playerStore.profile?.id || '',
          hotspotId: hotspotId,
          resources,
          result,
          timestamp: new Date().toISOString()
        };

        this.recentActions.unshift(newAction);

        // Force UI to update with new timer data
        this.timerRefreshCounter++;

        return result;
      } catch (error) {
        this.error = 'Failed to perform territory action';
        console.error('Error performing territory action:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    // Ensures all hotspots have both lastIncomeTime and nextIncomeTime properly set
    ensureAllIncomeTimes() {
      const playerStore = usePlayerStore();
      const playerId = playerStore.profile?.id;

      this.hotspots.forEach(hotspot => {
        // Only process controlled hotspots
        if (hotspot.controller === playerId) {
          // If we have lastIncomeTime but not nextIncomeTime, calculate it
          if (hotspot.lastIncomeTime && !hotspot.nextIncomeTime) {
            const lastIncomeTime = new Date(hotspot.lastIncomeTime);
            const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
            hotspot.nextIncomeTime = nextIncomeTime.toISOString();
            console.log(`Calculated nextIncomeTime for ${hotspot.name}: ${hotspot.nextIncomeTime}`);
          }
          // If both are missing, initialize with current time
          else if (!hotspot.lastIncomeTime && !hotspot.nextIncomeTime) {
            const now = new Date();
            hotspot.lastIncomeTime = now.toISOString();
            const nextIncomeTime = new Date(now.getTime() + 60 * 60 * 1000);
            hotspot.nextIncomeTime = nextIncomeTime.toISOString();
            console.log(`Initialized timing for ${hotspot.name}`);
          }
        }
      });

      // Force refresh the UI counter
      this.timerRefreshCounter++;
    },

    // Centralized timer starter - used on component mount and after fetch
    startIncomeTimer() {
      // Clean up existing timer if any
      if (this.incomeTimerInterval) {
        clearInterval(this.incomeTimerInterval);
        this.incomeTimerInterval = null;
      }

      // Set up new timer that updates every second
      this.incomeTimerInterval = window.setInterval(() => {
        // Increment the refresh counter to trigger reactivity
        this.timerRefreshCounter++;
      }, 1000);

      console.log('Income timer started');
    },

    // Stop income timer - used when component unmounts
    stopIncomeTimer() {
      if (this.incomeTimerInterval) {
        clearInterval(this.incomeTimerInterval);
        this.incomeTimerInterval = null;
        console.log('Income timer stopped');
      }
    },

    async collectHotspotIncome(hotspotId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await territoryService.collectHotspotIncome(hotspotId);
        if (!response.success || !response.data) {
          throw new Error('Failed to collect hotspot income');
        }

        // Extract the collection result
        let collectionResult;
        if ('result' in response.data) {
          collectionResult = response.data.result;
        } else {
          collectionResult = response.data;
        }

        // Update the hotspot's pending collection
        const hotspot = this.hotspots.find(h => h.id === hotspotId);
        if (hotspot) {
          hotspot.pendingCollection = 0;
          hotspot.lastCollectionTime = new Date().toISOString();
        }

        // Update filtered hotspots as well
        const filteredHotspot = this.filteredHotspots.find(h => h.id === hotspotId);
        if (filteredHotspot) {
          filteredHotspot.pendingCollection = 0;
          filteredHotspot.lastCollectionTime = new Date().toISOString();
        }

        // Update player money
        const playerStore = usePlayerStore();
        if (playerStore.profile && collectionResult.collectedAmount > 0) {
          playerStore.profile.money += collectionResult.collectedAmount;

          // Update player's pending collections total
          playerStore.profile.pendingCollections -= collectionResult.collectedAmount;
          if (playerStore.profile.pendingCollections < 0) {
            playerStore.profile.pendingCollections = 0;
          }
        }

        return collectionResult;
      } catch (error) {
        this.error = 'Failed to collect hotspot income';
        console.error('Error collecting hotspot income:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async collectAllHotspotIncome() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await territoryService.collectAllHotspotIncome();
        if (!response.success || !response.data) {
          throw new Error('Failed to collect all hotspot income');
        }

        // Extract the collection result
        let collectionResult;
        if ('result' in response.data) {
          collectionResult = response.data.result;
        } else {
          collectionResult = response.data;
        }

        // Update all controlled hotspots
        const controlledHotspots = this.hotspots.filter(h => h.controller === usePlayerStore().profile?.id);
        controlledHotspots.forEach(hotspot => {
          hotspot.pendingCollection = 0;
          hotspot.lastCollectionTime = new Date().toISOString();
        });

        // Update filtered hotspots as well
        const filteredControlledHotspots = this.filteredHotspots.filter(
          h => h.controller === usePlayerStore().profile?.id
        );
        filteredControlledHotspots.forEach(hotspot => {
          hotspot.pendingCollection = 0;
          hotspot.lastCollectionTime = new Date().toISOString();
        });

        // Update player money
        const playerStore = usePlayerStore();
        if (playerStore.profile && collectionResult.collectedAmount > 0) {
          playerStore.profile.money += collectionResult.collectedAmount;
          playerStore.profile.pendingCollections = 0;
        }

        return collectionResult;
      } catch (error) {
        this.error = 'Failed to collect all hotspot income';
        console.error('Error collecting all hotspot income:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    /**
     * Updates a hotspot's data with data from SSE event
     */
    updateHotspot(hotspotData: Partial<Hotspot>) {
      // Find the hotspot in the store
      const index = this.hotspots.findIndex(h => h.id === hotspotData.id);

      if (index !== -1) {
        // Merge the new data with existing hotspot
        this.hotspots[index] = {
          ...this.hotspots[index],
          ...hotspotData
        };

        // Ensure nextIncomeTime is set if we have lastIncomeTime
        if (this.hotspots[index].lastIncomeTime && !this.hotspots[index].nextIncomeTime) {
          const lastIncomeTime = new Date(this.hotspots[index].lastIncomeTime);
          const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
          this.hotspots[index].nextIncomeTime = nextIncomeTime.toISOString();
          console.log(
            `Set nextIncomeTime from SSE update for ${this.hotspots[index].name}: ${this.hotspots[index].nextIncomeTime}`
          );
        }
      }

      // Also update in filteredHotspots if present
      const filteredIndex = this.filteredHotspots.findIndex(h => h.id === hotspotData.id);

      if (filteredIndex !== -1) {
        this.filteredHotspots[filteredIndex] = {
          ...this.filteredHotspots[filteredIndex],
          ...hotspotData
        };

        // Ensure nextIncomeTime is set in filtered hotspots as well
        if (
          this.filteredHotspots[filteredIndex].lastIncomeTime &&
          !this.filteredHotspots[filteredIndex].nextIncomeTime
        ) {
          const lastIncomeTime = new Date(this.filteredHotspots[filteredIndex].lastIncomeTime);
          const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
          this.filteredHotspots[filteredIndex].nextIncomeTime = nextIncomeTime.toISOString();
        }
      }

      // Force timer refresh
      this.timerRefreshCounter++;
    }
  }
});

// Helper function to format remaining time
function formatTimeRemaining(nextIncomeTimeISO: string): string {
  try {
    const now = new Date();
    const nextIncomeTime = new Date(nextIncomeTimeISO);

    // Check for invalid date
    if (isNaN(nextIncomeTime.getTime())) {
      console.warn('Invalid date detected:', nextIncomeTimeISO);
      return 'Initializing...';
    }

    // Calculate time difference in milliseconds
    const diffMs = nextIncomeTime.getTime() - now.getTime();

    // If next income time has already passed, return "Now"
    if (diffMs <= 0) {
      return 'Now';
    }

    // Calculate hours, minutes, and seconds
    const diffSec = Math.floor(diffMs / 1000);
    const hours = Math.floor(diffSec / 3600);
    const minutes = Math.floor((diffSec % 3600) / 60);
    const seconds = diffSec % 60;

    // Format the time remaining
    if (hours > 0) {
      return `${hours}h ${minutes}m ${seconds}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds}s`;
    } else {
      return `${seconds}s`;
    }
  } catch (error) {
    console.error('Error calculating remaining time:', error, nextIncomeTimeISO);
    return 'Error';
  }
}
