// src/components/layout/AppHeader.vue



<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { usePlayerStore } from '@/stores/modules/player';
import {
    Notification,
    NotificationType
} from '@/types/player';

const route = useRoute();
const router = useRouter();
const playerStore = usePlayerStore();

// UI state
const currentRoute = computed(() => route.path);
const showNotifications = ref(false);
const showProfileMenu = ref(false);
const activeNotifTab = ref('all');

// Check if user is logged in
const isLoggedIn = computed(() => {
    return !!localStorage.getItem('auth_token');
});

// Player info
const playerName = computed(() => playerStore.profile?.name || 'Boss');
const playerTitle = computed(() => playerStore.profile?.title || 'Capo');
const playerAvatar = ref('🤵');

// Navigation
const navItems = [
    { name: 'Home', path: '/' },
    { name: 'Territory', path: '/territory' },
    { name: 'Operations', path: '/operations' },
    { name: 'Market', path: '/market' },
    { name: 'Rankings', path: '/rankings' },
    { name: 'NFT', path: '/nft' }
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
    const profileMenu = document.querySelector('.profile-link');

    if (notifBell && !notifBell.contains(event.target as Node)) {
        showNotifications.value = false;
    }

    if (profileMenu && !profileMenu.contains(event.target as Node)) {
        showProfileMenu.value = false;
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
    showProfileMenu.value = false;
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
    const diff = Math.floor((now.getTime() - new Date(date).getTime()) / 60000); // difference in minutes

    if (diff < 1) return 'Just now';
    if (diff < 60) return `${diff} min ago`;
    if (diff < 1440) return `${Math.floor(diff / 60)} hours ago`;
    return `${Math.floor(diff / 1440)} days ago`;
}

// Action functions
function toggleNotifications(event: MouseEvent) {
    event.stopPropagation();
    showNotifications.value = !showNotifications.value;

    // Close profile menu if open
    if (showNotifications.value) {
        showProfileMenu.value = false;
    }
}

function toggleProfileMenu(event: MouseEvent) {
    event.stopPropagation();
    showProfileMenu.value = !showProfileMenu.value;

    // Close notifications if open
    if (showProfileMenu.value) {
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

    // Redirect to login page
    router.push('/login');
}
</script>

<template>
    <header class="app-header">
        <div class="logo">
            <router-link to="/">
                <h1>Mafia Wars: <span class="gold-text">Criminal Empire</span></h1>
            </router-link>
        </div>

        <nav class="main-nav">
            <router-link v-for="item in navItems" :key="item.path" :to="item.path" class="nav-item"
                :class="{ active: currentRoute === item.path }">
                {{ item.name }}
            </router-link>
        </nav>

        <div class="user-controls">
            <div class="user-menu" v-if="isLoggedIn">
                <div class="profile-link" @click="toggleProfileMenu">
                    <div class="user-avatar">{{ playerAvatar }}</div>
                    <span class="user-name">{{ playerName }}</span>
                    <span class="menu-toggle" :class="{ open: showProfileMenu }">▼</span>

                    <div v-if="showProfileMenu" class="profile-dropdown">
                        <div class="dropdown-header">
                            <div class="user-info">
                                <div class="user-avatar large">{{ playerAvatar }}</div>
                                <div class="user-details">
                                    <div class="user-name">{{ playerName }}</div>
                                    <div class="user-title">{{ playerTitle }}</div>
                                </div>
                            </div>
                        </div>
                        <div class="dropdown-menu">
                            <router-link to="/profile" class="menu-item" @click="showProfileMenu = false">
                                <span class="item-icon">👤</span>
                                <span>My Profile</span>
                            </router-link>
                            <router-link to="/settings" class="menu-item" @click="showProfileMenu = false">
                                <span class="item-icon">⚙️</span>
                                <span>Settings</span>
                            </router-link>
                            <div class="divider"></div>
                            <div @click="logout" class="menu-item">
                                <span class="item-icon">🚪</span>
                                <span>Sign Out</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="notification-bell" @click="toggleNotifications">
                <div class="bell-icon">
                    <i class="notification-icon">🔔</i>
                    <span v-if="unreadNotifications" class="notification-badge">{{ unreadNotifications }}</span>
                </div>

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
                            Operations
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
                                'system': notification.type === 'system'
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
                            View All Notifications
                        </router-link>
                    </div>
                </div>
            </div>

            <div class="action-buttons" v-if="!isLoggedIn">
                <router-link to="/login" class="login-btn">
                    Sign In
                </router-link>
                <router-link to="/register" class="register-btn">
                    Sign Up
                </router-link>
            </div>
        </div>
    </header>
</template>

<style lang="scss">
.app-header {
    @include flex-between;
    background-color: $background-lighter;
    padding: $spacing-md $spacing-lg;
    box-shadow: $shadow-md;
    position: sticky;
    top: 0;
    z-index: 100;
    border-bottom: 1px solid $border-color;

    .logo {
        h1 {
            font-size: 24px;
            margin: 0;

            @include respond-to(md) {
                font-size: 28px;
            }
        }

        a {
            text-decoration: none;
            color: $text-color;
        }
    }

    .main-nav {
        display: flex;
        gap: $spacing-md;

        @include respond-to(xs) {
            display: none;
        }

        .nav-item {
            color: $text-secondary;
            text-decoration: none;
            padding: $spacing-sm;
            font-weight: 500;
            transition: $transition-base;
            position: relative;
            text-transform: uppercase;
            font-size: $font-size-sm;
            letter-spacing: 1px;

            &:hover,
            &.active {
                color: $secondary-color;
            }

            &.active:after {
                content: '';
                position: absolute;
                bottom: -4px;
                left: 0;
                width: 100%;
                height: 2px;
                background-color: $secondary-color;
                box-shadow: 0 0 8px rgba($secondary-color, 0.5);
            }
        }
    }

    .user-controls {
        display: flex;
        align-items: center;
        gap: $spacing-md;

        .user-menu {
            position: relative;

            .profile-link {
                display: flex;
                align-items: center;
                gap: $spacing-sm;
                padding: $spacing-sm;
                border-radius: $border-radius-sm;
                cursor: pointer;
                transition: $transition-base;

                &:hover {
                    background-color: rgba(255, 255, 255, 0.05);
                }

                .user-avatar {
                    width: 32px;
                    height: 32px;
                    background-color: $primary-color;
                    border-radius: 50%;
                    @include flex-center;
                    font-size: 16px;

                    &.large {
                        width: 48px;
                        height: 48px;
                        font-size: 24px;
                    }
                }

                .user-name {
                    @include respond-to(xs) {
                        display: none;
                    }
                }

                .menu-toggle {
                    font-size: 10px;
                    color: $text-secondary;
                    transition: $transition-base;

                    &.open {
                        transform: rotate(180deg);
                    }
                }
            }

            .profile-dropdown {
                position: absolute;
                top: 100%;
                right: 0;
                width: 250px;
                background-color: $background-card;
                border-radius: $border-radius-md;
                box-shadow: $shadow-lg;
                border: 1px solid $border-color;
                margin-top: $spacing-sm;
                overflow: hidden;
                z-index: $z-index-dropdown;
                animation: fadeIn 0.2s ease;

                &:before {
                    content: '';
                    position: absolute;
                    top: -5px;
                    right: 24px;
                    width: 10px;
                    height: 10px;
                    background-color: $background-card;
                    transform: rotate(45deg);
                    border-top: 1px solid $border-color;
                    border-left: 1px solid $border-color;
                }

                .dropdown-header {
                    padding: $spacing-md;
                    border-bottom: 1px solid $border-color;

                    .user-info {
                        display: flex;
                        gap: $spacing-md;

                        .user-details {
                            .user-name {
                                font-weight: 600;
                                margin-bottom: 2px;
                            }

                            .user-title {
                                font-size: $font-size-sm;
                                color: $text-secondary;
                            }
                        }
                    }
                }

                .dropdown-menu {
                    padding: $spacing-sm 0;

                    .menu-item {
                        display: flex;
                        align-items: center;
                        gap: $spacing-md;
                        padding: $spacing-sm $spacing-md;
                        color: $text-color;
                        text-decoration: none;
                        transition: $transition-base;
                        cursor: pointer;

                        &:hover {
                            background-color: rgba(255, 255, 255, 0.05);
                        }

                        .item-icon {
                            font-size: 18px;
                        }
                    }

                    .divider {
                        height: 1px;
                        background-color: $border-color;
                        margin: $spacing-xs 0;
                    }
                }
            }
        }

        .notification-bell {
            position: relative;
            cursor: pointer;

            .bell-icon {
                padding: $spacing-sm;
                border-radius: $border-radius-sm;
                transition: $transition-base;
                position: relative;

                &:hover {
                    background-color: rgba(255, 255, 255, 0.05);
                }

                .notification-icon {
                    font-size: 24px;
                    color: $text-secondary;
                    transition: $transition-base;
                }

                .notification-badge {
                    position: absolute;
                    top: 4px;
                    right: 4px;
                    width: 18px;
                    height: 18px;
                    background-color: $danger-color;
                    color: white;
                    border-radius: 50%;
                    font-size: 12px;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-weight: bold;
                    box-shadow: 0 0 5px rgba(0, 0, 0, 0.3);
                }
            }

            .notifications-dropdown {
                position: absolute;
                top: 100%;
                right: 0;
                width: 320px;
                max-width: 90vw;
                max-height: 80vh;
                display: flex;
                flex-direction: column;
                background-color: $background-card;
                border-radius: $border-radius-md;
                box-shadow: $shadow-lg;
                border: 1px solid $border-color;
                margin-top: $spacing-sm;
                z-index: $z-index-dropdown;
                animation: fadeIn 0.2s ease;
                overflow: hidden;

                &:before {
                    content: '';
                    position: absolute;
                    top: -5px;
                    right: 16px;
                    width: 10px;
                    height: 10px;
                    background-color: $background-card;
                    transform: rotate(45deg);
                    border-top: 1px solid $border-color;
                    border-left: 1px solid $border-color;
                }

                .notifications-header {
                    @include flex-between;
                    padding: $spacing-md;
                    border-bottom: 1px solid $border-color;

                    h4 {
                        margin: 0;
                    }

                    .mark-read-btn {
                        background: none;
                        border: none;
                        color: $secondary-color;
                        cursor: pointer;
                        font-size: $font-size-sm;

                        &:hover {
                            text-decoration: underline;
                        }
                    }
                }

                .notifications-tabs {
                    display: flex;
                    border-bottom: 1px solid $border-color;

                    .tab-btn {
                        flex: 1;
                        background: none;
                        border: none;
                        padding: $spacing-sm;
                        color: $text-secondary;
                        font-size: $font-size-sm;
                        font-weight: 500;
                        cursor: pointer;
                        transition: $transition-base;

                        &:hover {
                            background-color: rgba(255, 255, 255, 0.05);
                        }

                        &.active {
                            color: $secondary-color;
                            box-shadow: inset 0 -2px 0 $secondary-color;
                        }
                    }
                }

                .notifications-list {
                    flex: 1;
                    overflow-y: auto;
                    max-height: 400px;
                    padding: $spacing-xs;

                    .notification-item {
                        display: flex;
                        gap: $spacing-sm;
                        padding: $spacing-sm;
                        border-radius: $border-radius-sm;
                        border-left: 3px solid transparent;
                        transition: $transition-base;
                        position: relative;

                        &:not(:last-child) {
                            margin-bottom: $spacing-xs;
                        }

                        &:hover {
                            background-color: rgba(255, 255, 255, 0.05);
                        }

                        &.unread {
                            background-color: rgba($primary-color, 0.1);

                            &:hover {
                                background-color: rgba($primary-color, 0.15);
                            }

                            &:before {
                                content: '';
                                position: absolute;
                                top: $spacing-sm;
                                left: -8px;
                                width: 8px;
                                height: 8px;
                                background-color: $primary-color;
                                border-radius: 50%;
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
                            width: 32px;
                            height: 32px;
                            @include flex-center;
                        }

                        .notification-content {
                            flex: 1;

                            p {
                                margin: 0 0 $spacing-xs 0;
                                font-size: $font-size-sm;
                            }

                            .notification-time {
                                font-size: 11px;
                                color: $text-secondary;
                                display: block;
                            }
                        }

                        .notification-actions {
                            align-self: flex-start;

                            .action-btn {
                                width: 20px;
                                height: 20px;
                                background-color: rgba(255, 255, 255, 0.1);
                                border: none;
                                border-radius: 50%;
                                @include flex-center;
                                font-size: 12px;
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

                .empty-notifications {
                    padding: $spacing-xl;
                    text-align: center;
                    color: $text-secondary;

                    .empty-icon {
                        font-size: 36px;
                        margin-bottom: $spacing-md;
                        opacity: 0.5;
                    }

                    p {
                        margin: 0;
                    }
                }

                .notifications-footer {
                    padding: $spacing-sm;
                    border-top: 1px solid $border-color;
                    text-align: center;

                    .view-all-link {
                        color: $secondary-color;
                        font-size: $font-size-sm;
                        text-decoration: none;

                        &:hover {
                            text-decoration: underline;
                        }
                    }
                }
            }
        }

        .action-buttons {
            display: flex;
            gap: $spacing-sm;

            .login-btn,
            .register-btn {
                padding: $spacing-xs $spacing-md;
                border-radius: $border-radius-sm;
                font-size: $font-size-sm;
                font-weight: 600;
                text-decoration: none;
                text-transform: uppercase;
                letter-spacing: 0.5px;
                transition: $transition-base;
            }

            .login-btn {
                color: $secondary-color;
                border: 1px solid $secondary-color;

                &:hover {
                    background-color: rgba($secondary-color, 0.1);
                }
            }

            .register-btn {
                background-color: $secondary-color;
                color: $background-color;

                &:hover {
                    background-color: lighten($secondary-color, 5%);
                }
            }
        }
    }
}

@keyframes fadeIn {
    from {
        opacity: 0;
    }

    to {
        opacity: 1;
    }
}
</style>