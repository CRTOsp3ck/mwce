<!-- fe/src/components/campaign/CampaignMission.vue -->

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { Mission, Branch, CampaignOperation, CampaignPOI } from '@/types/campaign';
import { useCampaignStore } from '@/stores/modules/campaign';
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
const selectedBranchId = ref<string | null>(null);
const branchesProgress = ref<Record<string, boolean>>({});
const branchOperations = ref<Record<string, CampaignOperation[]>>({});
const branchPOIs = ref<Record<string, CampaignPOI[]>>({});
const loadingBranchData = ref<Record<string, boolean>>({});

// Get the selected branch or first branch
const selectedBranch = computed(() => {
  if (selectedBranchId.value) {
    return branches.value.find(b => b.id === selectedBranchId.value);
  }
  return branches.value[0];
});

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

  // Load data for ALL branches to ensure progress bars show correctly
  if (branches.value.length > 0) {
    await Promise.all(branches.value.map(branch => loadBranchData(branch)));
  }

  // For completed missions, find which branch was completed
  if (props.mission.is_completed && branches.value.length > 0) {
    // Check which branch was completed by looking at completedBranchIds
    const completedBranch = branches.value.find(branch =>
      campaignStore.isBranchComplete(branch.id)
    );

    if (completedBranch) {
      selectedBranchId.value = completedBranch.id;
    } else {
      // Fallback to first branch if we can't determine which was completed
      selectedBranchId.value = branches.value[0].id;
    }
  } else if (branches.value.length > 0) {
    // For current mission or if not completed, load first branch by default
    selectedBranchId.value = branches.value[0].id;
  }
});

async function selectBranch(branch: Branch) {
  selectedBranchId.value = branch.id;
  // Data should already be loaded on mount, but load if missing
  if (!branchOperations.value[branch.id] && !branchPOIs.value[branch.id]) {
    await loadBranchData(branch);
  }
}

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

// Helper to check if a branch was the one completed for this mission
function isBranchCompleted(branchId: string): boolean {
  return props.mission.is_completed === true && campaignStore.isBranchComplete(branchId);
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
      <div class="mission-title-section">
        <h3>Mission {{ mission.order }}: {{ mission.name }}</h3>
        <span class="mission-status-badge" v-if="mission.is_completed">
          <i class="fas fa-check-circle"></i> Completed
        </span>
      </div>
      <p class="description">{{ mission.description }}</p>
    </div>

    <div class="mission-branches">
      <div class="branch-selector" v-if="branches.length > 1">
        <h4>Choose Your Approach</h4>
        <div class="branch-tabs">
          <button
            v-for="(branch, index) in branches"
            :key="branch.id"
            class="branch-tab"
            :class="{
              active: selectedBranch?.id === branch.id,
              completed: isBranchCompleted(branch.id)
            }"
            @click="selectBranch(branch)"
          >
            <span class="branch-number">Path {{ index + 1 }}</span>
            <span class="branch-name">{{ branch.name }}</span>
            <span v-if="isBranchCompleted(branch.id)" class="branch-completed-badge">
              <i class="fas fa-check-circle"></i> Chosen
            </span>
            <div class="branch-progress-mini">
              <div class="progress-bar-mini">
                <div
                  class="progress-fill-mini"
                  :style="{ width: `${getBranchProgress(branch.id).percentage}%` }"
                ></div>
              </div>
              <span class="progress-text-mini">{{ getBranchProgress(branch.id).percentage }}%</span>
            </div>
          </button>
        </div>
      </div>

      <div v-if="selectedBranch" class="selected-branch-section">
        <div class="branch-header">
          <div class="branch-info">
            <h4 class="branch-title">{{ selectedBranch.name }}</h4>
            <p class="branch-description">{{ selectedBranch.description }}</p>
          </div>

          <div class="branch-progress">
            <div class="progress-text">
              <span>Progress: {{ getBranchProgress(selectedBranch.id).completed }} / {{ getBranchProgress(selectedBranch.id).total }} Complete</span>
              <span class="percentage">{{ getBranchProgress(selectedBranch.id).percentage }}%</span>
            </div>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: `${getBranchProgress(selectedBranch.id).percentage}%` }"
                :class="{ 'complete': getBranchProgress(selectedBranch.id).percentage === 100 }"></div>
            </div>
          </div>

          <BaseButton
            v-if="branchesProgress[selectedBranch.id] && !mission.is_completed"
            variant="primary"
            @click="completeBranch(selectedBranch)"
          >
            Complete This Path
          </BaseButton>
          <div v-else-if="mission.is_completed && isBranchCompleted(selectedBranch.id)" class="branch-completed-status">
            <i class="fas fa-check-circle"></i> You completed this path
          </div>
        </div>

        <div v-if="loadingBranchData[selectedBranch.id]" class="loading-state">
          <div class="spinner"></div>
          <p>Loading activities...</p>
        </div>

        <div v-else class="branch-activities">
          <!-- Operations -->
          <div v-if="branchOperations[selectedBranch.id]?.length > 0" class="operations-section">
            <h5><i class="fas fa-tasks"></i> Operations to Complete</h5>
            <p class="section-hint">Complete these operations from the Operations menu.</p>
            <div class="activities-grid">
              <div v-for="operation in branchOperations[selectedBranch.id]" :key="operation.id"
                class="activity-card operation-card"
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
          <div v-if="branchPOIs[selectedBranch.id]?.length > 0" class="pois-section">
            <h5><i class="fas fa-map-marker-alt"></i> Locations to Control</h5>
            <p class="section-hint">Visit these locations in the Territory view and take them over.</p>
            <div class="activities-grid">
              <div v-for="poi in branchPOIs[selectedBranch.id]" :key="poi.id"
                class="activity-card poi-card"
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
</template>

<style lang="scss" scoped>
.mission-view {
  .mission-header {
    margin-bottom: $spacing-xl;
    padding-bottom: $spacing-md;
    border-bottom: 1px solid $border-color;

    .mission-title-section {
      display: flex;
      align-items: center;
      gap: $spacing-md;
      margin-bottom: $spacing-xs;

      h3 {
        @include gold-accent;
        margin: 0;
      }

      .mission-status-badge {
        background: $success-color;
        color: white;
        padding: $spacing-xs $spacing-sm;
        border-radius: $border-radius-sm;
        font-size: $font-size-sm;
        display: flex;
        align-items: center;
        gap: $spacing-xs;
      }
    }

    .description {
      color: $text-secondary;
      line-height: 1.6;
    }
  }

  .branch-selector {
    margin-bottom: $spacing-xl;

    h4 {
      margin-bottom: $spacing-md;
      color: $text-primary;
    }

    .branch-tabs {
      display: flex;
      gap: $spacing-md;
      flex-wrap: wrap;

      .branch-tab {
        flex: 1;
        min-width: 250px;
        background: $background-card;
        border: 1px solid $border-color;
        border-radius: $border-radius;
        padding: $spacing-md;
        cursor: pointer;
        transition: $transition-base;
        position: relative;

        &:hover {
          border-color: $gold-color;
          transform: translateY(-2px);
        }

        &.active {
          border-color: $gold-color;
          background: rgba($gold-color, 0.1);

          &::before {
            content: 'Viewing';
            position: absolute;
            top: -10px;
            right: $spacing-md;
            background: $gold-color;
            color: $background-dark;
            padding: 2px 8px;
            border-radius: $border-radius-sm;
            font-size: $font-size-xs;
            font-weight: 600;
            text-transform: uppercase;
          }
        }

        &.completed {
          border-color: $success-color;
          background: rgba($success-color, 0.05);
        }

        .branch-completed-badge {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          color: $success-color;
          font-size: $font-size-sm;
          font-weight: 500;
          margin-top: $spacing-xs;

          i {
            font-size: 14px;
          }
        }

        .branch-number {
          display: block;
          color: $gold-color;
          font-size: $font-size-sm;
          font-weight: 600;
          text-transform: uppercase;
          margin-bottom: $spacing-xs;
        }

        .branch-name {
          display: block;
          color: $text-primary;
          font-weight: 500;
          margin-bottom: $spacing-sm;
        }

        .branch-progress-mini {
          .progress-bar-mini {
            height: 4px;
            background: rgba($gold-color, 0.2);
            border-radius: 2px;
            overflow: hidden;
            margin-bottom: $spacing-xs;

            .progress-fill-mini {
              height: 100%;
              background: $gold-color;
              transition: width 0.3s ease;
            }
          }

          .progress-text-mini {
            font-size: $font-size-xs;
            color: $text-secondary;
          }
        }
      }
    }
  }

  .selected-branch-section {
    .branch-header {
      @include card;
      margin-bottom: $spacing-xl;
      display: flex;
      flex-direction: column;
      gap: $spacing-md;

      .branch-info {
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
        .progress-text {
          @include flex-between;
          margin-bottom: $spacing-xs;
          font-size: $font-size-sm;
          color: $text-secondary;

          .percentage {
            font-weight: 600;
            color: $gold-color;
          }
        }

        .progress-bar {
          height: 12px;
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

      .branch-completed-status {
        color: $success-color;
        font-weight: 500;
        display: flex;
        align-items: center;
        gap: $spacing-xs;

        i {
          font-size: 18px;
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
      gap: $spacing-xl;

      .operations-section,
      .pois-section {
        h5 {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          margin-bottom: $spacing-xs;
          color: $text-primary;

          i {
            color: $gold-color;
          }
        }

        .section-hint {
          color: $text-tertiary;
          font-size: $font-size-sm;
          margin-bottom: $spacing-md;
        }

        .activities-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
          gap: $spacing-md;

          .activity-card {
            display: flex;
            align-items: flex-start;
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
              font-size: 24px;
              width: 40px;
              height: 40px;
              display: flex;
              align-items: center;
              justify-content: center;
              background-color: $background-dark;
              border-radius: $border-radius-sm;
              flex-shrink: 0;
            }

            .activity-info {
              flex: 1;

              h6 {
                margin-bottom: $spacing-xs;
                font-weight: 500;
                color: $text-primary;
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
</style>
