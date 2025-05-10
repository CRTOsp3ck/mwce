// src/components/layout/MobileNavBar.vue (Updated)

<template>
  <div class="mobile-nav-bar">
    <div class="nav-items-container" ref="navContainer">
      <!-- Left scroll arrow -->
      <button
        class="scroll-arrow left-arrow"
        :class="{ visible: canScrollLeft }"
        @click="scrollLeft"
        v-show="canScrollLeft"
      >
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 18l-6-6 6-6" />
        </svg>
      </button>

      <!-- Navigation items with horizontal scroll -->
      <div class="nav-items-wrapper" ref="navItemsWrapper">
        <div class="nav-items" ref="navItems">
          <div
            v-for="item in allNavItems"
            :key="item.id"
            class="nav-item"
            :class="{
              active: isActive(item.path),
              disabled: item.requiresRegion && !isInRegion
            }"
            @click="navigateTo(item)"
          >
            <span class="nav-icon">{{ item.icon }}</span>
            <span class="nav-label">{{ item.name }}</span>
          </div>

          <!-- Stats button -->
          <div class="nav-item stats-btn" @click="toggleStatsOverlay">
            <span class="nav-icon">
              👤
              <span v-if="pendingCollections > 0" class="stats-badge"></span>
            </span>
            <span class="nav-label">Stats</span>
          </div>
        </div>
      </div>

      <!-- Right scroll arrow -->
      <button
        class="scroll-arrow right-arrow"
        :class="{ visible: canScrollRight }"
        @click="scrollRight"
        v-show="canScrollRight"
      >
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 18l6-6-6-6" />
        </svg>
      </button>
    </div>
  </div>

  <!-- Player Stats Overlay -->
  <transition name="slide-up">
    <div v-if="showStatsOverlay" class="player-stats-overlay">
      <div class="overlay-backdrop" @click="showStatsOverlay = false"></div>
      <div class="overlay-content">
        <button class="close-overlay" @click="showStatsOverlay = false">×</button>

        <div class="overlay-header">
          <h3>{{ playerName }}</h3>
          <div>{{ playerTitle }}</div>
        </div>

        <div class="player-attributes">
          <div class="attribute">
            <div class="attribute-label">
              <span class="icon">💰</span>
              <span>Money</span>
            </div>
            <div class="attribute-value">${{ formatNumber(playerMoney) }}</div>
          </div>

          <div class="attribute">
            <div class="attribute-label">
              <span class="icon">👥</span>
              <span>Crew</span>
            </div>
            <div class="attribute-value">{{ playerCrew }} / {{ maxCrew }}</div>
          </div>

          <div class="attribute">
            <div class="attribute-label">
              <span class="icon">🔫</span>
              <span>Weapons</span>
            </div>
            <div class="attribute-value">{{ playerWeapons }} / {{ maxWeapons }}</div>
          </div>

          <div class="attribute">
            <div class="attribute-label">
              <span class="icon">🚗</span>
              <span>Vehicles</span>
            </div>
            <div class="attribute-value">{{ playerVehicles }} / {{ maxVehicles }}</div>
          </div>
        </div>

        <div class="player-stats">
          <div class="stat">
            <div class="stat-label">Respect</div>
            <div class="progress-bar">
              <div class="progress-fill respect" :style="{ width: `${playerRespect}%` }"></div>
            </div>
            <div class="stat-value">{{ playerRespect }}%</div>
          </div>

          <div class="stat">
            <div class="stat-label">Influence</div>
            <div class="progress-bar">
              <div class="progress-fill influence" :style="{ width: `${playerInfluence}%` }"></div>
            </div>
            <div class="stat-value">{{ playerInfluence }}%</div>
          </div>

          <div class="stat">
            <div class="stat-label">Heat</div>
            <div class="progress-bar">
              <div class="progress-fill heat" :style="{ width: `${playerHeat}%` }"></div>
            </div>
            <div class="stat-value">{{ playerHeat }}%</div>
          </div>
        </div>

        <div class="territory-control">
          <h4>Territory Control</h4>
          <div class="control-stats">
            <div class="control-stat">
              <div class="control-label">Hotspots Controlled</div>
              <div class="control-value">{{ controlledHotspots }} / {{ totalHotspots }}</div>
            </div>
            <div class="control-stat">
              <div class="control-label">Revenue per Hour</div>
              <div class="control-value">${{ formatNumber(hourlyRevenue) }}</div>
            </div>
          </div>

          <div v-if="pendingCollections > 0" class="sidebar-actions">
            <button class="action-btn collect-all" @click="collectAllPending"
              :disabled="pendingCollections <= 0 || isLoading">
              <div style="display: flex; flex-direction: column; gap: 8px; align-items: center;">
                <div class="icon">💼</div>
                <div>Collect All</div>
                <div>${{ formatNumber(pendingCollections) }}</div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { useTravelStore } from '@/stores/modules/travel';
import { navigationConfig } from '@/config/navigationConfig';

const router = useRouter();
const route = useRoute();
const playerStore = usePlayerStore();
const territoryStore = useTerritoryStore();
const travelStore = useTravelStore();

// Get ALL navigation items (both primary and secondary) for mobile
const allNavItems = computed(() => {
  return navigationConfig
    .sort((a, b) => (a.priority || 0) - (b.priority || 0));
});

// Stats overlay state
const showStatsOverlay = ref(false);

// Scroll refs and state
const navContainer = ref<HTMLElement | null>(null);
const navItemsWrapper = ref<HTMLElement | null>(null);
const navItems = ref<HTMLElement | null>(null);
const canScrollLeft = ref(false);
const canScrollRight = ref(false);

// Access the travel store
const currentRegion = computed(() => travelStore.currentRegion);
const isInRegion = computed(() => !!currentRegion.value);

// Get player basic info from player store
const playerName = computed(() => playerStore.profile?.name || 'Unnamed Boss');
const playerTitle = computed(() => playerStore.profile?.title || 'Associate');

// Get player resources from player store
const playerMoney = computed(() => playerStore.playerMoney);
const playerCrew = computed(() => playerStore.playerCrew);
const playerWeapons = computed(() => playerStore.playerWeapons);
const playerVehicles = computed(() => playerStore.playerVehicles);

// Get player max resources from player store
const maxCrew = computed(() => playerStore.profile?.maxCrew || 0);
const maxWeapons = computed(() => playerStore.profile?.maxWeapons || 0);
const maxVehicles = computed(() => playerStore.profile?.maxVehicles || 0);

// Get player stats from player store
const playerRespect = computed(() => playerStore.playerRespect);
const playerInfluence = computed(() => playerStore.playerInfluence);
const playerHeat = computed(() => playerStore.playerHeat);

// Get territory control data from player store
const controlledHotspots = computed(() => playerStore.profile?.controlledHotspots || 0);
const totalHotspots = computed(() => playerStore.profile?.totalHotspotCount || 0);
const hourlyRevenue = computed(() => playerStore.profile?.hourlyRevenue || 0);
const pendingCollections = computed(() => playerStore.profile?.pendingCollections || 0);

// Get loading state
const isLoading = computed(() => playerStore.isLoading);

// Check if a menu item is active
function isActive(path: string): boolean {
  return route.path === path;
}

// Navigate to a path, respecting region requirements
function navigateTo(item: { path: string, requiresRegion: boolean }): void {
  if (item.requiresRegion && !isInRegion.value) {
    // If requires region but player has no region, go to travel view
    router.push({
      path: '/travel',
      query: {
        returnTo: item.path,
        message: 'You need to travel to a region first'
      }
    });
  } else {
    router.push(item.path);

    // Hide stats overlay if open
    showStatsOverlay.value = false;
  }
}

// Toggle stats overlay
function toggleStatsOverlay(): void {
  showStatsOverlay.value = !showStatsOverlay.value;
}

// Scroll functions
function updateScrollArrows(): void {
  if (!navItemsWrapper.value || !navItems.value) return;

  const wrapper = navItemsWrapper.value;
  const scrollLeft = wrapper.scrollLeft;
  const scrollWidth = wrapper.scrollWidth;
  const clientWidth = wrapper.clientWidth;

  canScrollLeft.value = scrollLeft > 10;
  canScrollRight.value = scrollLeft < scrollWidth - clientWidth - 10;
}

function scrollLeft(): void {
  if (!navItemsWrapper.value) return;
  const itemWidth = 70; // Approximate width of each nav item
  navItemsWrapper.value.scrollBy({ left: -itemWidth * 3, behavior: 'smooth' });
}

function scrollRight(): void {
  if (!navItemsWrapper.value) return;
  const itemWidth = 70; // Approximate width of each nav item
  navItemsWrapper.value.scrollBy({ left: itemWidth * 3, behavior: 'smooth' });
}

// Collect all pending collections
const collectAllPending = async () => {
  if (pendingCollections.value > 0 && !isLoading.value) {
    await territoryStore.collectAllHotspotIncome();
  }
};

// Helper function to format numbers
function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M';
  } else if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K';
  }
  return value.toString();
}

// Lifecycle hooks
onMounted(() => {
  if (navItemsWrapper.value) {
    navItemsWrapper.value.addEventListener('scroll', updateScrollArrows);
    window.addEventListener('resize', updateScrollArrows);
    // Initial check after DOM is fully loaded
    setTimeout(updateScrollArrows, 200);
  }
});

onBeforeUnmount(() => {
  if (navItemsWrapper.value) {
    navItemsWrapper.value.removeEventListener('scroll', updateScrollArrows);
  }
  window.removeEventListener('resize', updateScrollArrows);
});

// Watch for changes that might affect scrollability
watch([allNavItems], () => {
  setTimeout(updateScrollArrows, 200);
});
</script>

<style lang="scss">
// Styles for the mobile navbar
.mobile-nav-bar {
  display: flex;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 60px;
  background-color: $background-lighter;
  border-top: 1px solid $border-color;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.2);
  z-index: 100;

  .nav-items-container {
    display: flex;
    width: 100%;
    position: relative;
    overflow: hidden;

    .scroll-arrow {
      position: absolute;
      top: 50%;
      transform: translateY(-50%);
      width: 36px;
      height: 50px;
      background: none;
      border: none;
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 10;
      color: $text-color;
      cursor: pointer;
      opacity: 0;
      transition: opacity 0.3s ease;
      padding: 0;

      &.visible {
        opacity: 1;
      }

      &.left-arrow {
        left: 0;
        background: linear-gradient(to right, $background-lighter, transparent);
        padding-right: 12px;
      }

      &.right-arrow {
        right: 0;
        background: linear-gradient(to left, $background-lighter, transparent);
        padding-left: 12px;
      }

      &:hover {
        opacity: 1 !important;
        color: $secondary-color;
      }

      svg {
        filter: drop-shadow(0 0 4px rgba(0, 0, 0, 0.5));
      }
    }

    .nav-items-wrapper {
      flex: 1;
      overflow-x: auto;
      overflow-y: hidden;
      scroll-behavior: smooth;
      scrollbar-width: none; /* Firefox */
      -ms-overflow-style: none; /* IE and Edge */

      &::-webkit-scrollbar {
        display: none; /* Chrome, Safari, and Opera */
      }

      .nav-items {
        display: flex;
        min-width: max-content;
        white-space: nowrap;
        padding: 0 40px; /* Space for arrows */

        .nav-item {
          flex: 0 0 auto;
          min-width: 65px;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: $spacing-xs $spacing-sm;
          color: $text-secondary;
          text-decoration: none;
          position: relative;
          transition: $transition-base;
          cursor: pointer;

          &.active {
            color: $secondary-color;

            &:after {
              content: '';
              position: absolute;
              bottom: 0;
              width: 40%;
              height: 2px;
              background-color: $secondary-color;
              box-shadow: 0 0 8px rgba($secondary-color, 0.5);
            }
          }

          &.disabled {
            opacity: 0.6;
          }

          // Stats button styling
          &.stats-btn {
            .nav-icon {
              position: relative;

              .stats-badge {
                position: absolute;
                top: -5px;
                right: -5px;
                width: 14px;
                height: 14px;
                background-color: $secondary-color;
                border-radius: 50%;
              }
            }
          }

          .nav-icon {
            font-size: 20px;
            margin-bottom: 2px;
          }

          .nav-label {
            font-size: 10px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
          }
        }
      }
    }
  }

  @include respond-to(md) {
    display: none; // Hide on desktop
  }
}

// Player stats overlay styles remain the same as before
.player-stats-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 60px;
  z-index: 101;

  .overlay-backdrop {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
  }

  // ... rest of the stats overlay styles ...
}

// Slide up animations
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  .overlay-content {
    transform: translateY(100%);
  }
  .overlay-backdrop {
    opacity: 0;
  }
}

.slide-up-enter-to,
.slide-up-leave-from {
  .overlay-content {
    transform: translateY(0);
  }
  .overlay-backdrop {
    opacity: 1;
  }
}

// Enhance touch targets for mobile
@media (hover: none) {
  .mobile-nav-bar {
    .nav-item {
      min-width: 70px;
      padding: $spacing-sm;

      &:active {
        background-color: rgba(255, 255, 255, 0.1);
      }
    }

    .scroll-arrow {
      &:active {
        background-color: rgba(255, 255, 255, 0.1);
      }
    }
  }
}
</style>
