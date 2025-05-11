<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseTooltip from '@/components/ui/BaseTooltip.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { useTravelStore } from '@/stores/modules/travel';
import { getSidebarNavItems } from '@/config/navigationConfig';
import type { ActionResult } from '@/types/territory';

const router = useRouter();
const route = useRoute();
const playerStore = usePlayerStore();
const territoryStore = useTerritoryStore();

// Access the travel store
const travelStore = useTravelStore();

// Collection modal states
const showCollectionResultModal = ref(false);
const collectionResult = ref<ActionResult | null>(null);
const isCollecting = ref(false);

// Get navigation items from central config
const menuItems = computed(() => getSidebarNavItems());

// Check if the player is in a region
const currentRegion = computed(() => travelStore.currentRegion);
const isInRegion = computed(() => !!currentRegion.value);

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

// Get all pending collections for display when not in a region
const totalPendingCollections = computed(() => playerStore.profile?.pendingCollections || 0);

// Load player data on component mount
onMounted(async () => {
  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }
});

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

// Use regional pending when in a region, otherwise use total
const pendingCollections = computed(() => {
  return isInRegion.value ? currentRegionPendingCollections.value : totalPendingCollections.value;
});

// Get loading state
const isLoading = computed(() => playerStore.isLoading);

// Helper function to format numbers
function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M';
  } else if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K';
  }
  return value.toString();
}

// Check if a menu item is active
function isActive(path: string): boolean {
  return route.path === path;
}

// Navigate to a path, respecting region requirements
function navigateTo(item: any): void {
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
  }
}

// Updated collect all with modal and regional support
async function collectAllPending() {
  if (pendingCollections.value <= 0 || isLoading.value) return;

  isCollecting.value = true;

  try {
    let result;
    const currentRegionName = travelStore.currentRegion?.name;

    if (isInRegion.value) {
      // Collect from current region
      result = await territoryStore.collectAllHotspotIncomeInCurrentRegion();
    } else {
      // Collect from all territories
      result = await territoryStore.collectAllHotspotIncome();
    }

    if (result) {
      // Show collection result modal
      collectionResult.value = {
        success: true,
        moneyGained: result.collectionResult.collectedAmount,
        message: result.gameMessage?.message ||
          `Successfully collected $${formatNumber(result.collectionResult.collectedAmount)} from ${result.collectionResult.hotspotsCount} businesses${currentRegionName ? ` in ${currentRegionName}` : ''}.`
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
  } finally {
    isCollecting.value = false;
  }
}

// Close collection result modal
function closeCollectionModal() {
  showCollectionResultModal.value = false;
  collectionResult.value = null;
}
</script>

<template>
  <aside class="app-sidebar">
    <div class="player-profile">
      <div class="player-avatar">
        <!-- <img src="@/assets/images/avatar-placeholder.png" alt="Player Avatar" /> -->
        <div class="player-level">{{ playerTitle }}</div>
      </div>
      <h3 class="player-name">{{ playerName }}</h3>
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
    </div>

    <nav class="sidebar-nav">
      <BaseTooltip v-for="item in menuItems" :key="item.id" :text="item.tooltip" position="right">
        <div class="nav-item" :class="{
          active: isActive(item.path),
          disabled: item.requiresRegion && !isInRegion
        }" @click="navigateTo(item)">
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-label">{{ item.name }}</span>
          <span v-if="item.requiresRegion && !isInRegion" class="region-required"
            title="You need to travel to a region first">🌎</span>
        </div>
      </BaseTooltip>
    </nav>

    <div class="sidebar-actions">
      <button class="action-btn collect-all" @click="collectAllPending"
        :disabled="pendingCollections <= 0 || isLoading || isCollecting">
        <BaseTooltip
          :text="isInRegion ? `Collect all pending income from businesses in ${currentRegion?.name}` : 'Collect all pending income from your territories'">
          <div style="display: flex; flex-direction: column; gap: 8px; align-items: center;">
            <div class="icon">💼</div>
            <div>{{ isInRegion ? 'Collect All' : 'Collect All' }}</div>
            <div style="display: flex; flex-direction: column; gap: 2px;">
              <div>${{ formatNumber(pendingCollections) }}</div>
              <div v-if="isInRegion" style="font-size: 0.8em; opacity: 0.8;">
                ({{ currentRegion?.name }})
              </div>
              <div v-else-if="!isInRegion && controlledHotspots > 0" style="font-size: 0.8em; opacity: 0.8;">
                (All Regions)
              </div>
            </div>
          </div>
        </BaseTooltip>
      </button>
    </div>

    <!-- Collection Result Modal -->
    <BaseModal v-model="showCollectionResultModal"
      :title="collectionResult?.success ? 'Collection Successful' : 'Collection Failed'"
      class="collection-result-modal">
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

      <template #footer>
        <BaseButton variant="primary" @click="closeCollectionModal">
          Continue
        </BaseButton>
      </template>
    </BaseModal>
  </aside>
</template>

<style lang="scss">
.app-sidebar {
  width: 280px;
  background-color: $background-lighter;
  border-right: 1px solid $border-color;
  padding: $spacing-md;
  @include flex-column;
  gap: $spacing-lg;
  height: 100%;
  overflow-y: auto;

  .player-profile {
    @include flex-column;
    align-items: center;
    text-align: center;
    padding-bottom: $spacing-md;
    border-bottom: 1px solid $border-color;

    .player-avatar {
      position: relative;
      margin-bottom: $spacing-sm;

      img {
        width: 80px;
        height: 80px;
        border-radius: 50%;
        object-fit: cover;
        border: 2px solid $secondary-color;
        box-shadow: 0 0 10px rgba($secondary-color, 0.5);
      }

      .player-level {
        background-color: $primary-color;
        color: white;
        font-size: 12px;
        padding: 2px 8px;
        border-radius: 10px;
        border: 1px solid $secondary-color;
      }
    }

    .player-name {
      margin: 0;
      @include gold-accent;
    }
  }

  .player-attributes {
    @include flex-column;
    gap: $spacing-sm;

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
    border-bottom: 1px solid $border-color;

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
  }

  /* Updated navigation menu styling with tooltip support */
  .sidebar-nav {
    @include flex-column;
    gap: $spacing-xs;
    margin: $spacing-md 0;

    .nav-item {
      display: flex;
      align-items: center;
      gap: $spacing-md;
      padding: $spacing-sm $spacing-md;
      border-radius: $border-radius-sm;
      cursor: pointer;
      transition: background-color 0.2s ease;
      position: relative;

      &:hover {
        background-color: rgba($background-lighter, 0.2);
      }

      &.active {
        background-color: rgba($primary-color, 0.2);

        &:before {
          content: '';
          position: absolute;
          left: 0;
          top: 0;
          bottom: 0;
          width: 3px;
          background-color: $primary-color;
        }
      }

      &.disabled {
        opacity: 0.6;
        cursor: pointer; // Still clickable, but will redirect to travel

        &:hover {
          background-color: rgba($warning-color, 0.1);
        }
      }

      .nav-icon {
        font-size: 20px;
        width: 24px;
        text-align: center;
      }

      .nav-label {
        flex: 1;
      }

      .region-required {
        font-size: $font-size-sm;
        color: $warning-color;
      }
    }
  }

  .sidebar-actions {
    margin-top: auto;
    padding-top: $spacing-md;

    .action-btn {
      @include button-base;
      width: 100%;
      background-color: $secondary-color;
      color: $background-color;
      padding: $spacing-md;
      position: relative;
      transition: all 0.2s ease;

      .icon {
        margin-right: $spacing-sm;
      }

      &:hover:not(:disabled) {
        background-color: lighten($secondary-color, 5%);
        transform: translateY(-1px);
        box-shadow: 0 4px 8px rgba($secondary-color, 0.3);
      }

      &:active:not(:disabled) {
        background-color: darken($secondary-color, 5%);
        transform: translateY(0);
      }

      &:disabled {
        opacity: 0.6;
        cursor: not-allowed;

        &:hover {
          background-color: $secondary-color;
          transform: none;
          box-shadow: none;
        }
      }
    }
  }
}

// Add styles for the collection result modal
.collection-result-modal {
  .collection-result {
    @include flex-column;
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
</style>
