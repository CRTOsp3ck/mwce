// src/views/ProfileView.vue

<template>
    <div class="profile-view">
        <div class="page-title">
            <h2>Criminal Profile</h2>
            <p class="subtitle">View and manage your criminal empire's details</p>
        </div>

        <div class="profile-content">
            <div class="profile-section">
                <BaseCard class="profile-card">
                    <div class="profile-header">
                        <div class="profile-avatar">
                            <div class="avatar-wrapper">
                                <span class="avatar-icon">{{ playerAvatar || '🤵' }}</span>
                            </div>
                            <button class="change-avatar-btn" @click="openAvatarModal">
                                <span class="btn-icon">📷</span>
                            </button>
                        </div>
                        <div class="profile-info">
                            <h3>{{ playerName }}</h3>
                            <div class="profile-title">{{ playerTitle }}</div>
                            <div class="profile-stats">
                                <div class="stat-badge respect">Respect: {{ playerRespect }}%</div>
                                <div class="stat-badge influence">Influence: {{ playerInfluence }}%</div>
                                <div class="stat-badge heat">Heat: {{ playerHeat }}%</div>
                            </div>
                        </div>
                    </div>

                    <div class="profile-details">
                        <div class="detail-item">
                            <div class="detail-label">Account Created</div>
                            <div class="detail-value">{{ formatDate(accountCreated) }}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Last Active</div>
                            <div class="detail-value">{{ formatDate(lastActive) }}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Main Territory</div>
                            <div class="detail-value">{{ mainTerritory }}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Email</div>
                            <div class="detail-value">{{ maskedEmail }}</div>
                        </div>
                    </div>
                </BaseCard>
            </div>

            <div class="stats-section">
                <BaseCard class="stats-card">
                    <template #header>
                        <h3>Criminal Stats</h3>
                    </template>

                    <div class="stats-grid">
                        <div class="stat-card">
                            <div class="stat-icon">💰</div>
                            <div class="stat-label">Total Money Earned</div>
                            <div class="stat-value money">${{ formatNumber(totalMoneyEarned) }}</div>
                        </div>

                        <div class="stat-card">
                            <div class="stat-icon">🏢</div>
                            <div class="stat-label">Total Hotspots Controlled</div>
                            <div class="stat-value">{{ totalHotspotsControlled }}</div>
                        </div>

                        <div class="stat-card">
                            <div class="stat-icon">🏆</div>
                            <div class="stat-label">Operations Completed</div>
                            <div class="stat-value">{{ totalOperationsCompleted }}</div>
                        </div>

                        <div class="stat-card">
                            <div class="stat-icon">🥇</div>
                            <div class="stat-label">Max Respect Achieved</div>
                            <div class="stat-value respect">{{ maxRespectAchieved }}%</div>
                        </div>

                        <div class="stat-card">
                            <div class="stat-icon">🗣️</div>
                            <div class="stat-label">Max Influence Achieved</div>
                            <div class="stat-value influence">{{ maxInfluenceAchieved }}%</div>
                        </div>

                        <div class="stat-card">
                            <div class="stat-icon">🤝</div>
                            <div class="stat-label">Successful Takeovers</div>
                            <div class="stat-value success">{{ successfulTakeovers }}</div>
                        </div>
                    </div>
                </BaseCard>
            </div>

            <div class="achievements-section">
                <BaseCard class="achievements-card">
                    <template #header>
                        <div class="section-header">
                            <h3>Achievements</h3>
                            <div class="achievement-progress">
                                {{ unlockedAchievements.length }}/{{ achievements.length }} Unlocked
                            </div>
                        </div>
                    </template>

                    <div class="achievements-grid">
                        <div v-for="achievement in achievements" :key="achievement.id" class="achievement-item"
                            :class="{ 'unlocked': isAchievementUnlocked(achievement.id) }"
                            @click="showAchievementDetails(achievement)">
                            <div class="achievement-icon">{{ achievement.icon }}</div>
                            <div class="achievement-info">
                                <div class="achievement-name">{{ achievement.name }}</div>
                                <div class="achievement-description">{{ achievement.description }}</div>
                            </div>
                            <div class="achievement-status">
                                <div v-if="isAchievementUnlocked(achievement.id)" class="status-icon unlocked">✓</div>
                                <div v-else class="status-icon locked">🔒</div>
                            </div>
                        </div>
                    </div>
                </BaseCard>
            </div>

            <div class="actions-section">
                <BaseCard class="actions-card">
                    <template #header>
                        <h3>Account Actions</h3>
                    </template>

                    <div class="account-actions">
                        <BaseButton variant="primary" @click="navigateToSettings">
                            Account Settings
                        </BaseButton>
                        <BaseButton variant="outline" @click="showChangePasswordModal">
                            Change Password
                        </BaseButton>
                        <BaseButton variant="danger" @click="showLogoutConfirmation">
                            Sign Out
                        </BaseButton>
                    </div>
                </BaseCard>
            </div>
        </div>

        <!-- Avatar Selection Modal -->
        <BaseModal v-model="showAvatarModal" title="Change Avatar">
            <div class="avatar-selection-modal">
                <p>Select your new criminal avatar:</p>
                <div class="avatar-grid">
                    <div v-for="(avatar, index) in avatarOptions" :key="index" class="avatar-option"
                        :class="{ selected: selectedAvatar === avatar }" @click="selectAvatar(avatar)">
                        <div class="avatar-icon">{{ avatar }}</div>
                    </div>
                </div>
            </div>
            <template #footer>
                <div class="modal-actions">
                    <BaseButton variant="text" @click="closeAvatarModal">
                        Cancel
                    </BaseButton>
                    <BaseButton variant="secondary" :disabled="!selectedAvatar || isUpdatingAvatar"
                        :loading="isUpdatingAvatar" @click="updateAvatar">
                        Save
                    </BaseButton>
                </div>
            </template>
        </BaseModal>

        <!-- Achievement Details Modal -->
        <BaseModal v-model="showAchievementModal" :title="selectedAchievement?.name || 'Achievement'">
            <div class="achievement-details-modal">
                <div class="achievement-header">
                    <div class="achievement-large-icon">{{ selectedAchievement?.icon }}</div>
                    <div class="achievement-status-label" :class="{
                        'unlocked': selectedAchievement && isAchievementUnlocked(selectedAchievement.id)
                    }">
                        {{ selectedAchievement && isAchievementUnlocked(selectedAchievement.id)
                            ? 'Unlocked' : 'Locked' }}
                    </div>
                </div>
                <div class="achievement-details-content">
                    <p>{{ selectedAchievement?.description }}</p>
                    <div class="achievement-criteria">
                        <h4>Requirements:</h4>
                        <ul>
                            <li v-for="(criteria, index) in selectedAchievement?.criteria" :key="index">
                                {{ criteria }}
                            </li>
                        </ul>
                    </div>
                    <div class="achievement-rewards" v-if="selectedAchievement?.rewards">
                        <h4>Rewards:</h4>
                        <ul>
                            <li v-for="(reward, index) in selectedAchievement?.rewards" :key="index">
                                {{ reward }}
                            </li>
                        </ul>
                    </div>
                    <div class="achievement-date"
                        v-if="selectedAchievement && isAchievementUnlocked(selectedAchievement.id)">
                        <p>Unlocked on: {{ formatDate(getAchievementUnlockDate(selectedAchievement.id)) }}</p>
                    </div>
                </div>
            </div>
            <template #footer>
                <BaseButton @click="closeAchievementModal">
                    Close
                </BaseButton>
            </template>
        </BaseModal>

        <!-- Change Password Modal -->
        <BaseModal v-model="showPasswordModal" title="Change Password">
            <div class="change-password-modal">
                <div class="form-group" :class="{ 'has-error': passwordErrors.currentPassword }">
                    <label for="current-password">Current Password</label>
                    <input type="password" id="current-password" v-model="passwordData.currentPassword"
                        placeholder="Enter your current password" />
                    <div class="error-message" v-if="passwordErrors.currentPassword">
                        {{ passwordErrors.currentPassword }}
                    </div>
                </div>

                <div class="form-group" :class="{ 'has-error': passwordErrors.newPassword }">
                    <label for="new-password">New Password</label>
                    <input type="password" id="new-password" v-model="passwordData.newPassword"
                        placeholder="Enter new password" />
                    <div class="error-message" v-if="passwordErrors.newPassword">
                        {{ passwordErrors.newPassword }}
                    </div>
                </div>

                <div class="form-group" :class="{ 'has-error': passwordErrors.confirmPassword }">
                    <label for="confirm-password">Confirm New Password</label>
                    <input type="password" id="confirm-password" v-model="passwordData.confirmPassword"
                        placeholder="Confirm new password" />
                    <div class="error-message" v-if="passwordErrors.confirmPassword">
                        {{ passwordErrors.confirmPassword }}
                    </div>
                </div>
            </div>
            <template #footer>
                <div class="modal-actions">
                    <BaseButton variant="text" @click="closePasswordModal">
                        Cancel
                    </BaseButton>
                    <BaseButton variant="primary" :disabled="!canChangePassword || isChangingPassword"
                        :loading="isChangingPassword" @click="changePassword">
                        Update Password
                    </BaseButton>
                </div>
            </template>
        </BaseModal>

        <!-- Logout Confirmation Modal -->
        <BaseModal v-model="showLogoutModal" title="Sign Out">
            <div class="logout-confirmation">
                <p>Are you sure you want to sign out from your criminal empire?</p>
                <p class="warning">Your progress is saved, but any ongoing operations will continue without supervision.
                </p>
            </div>
            <template #footer>
                <div class="modal-actions">
                    <BaseButton variant="text" @click="closeLogoutModal">
                        Cancel
                    </BaseButton>
                    <BaseButton variant="danger" :loading="isLoggingOut" @click="logout">
                        Sign Out
                    </BaseButton>
                </div>
            </template>
        </BaseModal>

        <BaseNotification v-if="showNotification" :type="notificationType" :message="notificationMessage"
            @close="closeNotification" />
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import BaseNotification from '@/components/ui/BaseNotification.vue';
import { usePlayerStore } from '@/stores/modules/player';

const router = useRouter();
const playerStore = usePlayerStore();

// Player data
const playerName = computed(() => playerStore.profile?.name || 'Don Corleone');
const playerTitle = computed(() => playerStore.profile?.title || 'Capo');
const playerAvatar = ref('🤵'); // This would come from the player profile
const playerRespect = computed(() => playerStore.playerRespect);
const playerInfluence = computed(() => playerStore.playerInfluence);
const playerHeat = computed(() => playerStore.playerHeat);
const accountCreated = computed(() => playerStore.profile?.createdAt || new Date().toISOString());
const lastActive = computed(() => playerStore.profile?.lastActive || new Date().toISOString());
const mainTerritory = ref('Downtown');

// Player statistics
const totalMoneyEarned = ref(1250000);
const totalHotspotsControlled = ref(15);
const totalOperationsCompleted = ref(47);
const maxRespectAchieved = ref(85);
const maxInfluenceAchieved = ref(72);
const successfulTakeovers = ref(32);

// Mock email
const playerEmail = ref('don.corleone@criminal.emp');
const maskedEmail = computed(() => {
    if (!playerEmail.value) return '';
    const [username, domain] = playerEmail.value.split('@');
    return `${username.substring(0, 3)}****@${domain}`;
});

// Modals state
const showAvatarModal = ref(false);
const showAchievementModal = ref(false);
const showPasswordModal = ref(false);
const showLogoutModal = ref(false);

// Avatar selection
const selectedAvatar = ref('');
const isUpdatingAvatar = ref(false);
const avatarOptions = ['🤵', '🕵️', '👨‍💼', '👩‍💼', '👨‍🦱', '👩‍🦱', '👨‍🦰', '👩‍🦰', '👨‍🦲', '👩‍🦲', '👴', '👵'];

// Change password form
const passwordData = ref({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
});

const passwordErrors = ref({
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
});

const isChangingPassword = ref(false);
const isLoggingOut = ref(false);

// Notification state
const showNotification = ref(false);
const notificationType = ref('info');
const notificationMessage = ref('');

// Mock achievements data
const achievements = [
    {
        id: 'first_blood',
        name: 'First Blood',
        icon: '🩸',
        description: 'Complete your first takeover',
        criteria: ['Successfully take over a business'],
        rewards: ['+5 Respect']
    },
    {
        id: 'money_maker',
        name: 'Money Maker',
        icon: '💰',
        description: 'Earn your first $100,000',
        criteria: ['Accumulate $100,000 in earnings'],
        rewards: ['+3 Influence', '+$5,000 bonus']
    },
    {
        id: 'crew_master',
        name: 'Crew Master',
        icon: '👥',
        description: 'Recruit a crew of 10 members',
        criteria: ['Have 10 crew members at once'],
        rewards: ['+10% crew efficiency']
    },
    {
        id: 'territory_boss',
        name: 'Territory Boss',
        icon: '🏙️',
        description: 'Control 5 businesses in the same district',
        criteria: ['Control 5 businesses in a single district'],
        rewards: ['+10 Respect', '+5 Influence']
    },
    {
        id: 'mastermind',
        name: 'Criminal Mastermind',
        icon: '🧠',
        description: 'Complete 10 special operations',
        criteria: ['Successfully complete 10 special operations'],
        rewards: ['+15 Respect', '+10 Influence', 'Unlocks "Godfather" title']
    },
    {
        id: 'heat_resistant',
        name: 'Heat Resistant',
        icon: '🔥',
        description: 'Survive with maximum heat for 24 hours',
        criteria: ['Maintain 100% heat for 24 continuous hours'],
        rewards: ['+15% heat resistance']
    }
];

// Mock unlocked achievements
const unlockedAchievements = ref(['first_blood', 'money_maker']);
const selectedAchievement = ref(null);

// Computed properties
const canChangePassword = computed(() => {
    return passwordData.value.currentPassword !== '' &&
        passwordData.value.newPassword !== '' &&
        passwordData.value.confirmPassword !== '' &&
        passwordData.value.newPassword === passwordData.value.confirmPassword &&
        passwordData.value.newPassword.length >= 6;
});

// Load data when component is mounted
onMounted(async () => {
    if (!playerStore.profile) {
        await playerStore.fetchProfile();
    }

    if (!playerStore.stats) {
        await playerStore.fetchStats();
    }
});

// Helper functions
function formatNumber(value: number): string {
    if (value >= 1000000) {
        return (value / 1000000).toFixed(1) + 'M';
    } else if (value >= 1000) {
        return (value / 1000).toFixed(1) + 'K';
    }
    return value.toString();
}

function formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });
}

function isAchievementUnlocked(achievementId: string): boolean {
    return unlockedAchievements.value.includes(achievementId);
}

function getAchievementUnlockDate(achievementId: string): string {
    // In a real app, this would come from the backend
    return new Date().toISOString();
}

// Modal functions
function openAvatarModal() {
    selectedAvatar.value = playerAvatar.value;
    showAvatarModal.value = true;
}

function closeAvatarModal() {
    showAvatarModal.value = false;
    selectedAvatar.value = '';
}

function selectAvatar(avatar: string) {
    selectedAvatar.value = avatar;
}

async function updateAvatar() {
    if (!selectedAvatar.value) return;

    isUpdatingAvatar.value = true;

    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Update avatar
        playerAvatar.value = selectedAvatar.value;

        // Show success notification
        showNotificationMessage('success', 'Avatar updated successfully');

        // Close modal
        closeAvatarModal();
    } catch (error) {
        showNotificationMessage('danger', 'Failed to update avatar');
    } finally {
        isUpdatingAvatar.value = false;
    }
}

function showAchievementDetails(achievement) {
    selectedAchievement.value = achievement;
    showAchievementModal.value = true;
}

function closeAchievementModal() {
    showAchievementModal.value = false;
    selectedAchievement.value = null;
}

function navigateToSettings() {
    router.push('/settings');
}

function showChangePasswordModal() {
    // Reset form
    passwordData.value = {
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
    };
    passwordErrors.value = {
        currentPassword: '',
        newPassword: '',
        confirmPassword: ''
    };

    showPasswordModal.value = true;
}

function closePasswordModal() {
    showPasswordModal.value = false;
}

async function changePassword() {
    if (!canChangePassword.value) return;

    isChangingPassword.value = true;

    try {
        // Validate current password (would be an API call in a real app)
        if (passwordData.value.currentPassword !== 'password123') {
            passwordErrors.value.currentPassword = 'Current password is incorrect';
            isChangingPassword.value = false;
            return;
        }

        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1500));

        // Show success notification
        showNotificationMessage('success', 'Password updated successfully');

        // Close modal
        closePasswordModal();
    } catch (error) {
        showNotificationMessage('danger', 'Failed to update password');
    } finally {
        isChangingPassword.value = false;
    }
}

function showLogoutConfirmation() {
    showLogoutModal.value = true;
}

function closeLogoutModal() {
    showLogoutModal.value = false;
}

async function logout() {
    isLoggingOut.value = true;

    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Clear auth token
        localStorage.removeItem('auth_token');

        // Redirect to login page
        router.push('/login');
    } catch (error) {
        showNotificationMessage('danger', 'Failed to log out');
        isLoggingOut.value = false;
    }
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

<style lang="scss">
.profile-view {
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

    .profile-content {
        display: grid;
        grid-template-columns: 1fr;
        gap: $spacing-xl;

        @include respond-to(lg) {
            grid-template-columns: 1fr 2fr;
        }

        .profile-section {
            @include respond-to(lg) {
                grid-column: 1;
                grid-row: 1 / span 2;
            }

            .profile-card {
                height: 100%;

                .profile-header {
                    display: flex;
                    gap: $spacing-lg;
                    margin-bottom: $spacing-xl;

                    @include respond-to(md) {
                        flex-direction: row;
                    }

                    @include respond-to(xs) {
                        flex-direction: column;
                        align-items: center;
                        text-align: center;
                    }

                    .profile-avatar {
                        position: relative;

                        .avatar-wrapper {
                            width: 100px;
                            height: 100px;
                            background-color: $background-lighter;
                            border: 3px solid $secondary-color;
                            border-radius: 50%;
                            display: flex;
                            align-items: center;
                            justify-content: center;
                            box-shadow: 0 0 15px rgba($secondary-color, 0.4);

                            .avatar-icon {
                                font-size: 48px;
                            }
                        }

                        .change-avatar-btn {
                            position: absolute;
                            bottom: 0;
                            right: 0;
                            width: 32px;
                            height: 32px;
                            background-color: $background-lighter;
                            border: 2px solid $secondary-color;
                            border-radius: 50%;
                            display: flex;
                            align-items: center;
                            justify-content: center;
                            cursor: pointer;
                            transition: $transition-base;

                            &:hover {
                                background-color: $secondary-color;

                                .btn-icon {
                                    color: $background-color;
                                }
                            }

                            .btn-icon {
                                font-size: 16px;
                                color: $secondary-color;
                                transition: $transition-base;
                            }
                        }
                    }

                    .profile-info {
                        flex: 1;

                        h3 {
                            margin: 0 0 $spacing-xs 0;
                            @include gold-accent;
                            font-size: $font-size-xl;
                        }

                        .profile-title {
                            font-size: $font-size-lg;
                            margin-bottom: $spacing-md;
                            color: $text-secondary;
                        }

                        .profile-stats {
                            display: flex;
                            flex-wrap: wrap;
                            gap: $spacing-sm;

                            .stat-badge {
                                padding: $spacing-xs $spacing-sm;
                                border-radius: $border-radius-sm;
                                font-size: $font-size-sm;
                                font-weight: 600;

                                &.respect {
                                    background-color: rgba($success-color, 0.2);
                                    color: $success-color;
                                    border: 1px solid rgba($success-color, 0.3);
                                }

                                &.influence {
                                    background-color: rgba($info-color, 0.2);
                                    color: $info-color;
                                    border: 1px solid rgba($info-color, 0.3);
                                }

                                &.heat {
                                    background-color: rgba($danger-color, 0.2);
                                    color: $danger-color;
                                    border: 1px solid rgba($danger-color, 0.3);
                                }
                            }
                        }
                    }
                }

                .profile-details {
                    .detail-item {
                        display: flex;
                        justify-content: space-between;
                        padding: $spacing-md 0;
                        border-bottom: 1px solid $border-color;

                        &:last-child {
                            border-bottom: none;
                        }

                        .detail-label {
                            color: $text-secondary;
                        }

                        .detail-value {
                            font-weight: 500;
                        }
                    }
                }
            }
        }

        .stats-section {
            @include respond-to(lg) {
                grid-column: 2;
                grid-row: 1;
            }

            .stats-grid {
                display: grid;
                grid-template-columns: repeat(1, 1fr);
                gap: $spacing-md;

                @include respond-to(sm) {
                    grid-template-columns: repeat(2, 1fr);
                }

                @include respond-to(md) {
                    grid-template-columns: repeat(3, 1fr);
                }

                .stat-card {
                    @include flex-column;
                    align-items: center;
                    justify-content: center;
                    text-align: center;
                    background-color: rgba($background-lighter, 0.5);
                    border-radius: $border-radius-md;
                    padding: $spacing-md;

                    .stat-icon {
                        font-size: 32px;
                        margin-bottom: $spacing-sm;
                    }

                    .stat-label {
                        color: $text-secondary;
                        font-size: $font-size-sm;
                        margin-bottom: $spacing-xs;
                    }

                    .stat-value {
                        font-size: $font-size-lg;
                        font-weight: 600;

                        &.money {
                            @include gold-accent;
                        }

                        &.respect {
                            color: $success-color;
                        }

                        &.influence {
                            color: $info-color;
                        }

                        &.success {
                            color: $success-color;
                        }
                    }
                }
            }
        }

        .achievements-section {
            @include respond-to(lg) {
                grid-column: 2;
                grid-row: 2;
            }

            .section-header {
                @include flex-between;
                width: 100%;

                h3 {
                    margin: 0;
                }

                .achievement-progress {
                    font-size: $font-size-sm;
                    color: $secondary-color;
                    font-weight: 600;
                }
            }

            .achievements-grid {
                @include flex-column;
                gap: $spacing-sm;

                .achievement-item {
                    display: flex;
                    align-items: center;
                    gap: $spacing-md;
                    background-color: rgba($background-lighter, 0.5);
                    border-radius: $border-radius-md;
                    padding: $spacing-md;
                    cursor: pointer;
                    transition: $transition-base;

                    &:hover {
                        background-color: rgba($background-lighter, 0.8);
                    }

                    &.unlocked {
                        border-left: 3px solid $success-color;
                    }

                    .achievement-icon {
                        font-size: 32px;
                        opacity: 0.7;

                        .unlocked & {
                            opacity: 1;
                        }
                    }

                    .achievement-info {
                        flex: 1;

                        .achievement-name {
                            font-weight: 600;
                            margin-bottom: $spacing-xs;
                            color: $text-color;

                            .unlocked & {
                                color: $text-color;
                            }
                        }

                        .achievement-description {
                            font-size: $font-size-sm;
                            color: $text-secondary;
                        }
                    }

                    .achievement-status {
                        .status-icon {
                            width: 24px;
                            height: 24px;
                            @include flex-center;
                            border-radius: 50%;

                            &.unlocked {
                                background-color: $success-color;
                                color: white;
                            }

                            &.locked {
                                background-color: $text-secondary;
                                color: $background-color;
                            }
                        }
                    }
                }
            }
        }

        .actions-section {
            @include respond-to(lg) {
                grid-column: 1 / -1;
                grid-row: 3;
            }

            .account-actions {
                display: flex;
                flex-wrap: wrap;
                gap: $spacing-md;
                justify-content: center;

                @include respond-to(md) {
                    justify-content: flex-start;
                }
            }
        }
    }

    // Modal styles
    .avatar-selection-modal {
        @include flex-column;
        gap: $spacing-md;
        align-items: center;
        text-align: center;

        .avatar-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: $spacing-md;

            @include respond-to(sm) {
                grid-template-columns: repeat(6, 1fr);
            }

            .avatar-option {
                width: 50px;
                height: 50px;
                @include flex-center;
                background-color: $background-lighter;
                border: 2px solid $border-color;
                border-radius: 50%;
                font-size: 24px;
                cursor: pointer;
                transition: $transition-base;

                &:hover {
                    background-color: lighten($background-lighter, 5%);
                }

                &.selected {
                    border-color: $secondary-color;
                    box-shadow: 0 0 10px rgba($secondary-color, 0.4);
                }
            }
        }
    }

    .achievement-details-modal {
        .achievement-header {
            @include flex-column;
            align-items: center;
            margin-bottom: $spacing-lg;

            .achievement-large-icon {
                font-size: 64px;
                margin-bottom: $spacing-md;
            }

            .achievement-status-label {
                padding: $spacing-xs $spacing-md;
                background-color: $background-lighter;
                border-radius: $border-radius-sm;
                color: $text-secondary;
                font-size: $font-size-sm;
                font-weight: 600;

                &.unlocked {
                    background-color: rgba($success-color, 0.2);
                    color: $success-color;
                }
            }
        }

        .achievement-details-content {
            @include flex-column;
            gap: $spacing-md;

            p {
                margin: 0;
            }

            .achievement-criteria,
            .achievement-rewards {
                h4 {
                    margin: 0 0 $spacing-sm 0;
                    color: $secondary-color;
                }

                ul {
                    margin: 0;
                    padding-left: 20px;

                    li {
                        margin-bottom: $spacing-xs;
                    }
                }
            }

            .achievement-date {
                margin-top: $spacing-md;
                padding-top: $spacing-md;
                border-top: 1px solid $border-color;
                font-size: $font-size-sm;
                color: $text-secondary;
            }
        }
    }

    .change-password-modal {
        @include flex-column;
        gap: $spacing-lg;

        .form-group {
            @include flex-column;
            gap: $spacing-xs;

            label {
                font-weight: 500;
            }

            input {
                background-color: rgba($background-lighter, 0.5);
                border: 1px solid $border-color;
                border-radius: $border-radius-sm;
                color: $text-color;
                padding: $spacing-md;
                transition: $transition-base;

                &:focus {
                    border-color: $secondary-color;
                    outline: none;
                    box-shadow: 0 0 0 2px rgba($secondary-color, 0.2);
                }
            }

            &.has-error {
                input {
                    border-color: $danger-color;

                    &:focus {
                        box-shadow: 0 0 0 2px rgba($danger-color, 0.2);
                    }
                }

                .error-message {
                    color: $danger-color;
                    font-size: $font-size-sm;
                    margin-top: 4px;
                }
            }
        }
    }

    .logout-confirmation {
        text-align: center;

        p {
            margin-bottom: $spacing-md;
        }

        .warning {
            color: $warning-color;
            font-style: italic;
        }
    }

    .modal-actions {
        @include flex-between;
        width: 100%;
    }
}
</style>