// src/views/HomeView.vue

<template>
  <div class="home-view">
    <div class="page-title">
      <h2>Dashboard</h2>
      <p class="welcome-message">Welcome back, {{ playerName }}. What's your next move?</p>
    </div>

    <div class="dashboard-grid">
      <!-- Overview Card -->
      <BaseCard title="Empire Overview" class="overview-card" gold-border>
        <div class="overview-stats">
          <div class="stat-item">
            <div class="stat-label">Total Money</div>
            <div class="stat-value money">${{ formatNumber(playerMoney) }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">Hourly Income</div>
            <div class="stat-value income">${{ formatNumber(hourlyRevenue) }}/hr</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">Controlled Territories</div>
            <div class="stat-value territories">{{ controlledHotspots }} / {{ totalHotspots }}</div>
          </div>
        </div>
      </BaseCard>

      <!-- Pending Collections Card -->
      <BaseCard title="Pending Collections" class="collections-card">
        <div v-if="controlledHotspots > 0" class="collections-content">
          <div class="total-pending">
            <div class="pending-label">Available to collect:</div>
            <div class="pending-value">${{ formatNumber(pendingCollections) }}</div>
          </div>
          <BaseButton variant="secondary" :disabled="pendingCollections <= 0" @click="collectAll">
            Collect All
          </BaseButton>
        </div>
        <div v-else class="empty-state">
          <p>You don't control any hotspots yet.</p>
          <BaseButton variant="outline" @click="goToTerritory">
            Expand Your Territory
          </BaseButton>
        </div>
      </BaseCard>

      <!-- Resources Card -->
      <BaseCard title="Resources" class="resources-card">
        <div class="resources-grid">
          <div class="resource-item">
            <div class="resource-icon">👥</div>
            <div class="resource-details">
              <div class="resource-name">Crew</div>
              <div class="resource-value">{{ playerCrew }} / {{ maxCrew }}</div>
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
              <div class="resource-value">{{ playerWeapons }} / {{ maxWeapons }}</div>
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
              <div class="resource-value">{{ playerVehicles }} / {{ maxVehicles }}</div>
            </div>
            <div class="resource-actions">
              <BaseButton variant="text" small @click="goToMarket(ResourceType.VEHICLES)">
                Buy
              </BaseButton>
            </div>
          </div>
        </div>
      </BaseCard>

      <!-- Operations Card -->
      <BaseCard title="Available Operations" class="operations-card">
        <div v-if="availableOperations.length > 0" class="operations-list">
          <div v-for="operation in visibleOperations" :key="operation.id" class="operation-item">
            <div class="operation-details">
              <div class="operation-name">{{ operation.name }}</div>
              <div class="operation-type">Type: {{ formatOperationType(operation.type) }}</div>
            </div>
            <BaseButton variant="outline" small @click="goToOperations(operation.id)">
              View
            </BaseButton>
          </div>
          <div class="view-all-link">
            <a @click.prevent="goToOperations()">View all operations</a>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>No operations available right now.</p>
          <BaseButton variant="outline" @click="goToOperations()">
            Check Operations
          </BaseButton>
        </div>
      </BaseCard>

      <!-- Territory Card -->
      <BaseCard title="Territory Status" class="territory-card">
        <div class="territory-stats">
          <div class="territory-region">
            <div class="region-name">Most Controlled Region:</div>
            <div class="region-value">{{ mostControlledRegion }}</div>
          </div>
          <div class="territory-progress">
            <div class="progress-label">Overall Territorial Control:</div>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: `${territorialControlPercentage}%` }"></div>
            </div>
            <div class="progress-value">{{ territorialControlPercentage }}%</div>
          </div>
        </div>
        <div class="territory-actions">
          <BaseButton variant="outline" @click="goToTerritory()">
            Manage Territory
          </BaseButton>
        </div>
      </BaseCard>

      <!-- Recent Actions Card -->
      <BaseCard title="Recent Activities" class="recent-actions-card">
        <div v-if="recentActions.length > 0" class="actions-list">
          <div v-for="action in visibleActions" :key="action.id" class="action-item"
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
          <p>No recent activities yet.</p>
        </div>
      </BaseCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { useOperationsStore } from '@/stores/modules/operations';
import {
  ResourceType
} from '@/types/market';
import {
  TerritoryActionType
} from '@/types/territory';
import {
  OperationType
} from '@/types/operations';

const router = useRouter();
const playerStore = usePlayerStore();
const territoryStore = useTerritoryStore();
const operationsStore = useOperationsStore();

const isLoading = ref(false);

// Computed properties from stores
const playerName = computed(() => playerStore.profile?.name || 'Boss');
const playerMoney = computed(() => playerStore.playerMoney);
const playerCrew = computed(() => playerStore.playerCrew);
const playerWeapons = computed(() => playerStore.playerWeapons);
const playerVehicles = computed(() => playerStore.playerVehicles);
const maxCrew = computed(() => playerStore.maxCrew);
const maxWeapons = computed(() => playerStore.maxWeapons);
const maxVehicles = computed(() => playerStore.maxVehicles);
const controlledHotspots = computed(() => playerStore.controlledHotspots);
const totalHotspots = computed(() => playerStore.totalHotspots);
const hourlyRevenue = computed(() => playerStore.hourlyRevenue);
const pendingCollections = computed(() => playerStore.pendingCollections);

// Territory stats
const territorialControlPercentage = computed(() => {
  if (totalHotspots.value === 0) return 0;
  return Math.round((controlledHotspots.value / totalHotspots.value) * 100);
});

// Mock data for now, will replace with real data later
const mostControlledRegion = ref('Downtown');

// Get recent actions from the territory store
const recentActions = computed(() => territoryStore.recentActions);

// Only show 3 most recent actions
const visibleActions = computed(() => {
  return recentActions.value.slice(0, 3);
});

// Get available operations from the operations store
const availableOperations = computed(() => {
  return operationsStore.availableOperations;
});

// Only show 3 most recent operations
const visibleOperations = computed(() => {
  return availableOperations.value.slice(0, 3);
});

// Load data when the component is mounted
onMounted(async () => {
  isLoading.value = true;

  // Fetch player profile if not already loaded
  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }

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

  isLoading.value = false;
});

// Helper function to format numbers
function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M';
  } else if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K';
  }
  return value.toString();
}

// Format timestamps to relative time
function formatTime(timestamp: string): string {
  const now = new Date();
  const date = new Date(timestamp);
  const diff = Math.floor((now.getTime() - date.getTime()) / 60000); // difference in minutes

  if (diff < 1) return 'Just now';
  if (diff < 60) return `${diff} min ago`;
  if (diff < 1440) return `${Math.floor(diff / 60)} hours ago`;
  return `${Math.floor(diff / 1440)} days ago`;
}

// Format action types to be more readable
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

// Get icon for action type
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

// Format operation types to be more readable
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

// Action functions

async function collectAll() {
  isLoading.value = true;

  try {
    // Use the territory store instead of player store
    await territoryStore.collectAllHotspotIncome();
  } catch (error) {
    console.error('Failed to collect all pending resources:', error);
  } finally {
    isLoading.value = false;
  }
}

// async function collectAll() {
//   isLoading.value = true;

//   try {
//     await playerStore.collectAllPending();
//   } catch (error) {
//     console.error('Failed to collect all pending resources:', error);
//   } finally {
//     isLoading.value = false;
//   }
// }
</script>

<style lang="scss">
.home-view {
  .page-title {
    margin-bottom: $spacing-xl;

    h2 {
      @include gold-accent;
      margin-bottom: $spacing-xs;
    }

    .welcome-message {
      color: $text-secondary;
      font-size: $font-size-lg;
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

    .overview-card {
      grid-column: 1 / -1;

      .overview-stats {
        display: flex;
        justify-content: space-between;
        flex-wrap: wrap;
        gap: $spacing-lg;

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

            &.money {
              @include gold-accent;
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
    }

    .collections-card,
    .resources-card,
    .operations-card,
    .territory-card,
    .recent-actions-card {
      display: flex;
      flex-direction: column;

      .empty-state {
        @include flex-column;
        align-items: center;
        justify-content: center;
        gap: $spacing-md;
        padding: $spacing-lg 0;
        text-align: center;
        color: $text-secondary;
      }
    }

    .collections-card {
      .collections-content {
        @include flex-column;
        gap: $spacing-lg;

        .total-pending {
          @include flex-between;

          .pending-label {
            color: $text-secondary;
          }

          .pending-value {
            font-size: $font-size-xl;
            font-weight: 600;
            @include gold-accent;
          }
        }
      }
    }

    .resources-card {
      .resources-grid {
        @include flex-column;
        gap: $spacing-md;

        .resource-item {
          @include flex-between;
          gap: $spacing-md;
          padding: $spacing-sm 0;
          border-bottom: 1px solid $border-color;

          &:last-child {
            border-bottom: none;
          }

          .resource-icon {
            font-size: 24px;
          }

          .resource-details {
            flex: 1;

            .resource-name {
              font-weight: 600;
            }

            .resource-value {
              color: $text-secondary;
              font-size: $font-size-sm;
            }
          }
        }
      }
    }

    .operations-card {
      .operations-list {
        @include flex-column;
        gap: $spacing-md;

        .operation-item {
          @include flex-between;
          gap: $spacing-md;
          padding: $spacing-sm 0;
          border-bottom: 1px solid $border-color;

          &:last-child {
            border-bottom: none;
          }

          .operation-details {
            flex: 1;

            .operation-name {
              font-weight: 600;
            }

            .operation-type {
              color: $text-secondary;
              font-size: $font-size-sm;
            }
          }
        }

        .view-all-link {
          text-align: center;
          margin-top: $spacing-md;

          a {
            cursor: pointer;

            &:hover {
              text-decoration: underline;
            }
          }
        }
      }
    }

    .territory-card {
      .territory-stats {
        @include flex-column;
        gap: $spacing-lg;

        .territory-region {
          @include flex-between;

          .region-name {
            color: $text-secondary;
          }

          .region-value {
            font-weight: 600;
            color: $info-color;
          }
        }

        .territory-progress {
          .progress-label {
            color: $text-secondary;
            margin-bottom: $spacing-xs;
          }

          .progress-bar {
            height: 8px;
            background-color: rgba(255, 255, 255, 0.1);
            border-radius: 4px;
            overflow: hidden;
            margin-bottom: $spacing-xs;

            .progress-fill {
              height: 100%;
              background-color: $secondary-color;
              border-radius: 4px;
            }
          }

          .progress-value {
            text-align: right;
            font-size: $font-size-sm;
          }
        }
      }

      .territory-actions {
        margin-top: $spacing-lg;
        text-align: center;
      }
    }

    .recent-actions-card {
      .actions-list {
        @include flex-column;
        gap: $spacing-md;

        .action-item {
          display: flex;
          gap: $spacing-md;
          padding: $spacing-sm;
          border-radius: $border-radius-sm;
          background-color: rgba(255, 255, 255, 0.05);

          &.success {
            background-color: rgba($success-color, 0.1);
            border-left: 3px solid $success-color;
          }

          &.failure {
            background-color: rgba($danger-color, 0.1);
            border-left: 3px solid $danger-color;
          }

          .action-icon {
            font-size: 24px;
          }

          .action-details {
            flex: 1;

            .action-type {
              font-weight: 600;
            }

            .action-result {
              font-size: $font-size-sm;
              margin: $spacing-xs 0;
            }

            .action-time {
              font-size: $font-size-sm;
              color: $text-secondary;
            }
          }
        }
      }
    }
  }
}
</style>