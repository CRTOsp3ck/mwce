// src/router/index.ts

import { createRouter, createWebHistory, RouteLocationNormalized } from 'vue-router';
import routes from './routes';
import { usePlayerStore } from '@/stores/modules/player';
import { useTravelStore } from '@/stores/modules/travel';

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    // Always scroll to top
    return { top: 0 };
  },
});

// Auth guard and region guard
router.beforeEach(async (to, from, next) => {
  // Update document title
  document.title = to.meta.title as string || 'Criminal Empire';

  // Check if route requires authentication
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth);
  const requiresGuest = to.matched.some(record => record.meta.requiresGuest);

  // Check if route requires player to be in a region
  const requiresRegion = to.matched.some(record => record.meta.requiresRegion);

  // Check authentication status
  const authToken = localStorage.getItem('auth_token');
  const isAuthenticated = !!authToken;

  // Handle authentication requirements
  if (requiresAuth && !isAuthenticated) {
    next({ name: 'Login' });
    return;
  }

  if (requiresGuest && isAuthenticated) {
    next({ name: 'Home' });
    return;
  }

  // If route requires a region, check if the player is in a region
  if (requiresRegion && isAuthenticated) {
    // Initialize the travel store
    const travelStore = useTravelStore();

    // Fetch the current region if not already loaded
    if (travelStore.currentRegion === null) {
      try {
        await travelStore.fetchCurrentRegion();
      } catch (error) {
        console.error('Error fetching current region:', error);
      }
    }

    // If player is not in a region, redirect to travel view
    if (!travelStore.currentRegion) {
      // Redirect to travel page with a returnTo query param
      next({
        name: 'Travel',
        query: {
          returnTo: to.fullPath,
          message: 'You need to travel to a region first'
        }
      });
      return;
    }
  }

  // Continue to the requested route
  next();
});

export default router;
