// src/router/index.ts

import { createRouter, createWebHistory } from 'vue-router';
import routes from './routes';

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else {
      return { top: 0 };
    }
  },
});

// Global navigation guard
router.beforeEach((to, from, next) => {
  // Update page title
  if (to.meta.title) {
    document.title = to.meta.title as string;
  }
  
  // Auth check
  const isAuthenticated = !!localStorage.getItem('auth_token');
  
  if (to.meta.requiresAuth && !isAuthenticated) {
    next({ path: '/login', query: { redirect: to.fullPath } });
  } else if (to.meta.requiresGuest && isAuthenticated) {
    next({ path: '/' });
  } else {
    next();
  }
});

export default router;