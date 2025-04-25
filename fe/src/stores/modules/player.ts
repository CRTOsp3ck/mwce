// src/stores/modules/player.ts

import { defineStore } from 'pinia';
import playerService from '@/services/playerService';
import {
  PlayerProfile,
  PlayerStats,
  Notification,
  PlayerTitle,
  NotificationType
} from '@/types/player';

interface PlayerState {
  profile: PlayerProfile | null;
  stats: PlayerStats | null;
  notifications: Notification[];
  isLoading: boolean;
  error: string | null;
}

export const usePlayerStore = defineStore('player', {
  state: (): PlayerState => ({
    profile: null,
    stats: null,
    notifications: [],
    isLoading: false,
    error: null,
  }),

  getters: {
    playerMoney: (state) => state.profile?.money || 0,
    playerCrew: (state) => state.profile?.crew || 0,
    playerWeapons: (state) => state.profile?.weapons || 0,
    playerVehicles: (state) => state.profile?.vehicles || 0,

    playerRespect: (state) => state.profile?.respect || 0,
    playerInfluence: (state) => state.profile?.influence || 0,
    playerHeat: (state) => state.profile?.heat || 0,

    playerTitle: (state) => state.profile?.title || PlayerTitle.ASSOCIATE,

    controlledHotspots: (state) => state.profile?.controlledHotspots || 0,
    totalHotspots: (state) => state.profile?.totalHotspotCount || 0,

    hourlyRevenue: (state) => state.profile?.hourlyRevenue || 0,
    pendingCollections: (state) => state.profile?.pendingCollections || 0,

    maxCrew: (state) => state.profile?.maxCrew || 0,
    maxWeapons: (state) => state.profile?.maxWeapons || 0,
    maxVehicles: (state) => state.profile?.maxVehicles || 0,

    unreadNotifications: (state) =>
      state.notifications.filter(n => !n.read).length,
  },

  actions: {
    async fetchProfile() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await playerService.getProfile();
        if (response.success && response.data) {
          this.profile = response.data;
        } else {
          throw new Error('Failed to load profile');
        }
      } catch (error) {
        this.error = 'Failed to load player profile';
        console.error('Error fetching player profile:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchStats() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await playerService.getStats();
        if (response.success && response.data) {
          this.stats = response.data;
        } else {
          throw new Error('Failed to load stats');
        }
      } catch (error) {
        this.error = 'Failed to load player stats';
        console.error('Error fetching player stats:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchNotifications() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await playerService.getNotifications();
        if (response.success && response.data) {
          this.notifications = response.data;
        } else {
          throw new Error('Failed to load notifications');
        }
      } catch (error) {
        this.error = 'Failed to load notifications';
        console.error('Error fetching notifications:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async markNotificationsAsRead() {
      try {
        const response = await playerService.markAllNotificationsAsRead();
        if (response.success) {
          this.notifications.forEach(notification => {
            notification.read = true;
          });
        }
      } catch (error) {
        console.error('Error marking notifications as read:', error);
      }
    },

    async collectAllPending() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await playerService.collectAllPending();

        if (!response.success || !response.data) {
          throw new Error('Failed to collect pending resources');
        }

        const collectedAmount = response.data.collectedAmount;

        // Update player money
        if (this.profile && collectedAmount) {
          this.profile.money += collectedAmount;
          this.profile.pendingCollections = 0;
        }

        return {
          collectedAmount,
          gameMessage: response.gameMessage
        };
      } catch (error) {
        this.error = 'Failed to collect pending resources';
        console.error('Error collecting pending resources:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    // Add a new notification to the list (used from SSE events)
    addNotification(notification: Notification) {
      this.notifications.unshift(notification);
    }
  }
});
