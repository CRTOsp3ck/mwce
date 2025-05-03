// src/App.vue

<template>
  <div class="app-container" :class="{ 'auth-layout': isAuthRoute }">
    <AppHeader v-if="!isAuthRoute" />
    <div class="main-content" :class="{ 'auth-content': isAuthRoute }">
      <AppSidebar v-if="!isAuthRoute" />
      <main class="content">
        <router-view />
      </main>
    </div>
    <AppFooter v-if="!isAuthRoute" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRoute } from 'vue-router';
import sseService from './services/sseService';
import { usePlayerStore } from '@/stores/modules/player';
import { useTravelStore } from '@/stores/modules/travel';
import AppHeader from '@/components/layout/AppHeader.vue';
import AppSidebar from '@/components/layout/AppSidebar.vue';
import AppFooter from '@/components/layout/AppFooter.vue';

const route = useRoute();
const playerStore = usePlayerStore();
const travelStore = useTravelStore();

// Check if current route is an authentication route (login or register)
const isAuthRoute = computed(() => {
  return ['/login', '/register'].includes(route.path);
});

// Initial data loading
onMounted(async () => {
  if (!isAuthRoute.value && localStorage.getItem('auth_token')) {
    // Load player data if we're on a non-auth page and have a token
    if (!playerStore.profile) {
      await playerStore.fetchProfile();
    }

    // Load current region information if we have a token
    await travelStore.fetchCurrentRegion();

    // Connect to SSE for real-time updates
    sseService.connect();
  }
});

// Disconnect from SSE when component in unmounted
onBeforeUnmount(() => {
  sseService.disconnect();
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
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: $background-color;
  color: $text-color;

  &.auth-layout {
    // Full page layout for auth pages
    background: none; // This allows the login/register backgrounds to show
  }
}

.main-content {
  display: flex;
  flex: 1;

  &.auth-content {
    // Auth pages don't need the flex layout with sidebar
    display: block;
  }
}

.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;

  .auth-layout & {
    padding: 0; // Remove padding for login/register pages
  }
}
</style>
