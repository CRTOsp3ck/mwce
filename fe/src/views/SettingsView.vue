// src/views/SettingsView.vue

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import BaseNotification from '@/components/ui/BaseNotification.vue';

const router = useRouter();

// Settings navigation
const settingsSections = [
    { id: 'profile', label: 'Profile', icon: '👤' },
    { id: 'notifications', label: 'Notifications', icon: '🔔' },
    { id: 'display', label: 'Display', icon: '🎨' },
    { id: 'privacy', label: 'Privacy', icon: '🔒' },
    { id: 'security', label: 'Security', icon: '🛡️' }
];
const activeSection = ref('profile');

// Profile settings
const profileSettings = ref({
    name: 'Don Corleone',
    email: 'don.corleone@criminal.emp',
    bio: 'The head of the most powerful crime family in the city. Respect and loyalty are everything to me.'
});
const isSavingProfile = ref(false);

// Notification settings
const notificationSettings = ref([
    {
        name: 'Territory Alerts',
        description: 'Get notified when your territories are under attack or when you can collect resources.',
        enabled: true
    },
    {
        name: 'Operation Notifications',
        description: 'Receive alerts when operations are completed or when special operations become available.',
        enabled: true
    },
    {
        name: 'Heat Warnings',
        description: 'Get alerts when your heat level is getting dangerously high.',
        enabled: true
    },
    {
        name: 'Market Updates',
        description: 'Receive notifications about significant market price changes.',
        enabled: false
    },
    {
        name: 'Rival Activity',
        description: 'Be informed when rival gangs make significant moves in your area.',
        enabled: true
    },
    {
        name: 'Email Notifications',
        description: 'Receive important game notifications via email.',
        enabled: false
    }
]);
const isSavingNotifications = ref(false);

// Display settings
const displaySettings = ref({
    theme: 'default',
    animations: 'full',
    fontSize: 1
});
const isSavingDisplay = ref(false);

// Privacy settings
const privacySettings = ref([
    {
        name: 'Profile Visibility',
        description: 'Allow other players to view your profile and statistics.',
        enabled: true
    },
    {
        name: 'Territory Visibility',
        description: 'Show your controlled territories on the public map.',
        enabled: true
    },
    {
        name: 'Achievement Sharing',
        description: 'Share your achievements with other players.',
        enabled: true
    },
    {
        name: 'Activity Status',
        description: 'Show when you were last active in the game.',
        enabled: false
    }
]);
const isSavingPrivacy = ref(false);

// Security settings
const lastPasswordChange = ref('2023-01-15T14:30:00Z');
const twoFactorEnabled = ref(false);
const loginSessions = ref([
    {
        id: 's1',
        device: 'Chrome on Windows',
        location: 'New York, USA',
        lastActive: new Date().toISOString(),
        current: true
    },
    {
        id: 's2',
        device: 'Safari on iPhone',
        location: 'New York, USA',
        lastActive: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
        current: false
    },
    {
        id: 's3',
        device: 'Firefox on Mac',
        location: 'Chicago, USA',
        lastActive: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
        current: false
    }
]);

// Change password modal
const showPasswordModal = ref(false);
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

// Delete account modal
const showDeleteModal = ref(false);
const deleteConfirmation = ref('');
const isDeletingAccount = ref(false);

// Notification state
const showNotification = ref(false);
const notificationType = ref('info');
const notificationMessage = ref('');

// Computed properties
const canChangePassword = computed(() => {
    return passwordData.value.currentPassword !== '' &&
        passwordData.value.newPassword !== '' &&
        passwordData.value.confirmPassword !== '' &&
        passwordData.value.newPassword === passwordData.value.confirmPassword &&
        passwordData.value.newPassword.length >= 6;
});

// Load data when component is mounted
onMounted(() => {
    // In a real app, this would fetch data from the server
});

// Helper functions
function formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });
}

function getFontSizeLabel(fontSizeValue: number): string {
    switch (Number(fontSizeValue)) {
        case 0: return 'Small';
        case 1: return 'Medium';
        case 2: return 'Large';
        default: return 'Medium';
    }
}

// Settings actions
async function saveProfileSettings() {
    isSavingProfile.value = true;

    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Show success notification
        showNotificationMessage('success', 'Profile settings saved successfully');
    } catch (error) {
        showNotificationMessage('danger', 'Failed to save profile settings');
    } finally {
        isSavingProfile.value = false;
    }
}

async function saveNotificationSettings() {
    isSavingNotifications.value = true;

    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Show success notification
        showNotificationMessage('success', 'Notification settings saved successfully');
    } catch (error) {
        showNotificationMessage('danger', 'Failed to save notification settings');
    } finally {
        isSavingNotifications.value = false;
    }
}

function selectTheme(theme: string) {
    displaySettings.value.theme = theme;
}

async function saveDisplaySettings() {
    isSavingDisplay.value = true;

    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Show success notification
        showNotificationMessage('success', 'Display settings saved successfully');
    } catch (error) {
        showNotificationMessage('danger', 'Failed to save display settings');
    } finally {
        isSavingDisplay.value = false;
    }
}

async function savePrivacySettings() {
    isSavingPrivacy.value = true;

    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Show success notification
        showNotificationMessage('success', 'Privacy settings saved successfully');
    } catch (error) {
        showNotificationMessage('danger', 'Failed to save privacy settings');
    } finally {
        isSavingPrivacy.value = false;
    }
}

// Security actions
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

        // Update last password change date
        lastPasswordChange.value = new Date().toISOString();

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

async function toggleTwoFactor() {
    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Toggle the state
        twoFactorEnabled.value = !twoFactorEnabled.value;

        // Show success notification
        showNotificationMessage(
            'success',
            twoFactorEnabled.value
                ? 'Two-factor authentication enabled successfully'
                : 'Two-factor authentication disabled'
        );
    } catch (error) {
        showNotificationMessage('danger', 'Failed to update two-factor authentication');
    }
}

async function terminateSession(sessionId: string) {
    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 500));

        // Remove the session from the list
        loginSessions.value = loginSessions.value.filter(session => session.id !== sessionId);

        // Show success notification
        showNotificationMessage('success', 'Session terminated successfully');
    } catch (error) {
        showNotificationMessage('danger', 'Failed to terminate session');
    }
}

async function terminateAllSessions() {
    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Keep only the current session
        loginSessions.value = loginSessions.value.filter(session => session.current);

        // Show success notification
        showNotificationMessage('success', 'All other sessions terminated successfully');
    } catch (error) {
        showNotificationMessage('danger', 'Failed to terminate sessions');
    }
}

// Account deletion
function showDeleteAccountConfirmation() {
    deleteConfirmation.value = '';
    showDeleteModal.value = true;
}

function closeDeleteModal() {
    showDeleteModal.value = false;
}

async function deleteAccount() {
    if (deleteConfirmation.value !== 'DELETE') return;

    isDeletingAccount.value = true;

    try {
        // In a real app, this would make an API call
        await new Promise(resolve => setTimeout(resolve, 2000));

        // Clear auth token
        localStorage.removeItem('auth_token');

        // Show success notification
        showNotificationMessage('success', 'Account deleted successfully');

        // Redirect to login page after a short delay
        setTimeout(() => {
            router.push('/login');
        }, 1500);
    } catch (error) {
        showNotificationMessage('danger', 'Failed to delete account');
        isDeletingAccount.value = false;
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

<template>
    <div class="settings-view">
        <div class="page-title">
            <h2>Account Settings</h2>
            <p class="subtitle">Customize your criminal experience</p>
        </div>

        <div class="settings-content">
            <div class="settings-sidebar">
                <div class="settings-navigation">
                    <button v-for="(section, index) in settingsSections" :key="index" class="nav-item"
                        :class="{ active: activeSection === section.id }" @click="activeSection = section.id">
                        <span class="nav-icon">{{ section.icon }}</span>
                        <span class="nav-label">{{ section.label }}</span>
                    </button>
                </div>
            </div>

            <div class="settings-main">
                <!-- Profile Settings -->
                <div v-if="activeSection === 'profile'" class="settings-section">
                    <h3 class="section-title">Profile Settings</h3>

                    <BaseCard>
                        <form @submit.prevent="saveProfileSettings">
                            <div class="form-group">
                                <label for="name">Crime Boss Name</label>
                                <input type="text" id="name" v-model="profileSettings.name"
                                    placeholder="Enter your boss name" />
                            </div>

                            <div class="form-group">
                                <label for="email">Email Address</label>
                                <input type="email" id="email" v-model="profileSettings.email"
                                    placeholder="Enter your email" />
                            </div>

                            <div class="form-group">
                                <label for="bio">Boss Bio</label>
                                <textarea id="bio" v-model="profileSettings.bio"
                                    placeholder="Tell others about your criminal empire" rows="4"></textarea>
                            </div>

                            <div class="form-actions">
                                <BaseButton variant="secondary" type="submit" :loading="isSavingProfile">
                                    Save Changes
                                </BaseButton>
                            </div>
                        </form>
                    </BaseCard>

                    <BaseCard class="delete-account-card">
                        <h4>Delete Account</h4>
                        <p class="warning-text">This action is irreversible. All your progress, assets, and criminal
                            empire will be lost forever.</p>
                        <BaseButton variant="danger" @click="showDeleteAccountConfirmation">
                            Delete Account
                        </BaseButton>
                    </BaseCard>
                </div>

                <!-- Notification Settings -->
                <div v-if="activeSection === 'notifications'" class="settings-section">
                    <h3 class="section-title">Notification Settings</h3>

                    <BaseCard>
                        <div class="notification-settings">
                            <div class="setting-item" v-for="(setting, index) in notificationSettings" :key="index">
                                <div class="setting-info">
                                    <div class="setting-name">{{ setting.name }}</div>
                                    <div class="setting-description">{{ setting.description }}</div>
                                </div>
                                <div class="setting-control">
                                    <label class="toggle-switch">
                                        <input type="checkbox" v-model="setting.enabled">
                                        <span class="toggle-slider"></span>
                                    </label>
                                </div>
                            </div>
                        </div>

                        <div class="form-actions">
                            <BaseButton variant="secondary" @click="saveNotificationSettings"
                                :loading="isSavingNotifications">
                                Save Changes
                            </BaseButton>
                        </div>
                    </BaseCard>
                </div>

                <!-- Display Settings -->
                <div v-if="activeSection === 'display'" class="settings-section">
                    <h3 class="section-title">Display Settings</h3>

                    <BaseCard>
                        <div class="display-settings">
                            <div class="form-group">
                                <label>Theme</label>
                                <div class="theme-options">
                                    <div class="theme-option" :class="{ selected: displaySettings.theme === 'default' }"
                                        @click="selectTheme('default')">
                                        <div class="theme-preview default-theme"></div>
                                        <div class="theme-name">Classic Noir</div>
                                    </div>
                                    <div class="theme-option" :class="{ selected: displaySettings.theme === 'modern' }"
                                        @click="selectTheme('modern')">
                                        <div class="theme-preview modern-theme"></div>
                                        <div class="theme-name">Modern Syndicate</div>
                                    </div>
                                    <div class="theme-option" :class="{ selected: displaySettings.theme === 'retro' }"
                                        @click="selectTheme('retro')">
                                        <div class="theme-preview retro-theme"></div>
                                        <div class="theme-name">Retro Mob</div>
                                    </div>
                                </div>
                            </div>

                            <div class="form-group">
                                <label>Animation Effects</label>
                                <div class="radio-group">
                                    <label class="radio-label">
                                        <input type="radio" name="animations" value="full"
                                            v-model="displaySettings.animations">
                                        <span>Full Animations</span>
                                    </label>
                                    <label class="radio-label">
                                        <input type="radio" name="animations" value="reduced"
                                            v-model="displaySettings.animations">
                                        <span>Reduced Animations</span>
                                    </label>
                                    <label class="radio-label">
                                        <input type="radio" name="animations" value="off"
                                            v-model="displaySettings.animations">
                                        <span>No Animations</span>
                                    </label>
                                </div>
                            </div>

                            <div class="form-group">
                                <label>Font Size</label>
                                <div class="font-size-control">
                                    <span class="font-size-label small">A</span>
                                    <input type="range" min="0" max="2" step="1" v-model="displaySettings.fontSize">
                                    <span class="font-size-label large">A</span>
                                </div>
                                <div class="font-size-value">
                                    {{ getFontSizeLabel(displaySettings.fontSize) }}
                                </div>
                            </div>
                        </div>

                        <div class="form-actions">
                            <BaseButton variant="secondary" @click="saveDisplaySettings" :loading="isSavingDisplay">
                                Save Changes
                            </BaseButton>
                        </div>
                    </BaseCard>
                </div>

                <!-- Privacy Settings -->
                <div v-if="activeSection === 'privacy'" class="settings-section">
                    <h3 class="section-title">Privacy Settings</h3>

                    <BaseCard>
                        <div class="privacy-settings">
                            <div class="setting-item" v-for="(setting, index) in privacySettings" :key="index">
                                <div class="setting-info">
                                    <div class="setting-name">{{ setting.name }}</div>
                                    <div class="setting-description">{{ setting.description }}</div>
                                </div>
                                <div class="setting-control">
                                    <label class="toggle-switch">
                                        <input type="checkbox" v-model="setting.enabled">
                                        <span class="toggle-slider"></span>
                                    </label>
                                </div>
                            </div>
                        </div>

                        <div class="form-actions">
                            <BaseButton variant="secondary" @click="savePrivacySettings" :loading="isSavingPrivacy">
                                Save Changes
                            </BaseButton>
                        </div>
                    </BaseCard>

                    <BaseCard>
                        <h4>Data & Privacy</h4>
                        <p>You can download all your data or request account deletion.</p>

                        <div class="privacy-actions">
                            <BaseButton variant="outline">
                                Download My Data
                            </BaseButton>
                            <BaseButton variant="danger" @click="showDeleteAccountConfirmation">
                                Delete Account
                            </BaseButton>
                        </div>
                    </BaseCard>
                </div>

                <!-- Security Settings -->
                <div v-if="activeSection === 'security'" class="settings-section">
                    <h3 class="section-title">Security Settings</h3>

                    <BaseCard>
                        <h4>Password</h4>
                        <div class="password-info">
                            <p>Last changed: {{ formatDate(lastPasswordChange) }}</p>
                            <BaseButton variant="outline" @click="showChangePasswordModal">
                                Change Password
                            </BaseButton>
                        </div>
                    </BaseCard>

                    <BaseCard>
                        <h4>Two-Factor Authentication</h4>
                        <div class="two-factor-section">
                            <div class="two-factor-status">
                                <div class="status-indicator" :class="{ enabled: twoFactorEnabled }"></div>
                                <div class="status-text">
                                    <p>{{ twoFactorEnabled ? 'Enabled' : 'Disabled' }}</p>
                                    <p class="status-description">{{ twoFactorEnabled ?
                                        'Your account is protected with two-factor authentication.' :
                                        'Add an extra layer of security to your account.' }}</p>
                                </div>
                            </div>
                            <BaseButton :variant="twoFactorEnabled ? 'danger' : 'primary'" @click="toggleTwoFactor">
                                {{ twoFactorEnabled ? 'Disable' : 'Enable' }} 2FA
                            </BaseButton>
                        </div>
                    </BaseCard>

                    <BaseCard>
                        <h4>Login Sessions</h4>
                        <div class="sessions-list">
                            <div class="session-item" v-for="(session, index) in loginSessions" :key="index">
                                <div class="session-info">
                                    <div class="session-device">{{ session.device }}</div>
                                    <div class="session-details">
                                        <span>{{ session.location }}</span>
                                        <span class="separator">•</span>
                                        <span>{{ formatDate(session.lastActive) }}</span>
                                    </div>
                                </div>
                                <div class="session-actions">
                                    <button class="text-btn danger" v-if="!session.current"
                                        @click="terminateSession(session.id)">
                                        Terminate
                                    </button>
                                    <span v-else class="current-badge">Current</span>
                                </div>
                            </div>
                        </div>
                        <div class="form-actions">
                            <BaseButton variant="danger" @click="terminateAllSessions">
                                Log Out All Devices
                            </BaseButton>
                        </div>
                    </BaseCard>
                </div>
            </div>
        </div>

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

        <!-- Delete Account Confirmation Modal -->
        <BaseModal v-model="showDeleteModal" title="Delete Account">
            <div class="delete-account-modal">
                <div class="warning-icon">⚠️</div>
                <h4>This action cannot be undone</h4>
                <p>Are you sure you want to permanently delete your account? All your progress, assets, and criminal
                    empire will
                    be lost forever.</p>

                <div class="confirmation-input">
                    <label for="confirmation">Type "DELETE" to confirm:</label>
                    <input type="text" id="confirmation" v-model="deleteConfirmation" placeholder="DELETE" />
                </div>
            </div>
            <template #footer>
                <div class="modal-actions">
                    <BaseButton variant="text" @click="closeDeleteModal">
                        Cancel
                    </BaseButton>
                    <BaseButton variant="danger" :disabled="deleteConfirmation !== 'DELETE' || isDeletingAccount"
                        :loading="isDeletingAccount" @click="deleteAccount">
                        Permanently Delete
                    </BaseButton>
                </div>
            </template>
        </BaseModal>

        <BaseNotification v-if="showNotification" :type="notificationType" :message="notificationMessage"
            @close="closeNotification" />
    </div>
</template>

<style lang="scss">
.settings-view {
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

    .settings-content {
        display: grid;
        grid-template-columns: 1fr;
        gap: $spacing-xl;

        @include respond-to(md) {
            grid-template-columns: 250px 1fr;
        }

        .settings-sidebar {
            @include respond-to(md) {
                grid-column: 1;
                grid-row: 1;
            }

            .settings-navigation {
                @include flex-column;
                gap: $spacing-xs;

                @include respond-to(xs) {
                    @include flex-between;
                    flex-direction: row;
                    flex-wrap: wrap;
                    gap: $spacing-xs;
                }

                .nav-item {
                    display: flex;
                    align-items: center;
                    gap: $spacing-md;
                    padding: $spacing-md;
                    background-color: $background-lighter;
                    border: 1px solid $border-color;
                    border-radius: $border-radius-md;
                    cursor: pointer;
                    transition: $transition-base;
                    text-align: left;

                    @include respond-to(xs) {
                        width: calc(50% - #{$spacing-xs} / 2);
                    }

                    &:hover {
                        background-color: lighten($background-lighter, 5%);
                    }

                    &.active {
                        border-color: $secondary-color;
                        background-color: rgba($secondary-color, 0.1);

                        .nav-label {
                            color: $secondary-color;
                            font-weight: 600;
                        }
                    }

                    .nav-icon {
                        font-size: 24px;
                    }

                    .nav-label {
                        font-size: $font-size-md;
                    }
                }
            }
        }

        .settings-main {
            @include respond-to(md) {
                grid-column: 2;
                grid-row: 1;
            }

            .settings-section {
                @include flex-column;
                gap: $spacing-lg;

                .section-title {
                    margin: 0 0 $spacing-md 0;
                    @include gold-accent;
                    border-bottom: 1px solid $border-color;
                    padding-bottom: $spacing-sm;
                }

                .form-group {
                    @include flex-column;
                    gap: $spacing-xs;
                    margin-bottom: $spacing-lg;

                    label {
                        font-weight: 500;
                    }

                    input[type="text"],
                    input[type="email"],
                    input[type="password"],
                    textarea {
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

                    textarea {
                        resize: vertical;
                        min-height: 100px;
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

                    .theme-options {
                        display: grid;
                        grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
                        gap: $spacing-md;

                        .theme-option {
                            @include flex-column;
                            align-items: center;
                            gap: $spacing-sm;
                            cursor: pointer;
                            transition: $transition-base;

                            .theme-preview {
                                width: 120px;
                                height: 80px;
                                border-radius: $border-radius-sm;
                                border: 2px solid $border-color;
                                transition: $transition-base;

                                &.default-theme {
                                    background: linear-gradient(135deg, #121212 0%, #8b0000 100%);
                                }

                                &.modern-theme {
                                    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
                                }

                                &.retro-theme {
                                    background: linear-gradient(135deg, #2c3639 0%, #3f4e4f 50%, #a27b5c 100%);
                                }
                            }

                            .theme-name {
                                font-size: $font-size-sm;
                            }

                            &.selected {
                                .theme-preview {
                                    border-color: $secondary-color;
                                    box-shadow: 0 0 10px rgba($secondary-color, 0.4);
                                }

                                .theme-name {
                                    color: $secondary-color;
                                    font-weight: 600;
                                }
                            }
                        }
                    }

                    .radio-group {
                        @include flex-column;
                        gap: $spacing-sm;

                        .radio-label {
                            display: flex;
                            align-items: center;
                            gap: $spacing-sm;
                            cursor: pointer;
                        }

                        input[type="radio"] {
                            accent-color: $secondary-color;
                        }
                    }

                    .font-size-control {
                        display: flex;
                        align-items: center;
                        gap: $spacing-md;

                        input[type="range"] {
                            flex: 1;
                            -webkit-appearance: none;
                            height: 8px;
                            border-radius: 4px;
                            background: $background-lighter;
                            outline: none;

                            &::-webkit-slider-thumb {
                                -webkit-appearance: none;
                                appearance: none;
                                width: 18px;
                                height: 18px;
                                border-radius: 50%;
                                background: $secondary-color;
                                cursor: pointer;
                            }

                            &::-moz-range-thumb {
                                width: 18px;
                                height: 18px;
                                border-radius: 50%;
                                background: $secondary-color;
                                cursor: pointer;
                                border: none;
                            }
                        }

                        .font-size-label {
                            &.small {
                                font-size: $font-size-sm;
                            }

                            &.large {
                                font-size: $font-size-lg;
                            }
                        }
                    }

                    .font-size-value {
                        text-align: center;
                        font-size: $font-size-sm;
                        color: $text-secondary;
                        margin-top: $spacing-xs;
                    }
                }

                .notification-settings,
                .privacy-settings {
                    @include flex-column;
                    gap: $spacing-lg;

                    .setting-item {
                        display: flex;
                        justify-content: space-between;
                        align-items: center;
                        gap: $spacing-md;
                        padding-bottom: $spacing-md;
                        border-bottom: 1px solid $border-color;

                        &:last-child {
                            border-bottom: none;
                            padding-bottom: 0;
                        }

                        .setting-info {
                            flex: 1;

                            .setting-name {
                                font-weight: 500;
                                margin-bottom: $spacing-xs;
                            }

                            .setting-description {
                                font-size: $font-size-sm;
                                color: $text-secondary;
                            }
                        }

                        .setting-control {
                            .toggle-switch {
                                position: relative;
                                display: inline-block;
                                width: 44px;
                                height: 24px;

                                input {
                                    opacity: 0;
                                    width: 0;
                                    height: 0;

                                    &:checked+.toggle-slider {
                                        background-color: $success-color;
                                    }

                                    &:checked+.toggle-slider:before {
                                        transform: translateX(20px);
                                    }
                                }

                                .toggle-slider {
                                    position: absolute;
                                    cursor: pointer;
                                    top: 0;
                                    left: 0;
                                    right: 0;
                                    bottom: 0;
                                    background-color: $background-lighter;
                                    transition: $transition-base;
                                    border-radius: 24px;

                                    &:before {
                                        position: absolute;
                                        content: "";
                                        height: 18px;
                                        width: 18px;
                                        left: 3px;
                                        bottom: 3px;
                                        background-color: white;
                                        transition: $transition-base;
                                        border-radius: 50%;
                                    }
                                }
                            }
                        }
                    }
                }

                .form-actions {
                    margin-top: $spacing-md;
                    display: flex;
                    justify-content: flex-end;
                }

                .delete-account-card {
                    margin-top: $spacing-lg;
                    border-left: 3px solid $danger-color;

                    h4 {
                        margin: 0 0 $spacing-md 0;
                        color: $danger-color;
                    }

                    .warning-text {
                        color: $text-secondary;
                        margin-bottom: $spacing-md;
                    }
                }

                .password-info {
                    @include flex-between;
                    align-items: center;

                    p {
                        margin: 0;
                        color: $text-secondary;
                    }
                }

                .two-factor-section {
                    @include flex-between;
                    align-items: center;
                    flex-wrap: wrap;
                    gap: $spacing-md;

                    .two-factor-status {
                        display: flex;
                        gap: $spacing-md;
                        flex: 1;

                        .status-indicator {
                            width: 16px;
                            height: 16px;
                            border-radius: 50%;
                            background-color: $danger-color;
                            margin-top: 4px;

                            &.enabled {
                                background-color: $success-color;
                            }
                        }

                        .status-text {
                            p {
                                margin: 0;

                                &:first-child {
                                    font-weight: 600;
                                }

                                &.status-description {
                                    font-size: $font-size-sm;
                                    color: $text-secondary;
                                }
                            }
                        }
                    }
                }

                .sessions-list {
                    @include flex-column;
                    gap: $spacing-md;

                    .session-item {
                        @include flex-between;
                        align-items: center;
                        padding: $spacing-md;
                        background-color: rgba($background-lighter, 0.5);
                        border-radius: $border-radius-sm;
                        gap: $spacing-md;

                        .session-info {
                            flex: 1;

                            .session-device {
                                font-weight: 500;
                                margin-bottom: $spacing-xs;
                            }

                            .session-details {
                                font-size: $font-size-sm;
                                color: $text-secondary;

                                .separator {
                                    margin: 0 $spacing-xs;
                                }
                            }
                        }

                        .session-actions {
                            .text-btn {
                                background: none;
                                border: none;
                                color: $text-secondary;
                                font-size: $font-size-sm;
                                cursor: pointer;
                                transition: $transition-base;

                                &:hover {
                                    text-decoration: underline;
                                }

                                &.danger {
                                    color: $danger-color;
                                }
                            }

                            .current-badge {
                                padding: $spacing-xs $spacing-sm;
                                background-color: rgba($success-color, 0.2);
                                color: $success-color;
                                border-radius: $border-radius-sm;
                                font-size: $font-size-sm;
                                font-weight: 600;
                            }
                        }
                    }
                }

                .privacy-actions {
                    display: flex;
                    gap: $spacing-md;
                    margin-top: $spacing-md;
                }
            }
        }
    }

    // Modal styles
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

    .delete-account-modal {
        @include flex-column;
        align-items: center;
        text-align: center;
        gap: $spacing-md;

        .warning-icon {
            font-size: 48px;
            color: $danger-color;
        }

        h4 {
            color: $danger-color;
            margin: 0;
        }

        p {
            color: $text-secondary;
            margin-bottom: $spacing-md;
        }

        .confirmation-input {
            width: 100%;
            @include flex-column;
            gap: $spacing-xs;
            margin-top: $spacing-md;

            label {
                font-weight: 500;
                color: $danger-color;
            }

            input {
                background-color: rgba($background-lighter, 0.5);
                border: 1px solid $danger-color;
                border-radius: $border-radius-sm;
                color: $text-color;
                padding: $spacing-md;
                text-align: center;
                font-weight: 600;

                &:focus {
                    outline: none;
                    box-shadow: 0 0 0 2px rgba($danger-color, 0.2);
                }
            }
        }
    }

    .modal-actions {
        @include flex-between;
        width: 100%;
    }
}
</style>