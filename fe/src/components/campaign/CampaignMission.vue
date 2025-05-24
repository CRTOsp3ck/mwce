<!-- fe/src/components/campaign/CampaignMission.vue -->

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Mission, Branch, CampaignOperation, CampaignPOI } from '@/types/campaign';
import { useCampaignStore } from '@/stores/modules/campaign';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import campaignService from '@/services/campaignService';

const props = defineProps<{
  mission: Mission;
}>();

const emit = defineEmits<{
  (e: 'completeBranch', branch: Branch): void;
}>();

const campaignStore = useCampaignStore();

const branches = computed(() => props.mission.branches || []);
const branchesProgress = ref<Record<string, boolean>>({});
const branchOperations = ref<Record<string, CampaignOperation[]>>({});
const branchPOIs = ref<Record<string, CampaignPOI[]>>({});
const loadingBranchData = ref<Record<string, boolean>>({});

onMounted(async () => {
  // Load branch progress
  try {
    const progressResponse = await campaignService.getMissionBranchesProgress(props.mission.id);
    if (progressResponse.data) {
      branchesProgress.value = progressResponse.data;
    }
  } catch (error) {
    console.error('Failed to load branch progress:', error);
  }

  // Load operations and POIs for each branch
  for (const branch of branches.value) {
    loadBranchData(branch);
  }
});

async function loadBranchData(branch: Branch) {
  loadingBranchData.value[branch.id] = true;
  
  try {
    // Load operations
    const opsResponse = await campaignService.getOperationsByBranchId(branch.id);
    if (opsResponse.data) {
      branchOperations.value[branch.id] = opsResponse.data;
    }

    // Load POIs
    const poisResponse = await campaignService.getPOIsByBranchId(branch.id);
    if (poisResponse.data) {
      branchPOIs.value[branch.id] = poisResponse.data;
    }
  } catch (error) {
    console.error(`Failed to load data for branch ${branch.id}:`, error);
  } finally {
    loadingBranchData.value[branch.id] = false;
  }
}

function completeBranch(branch: Branch) {
  emit('completeBranch', branch);
}

function getBranchProgress(branchId: string): { completed: number; total: number; percentage: number } {
  const operations = branchOperations.value[branchId] || [];
  const pois = branchPOIs.value[branchId] || [];
  const total = operations.length + pois.length;
  
  if (total === 0) {
    return { completed: 0, total: 0, percentage: 0 };
  }

  const completedOps = operations.filter(op => campaignStore.isOperationComplete(op.id)).length;
  const completedPOIs = pois.filter(poi => campaignStore.isPOIComplete(poi.id)).length;
  const completed = completedOps + completedPOIs;
  
  return {
    completed,
    total,
    percentage: Math.round((completed / total) * 100)
  };
}

function getLocationDisplay(poi: CampaignPOI): string {
  if (poi.metadata?.fullLocation) {
    return poi.metadata.fullLocation;
  }
  const parts = [];
  if (poi.metadata?.cityName) parts.push(poi.metadata.cityName);
  if (poi.metadata?.districtName) parts.push(poi.metadata.districtName);
  if (poi.metadata?.regionName) parts.push(poi.metadata.regionName);
  return parts.length > 0 ? parts.join(', ') : poi.cityId;
}

function getOperationRegionsDisplay(operation: CampaignOperation): string {
  return operation.metadata?.regionsDisplay || 'All Regions';
}
</script>

<template>
  <div class="mission-view">
    <div class="mission-header">
      <h3>{{ mission.name }}</h3>
      <p class="description">{{ mission.description }}</p>
    </div>

    <div class="mission-branches">
      <h4>Available Approaches</h4>
      <p class="subtitle">Complete all activities in any branch to progress. Each branch represents a different approach to this mission.</p>

      <div class="branches-list">
        <div v-for="branch in branches" :key="branch.id" class="branch-section">
          <div class="branch-header">
            <div class="branch-info">
              <h4 class="branch-title">{{ branch.name }}</h4>
              <p class="branch-description">{{ branch.description }}</p>
            </div>
            
            <div class="branch-progress">
              <div class="progress-text">
                <span>{{ getBranchProgress(branch.id).completed }} / {{ getBranchProgress(branch.id).total }} Complete</span>
                <span class="percentage">{{ getBranchProgress(branch.id).percentage }}%</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: `${getBranchProgress(branch.id).percentage}%` }"
                  :class="{ 'complete': getBranchProgress(branch.id).percentage === 100 }"></div>
              </div>
            </div>

            <BaseButton 
              v-if="branchesProgress[branch.id]" 
              variant="primary" 
              @click="completeBranch(branch)"
            >
              Complete Branch
            </BaseButton>
          </div>

          <div v-if="loadingBranchData[branch.id]" class="loading-state">
            <div class="spinner"></div>
            <p>Loading activities...</p>
          </div>

          <div v-else class="branch-activities">
            <!-- Operations -->
            <div v-if="branchOperations[branch.id]?.length > 0" class="operations-section">
              <h5>Operations</h5>
              <div class="activities-list">
                <div v-for="operation in branchOperations[branch.id]" :key="operation.id" 
                  class="activity-item" 
                  :class="{ 'completed': campaignStore.isOperationComplete(operation.id) }">
                  <div class="activity-icon">🎯</div>
                  <div class="activity-info">
                    <h6>{{ operation.name }}</h6>
                    <p>{{ operation.description }}</p>
                    <div class="activity-meta">
                      <span>⏱️ {{ Math.floor(operation.duration / 60) }} min</span>
                      <span>📊 {{ operation.successRate }}% Success</span>
                      <span>🏙️ {{ getOperationRegionsDisplay(operation) }}</span>
                    </div>
                  </div>
                  <div class="activity-status">
                    <span v-if="campaignStore.isOperationComplete(operation.id)" class="status-icon">✅</span>
                    <span v-else class="status-icon">⬜</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- POIs -->
            <div v-if="branchPOIs[branch.id]?.length > 0" class="pois-section">
              <h5>Points of Interest</h5>
              <p class="section-note">Visit these locations in the Territory view and take them over to complete.</p>
              <div class="activities-list">
                <div v-for="poi in branchPOIs[branch.id]" :key="poi.id" 
                  class="activity-item" 
                  :class="{ 'completed': campaignStore.isPOIComplete(poi.id) }">
                  <div class="activity-icon">📍</div>
                  <div class="activity-info">
                    <h6>{{ poi.name }}</h6>
                    <p>{{ poi.description }}</p>
                    <div class="activity-meta">
                      <span>🏙️ {{ getLocationDisplay(poi) }}</span>
                      <span>💼 {{ poi.businessType }}</span>
                    </div>
                  </div>
                  <div class="activity-status">
                    <span v-if="campaignStore.isPOIComplete(poi.id)" class="status-icon">✅</span>
                    <span v-else class="status-icon">⬜</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.mission-view {
  .mission-header {
    margin-bottom: $spacing-xl;
    padding-bottom: $spacing-md;
    border-bottom: 1px solid $border-color;

    h3 {
      @include gold-accent;
      margin-bottom: $spacing-xs;
    }

    .description {
      color: $text-secondary;
      line-height: 1.6;
    }
  }

  .mission-branches {
    h4 {
      margin-bottom: $spacing-xs;
    }

    .subtitle {
      color: $text-secondary;
      margin-bottom: $spacing-lg;
    }

    .branches-list {
      display: flex;
      flex-direction: column;
      gap: $spacing-xl;

      .branch-section {
        @include card;
        padding: $spacing-lg;

        .branch-header {
          display: flex;
          align-items: flex-start;
          gap: $spacing-md;
          margin-bottom: $spacing-lg;
          padding-bottom: $spacing-md;
          border-bottom: 1px solid $border-color;

          .branch-info {
            flex: 1;

            .branch-title {
              @include gold-accent;
              margin-bottom: $spacing-xs;
            }

            .branch-description {
              color: $text-secondary;
              line-height: 1.5;
            }
          }

          .branch-progress {
            min-width: 200px;

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

        .loading-state {
          @include flex-center;
          flex-direction: column;
          gap: $spacing-md;
          padding: $spacing-xl;

          .spinner {
            width: 30px;
            height: 30px;
            border: 3px solid rgba($gold-color, 0.1);
            border-radius: 50%;
            border-top: 3px solid $gold-color;
            animation: spin 1s linear infinite;
          }

          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        }

        .branch-activities {
          display: flex;
          flex-direction: column;
          gap: $spacing-lg;

          .operations-section,
          .pois-section {
            h5 {
              margin-bottom: $spacing-xs;
              color: $text-primary;
            }

            .section-note {
              color: $text-tertiary;
              font-size: $font-size-sm;
              font-style: italic;
              margin-bottom: $spacing-md;
            }

            .activities-list {
              display: flex;
              flex-direction: column;
              gap: $spacing-sm;

              .activity-item {
                display: flex;
                align-items: center;
                gap: $spacing-md;
                padding: $spacing-md;
                background-color: $background-darker;
                border-radius: $border-radius;
                transition: $transition-base;

                &.completed {
                  opacity: 0.7;
                  border-left: 3px solid $success-color;
                }

                &:hover {
                  background-color: $background-lighter;
                }

                .activity-icon {
                  font-size: 20px;
                  width: 36px;
                  height: 36px;
                  display: flex;
                  align-items: center;
                  justify-content: center;
                  background-color: $background-dark;
                  border-radius: $border-radius-sm;
                }

                .activity-info {
                  flex: 1;

                  h6 {
                    margin-bottom: $spacing-xs;
                    font-weight: 500;
                  }

                  p {
                    color: $text-secondary;
                    font-size: $font-size-sm;
                    margin-bottom: $spacing-xs;
                    line-height: 1.4;
                  }

                  .activity-meta {
                    display: flex;
                    gap: $spacing-md;
                    flex-wrap: wrap;
                    font-size: $font-size-xs;
                    color: $text-tertiary;

                    span {
                      display: flex;
                      align-items: center;
                      gap: 4px;
                    }
                  }
                }

                .activity-status {
                  .status-icon {
                    font-size: 20px;
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
</style>
