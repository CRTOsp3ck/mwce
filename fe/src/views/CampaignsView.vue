<!-- fe/src/views/CampaignsView.vue -->

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import CampaignsList from '@/components/campaign/CampaignsList.vue';
import CampaignMission from '@/components/campaign/CampaignMission.vue';
import { useCampaignStore } from '@/stores/modules/campaign';
import { Campaign, Branch } from '@/types/campaign';

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
          <h4>Campaign Chapters</h4>
          <div class="chapters-list">
            <div v-for="chapter in selectedCampaign.chapters" :key="chapter.id" class="chapter-card">
              <h5>{{ chapter.name }}</h5>
              <p>{{ chapter.description }}</p>
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

      <CampaignMission
        v-if="currentMission"
        :mission="currentMission"
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

        h5 {
          margin-bottom: $spacing-xs;
        }

        p {
          color: $text-secondary;
          font-size: $font-size-sm;
        }
      }
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
