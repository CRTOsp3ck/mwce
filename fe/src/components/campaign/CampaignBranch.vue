<!-- fe/src/components/campaign/CampaignBranch.vue -->

<script setup lang="ts">
import { computed } from 'vue';
import { Branch, Mission, CampaignOperation, CampaignPOI, InteractionType } from '@/types/campaign';
import { useCampaignStore } from '@/stores/modules/campaign';
import BaseButton from '@/components/ui/BaseButton.vue';

const props = defineProps<{
  branch: Branch;
  mission?: Mission | null;
  canComplete: boolean;
}>();

const emit = defineEmits<{
  (e: 'interactWithPOI', poiId: string, interactionType: InteractionType): void;
  (e: 'completeBranch'): void;
}>();

const campaignStore = useCampaignStore();

const operations = computed(() => campaignStore.branchOperations);
const pois = computed(() => campaignStore.branchPOIs);

const completedOperations = computed(() => operations.value.filter(op =>
  campaignStore.isOperationComplete(op.id)
));

const completedPOIs = computed(() => pois.value.filter(poi =>
  campaignStore.isPOIComplete(poi.id)
));

const completionPercentage = computed(() => {
  const totalItems = operations.value.length + pois.value.length;
  const completedItems = completedOperations.value.length + completedPOIs.value.length;
  return totalItems > 0 ? Math.round((completedItems / totalItems) * 100) : 0;
});

function interactWithPOI(poiId: string, interactionType: InteractionType) {
  emit('interactWithPOI', poiId, interactionType);
}

function completeBranch() {
  emit('completeBranch');
}

function isOperationComplete(operationId: string) {
  return campaignStore.isOperationComplete(operationId);
}

function isPOIComplete(poiId: string) {
  return campaignStore.isPOIComplete(poiId);
}
</script>

<template>
  <div class="branch-view">
    <div class="branch-header">
      <div class="branch-info">
        <h3>{{ branch.name }}</h3>
        <p class="description">{{ branch.description }}</p>
      </div>

      <div class="branch-progress">
        <div class="progress-text">
          <span>{{ completedOperations.length + completedPOIs.length }} / {{ operations.length + pois.length }} Complete</span>
          <span class="percentage">{{ completionPercentage }}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: `${completionPercentage}%` }"
            :class="{ 'complete': completionPercentage === 100 }"></div>
        </div>
      </div>

      <BaseButton v-if="canComplete" variant="primary" @click="completeBranch">
        Complete Branch
      </BaseButton>
    </div>

    <div class="branch-content">
      <div class="operations-section" v-if="operations.length > 0">
        <h4>Operations</h4>
        <p class="subtitle">Complete these operations to advance your mission.</p>

        <div class="operations-list">
          <div v-for="operation in operations" :key="operation.id" class="operation-item"
            :class="{ 'completed': isOperationComplete(operation.id) }">
            <div class="operation-icon">
              {{ operation.type === 'carjacking' ? '🚗' :
                 operation.type === 'goods_smuggling' ? '📦' :
                 operation.type === 'drug_trafficking' ? '💊' :
                 operation.type === 'official_bribing' ? '💼' :
                 operation.type === 'intelligence_gathering' ? '🔍' :
                 operation.type === 'crew_recruitment' ? '👥' : '🎯' }}
            </div>
            <div class="operation-details">
              <h5 class="operation-name">{{ operation.name }}</h5>
              <p class="operation-description">{{ operation.description }}</p>
              <div class="operation-stats">
                <div class="stat">
                  <span class="stat-icon">⏱️</span>
                  <span class="stat-value">{{ Math.floor(operation.duration / 60) }} min</span>
                </div>
                <div class="stat">
                  <span class="stat-icon">📊</span>
                  <span class="stat-value">{{ operation.successRate }}% Success</span>
                </div>
              </div>
            </div>
            <div class="operation-status">
              <span v-if="isOperationComplete(operation.id)" class="status-icon complete">✅</span>
              <span v-else class="status-icon incomplete">⬜</span>
            </div>
          </div>
        </div>
      </div>

      <div class="pois-section" v-if="pois.length > 0">
        <h4>Points of Interest</h4>
        <p class="subtitle">Visit these locations to gather information and make connections.</p>

        <div class="pois-list">
          <div v-for="poi in pois" :key="poi.id" class="poi-item"
            :class="{ 'completed': isPOIComplete(poi.id) }">
            <div class="poi-icon">
              {{ poi.type === 'Bar' ? '🍸' :
                 poi.type === 'Restaurant' ? '🍽️' :
                 poi.type === 'Club' ? '🎭' :
                 poi.type === 'Casino' ? '🎰' :
                 poi.type === 'Hotel' ? '🏨' :
                 poi.type === 'Warehouse' ? '🏭' :
                 poi.type === 'Dock' ? '🚢' :
                 poi.type === 'Factory' ? '🏭' :
                 poi.type === 'Shop' ? '🛒' :
                 poi.type === 'Construction' ? '🏗️' : '📍' }}
            </div>
            <div class="poi-details">
              <h5 class="poi-name">{{ poi.name }}</h5>
              <p class="poi-description">{{ poi.description }}</p>
              <div class="poi-stats">
                <div class="stat">
                  <span class="stat-icon">🏙️</span>
                  <span class="stat-value">{{ poi.cityId }}</span>
                </div>
                <div class="stat">
                  <span class="stat-icon">💼</span>
                  <span class="stat-value">{{ poi.businessType }}</span>
                </div>
              </div>
            </div>
            <div class="poi-status">
              <div v-if="isPOIComplete(poi.id)" class="status-complete">
                <span class="status-icon">✅</span>
                <span class="status-text">Complete</span>
              </div>
              <div v-else class="interaction-buttons">
                <BaseButton size="small" variant="text" @click="interactWithPOI(poi.id, InteractionType.Neutral)">
                  Neutral
                </BaseButton>
                <BaseButton size="small" variant="text" @click="interactWithPOI(poi.id, InteractionType.Convince)">
                  Convince
                </BaseButton>
                <BaseButton size="small" variant="text" @click="interactWithPOI(poi.id, InteractionType.Intimidate)">
                  Intimidate
                </BaseButton>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.branch-view {
  .branch-header {
    margin-bottom: $spacing-xl;
    padding-bottom: $spacing-md;
    border-bottom: 1px solid $border-color;
    display: flex;
    flex-direction: column;
    gap: $spacing-md;

    .branch-info {
      h3 {
        @include gold-accent;
        margin-bottom: $spacing-xs;
      }

      .description {
        color: $text-secondary;
        line-height: 1.6;
      }
    }

    .branch-progress {
      .progress-text {
        @include flex-between;
        margin-bottom: $spacing-xs;
        font-size: $font-size-sm;
        color: $text-secondary;

        .percentage {
          font-weight: 600;
        }
      }

      .progress-bar {
        height: 8px;
        background-color: $background-darker;
        border-radius: $border-radius-sm;
        overflow: hidden;

        .progress-fill {
          height: 100%;
          background-color: $info-color;
          transition: width 0.3s ease;

          &.complete {
            background-color: $success-color;
          }
        }
      }
    }
  }

  .branch-content {
    display: flex;
    flex-direction: column;
    gap: $spacing-xl;

    .operations-section,
    .pois-section {
      h4 {
        margin-bottom: $spacing-xs;
      }

      .subtitle {
        color: $text-secondary;
        margin-bottom: $spacing-md;
      }
    }

    .operations-list,
    .pois-list {
      display: flex;
      flex-direction: column;
      gap: $spacing-md;

      .operation-item,
      .poi-item {
        @include card;
        display: flex;
        align-items: flex-start;
        gap: $spacing-md;
        padding: $spacing-md;
        transition: $transition-base;

        &.completed {
          border-left: 4px solid $success-color;
          opacity: 0.8;
        }

        .operation-icon,
        .poi-icon {
          font-size: 24px;
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
          background-color: $background-darker;
          border-radius: $border-radius-sm;
        }

        .operation-details,
        .poi-details {
          flex: 1;

          .operation-name,
          .poi-name {
            margin-bottom: $spacing-xs;
          }

          .operation-description,
          .poi-description {
            color: $text-secondary;
            font-size: $font-size-sm;
            margin-bottom: $spacing-sm;
            line-height: 1.5;
          }

          .operation-stats,
          .poi-stats {
            display: flex;
            gap: $spacing-md;

            .stat {
              display: flex;
              align-items: center;
              gap: $spacing-xs;
              font-size: $font-size-sm;
              color: $text-secondary;
            }
          }
        }

        .operation-status,
        .poi-status {
          .status-icon {
            font-size: 24px;
          }

          .status-complete {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 4px;

            .status-text {
              font-size: $font-size-sm;
              color: $success-color;
            }
          }

          .interaction-buttons {
            display: flex;
            flex-direction: column;
            gap: $spacing-xs;
          }
        }
      }
    }
  }
}
</style>
