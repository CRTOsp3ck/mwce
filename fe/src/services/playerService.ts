// src/services/playerService.ts

import api, { ApiResponse } from './api';
import { PlayerProfile, PlayerStats, Notification } from '@/types/player';

// Define interfaces that match backend structures
export interface CollectAllResponse {
  collectedAmount: number;
  message: string;
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
  PROFILE: '/player/profile',
  STATS: '/player/stats',
  NOTIFICATIONS: '/player/notifications',
  MARK_NOTIFICATIONS_READ: '/player/notifications/read',
  MARK_NOTIFICATION_READ: '/player/notifications', // + /:id/read
  COLLECT_ALL: '/player/collect-all'
};

export default {
  /**
   * Get the player's profile
   */
  getProfile() {
    return api.get<ApiResponse<PlayerProfile>>(ENDPOINTS.PROFILE);
  },
  
  /**
   * Get the player's stats
   */
  getStats() {
    return api.get<ApiResponse<PlayerStats>>(ENDPOINTS.STATS);
  },
  
  /**
   * Get the player's notifications
   */
  getNotifications() {
    return api.get<ApiResponse<Notification[]>>(ENDPOINTS.NOTIFICATIONS);
  },
  
  /**
   * Mark all notifications as read
   */
  markAllNotificationsAsRead() {
    return api.post<ApiResponse<{message: string}>>(ENDPOINTS.MARK_NOTIFICATIONS_READ);
  },
  
  /**
   * Mark a specific notification as read
   */
  markNotificationAsRead(notificationId: string) {
    return api.post<ApiResponse<{message: string}>>(`${ENDPOINTS.MARK_NOTIFICATION_READ}/${notificationId}/read`);
  },
  
  /**
   * Collect all pending resources from controlled hotspots
   */
  collectAllPending() {
    return api.post<ApiResponse<GameMessageResponse<CollectAllResponse>>>(ENDPOINTS.COLLECT_ALL);
  }
};