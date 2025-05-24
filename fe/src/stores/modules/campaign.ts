// fe/src/stores/modules/campaign.ts

import { defineStore } from 'pinia';
import campaignService from '@/services/campaignService';
import {
  Campaign,
  Chapter,
  Mission,
  Branch,
  CampaignOperation,
  CampaignPOI,
  Dialogue,
  PlayerCampaignProgress,
  InteractionType,
  ResourceEffect
} from '@/types/campaign';
import { usePlayerStore } from './player';

interface CampaignState {
  campaigns: Campaign[];
  selectedCampaignId: string | null;
  playerProgress: Record<string, PlayerCampaignProgress | undefined>;
  currentMission: Mission | null;
  selectedBranch: Branch | null;
  branchOperations: CampaignOperation[];
  branchPOIs: CampaignPOI[];
  currentDialogue: Dialogue | null;
  isLoading: boolean;
  error: string | null;
}

interface BranchStatus {
  poiCompleted: number;
  poiTotal: number;
  operationsCompleted: number;
  operationsTotal: number;
}

export const useCampaignStore = defineStore('campaign', {
  state: (): CampaignState => ({
    campaigns: [],
    selectedCampaignId: null,
    playerProgress: {},
    currentMission: null,
    selectedBranch: null,
    branchOperations: [],
    branchPOIs: [],
    currentDialogue: null,
    isLoading: false,
    error: null
  }),

  getters: {
    selectedCampaign: (state): Campaign | null => {
      const campaign = state.campaigns.find(c => c.id === state.selectedCampaignId);
      if (!campaign) return null;
      
      // Enrich with completion status
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      if (progress) {
        // Deep clone to avoid mutating state
        const enrichedCampaign = JSON.parse(JSON.stringify(campaign)) as Campaign;
        
        // Add is_completed to missions
        enrichedCampaign.chapters.forEach(chapter => {
          chapter.missions.forEach(mission => {
            mission.is_completed = progress.completedMissionIds.includes(mission.id);
          });
        });
        
        // Check if entire campaign is completed
        enrichedCampaign.is_completed = enrichedCampaign.chapters.every(chapter => 
          chapter.missions.every(mission => mission.is_completed)
        );
        
        return enrichedCampaign;
      }
      
      return campaign;
    },

    currentCampaignProgress: (state): PlayerCampaignProgress | undefined => {
      return state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
    },

    hasCampaignStarted: (state): boolean => {
      return !!(state.selectedCampaignId && state.playerProgress[state.selectedCampaignId]);
    },

    isMissionComplete: (state) => (missionId: string): boolean => {
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      return progress ? progress.completedMissionIds.includes(missionId) : false;
    },

    isBranchComplete: (state) => (branchId: string): boolean => {
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      return progress ? progress.completedBranchIds.includes(branchId) : false;
    },

    isPOIComplete: (state) => (poiId: string): boolean => {
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      return progress ? progress.completedPoiIds.includes(poiId) : false;
    },

    isOperationComplete: (state) => (operationId: string): boolean => {
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      return progress ? progress.completedOperationIds.includes(operationId) : false;
    },

    incompletePOIs: (state): CampaignPOI[] => {
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      if (!progress) return [];
      return state.branchPOIs.filter(poi => !progress.completedPoiIds.includes(poi.id));
    },

    incompleteOperations: (state): CampaignOperation[] => {
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      if (!progress) return [];
      return state.branchOperations.filter(op => !progress.completedOperationIds.includes(op.id));
    },

    currentChapter: (state): Chapter | null => {
      if (!state.currentMission) return null;

      // Access the campaigns array directly from state instead of using the getter
      const selectedCampaign = state.campaigns.find(c => c.id === state.selectedCampaignId);
      if (!selectedCampaign) return null;

      const chapter = selectedCampaign.chapters.find(
        (ch: Chapter) => ch.id === state.currentMission?.chapterId
      );
      return chapter || null;
    },

    branchCompletionStatus: (state): BranchStatus => {
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      if (!progress || !state.selectedBranch) return { poiCompleted: 0, poiTotal: 0, operationsCompleted: 0, operationsTotal: 0 };

      const poiCompleted = state.branchPOIs.filter(poi => progress.completedPoiIds.includes(poi.id)).length;
      const operationsCompleted = state.branchOperations.filter(op => progress.completedOperationIds.includes(op.id)).length;

      return {
        poiCompleted,
        poiTotal: state.branchPOIs.length,
        operationsCompleted,
        operationsTotal: state.branchOperations.length
      };
    },

    isBranchCompletable: (state): boolean => {
      // Calculate the status directly instead of referencing this.branchCompletionStatus
      const progress = state.selectedCampaignId ? state.playerProgress[state.selectedCampaignId] : undefined;
      if (!progress || !state.selectedBranch) return false;

      const poiCompleted = state.branchPOIs.filter(poi => progress.completedPoiIds.includes(poi.id)).length;
      const operationsCompleted = state.branchOperations.filter(op => progress.completedOperationIds.includes(op.id)).length;

      return poiCompleted === state.branchPOIs.length &&
             operationsCompleted === state.branchOperations.length;
    },
  },

  actions: {
    async fetchCampaigns() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.getCampaigns();
        if (response.success && response.data) {
          this.campaigns = response.data;
        } else {
          throw new Error(response.error?.message || 'Failed to fetch campaigns');
        }
      } catch (error) {
        this.error = 'Failed to load campaigns';
        console.error('Error fetching campaigns:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async selectCampaign(campaignId: string) {
      this.selectedCampaignId = campaignId;
      await this.fetchPlayerProgress(campaignId);
    },

    async fetchPlayerProgress(campaignId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.getPlayerProgress(campaignId);
        if (response.success) {
          // Check if the response indicates campaign is not started
          if (response.data && 'started' in response.data && response.data.started === false) {
            // Campaign not started yet
            this.playerProgress[campaignId] = undefined;
          } else if (response.data) {
            // Campaign is started
            this.playerProgress[campaignId] = response.data as PlayerCampaignProgress;

            // If campaign is started and has a current mission, fetch it
            if (response.data.currentMissionId) {
              await this.fetchCurrentMission(campaignId);
            }
          }
        } else {
          throw new Error(response.error?.message || 'Failed to fetch player progress');
        }
      } catch (error) {
        this.error = 'Failed to load player progress';
        console.error('Error fetching player progress:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async startCampaign(campaignId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.startCampaign(campaignId);
        if (response.success && response.data) {
          this.playerProgress[campaignId] = response.data;

          // Fetch the current mission
          await this.fetchCurrentMission(campaignId);

          return response;
        } else {
          throw new Error(response.error?.message || 'Failed to start campaign');
        }
      } catch (error) {
        this.error = 'Failed to start campaign';
        console.error('Error starting campaign:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async fetchCurrentMission(campaignId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.getCurrentMission(campaignId);
        if (response.success && response.data) {
          this.currentMission = response.data;

          // If there's a current branch, fetch its details
          const progress = this.playerProgress[campaignId];
          if (progress && progress.currentBranchId) {
            await this.selectBranch(this.currentMission.id, progress.currentBranchId);
          }
        } else {
          this.currentMission = null;
        }
      } catch (error) {
        this.error = 'Failed to load current mission';
        console.error('Error fetching current mission:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async selectBranch(missionId: string, branchId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        // First select the branch on the server
        const response = await campaignService.selectBranch(missionId, branchId);
        if (!response.success) {
          throw new Error(response.error?.message || 'Failed to select branch');
        }

        // Then fetch the branch details
        const branchResponse = await campaignService.getBranch(branchId);
        if (branchResponse.success && branchResponse.data) {
          this.selectedBranch = branchResponse.data;

          // Update player progress to reflect the selected branch
          if (this.selectedCampaignId) {
            const progress = this.playerProgress[this.selectedCampaignId];
            if (progress) {
              progress.currentBranchId = branchId;
            }
          }

          // Fetch operations and POIs for the selected branch
          await Promise.all([
            this.fetchBranchOperations(branchId),
            this.fetchBranchPOIs(branchId)
          ]);

          return branchResponse;
        } else {
          throw new Error(branchResponse.error?.message || 'Failed to fetch branch details');
        }
      } catch (error) {
        this.error = 'Failed to select branch';
        console.error('Error selecting branch:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async fetchBranchOperations(branchId: string) {
      try {
        const response = await campaignService.getOperationsByBranchId(branchId);
        if (response.success && response.data) {
          this.branchOperations = response.data;
        } else {
          this.branchOperations = [];
        }
      } catch (error) {
        console.error('Error fetching branch operations:', error);
        this.branchOperations = [];
      }
    },

    async fetchBranchPOIs(branchId: string) {
      try {
        const response = await campaignService.getPOIsByBranchId(branchId);
        if (response.success && response.data) {
          this.branchPOIs = response.data;
        } else {
          this.branchPOIs = [];
        }
      } catch (error) {
        console.error('Error fetching branch POIs:', error);
        this.branchPOIs = [];
      }
    },

    async checkBranchCompletion(branchId: string) {
      try {
        const response = await campaignService.checkBranchCompletion(branchId);
        if (response.success && response.data) {
          return response.data.complete;
        }
        return false;
      } catch (error) {
        console.error('Error checking branch completion:', error);
        return false;
      }
    },

    async completeBranch(missionId: string, branchId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.completeBranch(missionId, branchId);
        if (response.success && response.data) {
          // Update player progress
          if (this.selectedCampaignId) {
            this.playerProgress[this.selectedCampaignId] = response.data;
          }

          // Clear current state
          this.selectedBranch = null;
          this.branchOperations = [];
          this.branchPOIs = [];

          // Fetch the new current mission if there is one
          if (response.data.currentMissionId && this.selectedCampaignId) {
            await this.fetchCurrentMission(this.selectedCampaignId);
          } else {
            // Campaign completed
            this.currentMission = null;
          }

          return response;
        } else {
          throw new Error(response.error?.message || 'Failed to complete branch');
        }
      } catch (error) {
        this.error = 'Failed to complete branch';
        console.error('Error completing branch:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async interactWithPOI(poiId: string, interactionType: InteractionType) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.interactWithPOI(poiId, interactionType);
        if (response.success && response.data) {
          this.currentDialogue = response.data.dialogue;

          // If there are resource effects, update player resources
          if (response.data.resourceEffect) {
            this.applyResourceEffect(response.data.resourceEffect);
          }

          // Check if POI is now completed
          if (this.selectedCampaignId && this.selectedBranch) {
            await this.fetchPlayerProgress(this.selectedCampaignId);
          }

          return response.data;
        } else {
          throw new Error(response.error?.message || 'Failed to interact with POI');
        }
      } catch (error) {
        this.error = 'Failed to interact with POI';
        console.error('Error interacting with POI:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async completePOI(poiId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.completePOI(poiId);
        if (response.success) {
          // Update player progress
          if (this.selectedCampaignId) {
            await this.fetchPlayerProgress(this.selectedCampaignId);
          }

          return response;
        } else {
          throw new Error(response.error?.message || 'Failed to complete POI');
        }
      } catch (error) {
        this.error = 'Failed to complete POI';
        console.error('Error completing POI:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async completeOperation(operationId: string, attemptId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.completeOperation(operationId, attemptId);
        if (response.success) {
          // Update player progress
          if (this.selectedCampaignId) {
            await this.fetchPlayerProgress(this.selectedCampaignId);
          }

          return response;
        } else {
          throw new Error(response.error?.message || 'Failed to complete operation');
        }
      } catch (error) {
        this.error = 'Failed to complete operation';
        console.error('Error completing operation:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    applyResourceEffect(effect: ResourceEffect) {
      const playerStore = usePlayerStore();
      if (!playerStore.profile) return;

      // Update player resources
      if (effect.money) {
        playerStore.profile.money += effect.money;
      }

      if (effect.crew) {
        playerStore.profile.crew += effect.crew;
      }

      if (effect.weapons) {
        playerStore.profile.weapons += effect.weapons;
      }

      if (effect.vehicles) {
        playerStore.profile.vehicles += effect.vehicles;
      }

      if (effect.respect) {
        playerStore.profile.respect += effect.respect;
      }

      if (effect.influence) {
        playerStore.profile.influence += effect.influence;
      }

      if (effect.heat) {
        playerStore.profile.heat += effect.heat;
      }
    },

    // Add a new hook to the store to handle operation completion in OperationsView
    async handleOperationCompleted(operationId: string, attemptId: string) {
      // First check if this is a campaign operation
      // This would typically be done by checking the metadata on the operation
      const isCampaignOperation = true; // This should be determined by checking metadata

      if (isCampaignOperation) {
        // If it's a campaign operation, mark it as completed
        await this.completeOperation(operationId, attemptId);
      }
    },

    // Similarly add a hook for POI interaction in TerritoryView
    async handlePOIInteraction(poiId: string) {
      // Check if this is a campaign POI
      const isCampaignPOI = true; // This should be determined by checking metadata

      if (isCampaignPOI) {
        // If it's a campaign POI, mark it as completed
        await this.completePOI(poiId);
      }
    },

    reset() {
      this.campaigns = [];
      this.selectedCampaignId = null;
      this.playerProgress = {};
      this.currentMission = null;
      this.selectedBranch = null;
      this.branchOperations = [];
      this.branchPOIs = [];
      this.currentDialogue = null;
      this.isLoading = false;
      this.error = null;
    }
  }
});
