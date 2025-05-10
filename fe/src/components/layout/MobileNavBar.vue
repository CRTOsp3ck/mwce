// src/components/layout/MobileNavBar.vue

<template>
  <div class="mobile-nav-bar">
    <div class="nav-items">
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
      </div>

      <!-- Player Stats Button -->
      <div class="nav-item stats-btn" @click="toggleStatsOverlay">
        <span class="nav-icon">
          👤
          <span v-if="pendingCollections > 0" class="stats-badge"></span>
        </span>
        <span class="nav-label">Stats</span>
      </div>
    </div>
  </div>

  <!-- Player Stats Overlay -->
  <div class="player-stats-overlay" :class="{ open: showStatsOverlay }">
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
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { useTravelStore } from '@/stores/modules/travel';

const router = useRouter();
const route = useRoute();
const playerStore = usePlayerStore();
const territoryStore = useTerritoryStore();
const travelStore = useTravelStore();

// Stats overlay state
const showStatsOverlay = ref(false);

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

// Navigation menu items with region requirements - mobile optimized with icons
const menuItems = computed(() => [
  { path: '/', name: 'Home', icon: '🏠', requiresRegion: false },
  { path: '/territory', name: 'Territory', icon: '🏙️', requiresRegion: true },
  { path: '/operations', name: 'Ops', icon: '🎯', requiresRegion: true },
  { path: '/market', name: 'Market', icon: '💹', requiresRegion: true },
  { path: '/travel', name: 'Travel', icon: '✈️', requiresRegion: false },
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

    // Hide stats overlay if open
    showStatsOverlay.value = false;
  }
}

// Toggle stats overlay
function toggleStatsOverlay(): void {
  showStatsOverlay.value = !showStatsOverlay.value;
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
</script>

<style lang="scss">
// Styles are in _responsive.scss
</style>
