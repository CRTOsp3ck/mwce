// src/components/layout/AppSidebar.vue

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { useTravelStore } from '@/stores/modules/travel';

const router = useRouter();
const route = useRoute();
const playerStore = usePlayerStore();
const territoryStore = useTerritoryStore();

// Access the travel store
const travelStore = useTravelStore();

// Check if the player is in a region
const currentRegion = computed(() => travelStore.currentRegion);
const isInRegion = computed(() => !!currentRegion.value);

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
const pendingCollections = computed(() => playerStore.profile?.pendingCollections || 0);

// Get loading state
const isLoading = computed(() => playerStore.isLoading);

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

// Navigation menu items with region requirements
const menuItems = computed(() => [
  { path: '/', name: 'Dashboard', icon: '📊', requiresRegion: false },
  { path: '/campaigns', name: 'Campaign', icon: '📜', requiresRegion: true },
  { path: '/travel', name: 'Travel Agency', icon: '✈️', requiresRegion: false },
  { path: '/territory', name: 'Territory', icon: '🏙️', requiresRegion: true },
  { path: '/operations', name: 'Operations', icon: '🎯', requiresRegion: true },
  { path: '/market', name: 'Market', icon: '💹', requiresRegion: true },
  { path: '/rankings', name: 'Rankings', icon: '🏆', requiresRegion: false },
  { path: '/nft', name: 'NFT', icon: '💎', requiresRegion: false },
]);

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
  }
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

    <!-- Updated Navigation Menu -->
    <nav class="sidebar-nav">
      <div
        v-for="item in menuItems"
        :key="item.path"
        class="nav-item"
        :class="{
          active: isActive(item.path),
          disabled: item.requiresRegion && !isInRegion
        }"
        @click="navigateTo(item)"
      >
        <span class="nav-icon">{{ item.icon }}</span>
        <span class="nav-label">{{ item.name }}</span>
        <span
          v-if="item.requiresRegion && !isInRegion"
          class="region-required"
          title="You need to travel to a region first"
        >🌎</span>
      </div>
    </nav>

    <div class="sidebar-actions">
      <button class="action-btn collect-all" @click="collectAllPending"
        :disabled="pendingCollections <= 0 || isLoading">
        <div style="display: flex; flex-direction: column; gap: 8px; align-items: center;">
          <div class="icon">💼</div>
          <div>Collect All</div>
          <div style="display: flex; flex-direction: column; gap: 2px;">
            <div>${{ formatNumber(pendingCollections) }}</div>
          </div>
        </div>
      </button>
    </div>
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

  /* Updated navigation menu styling */
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

      .icon {
        margin-right: $spacing-sm;
      }

      .amount {
        position: absolute;
        right: $spacing-md;
        font-weight: bold;
      }

      &:hover {
        background-color: lighten($secondary-color, 5%);
      }

      &:active {
        background-color: darken($secondary-color, 5%);
      }

      &:disabled {
        opacity: 0.6;
        cursor: not-allowed;

        &:hover {
          background-color: $secondary-color;
        }
      }
    }
  }
}
</style>
