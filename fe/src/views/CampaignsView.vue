<!-- fe/src/views/CampaignsView.vue -->

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import CampaignsList from '@/components/campaign/CampaignsList.vue';
import CampaignMission from '@/components/campaign/CampaignMission.vue';
import { useCampaignStore } from '@/stores/modules/campaign';
import { Campaign, Branch, Mission } from '@/types/campaign';
import campaignService from '@/services/campaignService';

const router = useRouter();
const route = useRoute();
const campaignStore = useCampaignStore();

// View state
const activeView = ref<'list' | 'detail' | 'mission'>('list');
const isLoading = ref(false);

// Modals
const showStartModal = ref(false);
const showCompleteModal = ref(false);
const selectedBranch = ref<Branch | null>(null);

// Computed properties
const campaigns = computed(() => campaignStore.campaigns);
const selectedCampaign = computed(() => campaignStore.selectedCampaign);
const currentMission = computed(() => campaignStore.currentMission);
const hasCampaignStarted = computed(() => campaignStore.hasCampaignStarted);

// Track if viewing a specific mission
const viewingMissionId = ref<string | null>(null);
const viewingMissionData = ref<Mission | null>(null);
const loadingMission = ref(false);

const viewingMission = computed(() => {
  if (!viewingMissionId.value) return null;
  
  // If the viewingMissionId is the current mission, return null to use currentMission instead
  if (currentMission.value && currentMission.value.id === viewingMissionId.value) {
    return null;
  }
  
  // Return the loaded mission data with full details
  return viewingMissionData.value;
});

// Load campaigns on mount and handle route params
onMounted(async () => {
  isLoading.value = true;
  await campaignStore.fetchCampaigns();
  
  // Check if we're on a specific campaign route
  if (route.params.campaignId) {
    const campaignId = route.params.campaignId as string;
    await campaignStore.selectCampaign(campaignId);
    
    // Check if we're on the mission route
    if (route.name === 'CampaignMission') {
      activeView.value = 'mission';
      
      // Check if viewing a specific mission
      if (route.query.missionId) {
        viewingMissionId.value = route.query.missionId as string;
        // Load mission details if it's not the current mission
        if (currentMission.value?.id !== viewingMissionId.value) {
          await loadMissionDetails(viewingMissionId.value);
        }
      }
    } else {
      activeView.value = 'detail';
    }
  }
  
  isLoading.value = false;
});

// Watch for route changes
watch(() => route.params.campaignId, async (newCampaignId) => {
  if (newCampaignId) {
    await campaignStore.selectCampaign(newCampaignId as string);
    
    if (route.name === 'CampaignMission') {
      activeView.value = 'mission';
    } else {
      activeView.value = 'detail';
    }
  } else {
    activeView.value = 'list';
    campaignStore.selectedCampaignId = null;
  }
});

// Watch for mission query changes
watch(() => route.query.missionId, async (newMissionId) => {
  viewingMissionId.value = newMissionId as string | null;
  
  // Load mission details if viewing a past mission
  if (newMissionId && currentMission.value?.id !== newMissionId) {
    await loadMissionDetails(newMissionId as string);
  } else {
    viewingMissionData.value = null;
  }
});

// Load full mission details with branches
async function loadMissionDetails(missionId: string) {
  loadingMission.value = true;
  try {
    const response = await campaignService.getMission(missionId);
    if (response.success && response.data) {
      // Add is_completed flag based on player progress
      const mission = response.data;
      mission.is_completed = campaignStore.isMissionComplete(missionId);
      viewingMissionData.value = mission;
    }
  } catch (error) {
    console.error('Failed to load mission details:', error);
  } finally {
    loadingMission.value = false;
  }
}

// Handle campaign selection
async function selectCampaign(campaign: Campaign) {
  await router.push({ name: 'CampaignDetail', params: { campaignId: campaign.id } });
}

// Handle campaign start
async function startCampaign() {
  if (!selectedCampaign.value) return;

  isLoading.value = true;
  const result = await campaignStore.startCampaign(selectedCampaign.value.id);
  isLoading.value = false;

  if (result && result.success) {
    showStartModal.value = false;

    // If there's a current mission, navigate to it
    if (campaignStore.currentMission && selectedCampaign.value) {
      await router.push({ 
        name: 'CampaignMission', 
        params: { campaignId: selectedCampaign.value.id } 
      });
    }
  }
}


// Show branch completion modal
function showCompleteBranchModal(branch: Branch) {
  selectedBranch.value = branch;
  showCompleteModal.value = true;
}

// Handle branch completion
async function completeBranch() {
  if (!currentMission.value || !selectedBranch.value) return;

  isLoading.value = true;
  const result = await campaignStore.completeBranch(currentMission.value.id, selectedBranch.value.id);
  isLoading.value = false;

  if (result && result.success) {
    showCompleteModal.value = false;
    selectedBranch.value = null;

    // If there's a new current mission, navigate to it
    if (campaignStore.currentMission && selectedCampaign.value) {
      await router.push({ 
        name: 'CampaignMission', 
        params: { campaignId: selectedCampaign.value.id } 
      });
    } else if (selectedCampaign.value) {
      // Campaign completed, go back to detail
      await router.push({ 
        name: 'CampaignDetail', 
        params: { campaignId: selectedCampaign.value.id } 
      });
    }
  }
}

// Navigation
function goBack() {
  if (activeView.value === 'detail') {
    router.push({ name: 'Campaigns' });
  } else if (activeView.value === 'mission' && selectedCampaign.value) {
    router.push({ 
      name: 'CampaignDetail', 
      params: { campaignId: selectedCampaign.value.id } 
    });
  }
}

// Get current chapter and mission info
const currentChapter = computed(() => {
  if (!selectedCampaign.value || !currentMission.value) return null;
  
  for (const chapter of selectedCampaign.value.chapters) {
    const mission = chapter.missions.find(m => m.id === currentMission.value!.id);
    if (mission) {
      return {
        chapter,
        missionIndex: chapter.missions.indexOf(mission),
        totalMissions: chapter.missions.length
      };
    }
  }
  return null;
});

// Get chapter info for any mission
const getMissionChapter = (missionToFind: any) => {
  if (!selectedCampaign.value || !missionToFind) return null;
  
  for (const chapter of selectedCampaign.value.chapters) {
    const mission = chapter.missions.find(m => m.id === missionToFind.id);
    if (mission) {
      return {
        chapter,
        missionIndex: chapter.missions.indexOf(mission),
        totalMissions: chapter.missions.length
      };
    }
  }
  return null;
};

// Check if chapter is complete
function isChapterComplete(chapter: any): boolean {
  return chapter.missions.every((m: any) => m.is_completed);
}

// Get completed branch name for a mission
function getCompletedBranch(mission: any): string | null {
  const progress = campaignStore.currentCampaignProgress;
  if (!progress || !mission.is_completed) return null;
  
  // Find which branch was completed for this mission
  // This would need backend support to track which branch was chosen
  // For now, return null
  return null;
}

// View a specific mission (completed or current)
function viewMission(mission: any) {
  if (mission.is_completed || currentMission.value?.id === mission.id) {
    // Navigate to mission view
    router.push({ 
      name: 'CampaignMission', 
      params: { campaignId: selectedCampaign.value!.id },
      query: { missionId: mission.id }
    });
  }
}

// Go back to current mission
function backToCurrent() {
  viewingMissionId.value = null;
  router.push({ 
    name: 'CampaignMission', 
    params: { campaignId: selectedCampaign.value!.id }
  });
}
</script>

<template>
  <div class="campaigns-view">
    <div class="page-header">
      <h2>Campaigns</h2>
      <p class="subtitle">Embark on narrative-driven missions to expand your criminal empire.</p>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading...</p>
    </div>

    <!-- Campaign List -->
    <div v-else-if="activeView === 'list'" class="campaigns-list">
      <CampaignsList :campaigns="campaigns" @select-campaign="selectCampaign" />
    </div>

    <!-- Campaign Detail -->
    <div v-else-if="activeView === 'detail'" class="campaign-detail">
      <div class="back-button">
        <BaseButton variant="text" icon="arrow-left" @click="goBack">Back to Campaigns</BaseButton>
      </div>

      <div v-if="selectedCampaign" class="campaign-content">
        <div class="campaign-header">
          <div class="campaign-info">
            <h3>{{ selectedCampaign.name }}</h3>
            <p class="description">{{ selectedCampaign.description }}</p>
          </div>

          <div class="campaign-actions">
            <BaseButton v-if="!hasCampaignStarted" @click="showStartModal = true">
              Start Campaign
            </BaseButton>
            <BaseButton v-else-if="currentMission" @click="router.push({ name: 'CampaignMission', params: { campaignId: selectedCampaign.id } })">
              Continue Campaign
            </BaseButton>
            <BaseButton v-else variant="outline" disabled>
              Campaign Completed
            </BaseButton>
          </div>
        </div>

        <div class="campaign-chapters">
          <h4>Campaign Progress</h4>
          <div class="chapters-list">
            <div v-for="chapter in selectedCampaign.chapters" :key="chapter.id" class="chapter-card">
              <div class="chapter-header">
                <h5>Chapter {{ chapter.order }}: {{ chapter.name }}</h5>
                <span class="chapter-status" v-if="isChapterComplete(chapter)">
                  <i class="fas fa-check-circle"></i> Complete
                </span>
              </div>
              <p class="chapter-description">{{ chapter.description }}</p>
              
              <div class="missions-list">
                <div 
                  v-for="mission in chapter.missions" 
                  :key="mission.id" 
                  class="mission-item"
                  :class="{
                    'completed': mission.is_completed,
                    'current': currentMission?.id === mission.id,
                    'clickable': mission.is_completed || currentMission?.id === mission.id
                  }"
                  @click="viewMission(mission)"
                >
                  <div class="mission-status">
                    <i v-if="mission.is_completed" class="fas fa-check-circle"></i>
                    <i v-else-if="currentMission?.id === mission.id" class="fas fa-play-circle"></i>
                    <i v-else class="fas fa-lock"></i>
                  </div>
                  <div class="mission-info">
                    <span class="mission-name">Mission {{ mission.order }}: {{ mission.name }}</span>
                    <span class="mission-path" v-if="getCompletedBranch(mission)">
                      <i class="fas fa-route"></i> {{ getCompletedBranch(mission) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Mission Detail -->
    <div v-else-if="activeView === 'mission'" class="mission-detail">
      <div class="back-button">
        <BaseButton variant="text" icon="arrow-left" @click="goBack">
          Back to Campaign
        </BaseButton>
      </div>

      <div v-if="viewingMission" class="mission-context viewing-mode">
        <p class="context-info">
          <i class="fas fa-eye"></i> Viewing Past Mission
        </p>
        <BaseButton variant="text" size="small" @click="backToCurrent">
          <i class="fas fa-arrow-left"></i> Back to Current Mission
        </BaseButton>
      </div>
      <div v-else-if="currentMission" class="mission-context">
        <template v-if="getMissionChapter(currentMission)">
          <p class="context-info">
            {{ selectedCampaign?.name }} - Chapter {{ getMissionChapter(currentMission)?.chapter.order }}: {{ getMissionChapter(currentMission)?.chapter.name }}
          </p>
          <p class="mission-progress">
            Mission {{ currentMission.order }} of {{ getMissionChapter(currentMission)?.totalMissions }}
          </p>
        </template>
      </div>

      <div v-if="loadingMission" class="loading-state">
        <div class="spinner"></div>
        <p>Loading mission details...</p>
      </div>
      
      <CampaignMission
        v-else-if="viewingMission || currentMission"
        :mission="viewingMission || currentMission!"
        @complete-branch="showCompleteBranchModal"
      />
    </div>

    <!-- Start Campaign Modal -->
    <BaseModal v-model="showStartModal" title="Start Campaign">
      <div v-if="selectedCampaign" class="start-campaign-modal">
        <p>Are you ready to embark on the "{{ selectedCampaign.name }}" campaign?</p>
        <p>This will be your first step into a larger criminal world. You'll need to make strategic decisions that will shape your future in the city.</p>
      </div>

      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton variant="text" @click="showStartModal = false">
            Cancel
          </BaseButton>
          <BaseButton variant="primary" :loading="isLoading" @click="startCampaign">
            Start Campaign
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- Complete Branch Modal -->
    <BaseModal v-model="showCompleteModal" title="Complete Branch">
      <div v-if="selectedBranch" class="complete-branch-modal">
        <p>You have completed all objectives for the "{{ selectedBranch.name }}" branch.</p>
        <p>Are you ready to move forward with your decision?</p>
        <p class="warning">Warning: This choice will be permanent and will affect your story progression.</p>
      </div>

      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton variant="text" @click="showCompleteModal = false">
            Not Yet
          </BaseButton>
          <BaseButton variant="primary" :loading="isLoading" @click="completeBranch">
            Complete Branch
          </BaseButton>
        </div>
      </template>
    </BaseModal>

  </div>
</template>

<style lang="scss">
.campaigns-view {
  .page-header {
    margin-bottom: $spacing-xl;

    h2 {
      @include gold-accent;
      margin-bottom: $spacing-xs;
    }

    .subtitle {
      color: $text-secondary;
    }
  }

  .back-button {
    margin-bottom: $spacing-md;
  }

  .loading-state {
    @include flex-center;
    flex-direction: column;
    gap: $spacing-md;
    padding: $spacing-xl 0;

    .spinner {
      width: 40px;
      height: 40px;
      border: 4px solid rgba($gold-color, 0.1);
      border-radius: 50%;
      border-top: 4px solid $gold-color;
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      0% {
        transform: rotate(0deg);
      }
      100% {
        transform: rotate(360deg);
      }
    }
  }

  .campaign-header {
    @include flex-between;
    margin-bottom: $spacing-lg;
    padding-bottom: $spacing-md;
    border-bottom: 1px solid $border-color;

    .campaign-info {
      h3 {
        @include gold-accent;
        margin-bottom: $spacing-xs;
      }

      .description {
        color: $text-secondary;
      }
    }
  }

  .campaign-chapters {
    h4 {
      margin-bottom: $spacing-md;
    }

    .chapters-list {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: $spacing-md;

      .chapter-card {
        @include card;
        padding: $spacing-md;

        .chapter-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: $spacing-xs;

          h5 {
            margin: 0;
          }

          .chapter-status {
            color: $success-color;
            font-size: $font-size-sm;
            display: flex;
            align-items: center;
            gap: $spacing-xs;
          }
        }

        .chapter-description {
          color: $text-secondary;
          font-size: $font-size-sm;
          margin-bottom: $spacing-md;
        }

        .missions-list {
          display: flex;
          flex-direction: column;
          gap: $spacing-sm;

          .mission-item {
            display: flex;
            align-items: center;
            gap: $spacing-sm;
            padding: $spacing-sm;
            background: rgba($background-dark, 0.5);
            border: 1px solid $border-color;
            border-radius: $border-radius-sm;
            transition: $transition-base;

            &.completed {
              opacity: 0.8;
              
              .mission-status {
                color: $success-color;
              }
            }

            &.current {
              border-color: $gold-color;
              background: rgba($gold-color, 0.1);
              
              .mission-status {
                color: $gold-color;
              }
            }

            &.clickable {
              cursor: pointer;

              &:hover {
                background: rgba($gold-color, 0.15);
                transform: translateX(4px);
              }
            }

            .mission-status {
              font-size: 16px;
              color: $text-secondary;
            }

            .mission-info {
              flex: 1;
              display: flex;
              justify-content: space-between;
              align-items: center;

              .mission-name {
                color: $text-primary;
                font-size: $font-size-sm;
              }

              .mission-path {
                color: $text-secondary;
                font-size: $font-size-xs;
                display: flex;
                align-items: center;
                gap: $spacing-xs;
              }
            }
          }
        }
      }
    }
  }

  .mission-context {
    background: rgba($background-dark, 0.6);
    border: 1px solid $border-color;
    border-radius: $border-radius;
    padding: $spacing-md;
    margin-bottom: $spacing-md;

    &.viewing-mode {
      background: rgba($info-color, 0.1);
      border-color: $info-color;
      display: flex;
      justify-content: space-between;
      align-items: center;

      .context-info {
        display: flex;
        align-items: center;
        gap: $spacing-sm;
        margin: 0;
        color: $info-color;
      }
    }

    .context-info {
      color: $text-primary;
      margin-bottom: $spacing-xs;
    }

    .mission-progress {
      color: $gold-color;
      font-weight: 500;
    }
  }

  .start-campaign-modal,
  .complete-branch-modal {
    p {
      margin-bottom: $spacing-md;
      line-height: 1.5;

      &.warning {
        color: $warning-color;
        font-weight: 500;
      }
    }
  }

  .modal-footer-actions {
    @include flex-between;
    width: 100%;
  }
}
</style>