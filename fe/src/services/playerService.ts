// src/services/playerService.ts

import api from './api';
import { PlayerProfile, PlayerStats, Notification } from '@/types/player';

// Endpoints
const ENDPOINTS = {
  PROFILE: '/player/profile',
  STATS: '/player/stats',
  NOTIFICATIONS: '/player/notifications',
  MARK_NOTIFICATIONS_READ: '/player/notifications/read',
  COLLECT_ALL: '/player/collect-all'
};

export default {
  /**
   * Get the player's profile
   */
  getProfile() {
    return api.get<PlayerProfile>(ENDPOINTS.PROFILE);
  },
  
  /**
   * Get the player's stats
   */
  getStats() {
    return api.get<PlayerStats>(ENDPOINTS.STATS);
  },
  
  /**
   * Get the player's notifications
   */
  getNotifications() {
    return api.get<Notification[]>(ENDPOINTS.NOTIFICATIONS);
  },
  
  /**
   * Mark all notifications as read
   */
  markAllNotificationsAsRead() {
    return api.post(ENDPOINTS.MARK_NOTIFICATIONS_READ);
  },
  
  /**
   * Mark a specific notification as read
   */
  markNotificationAsRead(notificationId: string) {
    return api.post(`${ENDPOINTS.NOTIFICATIONS}/${notificationId}/read`);
  },
  
  /**
   * Collect all pending resources from controlled hotspots
   */
  collectAllPending() {
    return api.post(ENDPOINTS.COLLECT_ALL);
  }
};