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
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import AppHeader from '@/components/layout/AppHeader.vue';
import AppSidebar from '@/components/layout/AppSidebar.vue';
import AppFooter from '@/components/layout/AppFooter.vue';

const route = useRoute();

// Check if current route is an authentication route (login or register)
const isAuthRoute = computed(() => {
  return ['/login', '/register'].includes(route.path);
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