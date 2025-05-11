<template>
  <div class="mobile-nav-bar">
    <div class="nav-items-container" ref="navContainer">
      <!-- Left scroll arrow -->
      <button class="scroll-arrow left-arrow" :class="{ visible: canScrollLeft }" @click="scrollLeft"
        v-show="canScrollLeft">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 18l-6-6 6-6" />
        </svg>
      </button>

      <!-- Navigation items with horizontal scroll -->
      <div class="nav-items-wrapper" ref="navItemsWrapper">
        <div class="nav-items" ref="navItems">
          <div v-for="item in allNavItems" :key="item.id" class="nav-item" :class="{
            active: isActive(item.path),
            disabled: item.requiresRegion && !isInRegion
          }" @click="navigateTo(item)">
            <span class="nav-icon">{{ item.icon }}</span>
            <span class="nav-label">{{ item.name }}</span>
          </div>

          <!-- Stats button -->
          <div class="nav-item stats-btn" @click="toggleStatsOverlay">
            <span class="nav-icon">
              👤
              <span v-if="currentRegionPendingCollections > 0" class="stats-badge"></span>
            </span>
            <span class="nav-label">Stats</span>
          </div>
        </div>
      </div>

      <!-- Right scroll arrow -->
      <button class="scroll-arrow right-arrow" :class="{ visible: canScrollRight }" @click="scrollRight"
        v-show="canScrollRight">
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

          <!-- Always show the button but keep it stateful -->
          <div class="sidebar-actions">
            <button class="action-btn collect-all" :class="{
              'has-pending': isInRegion && currentRegionPendingCollections > 0,
              'no-pending': isInRegion && currentRegionPendingCollections <= 0,
              'no-region': !isInRegion
            }" @click="collectAllPending"
              :disabled="!isInRegion || currentRegionPendingCollections <= 0 || isLoading">
              <div style="display: flex; flex-direction: column; gap: 8px; align-items: center;">
                <div class="icon">💼</div>
                <div>Collect All</div>
                <div style="display: flex; flex-direction: column; gap: 2px;">
                  <div v-if="!isInRegion" style="font-size: 0.8em; opacity: 0.8;">
                    Travel to Region
                  </div>
                  <div v-else-if="currentRegionPendingCollections <= 0" style="font-size: 0.8em; opacity: 0.8;">
                    No Income Available
                  </div>
                  <div v-else>
                    <div>${{ formatNumber(currentRegionPendingCollections) }}</div>
                    <div style="font-size: 0.8em; opacity: 0.8;">({{ currentRegion?.name }})</div>
                  </div>
                </div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </transition>

  <!-- Collection Result Modal -->
  <transition name="fade">
    <div v-if="showCollectionResultModal" class="modal-overlay">
      <div class="modal-container collection-result-modal">
        <div class="modal-header">
          <h3>{{ collectionResult?.success ? 'Collection Successful' : 'Collection Failed' }}</h3>
          <button class="close-btn" @click="closeCollectionModal">×</button>
        </div>

        <div class="modal-body">
          <div v-if="collectionResult" class="collection-result"
            :class="{ 'success': collectionResult.success, 'failure': !collectionResult.success }">
            <div class="result-icon">
              {{ collectionResult.success ? '✅' : '❌' }}
            </div>

            <div class="result-message">
              {{ collectionResult.message }}
            </div>

            <div v-if="collectionResult.success && collectionResult.moneyGained" class="result-details">
              <div class="result-item positive">
                <span class="item-icon">💵</span>
                <span class="item-value">+${{ formatNumber(collectionResult.moneyGained) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="primary-btn" @click="closeCollectionModal">
            Continue
          </button>
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
import type { ActionResult } from '@/types/territory';

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

// Collection modal states
const showCollectionResultModal = ref(false);
const collectionResult = ref<ActionResult | null>(null);

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

// Get current region controlled hotspots for collection
const currentRegionControlledHotspots = computed(() => {
  return territoryStore.currentRegionControlledHotspots;
});

// Get pending collections specifically from current region
const currentRegionPendingCollections = computed(() => {
  return currentRegionControlledHotspots.value.reduce((total, hotspot) => total + hotspot.pendingCollection, 0);
});

// Get collectable businesses count for current region
const currentRegionCollectableBusinesses = computed(() => {
  return currentRegionControlledHotspots.value.filter(h => h.pendingCollection > 0).length;
});

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

// Updated collect all to use regional collection with modal
const collectAllPending = async () => {
  if (!isInRegion.value || currentRegionPendingCollections.value <= 0 || isLoading.value) return;

  try {
    const result = await territoryStore.collectAllHotspotIncomeInCurrentRegion();

    if (result) {
      // Show collection result modal
      collectionResult.value = {
        success: true,
        moneyGained: result.collectionResult.collectedAmount,
        message: result.gameMessage?.message ||
          `Successfully collected $${formatNumber(result.collectionResult.collectedAmount)} from ${result.collectionResult.hotspotsCount} businesses in ${currentRegion.value?.name || 'this region'}.`
      };

      showCollectionResultModal.value = true;
    }
  } catch (error) {
    console.error('Failed to collect all pending resources:', error);

    // Show error modal
    collectionResult.value = {
      success: false,
      message: 'Failed to collect resources. Please try again.'
    };

    showCollectionResultModal.value = true;
  }
};

// Close collection result modal
function closeCollectionModal() {
  showCollectionResultModal.value = false;
  collectionResult.value = null;
}

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
      scrollbar-width: none;
      /* Firefox */
      -ms-overflow-style: none;
      /* IE and Edge */

      &::-webkit-scrollbar {
        display: none;
        /* Chrome, Safari, and Opera */
      }

      .nav-items {
        display: flex;
        min-width: max-content;
        white-space: nowrap;
        padding: 0 40px;
        /* Space for arrows */

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
                animation: pulse 2s infinite;
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

// Player stats overlay styles
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

  .overlay-content {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    background-color: $background-card;
    border-top-left-radius: $border-radius-lg;
    border-top-right-radius: $border-radius-lg;
    max-height: 80vh;
    overflow-y: auto;
    padding: $spacing-md;
    border: 1px solid $border-color;
    border-bottom: none;

    .close-overlay {
      position: absolute;
      top: $spacing-md;
      right: $spacing-md;
      background: none;
      border: none;
      font-size: 24px;
      cursor: pointer;
      color: $text-secondary;
      width: 32px;
      height: 32px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      transition: $transition-base;

      &:hover {
        color: $text-color;
        background-color: rgba($text-color, 0.1);
      }
    }

    .overlay-header {
      text-align: center;
      margin-bottom: $spacing-lg;
      padding-top: $spacing-lg;

      h3 {
        margin: 0;
        @include gold-accent;
      }

      div {
        color: $text-secondary;
        font-size: $font-size-sm;
        margin-top: $spacing-xs;
      }
    }

    .player-attributes {
      @include flex-column;
      gap: $spacing-sm;
      margin-bottom: $spacing-lg;

      .attribute {
        @include flex-between;
        padding: $spacing-xs 0;

        .attribute-label {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          color: $text-secondary;

          .icon {
            font-size: 20px;
          }
        }

        .attribute-value {
          font-weight: 600;
        }
      }
    }

    .player-stats {
      @include flex-column;
      gap: $spacing-md;
      margin-bottom: $spacing-lg;

      .stat {
        .stat-label {
          @include flex-between;
          margin-bottom: $spacing-xs;
          font-size: $font-size-sm;
          color: $text-secondary;
        }

        .progress-bar {
          height: 8px;
          background-color: rgba(255, 255, 255, 0.1);
          border-radius: 4px;
          overflow: hidden;
          margin-bottom: 4px;

          .progress-fill {
            height: 100%;
            border-radius: 4px;
            transition: width 0.3s ease;

            &.respect {
              background-color: $success-color;
            }

            &.influence {
              background-color: $info-color;
            }

            &.heat {
              background-color: $danger-color;
            }
          }
        }

        .stat-value {
          text-align: right;
          font-size: $font-size-sm;
        }
      }
    }

    .territory-control {
      padding: $spacing-md 0;
      border-top: 1px solid $border-color;

      h4 {
        margin-top: 0;
        margin-bottom: $spacing-md;
        color: $secondary-color;
      }

      .control-stats {
        @include flex-column;
        gap: $spacing-sm;

        .control-stat {
          @include flex-between;

          .control-label {
            font-size: $font-size-sm;
            color: $text-secondary;
          }

          .control-value {
            font-weight: 600;
          }
        }
      }

      .sidebar-actions {
        margin-top: $spacing-md;
        margin-bottom: $spacing-md;

        .action-btn {
          @include button-base;
          width: 100%;
          padding: $spacing-md;
          position: relative;
          transition: all 0.2s ease;

          &.has-pending {
            background-color: $secondary-color;
            color: $background-color;

            &:hover:not(:disabled) {
              background-color: lighten($secondary-color, 5%);
            }

            .icon {
              animation: pulse 2s infinite;
            }
          }

          &.no-pending {
            background-color: rgba($text-secondary, 0.2);
            color: $text-secondary;
            cursor: not-allowed;
          }

          &.no-region {
            background-color: rgba($warning-color, 0.2);
            color: $warning-color;
            cursor: not-allowed;
          }

          .icon {
            margin-right: $spacing-sm;
          }

          &:active:not(:disabled) {
            background-color: darken($secondary-color, 5%);
          }

          &:disabled {
            opacity: 0.8;

            &:hover {
              transform: none;
              box-shadow: none;
            }
          }
        }
      }
    }
  }
}

// Add styles for mobile modal
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: $spacing-md;
}

.modal-container {
  background-color: $background-card;
  border-radius: $border-radius-md;
  max-width: 90vw;
  max-height: 80vh;
  overflow-y: auto;
  border: 1px solid $border-color;

  .modal-header {
    padding: $spacing-md;
    border-bottom: 1px solid $border-color;
    display: flex;
    justify-content: space-between;
    align-items: center;

    h3 {
      margin: 0;
      @include gold-accent;
    }

    .close-btn {
      background: none;
      border: none;
      font-size: 24px;
      cursor: pointer;
      color: $text-secondary;

      &:hover {
        color: $text-color;
      }
    }
  }

  .modal-body {
    padding: $spacing-md;
  }

  .modal-footer {
    padding: $spacing-md;
    border-top: 1px solid $border-color;
    display: flex;
    justify-content: flex-end;

    .primary-btn {
      @include button-base;
      background-color: $primary-color;
      color: $text-color;
      padding: $spacing-sm $spacing-md;

      &:hover {
        background-color: lighten($primary-color, 10%);
      }
    }
  }
}

.collection-result-modal {
  .collection-result {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: $spacing-md;
    padding: $spacing-lg;
    border-radius: $border-radius-md;

    &.success {
      background-color: rgba($success-color, 0.1);
      border: 1px solid rgba($success-color, 0.2);
    }

    &.failure {
      background-color: rgba($danger-color, 0.1);
      border: 1px solid rgba($danger-color, 0.2);
    }

    .result-icon {
      font-size: 48px;
      margin-bottom: $spacing-sm;
    }

    .result-message {
      font-size: $font-size-lg;
      font-weight: 500;
      margin-bottom: $spacing-md;
      line-height: 1.4;
    }

    .result-details {
      display: flex;
      flex-direction: column;
      gap: $spacing-sm;
      width: 100%;

      .result-item {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: $spacing-sm;
        padding: $spacing-sm $spacing-md;
        background-color: rgba($background-lighter, 0.5);
        border-radius: $border-radius-sm;

        &.positive {
          color: $success-color;
          border: 1px solid rgba($success-color, 0.2);
        }

        .item-icon {
          font-size: 18px;
        }

        .item-value {
          font-weight: 600;
          font-size: $font-size-lg;
        }
      }
    }
  }
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

// Fade transition for modal
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
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

@keyframes pulse {
  0% {
    opacity: 1;
  }

  50% {
    opacity: 0.7;
  }

  100% {
    opacity: 1;
  }
}
</style>
