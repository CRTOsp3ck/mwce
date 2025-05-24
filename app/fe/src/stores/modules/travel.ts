// src/stores/modules/travel.ts

import { defineStore } from 'pinia';
import travelService, { TravelAttempt, TravelResponse } from '@/services/travelService';
import { Region } from '@/types/territory';
import { usePlayerStore } from './player';

interface TravelState {
  availableRegions: Region[];
  currentRegion: Region | null;
  travelHistory: TravelAttempt[];
  isLoading: boolean;
  error: string | null;
  lastTravelResponse: TravelResponse | null;
}

export const useTravelStore = defineStore('travel', {
  state: (): TravelState => ({
    availableRegions: [],
    currentRegion: null,
    travelHistory: [],
    isLoading: false,
    error: null,
    lastTravelResponse: null
  }),

  getters: {
    // Whether the player is in any region or at "headquarters"
    // isInRegion: state => !!state.currentRegion,

    // Whether the player is in a specific region
    isInRegion: state => (regionId: string) => {
      return state.currentRegion?.id === regionId;
    },

    // Get the player's location name
    currentLocationName: state => {
      return state.currentRegion?.name || 'Headquarters';
    },

    // Get the player's travel history, sorted by most recent
    recentTravelAttempts: state => {
      return [...state.travelHistory].sort((a, b) => {
        return new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime();
      });
    },

    // Get successful travel attempts
    successfulTravelAttempts: state => {
      return state.travelHistory.filter(attempt => attempt.success);
    },

    // Get failed travel attempts (caught by police)
    failedTravelAttempts: state => {
      return state.travelHistory.filter(attempt => !attempt.success);
    }
  },

  actions: {
    async fetchAvailableRegions() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await travelService.getAvailableRegions();
        if (response.success && response.data) {
          this.availableRegions = response.data;
        } else {
          throw new Error('Failed to load available regions');
        }
      } catch (error) {
        this.error = 'Failed to load available regions';
        console.error('Error fetching available regions:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchCurrentRegion() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await travelService.getCurrentRegion();
        if (response.success && response.data) {
          this.currentRegion = response.data;
        } else {
          throw new Error('Failed to load current region');
        }
      } catch (error) {
        this.error = 'Failed to load current region';
        console.error('Error fetching current region:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchTravelHistory(limit?: number) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await travelService.getTravelHistory(limit);
        if (response.success && response.data) {
          this.travelHistory = response.data;
        } else {
          throw new Error('Failed to load travel history');
        }
      } catch (error) {
        this.error = 'Failed to load travel history';
        console.error('Error fetching travel history:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async travel(regionId: string) {
      this.isLoading = true;
      this.error = null;
      this.lastTravelResponse = null;

      try {
        const response = await travelService.travel(regionId);
        if (!response.success || !response.data) {
          throw new Error('Travel failed');
        }

        this.lastTravelResponse = response.data;

        // If travel was successful, we don't need to update region here
        // The SSE event will handle updating all data when we arrive
        if (response.data.success) {
          // Just update the immediate effects (money and heat)
          const playerStore = usePlayerStore();
          if (playerStore.profile) {
            // Deduct travel cost
            playerStore.profile.money -= response.data.travelCost;

            // Apply heat reduction if successful
            if (response.data.heatReduction) {
              playerStore.profile.heat -= response.data.heatReduction;
              if (playerStore.profile.heat < 0) {
                playerStore.profile.heat = 0;
              }
            }
          }
        } else {
          // If travel failed (caught by police)
          const playerStore = usePlayerStore();
          if (playerStore.profile) {
            // Deduct fine
            if (response.data.fineAmount) {
              playerStore.profile.money -= response.data.fineAmount;
              if (playerStore.profile.money < 0) {
                playerStore.profile.money = 0;
              }
            }

            // Apply heat increase
            if (response.data.heatIncrease) {
              playerStore.profile.heat += response.data.heatIncrease;
            }
          }
        }

        // Refresh travel history to include this attempt
        await this.fetchTravelHistory();

        return {
          travelResult: response.data,
          gameMessage: response.gameMessage
        };
      } catch (error) {
        this.error = 'Travel failed';
        console.error('Error traveling:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    $reset() {
      this.availableRegions = [];
      this.currentRegion = null;
      this.travelHistory = [];
      this.isLoading = false;
      this.error = null;
      this.lastTravelResponse = null;
    }
  }
});
