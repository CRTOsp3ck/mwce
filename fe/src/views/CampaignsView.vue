<!-- fe/src/views/CampaignsView.vue -->

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import CampaignsList from '@/components/campaign/CampaignsList.vue';
import CampaignMission from '@/components/campaign/CampaignMission.vue';
import CampaignBranch from '@/components/campaign/CampaignBranch.vue';
import { useCampaignStore } from '@/stores/modules/campaign';
import { Campaign, Branch, Mission, InteractionType } from '@/types/campaign';

const router = useRouter();
const campaignStore = useCampaignStore();

// View state
const activeView = ref<'list' | 'detail' | 'mission' | 'branch'>('list');
const isLoading = ref(false);

// Modals
const showStartModal = ref(false);
const showCompleteModal = ref(false);
const showDialogueModal = ref(false);
const selectedInteractionType = ref<InteractionType | null>(null);
const dialogueText = ref('');
const dialogueResponse = ref('');
const resourceEffects = ref<{ [key: string]: number }>({});

// Computed properties
const campaigns = computed(() => campaignStore.campaigns);
const selectedCampaign = computed(() => campaignStore.selectedCampaign);
const currentMission = computed(() => campaignStore.currentMission);
const currentBranch = computed(() => campaignStore.selectedBranch);
const hasCampaignStarted = computed(() => campaignStore.hasCampaignStarted);
const isBranchCompletable = computed(() => campaignStore.isBranchCompletable);

// Load campaigns on mount
onMounted(async () => {
  isLoading.value = true;
  await campaignStore.fetchCampaigns();
  isLoading.value = false;
});

// Handle campaign selection
async function selectCampaign(campaign: Campaign) {
  await campaignStore.selectCampaign(campaign.id);
  activeView.value = 'detail';
}

// Handle campaign start
async function startCampaign() {
  if (!selectedCampaign.value) return;

  isLoading.value = true;
  const result = await campaignStore.startCampaign(selectedCampaign.value.id);
  isLoading.value = false;

  if (result && result.success) {
    showStartModal.value = false;

    // If there's a current mission, show it
    if (campaignStore.currentMission) {
      activeView.value = 'mission';
    }
  }
}

// Handle branch selection
async function selectBranch(mission: Mission, branch: Branch) {
  isLoading.value = true;
  const result = await campaignStore.selectBranch(mission.id, branch.id);
  isLoading.value = false;

  if (result && result.success) {
    activeView.value = 'branch';
  }
}

// Handle branch completion
async function completeBranch() {
  if (!currentMission.value || !currentBranch.value) return;

  isLoading.value = true;
  const result = await campaignStore.completeBranch(currentMission.value.id, currentBranch.value.id);
  isLoading.value = false;

  if (result && result.success) {
    showCompleteModal.value = false;

    // If there's a new current mission, show it
    if (campaignStore.currentMission) {
      activeView.value = 'mission';
    } else {
      // Campaign completed
      activeView.value = 'detail';
    }
  }
}

// Handle POI interaction
async function interactWithPOI(poiId: string, interactionType: InteractionType) {
  selectedInteractionType.value = interactionType;
  isLoading.value = true;

  const result = await campaignStore.interactWithPOI(poiId, interactionType);
  isLoading.value = false;

  if (result) {
    dialogueText.value = result.dialogue.text;
    dialogueResponse.value = "";

    // If there are resource effects, show them
    if (result.resourceEffect) {
      resourceEffects.value = {};

      if (result.resourceEffect.money) resourceEffects.value['money'] = result.resourceEffect.money;
      if (result.resourceEffect.crew) resourceEffects.value['crew'] = result.resourceEffect.crew;
      if (result.resourceEffect.weapons) resourceEffects.value['weapons'] = result.resourceEffect.weapons;
      if (result.resourceEffect.vehicles) resourceEffects.value['vehicles'] = result.resourceEffect.vehicles;
      if (result.resourceEffect.respect) resourceEffects.value['respect'] = result.resourceEffect.respect;
      if (result.resourceEffect.influence) resourceEffects.value['influence'] = result.resourceEffect.influence;
      if (result.resourceEffect.heat) resourceEffects.value['heat'] = result.resourceEffect.heat;
    }

    showDialogueModal.value = true;
  }
}

// Navigation
function goBack() {
  if (activeView.value === 'detail') {
    activeView.value = 'list';
    campaignStore.selectedCampaignId = null;
  } else if (activeView.value === 'mission') {
    activeView.value = 'detail';
  } else if (activeView.value === 'branch') {
    activeView.value = 'mission';
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
            <BaseButton v-else-if="currentMission" @click="activeView = 'mission'">
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
        @select-branch="selectBranch"
      />
    </div>

    <!-- Branch Detail -->
    <div v-else-if="activeView === 'branch'" class="branch-detail">
      <div class="back-button">
        <BaseButton variant="text" icon="arrow-left" @click="goBack">
          Back to Mission
        </BaseButton>
      </div>

      <CampaignBranch
        v-if="currentBranch"
        :branch="currentBranch"
        :mission="currentMission"
        @interact-with-poi="interactWithPOI"
        @complete-branch="showCompleteModal = true"
        :can-complete="isBranchCompletable"
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
      <div v-if="currentBranch" class="complete-branch-modal">
        <p>You have completed all objectives for the "{{ currentBranch.name }}" branch.</p>
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

    <!-- Dialogue Modal -->
    <BaseModal v-model="showDialogueModal" :title="selectedInteractionType ? `${selectedInteractionType} Interaction` : 'Dialogue'">
      <div class="dialogue-modal">
        <div class="dialogue-text">
          {{ dialogueText }}
        </div>

        <div v-if="Object.keys(resourceEffects).length > 0" class="resource-effects">
          <h4>Effects:</h4>
          <div class="effects-list">
            <div v-for="(value, key) in resourceEffects" :key="key" class="effect-item" :class="{ 'positive': value > 0, 'negative': value < 0 }">
              <span class="effect-icon">
                {{ key === 'money' ? '💰' : key === 'crew' ? '👥' : key === 'weapons' ? '🔫' : key === 'vehicles' ? '🚗' :
                   key === 'respect' ? '👊' : key === 'influence' ? '🏛️' : key === 'heat' ? '🔥' : '❓' }}
              </span>
              <span class="effect-label">{{ key }}:</span>
              <span class="effect-value">{{ value > 0 ? '+' : '' }}{{ value }}</span>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton variant="primary" @click="showDialogueModal = false">
            Continue
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

  .dialogue-modal {
    .dialogue-text {
      @include card;
      padding: $spacing-md;
      margin-bottom: $spacing-md;
      line-height: 1.6;
    }

    .resource-effects {
      h4 {
        margin-bottom: $spacing-sm;
      }

      .effects-list {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: $spacing-sm;

        .effect-item {
          // @include flex-align-center;
          gap: $spacing-xs;
          padding: $spacing-xs;
          border-radius: $border-radius-sm;

          &.positive {
            background-color: rgba($success-color, 0.1);
            color: $success-color;
          }

          &.negative {
            background-color: rgba($danger-color, 0.1);
            color: $danger-color;
          }

          .effect-icon {
            font-size: 1.2rem;
          }

          .effect-label {
            font-weight: 500;
            text-transform: capitalize;
          }
        }
      }
    }
  }

  .modal-footer-actions {
    @include flex-between;
    width: 100%;
  }
}
</style>
