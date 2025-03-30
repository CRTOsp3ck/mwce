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
        this.profile = response.data;
      } catch (error) {
        this.error = 'Failed to load player profile';
        console.error('Error fetching player profile:', error);
        
        // For development purposes, set mock data
        this.setMockProfile();
      } finally {
        this.isLoading = false;
      }
    },
    
    async fetchStats() {
      this.isLoading = true;
      this.error = null;
      
      try {
        const response = await playerService.getStats();
        this.stats = response.data;
      } catch (error) {
        this.error = 'Failed to load player stats';
        console.error('Error fetching player stats:', error);
        
        // For development purposes, set mock data
        this.setMockStats();
      } finally {
        this.isLoading = false;
      }
    },
    
    async fetchNotifications() {
      this.isLoading = true;
      this.error = null;
      
      try {
        const response = await playerService.getNotifications();
        this.notifications = response.data;
      } catch (error) {
        this.error = 'Failed to load notifications';
        console.error('Error fetching notifications:', error);
        
        // For development purposes, set mock data
        this.setMockNotifications();
      } finally {
        this.isLoading = false;
      }
    },
    
    async markNotificationsAsRead() {
      try {
        await playerService.markAllNotificationsAsRead();
        this.notifications.forEach(notification => {
          notification.read = true;
        });
      } catch (error) {
        console.error('Error marking notifications as read:', error);
      }
    },
    
    async collectAllPending() {
      this.isLoading = true;
      this.error = null;
      
      try {
        const response = await playerService.collectAllPending();
        
        // Update player money
        if (this.profile && response.data.collectedAmount) {
          this.profile.money += response.data.collectedAmount;
          this.profile.pendingCollections = 0;
        }
        
        return response.data;
      } catch (error) {
        this.error = 'Failed to collect pending resources';
        console.error('Error collecting pending resources:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },
    
    // For development and testing - set mock data
    setMockProfile() {
      this.profile = {
        id: '1',
        name: 'Don Corleone',
        title: PlayerTitle.CAPO,
        money: 250000,
        crew: 15,
        maxCrew: 25,
        weapons: 20,
        maxWeapons: 30,
        vehicles: 8,
        maxVehicles: 12,
        respect: 65,
        influence: 48,
        heat: 32,
        controlledHotspots: 7,
        totalHotspotCount: 30,
        hourlyRevenue: 12500,
        pendingCollections: 3750,
        createdAt: new Date().toISOString(),
        lastActive: new Date().toISOString()
      };
    },
    
    setMockStats() {
      this.stats = {
        totalOperationsCompleted: 47,
        totalMoneyEarned: 1250000,
        totalHotspotsControlled: 15,
        maxInfluenceAchieved: 72,
        maxRespectAchieved: 85,
        successfulTakeovers: 32,
        failedTakeovers: 18
      };
    },
    
    setMockNotifications() {
      this.notifications = [
        {
          id: '1',
          playerId: '1',
          message: 'Your territory in East District is under attack!',
          type: NotificationType.TERRITORY,
          timestamp: new Date(Date.now() - 10 * 60000).toISOString(),
          read: false
        },
        {
          id: '2',
          playerId: '1',
          message: 'Special operation "High Stakes Heist" is now available!',
          type: NotificationType.OPERATION,
          timestamp: new Date(Date.now() - 30 * 60000).toISOString(),
          read: false
        },
        {
          id: '3',
          playerId: '1',
          message: 'Daily collection from Downtown Speakeasy is available',
          type: NotificationType.COLLECTION,
          timestamp: new Date(Date.now() - 120 * 60000).toISOString(),
          read: true
        }
      ];
    }
  }
});