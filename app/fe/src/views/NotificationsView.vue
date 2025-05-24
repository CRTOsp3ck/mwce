// src/views/NotificationsView.vue

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseNotification from '@/components/ui/BaseNotification.vue';
import { usePlayerStore } from '@/stores/modules/player';
import {
  Notification,
  NotificationType
} from '@/types/player';

// Access the player store
const playerStore = usePlayerStore();

// View state
const isLoading = ref(false);
const activeFilter = ref<'all' | 'territory' | 'operation' | 'collection' | 'heat' | 'system'>('all');
const searchQuery = ref('');
const showNotification = ref(false);
const notificationType = ref('info');
const notificationMessage = ref('');
const isMarkingRead = ref(false);

// Get notifications from the player store
const notifications = computed(() => playerStore.notifications);

// Filtered notifications based on active tab and search
const filteredNotifications = computed(() => {
  let filtered = [...notifications.value];

  // Apply type filter
  if (activeFilter.value !== 'all') {
    if (activeFilter.value === 'territory') {
      filtered = filtered.filter(n => n.type === NotificationType.TERRITORY);
    } else if (activeFilter.value === 'operation') {
      filtered = filtered.filter(n => n.type === NotificationType.OPERATION);
    } else if (activeFilter.value === 'collection') {
      filtered = filtered.filter(n => n.type === NotificationType.COLLECTION);
    } else if (activeFilter.value === 'heat') {
      filtered = filtered.filter(n => n.type === NotificationType.HEAT);
    } else if (activeFilter.value === 'system') {
      filtered = filtered.filter(n => n.type === NotificationType.SYSTEM);
    }
  }

  // Apply search filter
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(n => n.message.toLowerCase().includes(query));
  }

  return filtered;
});

// Statistics
const totalNotifications = computed(() => notifications.value.length);
const unreadNotifications = computed(() => notifications.value.filter(n => !n.read).length);

// Type counts
const territoryCounts = computed(() =>
  notifications.value.filter(n => n.type === NotificationType.TERRITORY).length
);
const operationCounts = computed(() =>
  notifications.value.filter(n => n.type === NotificationType.OPERATION).length
);
const collectionCounts = computed(() =>
  notifications.value.filter(n => n.type === NotificationType.COLLECTION).length
);
const heatCounts = computed(() =>
  notifications.value.filter(n => n.type === NotificationType.HEAT).length
);
const systemCounts = computed(() =>
  notifications.value.filter(n => n.type === NotificationType.SYSTEM).length
);

// Load data when component is mounted
onMounted(async () => {
  if (notifications.value.length === 0) {
    isLoading.value = true;
    await playerStore.fetchNotifications();
    isLoading.value = false;
  }
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
    default:
      return '🔔';
  }
}

function formatTime(date: string): string {
  const now = new Date();
  const notificationDate = new Date(date);

  // Format as full date and time if more than a day old
  if (now.getTime() - notificationDate.getTime() > 24 * 60 * 60 * 1000) {
    return notificationDate.toLocaleDateString() + ' ' +
           notificationDate.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  // Otherwise use relative time
  const diff = Math.floor((now.getTime() - notificationDate.getTime()) / 60000); // difference in minutes

  if (diff < 1) return 'Just now';
  if (diff < 60) return `${diff} min ago`;
  return `${Math.floor(diff / 60)} hours ago`;
}

function getTypeLabel(type: NotificationType): string {
  switch (type) {
    case NotificationType.TERRITORY:
      return 'Territory';
    case NotificationType.OPERATION:
      return 'Operation';
    case NotificationType.COLLECTION:
      return 'Collection';
    case NotificationType.HEAT:
      return 'Heat';
    case NotificationType.SYSTEM:
      return 'System';
    default:
      return 'Notification';
  }
}

function getTypeClass(type: NotificationType): string {
  switch (type) {
    case NotificationType.TERRITORY:
      return 'territory';
    case NotificationType.OPERATION:
      return 'operation';
    case NotificationType.COLLECTION:
      return 'collection';
    case NotificationType.HEAT:
      return 'heat';
    case NotificationType.SYSTEM:
      return 'system';
    default:
      return '';
  }
}

// Action functions
async function markAllAsRead() {
  if (isMarkingRead.value) return;

  isMarkingRead.value = true;

  try {
    await playerStore.markNotificationsAsRead();

    // Show success notification
    showNotificationMessage('success', 'All notifications marked as read');
  } catch (error) {
    showNotificationMessage('danger', 'Failed to mark notifications as read');
  } finally {
    isMarkingRead.value = false;
  }
}

async function markAsRead(notificationId: string) {
  try {
    // In a real app, this would call the API through the store
    // For now, just updating the local state
    const notification = notifications.value.find(n => n.id === notificationId);
    if (notification) {
      notification.read = true;
    }
  } catch (error) {
    console.error('Error marking notification as read:', error);
  }
}

function resetFilters() {
  activeFilter.value = 'all';
  searchQuery.value = '';
}

// Notification helpers
function showNotificationMessage(type: string, message: string) {
  notificationType.value = type;
  notificationMessage.value = message;
  showNotification.value = true;
}

function closeNotification() {
  showNotification.value = false;
}
</script>

<template>
  <div class="notifications-view">
    <div class="page-title">
      <h2>Notifications</h2>
      <p class="subtitle">Stay updated on your criminal empire's activities</p>
    </div>

    <div class="notifications-content">
      <div class="notifications-sidebar">
        <BaseCard class="sidebar-card">
          <div class="sidebar-stats">
            <div class="stat">
              <div class="stat-value">{{ totalNotifications }}</div>
              <div class="stat-label">Total</div>
            </div>
            <div class="stat">
              <div class="stat-value">{{ unreadNotifications }}</div>
              <div class="stat-label">Unread</div>
            </div>
          </div>

          <div class="filter-categories">
            <h4>Filter by Type</h4>

            <button
              class="category-btn"
              :class="{ active: activeFilter === 'all' }"
              @click="activeFilter = 'all'"
            >
              <span class="category-icon">📋</span>
              <span class="category-label">All Notifications</span>
              <span class="category-count">{{ totalNotifications }}</span>
            </button>

            <button
              class="category-btn"
              :class="{ active: activeFilter === 'territory' }"
              @click="activeFilter = 'territory'"
            >
              <span class="category-icon">🏙️</span>
              <span class="category-label">Territory</span>
              <span class="category-count">{{ territoryCounts }}</span>
            </button>

            <button
              class="category-btn"
              :class="{ active: activeFilter === 'operation' }"
              @click="activeFilter = 'operation'"
            >
              <span class="category-icon">🎯</span>
              <span class="category-label">Operations</span>
              <span class="category-count">{{ operationCounts }}</span>
            </button>

            <button
              class="category-btn"
              :class="{ active: activeFilter === 'collection' }"
              @click="activeFilter = 'collection'"
            >
              <span class="category-icon">💰</span>
              <span class="category-label">Collections</span>
              <span class="category-count">{{ collectionCounts }}</span>
            </button>

            <button
              class="category-btn"
              :class="{ active: activeFilter === 'heat' }"
              @click="activeFilter = 'heat'"
            >
              <span class="category-icon">🔥</span>
              <span class="category-label">Heat Alerts</span>
              <span class="category-count">{{ heatCounts }}</span>
            </button>

            <button
              class="category-btn"
              :class="{ active: activeFilter === 'system' }"
              @click="activeFilter = 'system'"
            >
              <span class="category-icon">🔔</span>
              <span class="category-label">System</span>
              <span class="category-count">{{ systemCounts }}</span>
            </button>
          </div>

          <div class="sidebar-actions">
            <BaseButton
              variant="secondary"
              class="mark-read-btn"
              :disabled="unreadNotifications === 0 || isMarkingRead"
              :loading="isMarkingRead"
              @click="markAllAsRead"
            >
              Mark All as Read
            </BaseButton>
          </div>
        </BaseCard>
      </div>

      <div class="notifications-main">
        <BaseCard class="notifications-card">
          <template #header>
            <div class="card-header-content">
              <h3>{{ activeFilter === 'all' ? 'All Notifications' : getTypeLabel(activeFilter as NotificationType) + ' Notifications' }}</h3>

              <div class="search-filter">
                <input
                  type="text"
                  v-model="searchQuery"
                  placeholder="Search notifications..."
                  class="search-input"
                />
                <button
                  v-if="searchQuery"
                  @click="searchQuery = ''"
                  class="clear-search"
                >
                  ✕
                </button>
              </div>
            </div>
          </template>

          <div v-if="isLoading" class="loading-state">
            <div class="loading-spinner"></div>
            <p>Loading notifications...</p>
          </div>

          <div v-else-if="filteredNotifications.length > 0" class="notifications-list">
            <div
              v-for="notification in filteredNotifications"
              :key="notification.id"
              class="notification-item"
              :class="{
                'unread': !notification.read,
                [getTypeClass(notification.type)]: true
              }"
            >
              <div class="notification-icon">
                {{ getNotificationIcon(notification.type) }}
              </div>

              <div class="notification-content">
                <div class="notification-header">
                  <div class="notification-type-badge" :class="getTypeClass(notification.type)">
                    {{ getTypeLabel(notification.type) }}
                  </div>
                  <div class="notification-time">
                    {{ formatTime(notification.timestamp) }}
                  </div>
                </div>

                <div class="notification-message">
                  {{ notification.message }}
                </div>
              </div>

              <div class="notification-actions">
                <button
                  v-if="!notification.read"
                  class="action-btn mark-read"
                  @click="markAsRead(notification.id)"
                  title="Mark as read"
                >
                  ✓
                </button>
              </div>
            </div>
          </div>

          <div v-else class="empty-state">
            <div class="empty-icon">🔍</div>
            <h4>No notifications found</h4>
            <p>{{ searchQuery ? 'Try adjusting your search or filters' : 'You have no notifications at this time' }}</p>

            <div v-if="searchQuery || activeFilter !== 'all'">
              <BaseButton variant="outline" @click="resetFilters">
                Reset Filters
              </BaseButton>
            </div>
          </div>
        </BaseCard>
      </div>
    </div>

    <BaseNotification
      v-if="showNotification"
      :type="notificationType"
      :message="notificationMessage"
      @close="closeNotification"
    />
  </div>
</template>

<style lang="scss">
.notifications-view {
  .page-title {
    margin-bottom: $spacing-xl;

    h2 {
      @include gold-accent;
      margin-bottom: $spacing-xs;
    }

    .subtitle {
      color: $text-secondary;
    }
  }

  .notifications-content {
    display: grid;
    grid-template-columns: 1fr;
    gap: $spacing-xl;

    @include respond-to(md) {
      grid-template-columns: 280px 1fr;
    }

    .notifications-sidebar {
      @include respond-to(md) {
        grid-column: 1;
        grid-row: 1;
      }

      .sidebar-card {
        height: 100%;
        display: flex;
        flex-direction: column;

        .sidebar-stats {
          display: flex;
          justify-content: space-around;
          padding-bottom: $spacing-md;
          border-bottom: 1px solid $border-color;

          .stat {
            text-align: center;

            .stat-value {
              font-size: $font-size-xl;
              font-weight: 600;
              @include gold-accent;
            }

            .stat-label {
              color: $text-secondary;
              font-size: $font-size-sm;
            }
          }
        }

        .filter-categories {
          padding: $spacing-md 0;
          flex: 1;

          h4 {
            margin: 0 0 $spacing-md 0;
            color: $text-secondary;
          }

          .category-btn {
            display: flex;
            align-items: center;
            width: 100%;
            background: none;
            border: none;
            padding: $spacing-sm;
            text-align: left;
            border-radius: $border-radius-sm;
            cursor: pointer;
            transition: $transition-base;
            color: $text-color;
            margin-bottom: $spacing-xs;

            &:hover {
              background-color: rgba(255, 255, 255, 0.05);
            }

            &.active {
              background-color: rgba($primary-color, 0.2);
            }

            .category-icon {
              margin-right: $spacing-sm;
              font-size: 18px;
            }

            .category-label {
              flex: 1;
            }

            .category-count {
              background-color: rgba(255, 255, 255, 0.1);
              padding: 2px 8px;
              border-radius: 12px;
              font-size: $font-size-sm;
            }
          }
        }

        .sidebar-actions {
          margin-top: auto;
          padding-top: $spacing-md;
          border-top: 1px solid $border-color;

          .mark-read-btn {
            width: 100%;
          }
        }
      }
    }

    .notifications-main {
      @include respond-to(md) {
        grid-column: 2;
        grid-row: 1;
      }

      .notifications-card {
        .card-header-content {
          @include flex-between;
          width: 100%;

          h3 {
            margin: 0;
          }

          .search-filter {
            position: relative;

            .search-input {
              background-color: rgba($background-lighter, 0.5);
              border: 1px solid $border-color;
              border-radius: $border-radius-sm;
              color: $text-color;
              padding: $spacing-xs $spacing-md;
              width: 250px;
              transition: $transition-base;

              &:focus {
                border-color: $secondary-color;
                outline: none;
                box-shadow: 0 0 0 2px rgba($secondary-color, 0.2);
              }

              &::placeholder {
                color: $text-secondary;
              }
            }

            .clear-search {
              position: absolute;
              right: $spacing-sm;
              top: 50%;
              transform: translateY(-50%);
              background: none;
              border: none;
              color: $text-secondary;
              cursor: pointer;
              font-size: 14px;

              &:hover {
                color: $text-color;
              }
            }
          }
        }

        .loading-state {
          @include flex-column;
          align-items: center;
          justify-content: center;
          padding: $spacing-xl;
          color: $text-secondary;

          .loading-spinner {
            width: 40px;
            height: 40px;
            border: 3px solid rgba(255, 255, 255, 0.1);
            border-radius: 50%;
            border-top-color: $secondary-color;
            animation: spin 1s ease-in-out infinite;
            margin-bottom: $spacing-md;
          }
        }

        .notifications-list {
          @include flex-column;
          gap: $spacing-md;

          .notification-item {
            display: flex;
            gap: $spacing-md;
            padding: $spacing-md;
            background-color: rgba($background-lighter, 0.5);
            border-radius: $border-radius-md;
            position: relative;
            border-left: 4px solid transparent;
            transition: $transition-base;

            &:hover {
              background-color: rgba($background-lighter, 0.8);
            }

            &.unread {
              background-color: rgba($primary-color, 0.1);

              &:hover {
                background-color: rgba($primary-color, 0.15);
              }
            }

            &.territory {
              border-left-color: $info-color;
            }

            &.operation {
              border-left-color: $warning-color;
            }

            &.collection {
              border-left-color: $secondary-color;
            }

            &.heat {
              border-left-color: $danger-color;
            }

            &.system {
              border-left-color: $text-secondary;
            }

            .notification-icon {
              font-size: 24px;
              flex-shrink: 0;
              width: 40px;
              height: 40px;
              @include flex-center;
            }

            .notification-content {
              flex: 1;

              .notification-header {
                @include flex-between;
                margin-bottom: $spacing-xs;

                .notification-type-badge {
                  font-size: $font-size-sm;
                  padding: 2px 8px;
                  border-radius: 12px;
                  font-weight: 500;

                  &.territory {
                    background-color: rgba($info-color, 0.2);
                    color: $info-color;
                  }

                  &.operation {
                    background-color: rgba($warning-color, 0.2);
                    color: $warning-color;
                  }

                  &.collection {
                    background-color: rgba($secondary-color, 0.2);
                    color: $secondary-color;
                  }

                  &.heat {
                    background-color: rgba($danger-color, 0.2);
                    color: $danger-color;
                  }

                  &.system {
                    background-color: rgba($text-secondary, 0.2);
                    color: $text-secondary;
                  }
                }

                .notification-time {
                  font-size: $font-size-sm;
                  color: $text-secondary;
                }
              }

              .notification-message {
                font-size: $font-size-md;
              }
            }

            .notification-actions {
              align-self: center;

              .action-btn {
                width: 32px;
                height: 32px;
                background-color: rgba(255, 255, 255, 0.1);
                border: none;
                border-radius: 50%;
                @include flex-center;
                color: $text-secondary;
                cursor: pointer;
                transition: $transition-base;

                &:hover {
                  background-color: rgba(255, 255, 255, 0.2);
                  color: $text-color;
                }

                &.mark-read:hover {
                  background-color: rgba($success-color, 0.2);
                  color: $success-color;
                }
              }
            }
          }
        }

        .empty-state {
          @include flex-column;
          align-items: center;
          justify-content: center;
          gap: $spacing-md;
          padding: $spacing-xl;
          text-align: center;
          color: $text-secondary;

          .empty-icon {
            font-size: 48px;
            margin-bottom: $spacing-sm;
          }

          h4 {
            margin: 0;
            color: $text-color;
          }
        }
      }
    }
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
