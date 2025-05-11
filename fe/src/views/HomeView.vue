// src/views/HomeView.vue

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { useOperationsStore } from '@/stores/modules/operations';
import { useTravelStore } from '@/stores/modules/travel';
import { ResourceType } from '@/types/market';
import { TerritoryActionType } from '@/types/territory';
import { OperationType } from '@/types/operations';

const router = useRouter();
const playerStore = usePlayerStore();
const territoryStore = useTerritoryStore();
const operationsStore = useOperationsStore();
const travelStore = useTravelStore();

const isLoading = ref(false);

// Computed properties from stores
const playerName = computed(() => playerStore.profile?.name || 'Boss');
const playerMoney = computed(() => playerStore.playerMoney);
const playerCrew = computed(() => playerStore.playerCrew);
const playerWeapons = computed(() => playerStore.playerWeapons);
const playerVehicles = computed(() => playerStore.playerVehicles);
const playerTitle = computed(() => playerStore.playerTitle);

// Location-based properties
const currentRegion = computed(() => playerStore.profile?.currentRegionName || null);
const isInRegion = computed(() => !!playerStore.profile?.currentRegionId);

// Regional stats
const regionalControlled = computed(() => playerStore.profile?.regionalControlled || 0);
const regionalTotal = computed(() => playerStore.profile?.regionalTotalHotspots || 0);
const regionalRevenue = computed(() => playerStore.profile?.regionalRevenue || 0);
const regionalPending = computed(() => playerStore.profile?.regionalPending || 0);
const regionalControlPercentage = computed(() => {
  if (!regionalTotal.value) return 0;
  return Math.round((regionalControlled.value / regionalTotal.value) * 100);
});

// Overall stats (when not in region or for comparison)
const totalControlled = computed(() => playerStore.controlledHotspots);
const totalHotspots = computed(() => playerStore.totalHotspots);
const totalRevenue = computed(() => playerStore.hourlyRevenue);
const totalPending = computed(() => playerStore.pendingCollections);

// Region-specific greetings and descriptions
const regionGreeting = computed(() => {
  if (!currentRegion.value) {
    return {
      greeting: `Welcome back to headquarters, ${playerTitle.value} ${playerName.value}`,
      description: 'Your empire spans across the city, but every criminal mastermind needs a base. From here, you coordinate your operations across all regions.',
      atmosphere: 'headquarters'
    };
  }

  // Customize greeting based on region (this would be expanded with actual region data)
  const regionData = getRegionData(currentRegion.value);
  return {
    greeting: `${regionData.greeting}, ${playerTitle.value} ${playerName.value}`,
    description: regionData.description,
    atmosphere: regionData.atmosphere
  };
});

// Get region-specific data (this would be expanded with actual region characteristics)
function getRegionData(regionName: string) {
  const regions: Record<string, { greeting: string; description: string; atmosphere: string }> = {
    'Downtown': {
      greeting: 'The concrete jungle awaits',
      description: 'The beating heart of the city pulses around you. Skyscrapers cast long shadows over your operations, while corporate executives rub shoulders with street hustlers. Every corner presents a new opportunity.',
      atmosphere: 'urban'
    },
    'Industrial District': {
      greeting: 'Steel and shadows welcome you',
      description: 'Factories belch smoke into the gray sky, masking your clandestine activities. The constant hum of machinery provides cover for your operations among the warehouses and loading docks.',
      atmosphere: 'industrial'
    },
    'Harbor District': {
      greeting: 'Salt air carries the scent of opportunity',
      description: 'Ships from across the globe dock here, bringing both legitimate cargo and illicit opportunities. The maze of shipping containers and waterfront warehouses offers perfect cover for your enterprises.',
      atmosphere: 'waterfront'
    },
    'Residential Quarter': {
      greeting: 'Quiet streets hide profitable secrets',
      description: 'Behind the peaceful facade of suburban homes lies a network of underground operations. Here, discretion is key, and your influence spreads through seemingly innocent businesses.',
      atmosphere: 'suburban'
    },
    'Entertainment District': {
      greeting: 'Neon lights illuminate your domain',
      description: 'Casinos, clubs, and bars line the streets, creating the perfect front for your criminal enterprises. The constant flow of tourists and locals provides endless opportunities for profit.',
      atmosphere: 'nightlife'
    }
  };

  return regions[regionName] || {
    greeting: `Welcome to ${regionName}`,
    description: 'This region offers unique opportunities for expansion and profit. Study the local landscape and plan your moves carefully.',
    atmosphere: 'general'
  };
}

// Get recent actions from the territory store
const recentActions = computed(() => territoryStore.recentActions.slice(0, 3));

// Get available operations
const availableOperations = computed(() => operationsStore.availableOperations.slice(0, 3));

// Get atmosphere-specific styling class
const atmosphereClass = computed(() => `atmosphere-${regionGreeting.value.atmosphere}`);

// Load data when the component is mounted
onMounted(async () => {
  isLoading.value = true;

  // Fetch player profile if not already loaded
  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }

  // Fetch current region
  await travelStore.fetchCurrentRegion();

  // Fetch territory data if not already loaded
  if (territoryStore.regions.length === 0) {
    await territoryStore.fetchTerritoryData();
  }

  // Fetch recent actions
  await territoryStore.fetchRecentActions();

  // Fetch operations data if not already loaded
  if (operationsStore.availableOperations.length === 0) {
    await operationsStore.fetchAvailableOperations();
  }

  // Fetch available regions for travel
  await travelStore.fetchAvailableRegions();

  isLoading.value = false;
});

// Helper functions
function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M';
  } else if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K';
  }
  return value.toString();
}

function formatTime(timestamp: string): string {
  const now = new Date();
  const date = new Date(timestamp);
  const diff = Math.floor((now.getTime() - date.getTime()) / 60000);

  if (diff < 1) return 'Just now';
  if (diff < 60) return `${diff} min ago`;
  if (diff < 1440) return `${Math.floor(diff / 60)} hours ago`;
  return `${Math.floor(diff / 1440)} days ago`;
}

function formatActionType(actionType: TerritoryActionType): string {
  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      return 'Extortion';
    case TerritoryActionType.TAKEOVER:
      return 'Takeover';
    case TerritoryActionType.COLLECTION:
      return 'Collection';
    case TerritoryActionType.DEFEND:
      return 'Defense';
    default:
      return actionType;
  }
}

function getActionIcon(actionType: TerritoryActionType): string {
  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      return '💰';
    case TerritoryActionType.TAKEOVER:
      return '🏢';
    case TerritoryActionType.COLLECTION:
      return '💼';
    case TerritoryActionType.DEFEND:
      return '🛡️';
    default:
      return '❓';
  }
}

function formatOperationType(operationType: OperationType): string {
  switch (operationType) {
    case OperationType.CARJACKING:
      return 'Carjacking';
    case OperationType.GOODS_SMUGGLING:
      return 'Smuggling';
    case OperationType.DRUG_TRAFFICKING:
      return 'Drug Trafficking';
    case OperationType.OFFICIAL_BRIBING:
      return 'Bribing';
    case OperationType.INTELLIGENCE_GATHERING:
      return 'Intelligence';
    case OperationType.CREW_RECRUITMENT:
      return 'Recruitment';
    default:
      return operationType;
  }
}

// Navigation functions
function goToTerritory() {
  router.push('/territory');
}

function goToOperations(operationId?: string) {
  if (operationId) {
    router.push({ path: '/operations', query: { operation: operationId } });
  } else {
    router.push('/operations');
  }
}

function goToMarket(resourceType?: ResourceType) {
  if (resourceType) {
    router.push({ path: '/market', query: { resource: resourceType } });
  } else {
    router.push('/market');
  }
}

function goToTravel() {
  router.push('/travel');
}

// Action functions
async function collectAll() {
  isLoading.value = true;

  try {
    let result;

    if (isInRegion.value) {
      // Collect from current region
      result = await territoryStore.collectAllHotspotIncomeInCurrentRegion();
    } else {
      // Collect from all territories
      result = await territoryStore.collectAllHotspotIncome();
    }

    if (result) {
      // Show a success message
      // You could implement a toast system here
    }
  } catch (error) {
    console.error('Failed to collect all pending resources:', error);
  } finally {
    isLoading.value = false;
  }
}
</script>

<template>
  <div class="home-view" :class="atmosphereClass">
    <!-- Region-specific header -->
    <div class="page-header">
      <div class="page-title">
        <h2>{{ regionGreeting.greeting }}</h2>
        <p class="welcome-message">{{ regionGreeting.description }}</p>
      </div>

      <!-- Quick travel indicator if in region -->
      <div v-if="isInRegion" class="location-indicator">
        <div class="location-badge">
          <span class="location-icon">📍</span>
          <span class="location-name">{{ currentRegion }}</span>
        </div>
        <BaseButton variant="outline" small @click="goToTravel">
          <span>✈️</span> Travel
        </BaseButton>
      </div>
    </div>

    <div class="dashboard-grid">
      <!-- Regional Overview Card -->
      <BaseCard v-if="isInRegion" title="Regional Control" class="regional-card" gold-border>
        <div class="regional-stats">
          <div class="stat-item">
            <div class="stat-label">Region</div>
            <div class="stat-value region-name">{{ currentRegion }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">Controlled Businesses</div>
            <div class="stat-value territories">{{ regionalControlled }} / {{ regionalTotal }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">Regional Revenue</div>
            <div class="stat-value income">${{ formatNumber(regionalRevenue) }}/hr</div>
          </div>
        </div>

        <div class="control-progress">
          <div class="progress-label">Regional Dominance</div>
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: `${regionalControlPercentage}%` }"
                 :class="{ 'high-control': regionalControlPercentage > 60 }"></div>
          </div>
          <div class="progress-value">{{ regionalControlPercentage }}%</div>
        </div>
      </BaseCard>

      <!-- Empire Overview Card (for when not in region or as secondary info) -->
      <BaseCard :title="isInRegion ? 'Empire Overview' : 'Criminal Empire'"
                class="overview-card"
                :class="{ 'secondary': isInRegion }"
                :gold-border="!isInRegion">
        <div class="overview-stats">
          <div class="stat-item">
            <div class="stat-label">Total Money</div>
            <div class="stat-value money">${{ formatNumber(playerMoney) }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ isInRegion ? 'Total' : 'Hourly' }} Income</div>
            <div class="stat-value income">${{ formatNumber(totalRevenue) }}/hr</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ isInRegion ? 'Empire' : 'Total' }} Territories</div>
            <div class="stat-value territories">{{ totalControlled }} / {{ totalHotspots }}</div>
          </div>
        </div>
      </BaseCard>

      <!-- Pending Collections Card -->
      <BaseCard title="Pending Collections" class="collections-card">
        <div v-if="(isInRegion && regionalControlled > 0) || (!isInRegion && totalControlled > 0)"
             class="collections-content">
          <div class="total-pending">
            <div class="pending-label">
              {{ isInRegion ? 'Available in region:' : 'Available to collect:' }}
            </div>
            <div class="pending-value">
              ${{ formatNumber(isInRegion ? regionalPending : totalPending) }}
            </div>
          </div>
          <BaseButton variant="secondary"
                      :disabled="(isInRegion ? regionalPending : totalPending) <= 0"
                      @click="collectAll"
                      :loading="isLoading">
            {{ isInRegion ? 'Collect from Region' : 'Collect All' }}
          </BaseButton>
        </div>
        <div v-else class="empty-state">
          <p>{{ isInRegion ? 'You don\'t control any businesses in this region yet.' : 'You don\'t control any businesses yet.' }}</p>
          <BaseButton variant="outline" @click="goToTerritory">
            {{ isInRegion ? 'Expand in Region' : 'Expand Your Territory' }}
          </BaseButton>
        </div>
      </BaseCard>

      <!-- Resources Card (unchanged) -->
      <BaseCard title="Resources" class="resources-card">
        <div class="resources-grid">
          <div class="resource-item">
            <div class="resource-icon">👥</div>
            <div class="resource-details">
              <div class="resource-name">Crew</div>
              <div class="resource-value">{{ playerCrew }} / {{ playerStore.maxCrew }}</div>
            </div>
            <div class="resource-actions">
              <BaseButton variant="text" small @click="goToMarket(ResourceType.CREW)">
                Buy
              </BaseButton>
            </div>
          </div>
          <div class="resource-item">
            <div class="resource-icon">🔫</div>
            <div class="resource-details">
              <div class="resource-name">Weapons</div>
              <div class="resource-value">{{ playerWeapons }} / {{ playerStore.maxWeapons }}</div>
            </div>
            <div class="resource-actions">
              <BaseButton variant="text" small @click="goToMarket(ResourceType.WEAPONS)">
                Buy
              </BaseButton>
            </div>
          </div>
          <div class="resource-item">
            <div class="resource-icon">🚗</div>
            <div class="resource-details">
              <div class="resource-name">Vehicles</div>
              <div class="resource-value">{{ playerVehicles }} / {{ playerStore.maxVehicles }}</div>
            </div>
            <div class="resource-actions">
              <BaseButton variant="text" small @click="goToMarket(ResourceType.VEHICLES)">
                Buy
              </BaseButton>
            </div>
          </div>
        </div>
      </BaseCard>

      <!-- Location-Aware Operations Card -->
      <BaseCard :title="isInRegion ? `Operations in ${currentRegion}` : 'Available Operations'"
                class="operations-card">
        <div v-if="availableOperations.length > 0" class="operations-list">
          <div v-for="operation in availableOperations" :key="operation.id" class="operation-item">
            <div class="operation-details">
              <div class="operation-name">{{ operation.name }}</div>
              <div class="operation-type">{{ formatOperationType(operation.type) }}</div>
              <div v-if="operation.regionId" class="operation-region">🏙️ Regional Operation</div>
            </div>
            <BaseButton variant="outline" small @click="goToOperations(operation.id)">
              View
            </BaseButton>
          </div>
          <div class="view-all-link">
            <a @click.prevent="goToOperations()">
              {{ isInRegion ? 'View all regional operations' : 'View all operations' }}
            </a>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>{{ isInRegion ? 'No operations available in this region.' : 'No operations available right now.' }}</p>
          <BaseButton variant="outline" @click="goToOperations()">
            Check Operations
          </BaseButton>
        </div>
      </BaseCard>

      <!-- Location-Aware Territory Card -->
      <BaseCard :title="isInRegion ? `Territory Control in ${currentRegion}` : 'Territory Status'"
                class="territory-card">
        <div class="territory-stats">
          <div v-if="isInRegion" class="regional-control">
            <div class="control-stat">
              <div class="stat-label">Businesses Controlled:</div>
              <div class="stat-value">{{ regionalControlled }}/{{ regionalTotal }}</div>
            </div>
            <div class="control-stat">
              <div class="stat-label">Regional Influence:</div>
              <div class="stat-value">{{ regionalControlPercentage }}%</div>
            </div>
          </div>
          <div v-else class="empire-control">
            <div class="control-stat">
              <div class="stat-label">Total Territories:</div>
              <div class="stat-value">{{ totalControlled }}/{{ totalHotspots }}</div>
            </div>
            <div class="control-stat">
              <div class="stat-label">Overall Control:</div>
              <div class="stat-value">{{ Math.round((totalControlled / totalHotspots) * 100) }}%</div>
            </div>
          </div>
        </div>
        <div class="territory-actions">
          <BaseButton variant="outline" @click="goToTerritory()">
            {{ isInRegion ? 'Manage Regional Territory' : 'Manage Territory' }}
          </BaseButton>
        </div>
      </BaseCard>

      <!-- Recent Actions Card -->
      <BaseCard :title="isInRegion ? 'Recent Regional Activity' : 'Recent Activities'"
                class="recent-actions-card">
        <div v-if="recentActions.length > 0" class="actions-list">
          <div v-for="action in recentActions" :key="action.id" class="action-item"
            :class="{ 'success': action.result && action.result.success, 'failure': action.result && !action.result.success }">
            <div class="action-icon">
              {{ getActionIcon(action.type) }}
            </div>
            <div class="action-details">
              <div class="action-type">{{ formatActionType(action.type) }}</div>
              <div class="action-result">{{ action.result ? action.result.message : 'In progress...' }}</div>
              <div class="action-time">{{ formatTime(action.timestamp) }}</div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>{{ isInRegion ? 'No recent activities in this region.' : 'No recent activities yet.' }}</p>
        </div>
      </BaseCard>
    </div>
  </div>
</template>

<style lang="scss">
.home-view {
  // Base atmosphere styling
  &.atmosphere-headquarters {
    .page-header {
      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        height: 200px;
        background: linear-gradient(to bottom, rgba(0, 10, 20, 0.7) 0%, transparent 100%);
        z-index: 0;
      }
    }
  }

  &.atmosphere-urban {
    .page-header {
      background: linear-gradient(135deg, rgba(10, 10, 15, 0.9) 0%, rgba(5, 5, 10, 0.95) 100%);
      color: #f0f0f0;
    }

    .regional-card {
      background: linear-gradient(135deg, rgba(15, 20, 30, 0.95) 0%, rgba(10, 15, 25, 0.98) 100%);
    }
  }

  &.atmosphere-industrial {
    .page-header {
      background: linear-gradient(135deg, rgba(15, 10, 5, 0.9) 0%, rgba(10, 7, 4, 0.95) 100%);
    }

    .regional-card {
      border-left: 4px solid $warning-color;
    }
  }

  &.atmosphere-waterfront {
    .page-header {
      background: linear-gradient(135deg, rgba(0, 15, 20, 0.9) 0%, rgba(0, 10, 15, 0.95) 100%);
    }

    .regional-card {
      border-left: 4px solid $info-color;
    }
  }

  &.atmosphere-suburban {
    .page-header {
      background: linear-gradient(135deg, rgba(10, 15, 10, 0.9) 0%, rgba(8, 12, 8, 0.95) 100%);
    }
  }

  &.atmosphere-nightlife {
    .page-header {
      background: linear-gradient(135deg, rgba(20, 5, 15, 0.9) 0%, rgba(15, 4, 10, 0.95) 100%);
    }

    .regional-card {
      background: linear-gradient(135deg, rgba(30, 15, 25, 0.95) 0%, rgba(25, 10, 20, 0.98) 100%);
    }
  }

  .page-header {
    position: relative;
    @include flex-column;
    gap: $spacing-lg;
    margin-bottom: $spacing-xl;
    padding: $spacing-lg;
    border-radius: $border-radius-md;
    background: linear-gradient(135deg, rgba(10, 10, 15, 0.9) 0%, rgba(5, 5, 10, 0.95) 100%);
    overflow: hidden;

    .page-title {
      z-index: 1;

      h2 {
        @include gold-accent;
        margin-bottom: $spacing-xs;
        font-size: $font-size-xl;
        text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
      }

      .welcome-message {
        color: $text-secondary;
        font-size: $font-size-lg;
        line-height: 1.6;
        max-width: 80%;
      }
    }

    .location-indicator {
      position: absolute;
      top: $spacing-lg;
      right: $spacing-lg;
      display: flex;
      align-items: center;
      gap: $spacing-md;
      z-index: 1;

      .location-badge {
        display: flex;
        align-items: center;
        gap: $spacing-sm;
        background: rgba($background-darker, 0.7);
        padding: $spacing-sm $spacing-md;
        border-radius: $border-radius-md;
        border: 1px solid rgba($gold-color, 0.3);

        .location-icon {
          color: $gold-color;
        }

        .location-name {
          font-weight: 600;
          color: $text-color;
        }
      }
    }
  }

  .dashboard-grid {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    gap: $spacing-lg;

    @include respond-to(md) {
      grid-template-columns: repeat(2, 1fr);
    }

    @include respond-to(lg) {
      grid-template-columns: repeat(3, 1fr);
    }

    .regional-card {
      grid-column: 1 / -1;
      background: linear-gradient(135deg, rgba(15, 20, 30, 0.95) 0%, rgba(10, 15, 25, 0.98) 100%);
      border: 1px solid rgba($gold-color, 0.2);

      .regional-stats {
        display: flex;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: $spacing-lg;
        margin-bottom: $spacing-lg;

        .stat-item {
          flex: 1;
          min-width: 150px;

          .stat-label {
            color: $text-secondary;
            font-size: $font-size-sm;
            margin-bottom: $spacing-xs;
          }

          .stat-value {
            font-size: $font-size-xl;
            font-weight: 600;

            &.region-name {
              @include gold-accent;
              font-size: $font-size-lg;
            }

            &.income {
              color: $success-color;
            }

            &.territories {
              color: $info-color;
            }
          }
        }
      }

      .control-progress {
        .progress-label {
          color: $text-secondary;
          margin-bottom: $spacing-xs;
          font-size: $font-size-sm;
        }

        .progress-bar {
          height: 12px;
          background-color: rgba(255, 255, 255, 0.1);
          border-radius: 6px;
          overflow: hidden;
          margin-bottom: $spacing-xs;

          .progress-fill {
            height: 100%;
            background: linear-gradient(to right, $primary-color, $secondary-color);
            border-radius: 6px;
            transition: width 0.3s ease;

            &.high-control {
              background: linear-gradient(to right, $secondary-color, $success-color);
            }
          }
        }

        .progress-value {
          text-align: right;
          font-size: $font-size-sm;
          font-weight: 600;
        }
      }
    }

    .overview-card {
      &.secondary {
        grid-column: span 1;

        .overview-stats {
          .stat-value {
            font-size: $font-size-lg;
          }
        }
      }
    }

    .operation-item {
      .operation-region {
        font-size: $font-size-xs;
        color: $gold-color;
        margin-top: $spacing-xs;
      }
    }
  }
}
</style>
