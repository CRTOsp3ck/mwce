// src/views/TravelView.vue

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTravelStore } from '@/stores/modules/travel';
import { Region } from '@/types/territory';

const router = useRouter();
const playerStore = usePlayerStore();
const travelStore = useTravelStore();

// State
const isLoading = ref(false);
const selectedRegionId = ref<string | null>(null);
const showConfirmModal = ref(false);
const showResultModal = ref(false);
const isTraveling = ref(false);
const showHistoryModal = ref(false);

// Computed properties
const playerMoney = computed(() => playerStore.playerMoney);
const playerHeat = computed(() => playerStore.playerHeat);
const playerName = computed(() => playerStore.profile?.name || 'Boss');
const availableRegions = computed(() => travelStore.availableRegions);
const currentRegion = computed(() => travelStore.currentRegion);
const currentLocationName = computed(() => travelStore.currentLocationName);
const travelHistory = computed(() => travelStore.recentTravelAttempts);

const selectedRegion = computed(() => {
  if (!selectedRegionId.value) return null;
  return availableRegions.value.find(r => r.id === selectedRegionId.value) || null;
});

// Load data when component is mounted
onMounted(async () => {
  isLoading.value = true;

  // Fetch player profile if not already loaded
  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }

  // Fetch travel-related data
  try {
    await Promise.all([
      travelStore.fetchAvailableRegions(),
      travelStore.fetchCurrentRegion(),
      travelStore.fetchTravelHistory(10)
    ]);
  } catch (error) {
    console.error('Error loading travel data:', error);
  } finally {
    isLoading.value = false;
  }
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

function formatDate(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleString();
}

function formatTimeAgo(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.round(diffMs / 1000);
  const diffMin = Math.round(diffSec / 60);
  const diffHour = Math.round(diffMin / 60);
  const diffDay = Math.round(diffHour / 24);

  if (diffSec < 60) {
    return 'just now';
  } else if (diffMin < 60) {
    return `${diffMin} minute${diffMin > 1 ? 's' : ''} ago`;
  } else if (diffHour < 24) {
    return `${diffHour} hour${diffHour > 1 ? 's' : ''} ago`;
  } else {
    return `${diffDay} day${diffDay > 1 ? 's' : ''} ago`;
  }
}

function calculateCatchChance(baseChance: number, heat: number, multiplier: number, maxChance: number): number {
  const catchChance = baseChance + (heat * multiplier);
  return Math.min(catchChance, maxChance);
}

function getRegionImageUrl(regionId: string | null): string {
  if (!regionId) return '/images/regions/headquarters.jpg';

  // In a real implementation, you would map region IDs to their image URLs
  // For now, we'll just return a placeholder image
  const regionImages = {
    'north': '/images/regions/north.jpg',
    'south': '/images/regions/south.jpg',
    'east': '/images/regions/east.jpg',
    'west': '/images/regions/west.jpg',
    'downtown': '/images/regions/downtown.jpg',
    'harbor': '/images/regions/harbor.jpg',
    'suburbs': '/images/regions/suburbs.jpg',
    'industrial': '/images/regions/industrial.jpg',
  };

  return regionImages[regionId.toLowerCase()] || '/images/regions/default.jpg';
}

// Form functions
function selectRegion(regionId: string) {
  // If player is already in this region, don't do anything
  if (currentRegion.value?.id === regionId) {
    return;
  }

  selectedRegionId.value = regionId;
  showConfirmModal.value = true;
}

function closeConfirmModal() {
  showConfirmModal.value = false;
  selectedRegionId.value = null;
}

function closeResultModal() {
  showResultModal.value = false;
}

function openHistoryModal() {
  showHistoryModal.value = true;
}

function closeHistoryModal() {
  showHistoryModal.value = false;
}

// Travel logic
async function confirmTravel() {
  if (!selectedRegionId.value || isTraveling.value) return;

  isTraveling.value = true;

  try {
    const result = await travelStore.travel(selectedRegionId.value);

    if (result) {
      // Close confirm modal
      showConfirmModal.value = false;

      // Show result modal
      showResultModal.value = true;
    }
  } catch (error) {
    console.error('Error traveling:', error);
  } finally {
    isTraveling.value = false;
  }
}

// Navigation function to go to main dashboard after travel
function goToDashboard() {
  router.push('/');
}

// Constants for the travel mechanics demo
const TRAVEL_CONFIG = {
  BASE_COST: 1000,
  BASE_CATCH_CHANCE: 10.0,  // 10% base chance
  HEAT_MULTIPLIER: 0.5,     // Each heat point adds 0.5% to catch chance
  MAX_CATCH_CHANCE: 75.0,   // 75% maximum catch chance
  BASE_FINE_FACTOR: 0.15,   // 15% of player's money
  MINIMUM_FINE: 500,        // Minimum fine amount
  MAX_FINE_PERCENT: 0.5,    // Maximum 50% of player's money
  CAUGHT_HEAT_INCREASE: 20, // Heat increase when caught
  SUCCESS_HEAT_REDUCTION: 5 // Heat reduction on successful travel
};

const estimatedCatchChance = computed(() => {
  return calculateCatchChance(
    TRAVEL_CONFIG.BASE_CATCH_CHANCE,
    playerHeat.value,
    TRAVEL_CONFIG.HEAT_MULTIPLIER,
    TRAVEL_CONFIG.MAX_CATCH_CHANCE
  );
});

const travelCost = computed(() => {
  // Could be more complex based on distance, but we'll use a fixed cost for now
  return TRAVEL_CONFIG.BASE_COST;
});

const estimatedFine = computed(() => {
  // Calculate potential fine if caught
  const finePercent = TRAVEL_CONFIG.BASE_FINE_FACTOR;
  const fineAmount = Math.max(
    Math.round(playerMoney.value * finePercent),
    TRAVEL_CONFIG.MINIMUM_FINE
  );

  // Cap the fine to prevent wiping out the player
  const maxFine = Math.round(playerMoney.value * TRAVEL_CONFIG.MAX_FINE_PERCENT);
  return Math.min(fineAmount, maxFine);
});

const canAffordTravel = computed(() => {
  return playerMoney.value >= travelCost.value;
});

// Result-related computed properties
const travelResult = computed(() => travelStore.lastTravelResponse);

const resultIconClass = computed(() => {
  if (!travelResult.value) return '';
  return travelResult.value.success ? 'success-icon' : 'failure-icon';
});

const resultIcon = computed(() => {
  if (!travelResult.value) return '';
  return travelResult.value.success ? '✅' : '❌';
});
</script>

<template>
  <div class="travel-view">
    <div class="page-title">
      <h2>Mafia Travel Agency</h2>
      <p class="subtitle">Travel between regions to expand your criminal empire.</p>
    </div>

    <div class="current-location-card">
      <div class="location-icon">🌆</div>
      <div class="location-details">
        <h3>Current Location: {{ currentLocationName }}</h3>
        <p v-if="currentRegion">You're currently operating in {{ currentRegion.name }}.</p>
        <p v-else>You're currently at your criminal headquarters.</p>
      </div>
      <div class="location-actions">
        <BaseButton variant="text" @click="openHistoryModal">
          <span>Travel History</span>
        </BaseButton>
      </div>
    </div>

    <div class="travel-warning" v-if="playerHeat > 40">
      <div class="warning-icon">⚠️</div>
      <div class="warning-message">
        <strong>Warning:</strong> Your heat level is high ({{ playerHeat }}). This increases the chance of getting caught while traveling.
      </div>
    </div>

    <h3 class="section-title">Available Destinations</h3>

    <div class="regions-grid">
      <BaseCard
        v-for="region in availableRegions"
        :key="region.id"
        class="region-card"
        :class="{ 'current-region': currentRegion?.id === region.id }"
      >
        <div class="region-image" :style="{ backgroundImage: `url(${getRegionImageUrl(region.id)})` }">
          <div class="region-name">{{ region.name }}</div>
        </div>

        <div class="region-details">
          <p class="region-description">
            {{ region.description || `The ${region.name} region is full of opportunities for a cunning criminal.` }}
          </p>

          <div class="region-stats">
            <div class="stat">
              <div class="stat-label">Travel Cost:</div>
              <div class="stat-value">${{ formatNumber(travelCost) }}</div>
            </div>
            <div class="stat">
              <div class="stat-label">Travel Time:</div>
              <div class="stat-value">Instant</div>
            </div>
          </div>
        </div>

        <template #footer>
          <div class="region-footer">
            <BaseButton
              v-if="currentRegion?.id !== region.id"
              :disabled="!canAffordTravel || isLoading"
              @click="selectRegion(region.id)"
            >
              Travel to {{ region.name }}
            </BaseButton>
            <div v-else class="current-badge">
              Currently Here
            </div>
          </div>
        </template>
      </BaseCard>
    </div>

    <!-- Confirm Travel Modal -->
    <BaseModal
      v-model="showConfirmModal"
      :title="`Travel to ${selectedRegion?.name || 'Destination'}`"
    >
      <div class="travel-confirm-content">
        <div class="destination-preview" :style="{ backgroundImage: `url(${getRegionImageUrl(selectedRegionId)})` }">
          <div class="destination-overlay">
            <h3>{{ selectedRegion?.name }}</h3>
          </div>
        </div>

        <div class="travel-details">
          <div class="travel-info">
            <div class="info-icon">💰</div>
            <div class="info-details">
              <div class="info-label">Travel Cost:</div>
              <div class="info-value">${{ formatNumber(travelCost) }}</div>
            </div>
          </div>

          <div class="travel-info">
            <div class="info-icon">🚨</div>
            <div class="info-details">
              <div class="info-label">Risk of Getting Caught:</div>
              <div class="info-value" :class="{
                'high-risk': estimatedCatchChance > 50,
                'medium-risk': estimatedCatchChance > 25 && estimatedCatchChance <= 50,
                'low-risk': estimatedCatchChance <= 25
              }">
                {{ estimatedCatchChance.toFixed(1) }}%
              </div>
            </div>
          </div>

          <div class="travel-info">
            <div class="info-icon">❄️</div>
            <div class="info-details">
              <div class="info-label">Heat Reduction (if successful):</div>
              <div class="info-value">{{ TRAVEL_CONFIG.SUCCESS_HEAT_REDUCTION }} points</div>
            </div>
          </div>

          <div class="travel-info danger">
            <div class="info-icon">👮</div>
            <div class="info-details">
              <div class="info-label">If Caught:</div>
              <div class="info-value">
                <div>Fine: ${{ formatNumber(estimatedFine) }}</div>
                <div>Heat Increase: {{ TRAVEL_CONFIG.CAUGHT_HEAT_INCREASE }} points</div>
              </div>
            </div>
          </div>
        </div>

        <div class="travel-warning" v-if="playerMoney < travelCost">
          <div class="warning-icon">⚠️</div>
          <div class="warning-message">
            You don't have enough money to travel. You need ${{ formatNumber(travelCost) }}.
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton
            variant="text"
            @click="closeConfirmModal"
          >
            Cancel
          </BaseButton>
          <BaseButton
            :disabled="!canAffordTravel || isTraveling"
            :loading="isTraveling"
            @click="confirmTravel"
          >
            Confirm Travel
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- Travel Result Modal -->
    <BaseModal
      v-model="showResultModal"
      :title="travelResult?.success ? 'Travel Successful' : 'Travel Failed'"
    >
      <div class="travel-result">
        <div class="result-icon" :class="resultIconClass">
          {{ resultIcon }}
        </div>

        <div class="result-message">
          {{ travelResult?.message || 'Your travel attempt has been processed.' }}
        </div>

        <div class="result-details">
          <div v-if="travelResult?.success" class="success-details">
            <div class="detail-item">
              <div class="detail-label">Destination:</div>
              <div class="detail-value">{{ travelResult.regionName }}</div>
            </div>
            <div class="detail-item">
              <div class="detail-label">Travel Cost:</div>
              <div class="detail-value">${{ formatNumber(travelResult.travelCost) }}</div>
            </div>
            <div v-if="travelResult.heatReduction" class="detail-item">
              <div class="detail-label">Heat Reduced:</div>
              <div class="detail-value">-{{ travelResult.heatReduction }} points</div>
            </div>
          </div>

          <div v-else class="failure-details">
            <div class="detail-item">
              <div class="detail-label">Attempted Destination:</div>
              <div class="detail-value">{{ travelResult?.regionName }}</div>
            </div>
            <div v-if="travelResult?.fineAmount" class="detail-item">
              <div class="detail-label">Fine Paid:</div>
              <div class="detail-value">${{ formatNumber(travelResult.fineAmount) }}</div>
            </div>
            <div v-if="travelResult?.heatIncrease" class="detail-item">
              <div class="detail-label">Heat Increased:</div>
              <div class="detail-value">+{{ travelResult.heatIncrease }} points</div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton @click="closeResultModal">
            Continue
          </BaseButton>
          <BaseButton variant="secondary" v-if="travelResult?.success" @click="goToDashboard">
            Go to Dashboard
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- Travel History Modal -->
    <BaseModal
      v-model="showHistoryModal"
      title="Travel History"
    >
      <div class="travel-history">
        <div v-if="travelHistory.length > 0" class="history-list">
          <div
            v-for="attempt in travelHistory"
            :key="attempt.id"
            class="history-item"
            :class="{ 'successful': attempt.success, 'failed': !attempt.success }"
          >
            <div class="history-icon">
              {{ attempt.success ? '✅' : '❌' }}
            </div>

            <div class="history-details">
              <div class="history-header">
                <div class="history-title">
                  {{ attempt.success ? 'Successful Travel' : 'Caught by Police' }}
                </div>
                <div class="history-time">{{ formatTimeAgo(attempt.timestamp) }}</div>
              </div>

              <div class="travel-route">
                {{ attempt.fromRegionId ? `From region: ${attempt.fromRegionId}` : 'From: Headquarters' }}
                <span class="route-arrow">→</span>
                To region: {{ attempt.toRegionId }}
              </div>

              <div class="travel-outcome">
                <div class="outcome-item">
                  <span class="outcome-label">Cost:</span>
                  <span class="outcome-value">${{ formatNumber(attempt.travelCost) }}</span>
                </div>

                <div v-if="!attempt.success" class="outcome-item">
                  <span class="outcome-label">Fine:</span>
                  <span class="outcome-value">${{ formatNumber(attempt.fineAmount) }}</span>
                </div>

                <div class="outcome-item">
                  <span class="outcome-label">Heat Change:</span>
                  <span class="outcome-value" :class="{ 'heat-reduced': attempt.heatChange < 0, 'heat-increased': attempt.heatChange > 0 }">
                    {{ attempt.heatChange > 0 ? '+' : '' }}{{ attempt.heatChange }}
                  </span>
                </div>

                <div class="outcome-date">
                  {{ formatDate(attempt.timestamp) }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="empty-history">
          <div class="empty-icon">🧳</div>
          <p>No travel history found.</p>
          <p>Start traveling to build your criminal empire across the city!</p>
        </div>
      </div>

      <template #footer>
        <BaseButton @click="closeHistoryModal">Close</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<style lang="scss">
.travel-view {
  .page-title {
    margin-bottom: $spacing-xl;

    h2 {
      @include gold-accent;
      margin-bottom: $spacing-xs;
    }

    .subtitle {
      color: $text-secondary;
    }
  }

  .current-location-card {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    padding: $spacing-md;
    background-color: $background-card;
    border-radius: $border-radius-md;
    margin-bottom: $spacing-lg;

    .location-icon {
      font-size: 32px;
    }

    .location-details {
      flex: 1;

      h3 {
        margin: 0 0 $spacing-xs 0;
      }

      p {
        margin: 0;
        color: $text-secondary;
      }
    }
  }

  .travel-warning {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    padding: $spacing-md;
    background-color: rgba($warning-color, 0.1);
    border-left: 3px solid $warning-color;
    border-radius: $border-radius-sm;
    margin-bottom: $spacing-lg;

    .warning-icon {
      font-size: 24px;
    }

    .warning-message {
      color: $text-color;

      strong {
        color: $warning-color;
      }
    }
  }

  .section-title {
    margin-bottom: $spacing-md;
    position: relative;

    &:after {
      content: '';
      position: absolute;
      bottom: -8px;
      left: 0;
      width: 50px;
      height: 3px;
      background-color: $primary-color;
    }
  }

  .regions-grid {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    gap: $spacing-lg;

    @include respond-to(sm) {
      grid-template-columns: repeat(2, 1fr);
    }

    @include respond-to(md) {
      grid-template-columns: repeat(3, 1fr);
    }

    @include respond-to(lg) {
      grid-template-columns: repeat(4, 1fr);
    }

    .region-card {
      overflow: hidden;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-5px);
        box-shadow: 0 10px 20px rgba(0, 0, 0, 0.3);
      }

      &.current-region {
        border: 2px solid $gold-color;
        box-shadow: 0 0 15px rgba($gold-color, 0.5);
      }

      .region-image {
        height: 150px;
        background-size: cover;
        background-position: center;
        position: relative;

        &:before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: linear-gradient(to bottom, rgba(0, 0, 0, 0.1), rgba(0, 0, 0, 0.7));
        }

        .region-name {
          position: absolute;
          bottom: $spacing-md;
          left: $spacing-md;
          color: white;
          font-size: $font-size-lg;
          font-weight: 600;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
        }
      }

      .region-details {
        padding: $spacing-md;

        .region-description {
          color: $text-secondary;
          margin-bottom: $spacing-md;
        }

        .region-stats {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: $spacing-md;

          .stat {
            .stat-label {
              color: $text-secondary;
              font-size: $font-size-sm;
              margin-bottom: 2px;
            }

            .stat-value {
              font-weight: 600;
            }
          }
        }
      }

      .region-footer {
        display: flex;
        justify-content: center;

        .current-badge {
          background-color: $gold-color;
          color: $background-darker;
          padding: $spacing-sm $spacing-md;
          border-radius: $border-radius-sm;
          font-weight: 600;
        }
      }
    }
  }

  .travel-confirm-content {
    @include flex-column;
    gap: $spacing-lg;

    .destination-preview {
      height: 180px;
      background-size: cover;
      background-position: center;
      border-radius: $border-radius-md;
      position: relative;
      overflow: hidden;

      .destination-overlay {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        padding: $spacing-md;
        background: linear-gradient(to top, rgba(0, 0, 0, 0.8), transparent);

        h3 {
          margin: 0;
          color: white;
          font-size: $font-size-xl;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
        }
      }
    }

    .travel-details {
      @include flex-column;
      gap: $spacing-md;

      .travel-info {
        display: flex;
        align-items: center;
        gap: $spacing-md;
        padding: $spacing-sm;
        border-radius: $border-radius-sm;
        background-color: rgba($background-lighter, 0.1);

        &.danger {
          background-color: rgba($danger-color, 0.1);
        }

        .info-icon {
          font-size: 24px;
        }

        .info-details {
          flex: 1;

          .info-label {
            color: $text-secondary;
            font-size: $font-size-sm;
            margin-bottom: 2px;
          }

          .info-value {
            font-weight: 600;

            &.high-risk {
              color: $danger-color;
            }

            &.medium-risk {
              color: $warning-color;
            }

            &.low-risk {
              color: $success-color;
            }
          }
        }
      }
    }
  }

  .travel-result {
    @include flex-column;
    align-items: center;
    text-align: center;
    gap: $spacing-lg;

    .result-icon {
      font-size: 48px;

      &.success-icon {
        color: $success-color;
      }

      &.failure-icon {
        color: $danger-color;
      }
    }

    .result-message {
      font-size: $font-size-lg;
      font-weight: 500;
    }

    .result-details {
      width: 100%;
      background-color: rgba($background-lighter, 0.1);
      border-radius: $border-radius-md;
      padding: $spacing-md;

      .success-details,
      .failure-details {
        @include flex-column;
        gap: $spacing-sm;

        .detail-item {
          @include flex-between;

          .detail-label {
            color: $text-secondary;
          }

          .detail-value {
            font-weight: 600;
          }
        }
      }
    }
  }

  .travel-history {
    max-height: 600px;
    overflow-y: auto;

    .history-list {
      @include flex-column;
      gap: $spacing-md;

      .history-item {
        display: flex;
        gap: $spacing-md;
        padding: $spacing-md;
        background-color: rgba($background-lighter, 0.1);
        border-radius: $border-radius-md;

        &.successful {
          border-left: 3px solid $success-color;
        }

        &.failed {
          border-left: 3px solid $danger-color;
        }

        .history-icon {
          font-size: 24px;
        }

        .history-details {
          flex: 1;

          .history-header {
            @include flex-between;
            margin-bottom: $spacing-xs;

            .history-title {
              font-weight: 600;
            }

            .history-time {
              color: $text-secondary;
              font-size: $font-size-sm;
            }
          }

          .travel-route {
            margin-bottom: $spacing-sm;

            .route-arrow {
              color: $primary-color;
              margin: 0 $spacing-xs;
              font-weight: 600;
            }
          }

          .travel-outcome {
            display: flex;
            flex-wrap: wrap;
            gap: $spacing-sm $spacing-md;

            .outcome-item {
              .outcome-label {
                color: $text-secondary;
                margin-right: $spacing-xs;
              }

              .outcome-value {
                font-weight: 600;

                &.heat-reduced {
                  color: $success-color;
                }

                &.heat-increased {
                  color: $danger-color;
                }
              }
            }

            .outcome-date {
              width: 100%;
              margin-top: $spacing-xs;
              color: $text-secondary;
              font-size: $font-size-sm;
            }
          }
        }
      }
    }

    .empty-history {
      @include flex-column;
      align-items: center;
      justify-content: center;
      gap: $spacing-md;
      padding: $spacing-xl 0;
      text-align: center;
      color: $text-secondary;

      .empty-icon {
        font-size: 48px;
        margin-bottom: $spacing-sm;
      }
    }
  }

  .modal-footer-actions {
    @include flex-between;
    width: 100%;
  }
}
</style>
