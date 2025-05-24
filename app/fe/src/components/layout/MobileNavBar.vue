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

// Function to get status level based on absolute values
function getStatusLevel(value: number, type: 'respect' | 'influence' | 'heat'): string {
  // These thresholds would be adjusted based on your game design
  if (type === 'heat') {
    // Heat is bad, so higher values mean more dangerous status
    if (value >= 800) return 'critical';
    if (value >= 600) return 'high';
    if (value >= 300) return 'moderate';
    if (value >= 100) return 'low';
    return 'minimal';
  } else {
    // Respect and influence are good, so higher values mean better status
    if (value >= 800) return 'legendary';
    if (value >= 600) return 'elite';
    if (value >= 300) return 'established';
    if (value >= 100) return 'rising';
    return 'newcomer';
  }
}

// Function to get the number of status pips to display (out of 5)
function getStatusPips(value: number, type: 'respect' | 'influence' | 'heat'): number {
  // Maps status levels to a number of pips (0-5)
  const statusLevel = getStatusLevel(value, type);

  if (type === 'heat') {
    switch(statusLevel) {
      case 'critical': return 5;
      case 'high': return 4;
      case 'moderate': return 3;
      case 'low': return 2;
      default: return 1; // minimal
    }
  } else {
    switch(statusLevel) {
      case 'legendary': return 5;
      case 'elite': return 4;
      case 'established': return 3;
      case 'rising': return 2;
      default: return 1; // newcomer
    }
  }
}

// Function to generate an array of status indicators
function generateStatusIndicators(value: number, type: 'respect' | 'influence' | 'heat'): number[] {
  const pips = getStatusPips(value, type);
  return Array.from({ length: 5 }, (_, index) => index < pips ? 1 : 0);
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
          <!-- Respect with diamond ranking -->
          <div class="stat">
            <div class="stat-header">
              <span class="stat-name">Respect</span>
              <span class="stat-value">{{ playerRespect }}</span>
            </div>
            <div class="stat-indicator respect" :class="getStatusLevel(playerRespect, 'respect')">
              <div class="status-diamonds">
                <div v-for="(active, index) in generateStatusIndicators(playerRespect, 'respect')"
                     :key="index"
                     class="diamond"
                     :class="{ active: active === 1 }"></div>
              </div>
              <div class="status-level">{{ getStatusLevel(playerRespect, 'respect') }}</div>
            </div>
          </div>

          <!-- Influence with diamond ranking -->
          <div class="stat">
            <div class="stat-header">
              <span class="stat-name">Influence</span>
              <span class="stat-value">{{ playerInfluence }}</span>
            </div>
            <div class="stat-indicator influence" :class="getStatusLevel(playerInfluence, 'influence')">
              <div class="status-diamonds">
                <div v-for="(active, index) in generateStatusIndicators(playerInfluence, 'influence')"
                     :key="index"
                     class="diamond"
                     :class="{ active: active === 1 }"></div>
              </div>
              <div class="status-level">{{ getStatusLevel(playerInfluence, 'influence') }}</div>
            </div>
          </div>

          <!-- Heat with flame indicators -->
          <div class="stat">
            <div class="stat-header">
              <span class="stat-name">Heat</span>
              <span class="stat-value">{{ playerHeat }}</span>
            </div>
            <div class="stat-indicator heat" :class="getStatusLevel(playerHeat, 'heat')">
              <div class="status-flames">
                <div v-for="(active, index) in generateStatusIndicators(playerHeat, 'heat')"
                     :key="index"
                     class="flame"
                     :class="{ active: active === 1 }">
                  <div class="flame-inner"></div>
                </div>
              </div>
              <div class="status-level">{{ getStatusLevel(playerHeat, 'heat') }}</div>
            </div>
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
        .stat-header {
          @include flex-between;
          margin-bottom: $spacing-xs;

          .stat-name {
            font-size: $font-size-sm;
            color: $text-secondary;
          }

          .stat-value {
            font-weight: 600;
            color: $text-color;
          }
        }

        .stat-indicator {
          display: flex;
          flex-direction: column;
          gap: 4px;

          .status-diamonds {
            display: flex;
            gap: 4px;
            // justify-content: space-between;

            .diamond {
              width: 16px;
              height: 16px;
              background-color: rgba($text-secondary, 0.2);
              clip-path: polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%);
              transition: all 0.3s ease;
              position: relative;

              &.active {
                background-color: $text-color;
              }
            }
          }

          .status-flames {
            display: flex;
            gap: 4px;
            // justify-content: space-between;

            .flame {
              width: 16px;
              height: 16px;
              position: relative;
              background-color: rgba($text-secondary, 0.2);
              clip-path: polygon(50% 0%, 80% 30%, 100% 60%, 80% 100%, 20% 100%, 0% 60%, 20% 30%);
              transition: all 0.3s ease;

              .flame-inner {
                position: absolute;
                top: 3px;
                left: 3px;
                right: 3px;
                bottom: 3px;
                background-color: $background-lighter;
                clip-path: polygon(50% 10%, 75% 35%, 90% 60%, 75% 90%, 25% 90%, 10% 60%, 25% 35%);
                transition: all 0.3s ease;
              }

              &.active {
                background-color: $danger-color;

                .flame-inner {
                  background-color: lighten($danger-color, 15%);
                }

                &:nth-child(5) {
                  animation: flicker 2s infinite alternate;
                }

                &:nth-child(4) {
                  animation: flicker 3s 0.3s infinite alternate;
                }

                &:nth-child(3) {
                  animation: flicker 2.5s 0.6s infinite alternate;
                }

                &:nth-child(2) {
                  animation: flicker 2.7s 0.9s infinite alternate;
                }

                &:nth-child(1) {
                  animation: flicker 3.2s 1.2s infinite alternate;
                }
              }
            }
          }

          .status-level {
            text-align: right;
            font-size: 10px;
            text-transform: uppercase;
            font-weight: 600;
            letter-spacing: 0.5px;
            color: $text-secondary;
          }

          // Respect indicator styling
          &.respect {
            .status-diamonds .diamond.active {
              background-color: $success-color;
            }

            &.legendary {
              .status-diamonds .diamond.active {
                background-color: $success-color;
                box-shadow: 0 0 5px $success-color;
                animation: pulse 2s infinite alternate;
              }

              .status-level {
                color: $success-color;
                text-shadow: 0 0 5px rgba($success-color, 0.5);
              }
            }

            &.elite {
              .status-diamonds .diamond.active {
                background-color: lighten($success-color, 10%);
              }

              .status-level {
                color: lighten($success-color, 10%);
              }
            }

            &.established {
              .status-diamonds .diamond.active {
                background-color: lighten($success-color, 20%);
              }

              .status-level {
                color: $text-color;
              }
            }
          }

          // Influence indicator styling
          &.influence {
            .status-diamonds .diamond.active {
              background-color: $info-color;
            }

            &.legendary {
              .status-diamonds .diamond.active {
                background-color: $info-color;
                box-shadow: 0 0 5px $info-color;
                animation: pulse 2s infinite alternate;
              }

              .status-level {
                color: $info-color;
                text-shadow: 0 0 5px rgba($info-color, 0.5);
              }
            }

            &.elite {
              .status-diamonds .diamond.active {
                background-color: lighten($info-color, 10%);
              }

              .status-level {
                color: lighten($info-color, 10%);
              }
            }

            &.established {
              .status-diamonds .diamond.active {
                background-color: lighten($info-color, 20%);
              }

              .status-level {
                color: $text-color;
              }
            }
          }

          // Heat indicator styling
          &.heat {
            &.critical {
              .status-flames .flame.active {
                background-color: $danger-color;
                box-shadow: 0 0 5px $danger-color;
              }

              .status-level {
                color: $danger-color;
                text-shadow: 0 0 5px rgba($danger-color, 0.5);
                animation: pulse 1.5s infinite;
              }
            }

            &.high {
              .status-flames .flame.active {
                background-color: lighten($danger-color, 10%);
              }

              .status-level {
                color: lighten($danger-color, 10%);
              }
            }

            &.moderate {
              .status-flames .flame.active {
                background-color: lighten($danger-color, 20%);
              }

              .status-level {
                color: $text-color;
              }
            }
          }
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

@keyframes flicker {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}
</style>
