// src/stores/modules/campaign.ts

import { defineStore } from 'pinia';
import campaignService from '@/services/campaignService';
import { usePlayerStore } from './player';
import {
  Campaign,
  Chapter,
  Mission,
  PlayerCampaignProgress,
  PlayerMissionProgress,
  MissionStatus,
  MissionCompleteResult
} from '@/types/campaign';

interface CampaignState {
  campaigns: Campaign[];
  campaignProgress: { [key: string]: PlayerCampaignProgress };
  missionProgress: { [key: string]: PlayerMissionProgress };
  currentCampaign: Campaign | null;
  currentChapter: Chapter | null;
  currentMission: Mission | null;
  isLoading: boolean;
  error: string | null;
}

export const useCampaignStore = defineStore('campaign', {
  state: (): CampaignState => ({
    campaigns: [],
    campaignProgress: {},
    missionProgress: {},
    currentCampaign: null,
    currentChapter: null,
    currentMission: null,
    isLoading: false,
    error: null
  }),

  getters: {
    availableCampaigns: (state) => {
      return state.campaigns.filter(c => c.isActive);
    },

    inProgressCampaigns: (state) => {
      const result: Campaign[] = [];

      for (const campaignId in state.campaignProgress) {
        const progress = state.campaignProgress[campaignId];
        if (!progress.isCompleted) {
          const campaign = state.campaigns.find(c => c.id === campaignId);
          if (campaign) {
            result.push(campaign);
          }
        }
      }

      return result;
    },

    completedCampaigns: (state) => {
      const result: Campaign[] = [];

      for (const campaignId in state.campaignProgress) {
        const progress = state.campaignProgress[campaignId];
        if (progress.isCompleted) {
          const campaign = state.campaigns.find(c => c.id === campaignId);
          if (campaign) {
            result.push(campaign);
          }
        }
      }

      return result;
    },

    chapterMissions: (state) => {
      if (!state.currentChapter) return [];
      return state.currentChapter.missions || [];
    },

    missionChoices: (state) => {
      if (!state.currentMission) return [];
      return state.currentMission.choices || [];
    },

    // Helper getter to get mission status
    getMissionStatus: (state) => (missionId: string) => {
      const progress = state.missionProgress[missionId];
      if (!progress) return MissionStatus.NOT_STARTED;
      return progress.status;
    },

    // Helper to check if a mission is locked
    isMissionLocked: (state) => (missionId: string) => {
      const mission = state.currentChapter?.missions?.find(m => m.id === missionId);
      if (!mission) return true;
      return mission.isLocked;
    }
  },

  actions: {
    async fetchCampaigns() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.getCampaigns();

        if (response.success && response.data) {
          this.campaigns = response.data.campaigns;
          this.campaignProgress = response.data.progress;
        } else {
          throw new Error('Failed to load campaigns');
        }
      } catch (error) {
        this.error = 'Failed to load campaigns';
        console.error('Error fetching campaigns:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchCampaignDetail(campaignId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.getCampaignDetail(campaignId);

        if (response.success && response.data) {
          this.currentCampaign = response.data.campaign;

          if (response.data.progress) {
            this.campaignProgress[campaignId] = response.data.progress;
          }

          // Update mission progress
          for (const missionId in response.data.missionProgress) {
            this.missionProgress[missionId] = response.data.missionProgress[missionId];
          }

          // If campaign has progress, load current chapter
          if (response.data.progress?.currentChapterId) {
            const chapterId = response.data.progress.currentChapterId;
            await this.fetchChapter(chapterId);
          } else if (this.currentCampaign.chapters && this.currentCampaign.chapters.length > 0) {
            // If no progress, set the first chapter as current
            this.currentChapter = this.currentCampaign.chapters[0];
          }
        } else {
          throw new Error('Failed to load campaign details');
        }
      } catch (error) {
        this.error = 'Failed to load campaign details';
        console.error('Error fetching campaign details:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchChapter(chapterId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.getChapter(chapterId);

        if (response.success && response.data) {
          this.currentChapter = response.data.chapter;

          // Update mission progress
          for (const missionId in response.data.missionProgress) {
            this.missionProgress[missionId] = response.data.missionProgress[missionId];
          }

          // If campaign progress has a current mission, load it
          const campaignId = this.currentChapter.campaignId;
          const progress = this.campaignProgress[campaignId];

          if (progress && progress.currentMissionId) {
            await this.fetchMission(progress.currentMissionId);
          } else if (this.currentChapter.missions && this.currentChapter.missions.length > 0) {

            // If no progress, set the first mission as current
            await this.fetchMission(this.currentChapter.missions[0].id);
          }
        } else {
          throw new Error('Failed to load chapter details');
        }
      } catch (error) {
        this.error = 'Failed to load chapter details';
        console.error('Error fetching chapter details:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchMission(missionId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.getMission(missionId);

        if (response.success && response.data) {
          this.currentMission = response.data.mission;

          // Update mission progress
          if (response.data.progress) {
            this.missionProgress[missionId] = response.data.progress;
          }
        } else {
          throw new Error('Failed to load mission details');
        }
      } catch (error) {
        this.error = 'Failed to load mission details';
        console.error('Error fetching mission details:', error);
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
          this.campaignProgress[campaignId] = response.data;
          await this.fetchCampaignDetail(campaignId);
          return {
            success: true,
            message: response.gameMessage?.message || 'Campaign started successfully'
          };
        } else {
          throw new Error('Failed to start campaign');
        }
      } catch (error) {
        this.error = 'Failed to start campaign';
        console.error('Error starting campaign:', error);
        return {
          success: false,
          message: error instanceof Error ? error.message : 'An unknown error occurred'
        };
      } finally {
        this.isLoading = false;
      }
    },

    async startMission(missionId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.startMission(missionId);

        if (response.success && response.data) {
          this.missionProgress[missionId] = response.data;
          await this.fetchMission(missionId);
          return {
            success: true,
            message: response.gameMessage?.message || 'Mission started successfully'
          };
        } else {
          throw new Error('Failed to start mission');
        }
      } catch (error) {
        this.error = 'Failed to start mission';
        console.error('Error starting mission:', error);
        return {
          success: false,
          message: error instanceof Error ? error.message : 'An unknown error occurred'
        };
      } finally {
        this.isLoading = false;
      }
    },

    async completeMission(missionId: string, choiceId?: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await campaignService.completeMission(missionId, choiceId);

        if (response.success && response.data) {
          // Update mission progress for completed mission
          this.missionProgress[missionId] = response.data.progress;

          // If there's a next mission, update current mission and fetch it
          if (response.data.nextMission) {
            const nextMissionId = response.data.nextMission.id;

            // Create an entry for the next mission's progress if needed
            if (!this.missionProgress[nextMissionId]) {
              this.missionProgress[nextMissionId] = {
                id: '',
                playerId: '',
                missionId: nextMissionId,
                status: MissionStatus.NOT_STARTED
              };
            }

            // Update campaign progress
            const currentCampaignId = this.currentCampaign?.id;
            if (currentCampaignId && this.campaignProgress[currentCampaignId]) {
              this.campaignProgress[currentCampaignId].currentMissionId = nextMissionId;

              // If next mission is in a different chapter, update current chapter
              if (response.data.nextMission.chapterId !== this.currentChapter?.id) {
                this.campaignProgress[currentCampaignId].currentChapterId = response.data.nextMission.chapterId;
                await this.fetchChapter(response.data.nextMission.chapterId);
              }
            }

            // Fetch the next mission
            await this.fetchMission(nextMissionId);
          }

          // Apply rewards to player resources
          this.applyRewards(response.data.rewards);

          return {
            success: true,
            result: response.data,
            message: response.gameMessage?.message || 'Mission completed successfully'
          };
        } else {
          throw new Error('Failed to complete mission');
        }
      } catch (error) {
        this.error = 'Failed to complete mission';
        console.error('Error completing mission:', error);
        return {
          success: false,
          message: error instanceof Error ? error.message : 'An unknown error occurred'
        };
      } finally {
        this.isLoading = false;
      }
    },

    // Helper method to apply rewards to player resources
    applyRewards(rewards: MissionCompleteResult['rewards']) {
      const playerStore = usePlayerStore();

      if (!playerStore.profile) return;

      if (rewards.money) {
        playerStore.profile.money += rewards.money;
      }

      if (rewards.crew) {
        playerStore.profile.crew += rewards.crew;
      }

      if (rewards.weapons) {
        playerStore.profile.weapons += rewards.weapons;
      }

      if (rewards.vehicles) {
        playerStore.profile.vehicles += rewards.vehicles;
      }

      if (rewards.respect) {
        playerStore.profile.respect += rewards.respect;
      }

      if (rewards.influence) {
        playerStore.profile.influence += rewards.influence;
      }

      if (rewards.heatReduction) {
        playerStore.profile.heat = Math.max(0, playerStore.profile.heat - rewards.heatReduction);
      }
    }
  }
});
