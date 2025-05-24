// src/views/LoginView.vue

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseNotification from '@/components/ui/BaseNotification.vue';
import authService from '@/services/authService';
import { usePlayerStore } from '@/stores/modules/player';

const router = useRouter();
const route = useRoute();
const playerStore = usePlayerStore();

// Form data
const email = ref('');
const password = ref('');
const rememberMe = ref(false);
const showPassword = ref(false);
const isLoading = ref(false);

// Form validation
const errors = ref({
  email: '',
  password: '',
  server: ''
});

// Notification state
const showNotification = ref(false);
const notificationType = ref('info');
const notificationMessage = ref('');

// Computed property for form validation
const isFormValid = computed(() => {
  return email.value.trim() !== '' &&
         password.value.trim() !== '' &&
         errors.value.email === '' &&
         errors.value.password === '';
});

// Form validation methods
function validateEmail() {
  const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (email.value.trim() === '') {
    errors.value.email = 'Email is required';
  } else if (!emailPattern.test(email.value)) {
    errors.value.email = 'Please enter a valid email address';
  } else {
    errors.value.email = '';
  }
}

function validatePassword() {
  if (password.value.trim() === '') {
    errors.value.password = 'Password is required';
  } else if (password.value.length < 6) {
    errors.value.password = 'Password must be at least 6 characters';
  } else {
    errors.value.password = '';
  }
}

function togglePasswordVisibility() {
  showPassword.value = !showPassword.value;
}

// Login function
async function login() {
  if (!isFormValid.value) return;

  isLoading.value = true;
  errors.value.server = '';

  try {
    const response = await authService.login({
      email: email.value,
      password: password.value
    });

    // Store authentication token
    const { player, token } = response.data;
    localStorage.setItem('auth_token', token);

    // Set up user state if needed
    if (playerStore.profile === null) {
      // Fetch player profile after login
      await playerStore.fetchProfile();
    }

    // Show success notification
    showNotificationMessage('success', 'Login successful! Welcome back.');

    // Redirect to home page or previous page
    const redirectPath = route.query.redirect ? String(route.query.redirect) : '/';
    setTimeout(() => {
      router.push(redirectPath);
    }, 1000);
  } catch (error: any) {
    // Handle error responses from the API
    if (error.message) {
      errors.value.server = error.message;
      showNotificationMessage('danger', error.message || 'Login failed. Invalid credentials.');
    } else {
      showNotificationMessage('danger', 'Login failed. Please check your credentials and try again.');
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

<template>
  <div class="login-view">
    <div class="auth-container">
      <div class="auth-card">
        <div class="auth-header">
          <h2>Mafia Wars: <span class="gold-text">Criminal Empire</span></h2>
          <p class="auth-subtitle">Sign in to continue your criminal empire</p>
        </div>

        <div class="auth-form">
          <div class="form-group" :class="{ 'has-error': errors.email }">
            <label for="email">Email</label>
            <input
              type="email"
              id="email"
              v-model="email"
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
                v-model="password"
                placeholder="Enter your password"
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

          <div class="form-options">
            <div class="remember-me">
              <input type="checkbox" id="remember" v-model="rememberMe" />
              <label for="remember">Remember me</label>
            </div>
            <router-link to="/forgot-password" class="forgot-password">Forgot password?</router-link>
          </div>

          <div class="auth-actions">
            <BaseButton
              variant="secondary"
              class="login-btn"
              :disabled="!isFormValid || isLoading"
              :loading="isLoading"
              @click="login"
            >
              Sign In
            </BaseButton>
          </div>

          <div class="auth-divider">
            <span>OR</span>
          </div>

          <div class="social-login">
            <button class="social-btn google">
              Google
            </button>
            <button class="social-btn facebook">
              Facebook
            </button>
          </div>
        </div>

        <div class="auth-footer">
          <p>Don't have an account? <router-link to="/register" class="register-link">Sign Up</router-link></p>
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



<style lang="scss">
.login-view {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-image: url('@/assets/images/city-skyline.png');
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
    background-color: rgba(37, 34, 34, 0.7);
    z-index: 1;
  }

  .auth-container {
    position: relative;
    z-index: 2;
    width: 100%;
    max-width: 480px;
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
    margin-bottom: $spacing-xl;

    h2 {
      font-size: $font-size-xxl;
      margin-bottom: $spacing-sm;
    }

    .auth-subtitle {
      color: $text-secondary;
      font-size: $font-size-md;
    }
  }

  .auth-form {
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

    .form-options {
      @include flex-between;

      .remember-me {
        display: flex;
        align-items: center;
        gap: $spacing-xs;

        input[type="checkbox"] {
          accent-color: $secondary-color;
        }
      }

      .forgot-password {
        color: $text-secondary;
        text-decoration: none;
        transition: $transition-base;

        &:hover {
          color: $secondary-color;
          text-decoration: underline;
        }
      }
    }

    .auth-actions {
      margin-top: $spacing-md;

      .login-btn {
        width: 100%;
      }
    }

    .auth-divider {
      display: flex;
      align-items: center;
      margin: $spacing-md 0;

      &:before,
      &:after {
        content: '';
        flex: 1;
        height: 1px;
        background-color: $border-color;
      }

      span {
        padding: 0 $spacing-md;
        color: $text-secondary;
        font-size: $font-size-sm;
      }
    }

    .social-login {
      display: flex;
      gap: $spacing-md;

      .social-btn {
        flex: 1;
        @include button-base;
        background-color: $background-lighter;
        border: 1px solid $border-color;
        padding: $spacing-md;

        &:hover {
          background-color: lighten($background-lighter, 5%);
        }

        &.google {
          &:hover {
            border-color: #DB4437;
          }
        }

        &.facebook {
          &:hover {
            border-color: #4267B2;
          }
        }
      }
    }
  }

  .auth-footer {
    text-align: center;
    margin-top: $spacing-xl;
    color: $text-secondary;

    .register-link {
      color: $secondary-color;
      font-weight: 600;
      text-decoration: none;
      transition: $transition-base;

      &:hover {
        text-decoration: underline;
      }
    }
  }
}
</style>
