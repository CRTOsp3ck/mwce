// src/App.vue

<template>
  <div class="app-container" :class="{ 'auth-layout': isAuthRoute }">
    <!-- Use conditional components for mobile/desktop layouts -->
    <MobileHeader v-if="!isAuthRoute && isMobile" />
    <AppHeader v-else-if="!isAuthRoute" />

    <div class="main-content" :class="{ 'auth-content': isAuthRoute }">
      <!-- Sidebar only visible on desktop -->
      <AppSidebar v-if="!isAuthRoute && !isMobile" />

      <!-- Main content area with page transitions -->
      <main class="content">
        <router-view v-slot="{ Component, route }">
          <transition name="page-transition" mode="out-in">
            <!-- CHANGED: Using route.path as key instead of $route.fullPath -->
            <div :key="route.path" class="view-wrapper">
              <component :is="Component" />
            </div>
          </transition>
        </router-view>
      </main>
    </div>

    <!-- Mobile navigation bar (only on mobile) -->
    <MobileNavBar v-if="!isAuthRoute && isMobile" />

    <!-- Footer always appears -->
    <AppFooter v-if="!isAuthRoute" />

    <!-- Loading indicator for mobile -->
    <div v-if="isLoading" class="mobile-loading"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, watch, ref, onUpdated } from 'vue';
import { useRoute } from 'vue-router';
import sseService from './services/sseService';
import { usePlayerStore } from '@/stores/modules/player';
import { useTravelStore } from '@/stores/modules/travel';
import AppHeader from '@/components/layout/AppHeader.vue';
import AppSidebar from '@/components/layout/AppSidebar.vue';
import AppFooter from '@/components/layout/AppFooter.vue';
import MobileHeader from '@/components/layout/MobileHeader.vue';
import MobileNavBar from '@/components/layout/MobileNavBar.vue';

const route = useRoute();
const playerStore = usePlayerStore();
const travelStore = useTravelStore();

// Mobile detection
const isMobile = ref(false);
// Loading state
const isLoading = ref(false);

// Track window dimensions and update mobile state
function checkMobile() {
  isMobile.value = window.innerWidth < 768; // Same as md breakpoint
}

// Check if current route is an authentication route (login or register)
const isAuthRoute = computed(() => {
  return ['/login', '/register'].includes(route.path);
});

// Initial setup and responsive listeners
onMounted(async () => {
  // Initialize mobile detection
  checkMobile();
  window.addEventListener('resize', checkMobile);

  // Create a global loading state event handler
  window.addEventListener('app:loading:start', () => { isLoading.value = true; });
  window.addEventListener('app:loading:end', () => { isLoading.value = false; });

  if (!isAuthRoute.value && localStorage.getItem('auth_token')) {
    isLoading.value = true;
    try {
      // Load player data if we're on a non-auth page and have a token
      if (!playerStore.profile) {
        await playerStore.fetchProfile();
      }

      // Load current region information if we have a token
      await travelStore.fetchCurrentRegion();

      // Connect to SSE for real-time updates
      sseService.connect();
    } catch (error) {
      console.error('Error loading initial data:', error);
    } finally {
      isLoading.value = false;
    }
  }
});

// Disconnect from SSE when component in unmounted
onBeforeUnmount(() => {
  sseService.disconnect();
  window.removeEventListener('resize', checkMobile);
  window.removeEventListener('app:loading:start', () => { isLoading.value = true; });
  window.removeEventListener('app:loading:end', () => { isLoading.value = false; });
});

// Add touch class to body for mobile-specific styles
onUpdated(() => {
  document.body.classList.toggle('is-touch-device', isMobile.value);
});

// Watch for route changes to update data as needed
watch(route, async (newRoute) => {
  if (!isAuthRoute.value && localStorage.getItem('auth_token')) {
    // Check if we need to load the player's current region
    if (newRoute.meta.requiresRegion && !travelStore.currentRegion) {
      await travelStore.fetchCurrentRegion();
    }
  }
});
</script>

<style lang="scss">
// Import responsive styles
@import './assets/styles/_responsive.scss';

// Core app styling
.app-container {
  position: relative;
  min-height: 100vh;
  background-color: $background-color;
  color: $text-color;

  &.auth-layout {
    // Full page layout for auth pages
    background: none; // This allows the login/register backgrounds to show
  }
}

// View wrapper to fix the transition issue
.view-wrapper {
  position: relative;
  width: 100%;
  min-height: 50vh;
}

// Apply a standard touch effect to all clickable elements on mobile
.is-touch-device {
  .nav-item, .button, a, .card {
    &:active {
      transform: scale(0.98);
    }
  }
}

// Game-like animation when clicking buttons
.base-button:active {
  transform: translateY(2px);
  transition: transform 0.1s;
}

// Improved page transitions
.page-transition-enter-active,
.page-transition-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.page-transition-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.page-transition-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

// Loading indicator styling
.mobile-loading {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(
    to right,
    $secondary-color 0%,
    $primary-color 50%,
    $secondary-color 100%
  );
  background-size: 200% 100%;
  animation: loading-slide 1.5s infinite linear;
  z-index: $z-index-modal + 10;
}

@keyframes loading-slide {
  0% { background-position: 200% 0; }
  100% { background-position: 0 0; }
}
</style>
