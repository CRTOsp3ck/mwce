// src/components/layout/MobileHeader.vue

<template>
  <header class="app-header">
    <!-- Hamburger Menu -->
    <button class="mobile-nav-toggle" @click="toggleMobileDrawer">
      ☰
    </button>

    <!-- Logo -->
    <div class="logo">
      <router-link to="/">
        <h1>Mafioso: <span class="gold-text">CE</span></h1>
      </router-link>
    </div>

    <!-- Region indicator -->
    <div class="region-indicator" @click="router.push('/travel')">
      <div class="region-icon">{{ currentRegion ? '🌆' : '🏠' }}</div>
      <div class="region-name">{{ currentLocationName }}</div>
    </div>

    <!-- User Controls -->
    <div class="user-controls">
      <!-- Notification Bell -->
      <div class="notification-bell" @click="toggleNotifications">
        <div class="bell-icon">
          <i class="notification-icon">🔔</i>
          <span v-if="unreadNotifications" class="notification-badge">{{ unreadNotifications }}</span>
        </div>

        <!-- Notifications Dropdown -->
        <div v-if="showNotifications" class="notifications-dropdown">
          <div class="notifications-header">
            <h4>Notifications</h4>
            <button @click.stop="markAllAsRead" class="mark-read-btn">Mark all as read</button>
          </div>

          <div class="notifications-tabs">
            <button class="tab-btn" :class="{ active: activeNotifTab === 'all' }"
              @click.stop="activeNotifTab = 'all'">
              All
            </button>
            <button class="tab-btn" :class="{ active: activeNotifTab === 'territory' }"
              @click.stop="activeNotifTab = 'territory'">
              Territory
            </button>
            <button class="tab-btn" :class="{ active: activeNotifTab === 'operations' }"
              @click.stop="activeNotifTab = 'operations'">
              Ops
            </button>
          </div>

          <div class="notifications-list" v-if="filteredNotifications.length > 0">
            <div v-for="notification in filteredNotifications" :key="notification.id"
              class="notification-item" :class="{
                'unread': !notification.read,
                'territory': notification.type === 'territory',
                'operation': notification.type === 'operation',
                'collection': notification.type === 'collection',
                'heat': notification.type === 'heat',
                'system': notification.type === 'system',
                'travel': notification.type === 'travel'
              }">
              <div class="notification-icon">
                {{ getNotificationIcon(notification.type) }}
              </div>
              <div class="notification-content">
                <p>{{ notification.message }}</p>
                <span class="notification-time">{{ formatTime(notification.timestamp) }}</span>
              </div>
              <div class="notification-actions">
                <button v-if="!notification.read" class="action-btn mark-read"
                  @click.stop="markAsRead(notification.id)">
                  ✓
                </button>
              </div>
            </div>
          </div>

          <div v-else class="empty-notifications">
            <div class="empty-icon">🔍</div>
            <p>No notifications</p>
          </div>

          <div class="notifications-footer">
            <router-link to="/notifications" class="view-all-link" @click="showNotifications = false">
              View All
            </router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Mobile Drawer Menu -->
    <div class="mobile-drawer" :class="{ open: showMobileDrawer }">
      <div class="drawer-header">
        <h3>Main Menu</h3>
        <button class="close-drawer" @click="showMobileDrawer = false">×</button>
      </div>
      <div class="drawer-content">
        <!-- User Profile Section -->
        <div class="drawer-profile">
          <div class="user-avatar large">{{ playerAvatar }}</div>
          <div class="user-details">
            <div class="user-name">{{ playerName }}</div>
            <div class="user-title">{{ playerTitle }}</div>
          </div>
        </div>

        <!-- Main Navigation -->
        <nav class="drawer-nav">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="drawer-nav-item"
            @click="showMobileDrawer = false"
          >
            <span class="item-icon">{{ item.icon }}</span>
            <span class="item-label">{{ item.name }}</span>
          </router-link>
        </nav>

        <!-- Additional Links -->
        <div class="drawer-section">
          <h4>Quick Access</h4>
          <router-link to="/profile" class="drawer-link" @click="showMobileDrawer = false">
            <span class="link-icon">👤</span>
            <span>My Profile</span>
          </router-link>
          <router-link to="/settings" class="drawer-link" @click="showMobileDrawer = false">
            <span class="link-icon">⚙️</span>
            <span>Settings</span>
          </router-link>
          <router-link to="/help" class="drawer-link" @click="showMobileDrawer = false">
            <span class="link-icon">❓</span>
            <span>Help</span>
          </router-link>
        </div>

        <!-- Logout -->
        <div class="drawer-footer">
          <button class="logout-btn" @click="logout">
            <span class="btn-icon">🚪</span>
            <span>Sign Out</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Drawer Backdrop -->
    <div class="mobile-drawer-backdrop" :class="{ open: showMobileDrawer }" @click="showMobileDrawer = false"></div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { usePlayerStore } from '@/stores/modules/player';
import { useTravelStore } from '@/stores/modules/travel';
import {
    Notification,
    NotificationType
} from '@/types/player';

const route = useRoute();
const router = useRouter();
const playerStore = usePlayerStore();
const travelStore = useTravelStore();

// UI state
const currentRoute = computed(() => route.path);
const showNotifications = ref(false);
const showMobileDrawer = ref(false);
const activeNotifTab = ref('all');

// Check if user is logged in
const isLoggedIn = computed(() => {
  return !!localStorage.getItem('auth_token');
});

// Player info
const playerName = computed(() => playerStore.profile?.name || 'Boss');
const playerTitle = computed(() => playerStore.profile?.title || 'Capo');
const playerAvatar = ref('🤵');

// Current region info
const currentRegion = computed(() => travelStore.currentRegion);
const currentLocationName = computed(() => travelStore.currentLocationName);

// Navigation items for drawer
const navItems = [
    { name: 'Home', path: '/', icon: '🏠' },
    { name: 'Territory', path: '/territory', icon: '🏙️' },
    { name: 'Operations', path: '/operations', icon: '🎯' },
    { name: 'Market', path: '/market', icon: '💹' },
    { name: 'Travel', path: '/travel', icon: '✈️' },
    { name: 'Rankings', path: '/rankings', icon: '🏆' },
    { name: 'NFT', path: '/nft', icon: '💎' }
];

// Notifications
const notifications = computed(() => playerStore.notifications);
const unreadNotifications = computed(() => playerStore.unreadNotifications);

// Filtered notifications based on active tab
const filteredNotifications = computed(() => {
    if (activeNotifTab.value === 'all') {
        return notifications.value;
    } else if (activeNotifTab.value === 'territory') {
        return notifications.value.filter(n =>
            n.type === NotificationType.TERRITORY ||
            n.type === NotificationType.COLLECTION
        );
    } else if (activeNotifTab.value === 'operations') {
        return notifications.value.filter(n =>
            n.type === NotificationType.OPERATION ||
            n.type === NotificationType.HEAT
        );
    }
    return notifications.value;
});

// Event listeners for clicking outside
let clickOutsideHandler = (event: MouseEvent) => {
    const notifBell = document.querySelector('.notification-bell');

    if (notifBell && !notifBell.contains(event.target as Node)) {
        showNotifications.value = false;
    }
};

// Lifecycle hooks
onMounted(() => {
    // Fetch notifications
    if (isLoggedIn.value && playerStore.notifications.length === 0) {
        playerStore.fetchNotifications();
    }

    // Add click outside listener
    document.addEventListener('click', clickOutsideHandler);
});

onBeforeUnmount(() => {
    // Remove click outside listener
    document.removeEventListener('click', clickOutsideHandler);
});

// Watch for route changes to close dropdowns
watch(route, () => {
    showNotifications.value = false;
    showMobileDrawer.value = false;
});

// Helper functions
function getNotificationIcon(type: NotificationType): string {
    switch (type) {
        case NotificationType.TERRITORY:
            return '🏙️';
        case NotificationType.OPERATION:
            return '🎯';
        case NotificationType.COLLECTION:
            return '💰';
        case NotificationType.HEAT:
            return '🔥';
        case NotificationType.SYSTEM:
            return '🔔';
        case 'travel' as NotificationType:
            return '✈️';
        default:
            return '🔔';
    }
}

function formatTime(date: string): string {
    const now = new Date();
    const diff = Math.floor((now.getTime() - new Date(date).getTime()) / 60000); // difference in minutes

    if (diff < 1) return 'Now';
    if (diff < 60) return `${diff}m`;
    if (diff < 1440) return `${Math.floor(diff / 60)}h`;
    return `${Math.floor(diff / 1440)}d`;
}

// Action functions
function toggleNotifications(event: MouseEvent) {
    event.stopPropagation();
    showNotifications.value = !showNotifications.value;
}

function toggleMobileDrawer() {
    showMobileDrawer.value = !showMobileDrawer.value;

    // Close notifications if open
    if (showMobileDrawer.value) {
        showNotifications.value = false;
    }
}

function markAllAsRead() {
    playerStore.markNotificationsAsRead();
}

function markAsRead(notificationId: string) {
    // In a real app, this would call the API
    const notification = notifications.value.find(n => n.id === notificationId);
    if (notification) {
        notification.read = true;
    }
}

async function logout() {
    // Clear auth token
    localStorage.removeItem('auth_token');

    // Close the drawer
    showMobileDrawer.value = false;

    // Redirect to login page
    router.push('/login');
}
</script>

<style lang="scss">
// Mobile header specific styles
.drawer-profile {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-md 0;
  margin-bottom: $spacing-md;
  border-bottom: 1px solid $border-color;

  .user-avatar {
    width: 48px;
    height: 48px;
    background-color: $primary-color;
    border-radius: 50%;
    @include flex-center;
    font-size: 24px;
  }

  .user-details {
    .user-name {
      font-weight: 600;
      @include gold-accent;
    }

    .user-title {
      font-size: $font-size-sm;
      color: $text-secondary;
    }
  }
}

.drawer-nav {
  @include flex-column;
  gap: $spacing-xs;
  margin-bottom: $spacing-lg;

  .drawer-nav-item {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    padding: $spacing-sm;
    border-radius: $border-radius-sm;
    text-decoration: none;
    color: $text-color;
    transition: $transition-base;

    &:hover {
      background-color: rgba(255, 255, 255, 0.05);
    }

    &.router-link-active {
      background-color: rgba($secondary-color, 0.1);
      color: $secondary-color;
    }

    .item-icon {
      font-size: 20px;
      width: 24px;
      text-align: center;
    }
  }
}

.drawer-section {
  margin-bottom: $spacing-lg;

  h4 {
    margin-bottom: $spacing-sm;
    color: $text-secondary;
    font-size: $font-size-sm;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .drawer-link {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    padding: $spacing-sm;
    border-radius: $border-radius-sm;
    text-decoration: none;
    color: $text-color;
    transition: $transition-base;

    &:hover {
      background-color: rgba(255, 255, 255, 0.05);
    }

    .link-icon {
      font-size: 18px;
      width: 24px;
      text-align: center;
    }
  }
}

.drawer-footer {
  margin-top: auto;
  padding-top: $spacing-lg;
  border-top: 1px solid $border-color;

  .logout-btn {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    width: 100%;
    padding: $spacing-md;
    background-color: rgba($danger-color, 0.1);
    border: none;
    border-radius: $border-radius-sm;
    color: $danger-color;
    font-weight: 600;
    text-align: left;
    cursor: pointer;
    transition: $transition-base;

    &:hover {
      background-color: rgba($danger-color, 0.2);
    }

    .btn-icon {
      font-size: 18px;
    }
  }
}
</style>
