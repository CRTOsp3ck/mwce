// src/router/routes.ts

import { RouteRecordRaw } from 'vue-router';

// Lazy-load components for better performance
const HomeView = () => import('@/views/HomeView.vue');
const TravelView = () => import('@/views/TravelView.vue');  // New Travel view
const TerritoryView = () => import('@/views/TerritoryView.vue');
const OperationsView = () => import('@/views/OperationsView.vue');
const MarketView = () => import('@/views/MarketView.vue');
const RankingsView = () => import('@/views/RankingsView.vue');
const NftView = () => import('@/views/NftView.vue');
const LoginView = () => import('@/views/LoginView.vue');
const RegisterView = () => import('@/views/RegisterView.vue');
const ProfileView = () => import('@/views/ProfileView.vue');
const SettingsView = () => import('@/views/SettingsView.vue');
const NotificationsView = () => import('@/views/NotificationsView.vue');
const CampaignsView = () => import('@/views/CampaignsView.vue');

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'Home',
    component: HomeView,
    meta: {
      title: 'Dashboard - Criminal Empire',
      requiresAuth: true
    },
  },
  {
    path: '/travel',
    name: 'Travel',
    component: TravelView,
    meta: {
      title: 'Travel - Criminal Empire',
      requiresAuth: true
    },
  },
  {
    path: '/territory',
    name: 'Territory',
    component: TerritoryView,
    meta: {
      title: 'Territory - Criminal Empire',
      requiresAuth: true,
      requiresRegion: true  // New meta field to check if player is in a region
    },
  },
  {
    path: '/operations',
    name: 'Operations',
    component: OperationsView,
    meta: {
      title: 'Operations - Criminal Empire',
      requiresAuth: true,
      requiresRegion: true  // New meta field to check if player is in a region
    },
  },
  {
    path: '/market',
    name: 'Market',
    component: MarketView,
    meta: {
      title: 'Market - Criminal Empire',
      requiresAuth: true,
      requiresRegion: true  // New meta field to check if player is in a region
    },
  },
  {
    path: '/rankings',
    name: 'Rankings',
    component: RankingsView,
    meta: {
      title: 'Rankings - Criminal Empire',
      requiresAuth: true
    },
  },
  {
    path: '/nft',
    name: 'NFT',
    component: NftView,
    meta: {
      title: 'NFT - Criminal Empire',
      requiresAuth: true
    },
  },
  {
    path: '/notifications',
    name: 'Notifications',
    component: NotificationsView,
    meta: {
      title: 'Notifications - Criminal Empire',
      requiresAuth: true
    },
  },
  {
    path: '/login',
    name: 'Login',
    component: LoginView,
    meta: {
      title: 'Sign In - Criminal Empire',
      requiresGuest: true
    },
  },
  {
    path: '/register',
    name: 'Register',
    component: RegisterView,
    meta: {
      title: 'Sign Up - Criminal Empire',
      requiresGuest: true
    },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: ProfileView,
    meta: {
      title: 'My Profile - Criminal Empire',
      requiresAuth: true
    },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: SettingsView,
    meta: {
      title: 'Account Settings - Criminal Empire',
      requiresAuth: true
    },
  },
  {
    path: '/campaigns',
    name: 'Campaigns',
    component: CampaignsView,
    meta: {
      title: 'Campaigns - Criminal Empire',
      requiresAuth: true
    },
  },
  // Catch-all route for 404
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
];

export default routes;
