// src/views/RegisterView.vue

<template>
  <div class="register-view">
    <div class="auth-container">
      <div class="auth-card">
        <div class="auth-header">
          <h2>Mafia Wars: <span class="gold-text">Criminal Empire</span></h2>
          <p class="auth-subtitle">Create your criminal empire</p>
        </div>

        <div class="auth-steps">
          <div 
            v-for="(step, index) in steps" 
            :key="index" 
            class="step" 
            :class="{ 
              'active': currentStep === index, 
              'completed': currentStep > index 
            }"
          >
            <div class="step-number">{{ index + 1 }}</div>
            <div class="step-label">{{ step }}</div>
          </div>
        </div>

        <div class="auth-form">
          <!-- Step 1: Account Information -->
          <div v-if="currentStep === 0" class="step-content slide-in">
            <div class="form-group" :class="{ 'has-error': errors.email }">
              <label for="email">Email</label>
              <input 
                type="email" 
                id="email" 
                v-model="formData.email" 
                placeholder="Enter your email"
                @input="validateEmail"
              />
              <div class="error-message" v-if="errors.email">{{ errors.email }}</div>
            </div>

            <div class="form-group" :class="{ 'has-error': errors.password }">
              <label for="password">Password</label>
              <div class="password-input">
                <input 
                  :type="showPassword ? 'text' : 'password'" 
                  id="password" 
                  v-model="formData.password" 
                  placeholder="Create a password"
                  @input="validatePassword"
                />
                <button 
                  type="button" 
                  class="password-toggle" 
                  @click="togglePasswordVisibility"
                >
                  {{ showPassword ? '👁️' : '👁️‍🗨️' }}
                </button>
              </div>
              <div class="error-message" v-if="errors.password">{{ errors.password }}</div>
            </div>

            <div class="form-group" :class="{ 'has-error': errors.confirmPassword }">
              <label for="confirm-password">Confirm Password</label>
              <input 
                type="password" 
                id="confirm-password" 
                v-model="formData.confirmPassword" 
                placeholder="Confirm your password"
                @input="validateConfirmPassword"
              />
              <div class="error-message" v-if="errors.confirmPassword">{{ errors.confirmPassword }}</div>
            </div>
          </div>

          <!-- Step 2: Personal Information -->
          <div v-if="currentStep === 1" class="step-content slide-in">
            <div class="form-group" :class="{ 'has-error': errors.name }">
              <label for="name">Crime Boss Name</label>
              <input 
                type="text" 
                id="name" 
                v-model="formData.name" 
                placeholder="Enter your boss name"
                @input="validateName"
              />
              <div class="error-message" v-if="errors.name">{{ errors.name }}</div>
            </div>

            <div class="form-group">
              <label for="avatar">Avatar</label>
              <div class="avatar-selection">
                <div 
                  v-for="(avatar, index) in avatarOptions" 
                  :key="index"
                  class="avatar-option"
                  :class="{ selected: formData.avatar === avatar }"
                  @click="selectAvatar(avatar)"
                >
                  <div class="avatar-icon">{{ avatar }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 3: Starting Territory -->
          <div v-if="currentStep === 2" class="step-content slide-in">
            <h3 class="step-title">Choose Your Starting Territory</h3>
            <p class="step-description">Select the district where your criminal empire will begin:</p>

            <div class="territory-options">
              <div 
                v-for="(territory, index) in territoryOptions" 
                :key="index"
                class="territory-option"
                :class="{ selected: formData.territory === territory.id }"
                @click="selectTerritory(territory.id)"
              >
                <div class="territory-icon">{{ territory.icon }}</div>
                <div class="territory-info">
                  <div class="territory-name">{{ territory.name }}</div>
                  <div class="territory-description">{{ territory.description }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 4: Terms and Conditions -->
          <div v-if="currentStep === 3" class="step-content slide-in">
            <h3 class="step-title">Terms of Service</h3>
            <div class="terms-container">
              <p>By creating an account, you agree to the following terms:</p>
              <ol>
                <li>You are responsible for maintaining the confidentiality of your account.</li>
                <li>You will not use the service for any illegal activities (in real life).</li>
                <li>Sharing account credentials is prohibited.</li>
                <li>The game administrators reserve the right to terminate accounts that violate the terms.</li>
                <li>All in-game purchases are final and non-refundable.</li>
              </ol>
            </div>

            <div class="form-group checkbox">
              <input type="checkbox" id="terms" v-model="formData.agreedToTerms" />
              <label for="terms">I agree to the Terms of Service and Privacy Policy</label>
              <div class="error-message" v-if="errors.agreedToTerms">{{ errors.agreedToTerms }}</div>
            </div>
          </div>

          <!-- Error message for server errors -->
          <div v-if="errors.server" class="server-error-message">
            {{ errors.server }}
          </div>

          <div class="auth-actions">
            <BaseButton 
              v-if="currentStep > 0" 
              variant="outline" 
              @click="prevStep"
            >
              Back
            </BaseButton>
            <BaseButton 
              v-if="currentStep < steps.length - 1" 
              variant="primary" 
              @click="nextStep"
              :disabled="!canProceed"
            >
              Continue
            </BaseButton>
            <BaseButton 
              v-if="currentStep === steps.length - 1" 
              variant="secondary" 
              :disabled="!canRegister || isLoading"
              :loading="isLoading"
              @click="register"
            >
              Create Account
            </BaseButton>
          </div>
        </div>

        <div class="auth-footer">
          <p>Already have an account? <router-link to="/login" class="login-link">Sign In</router-link></p>
        </div>
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

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseNotification from '@/components/ui/BaseNotification.vue';
import authService from '@/services/authService';
import { usePlayerStore } from '@/stores/modules/player';

const router = useRouter();
const playerStore = usePlayerStore();

// Multi-step form data
const steps = [
  'Account Information',
  'Personal Information',
  'Starting Territory',
  'Terms & Conditions'
];
const currentStep = ref(0);

// Form data
const formData = ref({
  email: '',
  password: '',
  confirmPassword: '',
  name: '',
  avatar: '🤵',
  territory: '',
  agreedToTerms: false
});

// UI state
const showPassword = ref(false);
const isLoading = ref(false);

// Avatar options
const avatarOptions = ['🤵', '🕵️', '👨‍💼', '👩‍💼', '👨‍🦱', '👩‍🦱', '👨‍🦰', '👩‍🦰'];

// Territory options
const territoryOptions = [
  {
    id: 'downtown',
    name: 'Downtown',
    icon: '🏙️',
    description: 'The heart of the city. High risk but high reward with plenty of business opportunities.'
  },
  {
    id: 'docks',
    name: 'The Docks',
    icon: '⚓',
    description: 'Control the flow of goods in and out of the city. Great for smuggling operations.'
  },
  {
    id: 'suburbs',
    name: 'The Suburbs',
    icon: '🏘️',
    description: 'Quiet neighborhoods with less competition but lower initial profits.'
  },
  {
    id: 'industrial',
    name: 'Industrial District',
    icon: '🏭',
    description: 'Factories and warehouses provide cover for various operations.'
  }
];

// Form validation
const errors = ref({
  email: '',
  password: '',
  confirmPassword: '',
  name: '',
  agreedToTerms: '',
  server: ''
});

// Notification state
const showNotification = ref(false);
const notificationType = ref('info');
const notificationMessage = ref('');

// Computed properties for form validation
const isStep1Valid = computed(() => {
  return formData.value.email !== '' && 
         formData.value.password !== '' && 
         formData.value.confirmPassword !== '' && 
         errors.value.email === '' && 
         errors.value.password === '' && 
         errors.value.confirmPassword === '';
});

const isStep2Valid = computed(() => {
  return formData.value.name !== '' && errors.value.name === '';
});

const isStep3Valid = computed(() => {
  return formData.value.territory !== '';
});

const isStep4Valid = computed(() => {
  return formData.value.agreedToTerms;
});

const canProceed = computed(() => {
  switch (currentStep.value) {
    case 0: return isStep1Valid.value;
    case 1: return isStep2Valid.value;
    case 2: return isStep3Valid.value;
    case 3: return isStep4Valid.value;
    default: return false;
  }
});

const canRegister = computed(() => {
  return isStep1Valid.value && isStep2Valid.value && isStep3Valid.value && isStep4Valid.value;
});

// Form validation methods
function validateEmail() {
  const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (formData.value.email.trim() === '') {
    errors.value.email = 'Email is required';
  } else if (!emailPattern.test(formData.value.email)) {
    errors.value.email = 'Please enter a valid email address';
  } else {
    errors.value.email = '';
  }
}

function validatePassword() {
  if (formData.value.password.trim() === '') {
    errors.value.password = 'Password is required';
  } else if (formData.value.password.length < 6) {
    errors.value.password = 'Password must be at least 6 characters';
  } else {
    errors.value.password = '';
  }

  // Also validate confirm password if it has a value
  if (formData.value.confirmPassword) {
    validateConfirmPassword();
  }
}

function validateConfirmPassword() {
  if (formData.value.confirmPassword.trim() === '') {
    errors.value.confirmPassword = 'Please confirm your password';
  } else if (formData.value.confirmPassword !== formData.value.password) {
    errors.value.confirmPassword = 'Passwords do not match';
  } else {
    errors.value.confirmPassword = '';
  }
}

function validateName() {
  if (formData.value.name.trim() === '') {
    errors.value.name = 'Boss name is required';
  } else if (formData.value.name.length < 3) {
    errors.value.name = 'Boss name must be at least 3 characters';
  } else {
    errors.value.name = '';
  }
}

function togglePasswordVisibility() {
  showPassword.value = !showPassword.value;
}

function selectAvatar(avatar: string) {
  formData.value.avatar = avatar;
}

function selectTerritory(territoryId: string) {
  formData.value.territory = territoryId;
}

// Step navigation
function nextStep() {
  if (currentStep.value < steps.length - 1 && canProceed.value) {
    currentStep.value++;
  }
}

function prevStep() {
  if (currentStep.value > 0) {
    currentStep.value--;
  }
}

// Register function
async function register() {
  if (!canRegister.value) {
    if (!formData.value.agreedToTerms) {
      errors.value.agreedToTerms = 'You must agree to the terms to continue';
    }
    return;
  }

  isLoading.value = true;
  errors.value.server = '';
  
  try {
    // Call the registration endpoint with the required data
    const response = await authService.register({
      name: formData.value.name,
      email: formData.value.email,
      password: formData.value.password,
      confirmPassword: formData.value.confirmPassword,
      territory: formData.value.territory
    });

    // Store authentication token
    const { token, player } = response.data;
    localStorage.setItem('auth_token', token);
    
    // Initialize player profile
    await playerStore.fetchProfile();

    // Show success notification
    showNotificationMessage('success', 'Registration successful! Welcome to Criminal Empire.');

    // Redirect to home page after a short delay
    setTimeout(() => {
      router.push('/');
    }, 1500);
  } catch (error: any) {
    // Handle error responses from the API
    errors.value.server = error.message || 'Registration failed. Please try again.';
    
    // Show error notification
    showNotificationMessage('danger', error.message || 'Registration failed. Please try again.');
    
    // If there's a specific field error, highlight it and go to that step
    if (error.errors) {
      if (error.errors.email || error.errors.password || error.errors.confirmPassword) {
        currentStep.value = 0;
      } else if (error.errors.name) {
        currentStep.value = 1;
      } else if (error.errors.territory) {
        currentStep.value = 2;
      }
    }
  } finally {
    isLoading.value = false;
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
// Continuing the styles for RegisterView.vue

.register-view {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-image: url('https://images.unsplash.com/photo-1582574501989-b88675f0a25a?ixlib=rb-1.2.1&auto=format&fit=crop&w=1950&q=80');
  background-size: cover;
  background-position: center;
  position: relative;

  &:before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    z-index: 1;
  }

  .auth-container {
    position: relative;
    z-index: 2;
    width: 100%;
    max-width: 580px;
    padding: $spacing-md;
  }

  .auth-card {
    @include glass-effect;
    border-radius: $border-radius-lg;
    overflow: hidden;
    padding: $spacing-xl;
  }

  .auth-header {
    text-align: center;
    margin-bottom: $spacing-lg;

    h2 {
      font-size: $font-size-xxl;
      margin-bottom: $spacing-sm;
    }

    .auth-subtitle {
      color: $text-secondary;
      font-size: $font-size-md;
    }
  }

  .auth-steps {
    display: flex;
    justify-content: space-between;
    margin-bottom: $spacing-xl;
    position: relative;

    &:before {
      content: '';
      position: absolute;
      top: 16px;
      left: 20px;
      right: 20px;
      height: 2px;
      background-color: $border-color;
      z-index: 1;
    }

    .step {
      position: relative;
      z-index: 2;
      @include flex-column;
      align-items: center;
      gap: $spacing-xs;
      flex: 1;

      .step-number {
        width: 32px;
        height: 32px;
        @include flex-center;
        background-color: $background-lighter;
        border: 2px solid $border-color;
        border-radius: 50%;
        font-weight: 600;
        transition: $transition-base;
      }

      .step-label {
        font-size: $font-size-sm;
        color: $text-secondary;
        text-align: center;
        transition: $transition-base;

        @include respond-to(sm) {
          font-size: $font-size-md;
        }
      }

      &.active {
        .step-number {
          background-color: $primary-color;
          border-color: $primary-color;
        }

        .step-label {
          color: $text-color;
          font-weight: 500;
        }
      }

      &.completed {
        .step-number {
          background-color: $secondary-color;
          border-color: $secondary-color;
        }
      }
    }
  }

  .auth-form {
    .step-content {
      @include flex-column;
      gap: $spacing-lg;
      margin-bottom: $spacing-xl;
    }

    .step-title {
      margin: 0 0 $spacing-sm 0;
      @include gold-accent;
    }

    .step-description {
      color: $text-secondary;
      margin-bottom: $spacing-md;
    }

    .form-group {
      @include flex-column;
      gap: $spacing-xs;

      label {
        font-weight: 500;
      }

      input[type="text"],
      input[type="email"],
      input[type="password"] {
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

      &.checkbox {
        flex-direction: row;
        align-items: center;
        gap: $spacing-sm;
        
        input[type="checkbox"] {
          accent-color: $secondary-color;
        }
      }

      .password-input {
        position: relative;

        input {
          width: 100%;
          padding-right: 40px;
        }

        .password-toggle {
          position: absolute;
          right: $spacing-sm;
          top: 50%;
          transform: translateY(-50%);
          background: none;
          border: none;
          color: $text-secondary;
          cursor: pointer;

          &:hover {
            color: $text-color;
          }
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

    .server-error-message {
      padding: $spacing-md;
      background-color: rgba($danger-color, 0.1);
      border-left: 3px solid $danger-color;
      color: $danger-color;
      margin-bottom: $spacing-md;
      border-radius: $border-radius-sm;
    }

    .avatar-selection {
      display: flex;
      flex-wrap: wrap;
      gap: $spacing-md;
      justify-content: center;

      .avatar-option {
        width: 60px;
        height: 60px;
        @include flex-center;
        background-color: $background-lighter;
        border: 2px solid $border-color;
        border-radius: 50%;
        font-size: 32px;
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

    .territory-options {
      @include flex-column;
      gap: $spacing-md;

      .territory-option {
        display: flex;
        gap: $spacing-md;
        background-color: rgba($background-lighter, 0.5);
        border: 1px solid $border-color;
        border-radius: $border-radius-md;
        padding: $spacing-md;
        cursor: pointer;
        transition: $transition-base;

        &:hover {
          background-color: rgba($background-lighter, 0.8);
        }

        &.selected {
          border-color: $secondary-color;
          background-color: rgba($secondary-color, 0.1);
        }

        .territory-icon {
          font-size: 32px;
          align-self: center;
        }

        .territory-info {
          flex: 1;

          .territory-name {
            font-weight: 600;
            margin-bottom: $spacing-xs;
          }

          .territory-description {
            font-size: $font-size-sm;
            color: $text-secondary;
          }
        }
      }
    }

    .terms-container {
      height: 200px;
      overflow-y: auto;
      background-color: rgba($background-lighter, 0.5);
      border: 1px solid $border-color;
      border-radius: $border-radius-sm;
      padding: $spacing-md;
      margin-bottom: $spacing-md;

      p, ol, li {
        margin-bottom: $spacing-sm;
        color: $text-secondary;
      }

      ol {
        padding-left: 20px;
      }
    }

    .auth-actions {
      display: flex;
      justify-content: space-between;

      button {
        min-width: 120px;
      }
    }
  }

  .auth-footer {
    text-align: center;
    margin-top: $spacing-lg;
    color: $text-secondary;

    .login-link {
      color: $secondary-color;
      font-weight: 600;
      text-decoration: none;
      transition: $transition-base;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  .slide-in {
    animation: slideIn 0.3s ease forwards;
  }

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateX(20px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }
}
</style>