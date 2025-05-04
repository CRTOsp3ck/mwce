<!-- src/views/CampaignsView.vue -->

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useCampaignStore } from '@/stores/modules/campaign';
import { Campaign, PlayerCampaignProgress } from '@/types/campaign';

const router = useRouter();
const playerStore = usePlayerStore();
const campaignStore = useCampaignStore();

// State
const isLoading = ref(false);
const activeTab = ref('available');
const showConfirmModal = ref(false);
const selectedCampaign = ref<Campaign | null>(null);
const isStarting = ref(false);

// Computed properties
const availableCampaigns = computed(() => campaignStore.availableCampaigns);
const inProgressCampaigns = computed(() => campaignStore.inProgressCampaigns);
const completedCampaigns = computed(() => campaignStore.completedCampaigns);

// Functions
function handleStartCampaign(campaign: Campaign) {
  selectedCampaign.value = campaign;
  showConfirmModal.value = true;
}

function closeConfirmModal() {
  showConfirmModal.value = false;
  selectedCampaign.value = null;
}

async function confirmStartCampaign() {
  if (!selectedCampaign.value || isStarting.value) return;

  isStarting.value = true;

  try {
    const result = await campaignStore.startCampaign(selectedCampaign.value.id);

    if (result.success) {
      // Navigate to campaign detail
      router.push(`/campaigns/${selectedCampaign.value.id}`);
    } else {
      // Show error
      console.error('Failed to start campaign:', result.message);
    }
  } catch (error) {
    console.error('Error starting campaign:', error);
  } finally {
    isStarting.value = false;
    showConfirmModal.value = false;
  }
}

function handleContinueCampaign(campaign: Campaign) {
  router.push(`/campaigns/${campaign.id}`);
}

function switchTab(tab: string) {
  activeTab.value = tab;
}

// Load data on mount
onMounted(async () => {
  isLoading.value = true;

  try {
    await campaignStore.fetchCampaigns();
  } catch (error) {
    console.error('Error loading campaigns:', error);
  } finally {
    isLoading.value = false;
  }
});

// Helper functions
function getProgressPercent(campaign: Campaign): number {
  // Placeholder implementation - in a real app you'd need to calculate
  // based on missions completed vs total missions
  return Math.floor(Math.random() * 100);
}

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString();
}
</script>

<template>
  <div class="campaigns-view">
    <div class="page-title">
      <h2>Campaigns</h2>
      <p class="subtitle">Experience story-driven missions to build your criminal empire.</p>
    </div>

    <div class="campaigns-tabs">
      <button
        class="tab-button"
        :class="{ active: activeTab === 'available' }"
        @click="switchTab('available')"
      >
        Available
      </button>
      <button
        class="tab-button"
        :class="{ active: activeTab === 'in-progress' }"
        @click="switchTab('in-progress')"
      >
        In Progress
      </button>
      <button
        class="tab-button"
        :class="{ active: activeTab === 'completed' }"
        @click="switchTab('completed')"
      >
        Completed
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading campaigns...</p>
    </div>

    <!-- Available Campaigns Tab -->
    <div v-else-if="activeTab === 'available'" class="campaigns-grid">
      <BaseCard
        v-for="campaign in availableCampaigns"
        :key="campaign.id"
        class="campaign-card"
      >
        <div
          class="campaign-image"
          :style="{ backgroundImage: `url(${campaign.imageUrl || '/images/campaigns/default.jpg'})` }"
        >
          <div class="campaign-title">{{ campaign.title }}</div>
        </div>

        <div class="campaign-details">
          <p class="campaign-description">{{ campaign.description }}</p>
        </div>

        <template #footer>
          <div class="campaign-footer">
            <BaseButton @click="handleStartCampaign(campaign)">Start Campaign</BaseButton>
          </div>
        </template>
      </BaseCard>

      <div v-if="availableCampaigns.length === 0" class="empty-state">
        <div class="empty-icon">📜</div>
        <h4>No Campaigns Available</h4>
        <p>Check back later for new story campaigns.</p>
      </div>
    </div>

    <!-- In Progress Campaigns Tab -->
    <div v-else-if="activeTab === 'in-progress'" class="campaigns-grid">
      <BaseCard
        v-for="campaign in inProgressCampaigns"
        :key="campaign.id"
        class="campaign-card"
      >
        <div
          class="campaign-image"
          :style="{ backgroundImage: `url(${campaign.imageUrl || '/images/campaigns/default.jpg'})` }"
        >
          <div class="campaign-title">{{ campaign.title }}</div>
        </div>

        <div class="campaign-details">
          <p class="campaign-description">{{ campaign.description }}</p>

          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: `${getProgressPercent(campaign)}%` }"
            ></div>
          </div>
          <div class="progress-text">{{ getProgressPercent(campaign) }}% Complete</div>
        </div>

        <template #footer>
          <div class="campaign-footer">
            <BaseButton @click="handleContinueCampaign(campaign)">Continue</BaseButton>
          </div>
        </template>
      </BaseCard>

      <div v-if="inProgressCampaigns.length === 0" class="empty-state">
        <div class="empty-icon">🎮</div>
        <h4>No Active Campaigns</h4>
        <p>Start a new campaign to begin your journey.</p>
      </div>
    </div>

    <!-- Completed Campaigns Tab -->
    <div v-else-if="activeTab === 'completed'" class="campaigns-grid">
      <BaseCard
        v-for="campaign in completedCampaigns"
        :key="campaign.id"
        class="campaign-card completed"
      >
        <div
          class="campaign-image"
          :style="{ backgroundImage: `url(${campaign.imageUrl || '/images/campaigns/default.jpg'})` }"
        >
          <div class="campaign-title">{{ campaign.title }}</div>
          <div class="completion-badge">Completed</div>
        </div>

        <div class="campaign-details">
          <p class="campaign-description">{{ campaign.description }}</p>

          <div class="completion-date">
            Completed: {{ formatDate(campaignStore.campaignProgress[campaign.id]?.completedAt || '') }}
          </div>
        </div>

        <template #footer>
          <div class="campaign-footer">
            <BaseButton variant="secondary" @click="handleStartCampaign(campaign)">Play Again</BaseButton>
          </div>
        </template>
      </BaseCard>

      <div v-if="completedCampaigns.length === 0" class="empty-state">
        <div class="empty-icon">🏆</div>
        <h4>No Completed Campaigns</h4>
        <p>Complete a campaign to see it here.</p>
      </div>
    </div>

    <!-- Start Campaign Confirmation Modal -->
    <BaseModal
      v-model="showConfirmModal"
      :title="`Start Campaign: ${selectedCampaign?.title || 'Campaign'}`"
    >
      <div class="start-campaign-modal">
        <div
          class="campaign-preview"
          :style="{ backgroundImage: `url(${selectedCampaign?.imageUrl || '/images/campaigns/default.jpg'})` }"
        >
          <div class="preview-overlay">
            <h3>{{ selectedCampaign?.title }}</h3>
          </div>
        </div>

        <p class="campaign-description">{{ selectedCampaign?.description }}</p>

        <div class="requirements-check">
          <h4>Campaign Requirements</h4>

          <div class="requirement-item">
            <span class="requirement-label">Required Title:</span>
            <span class="requirement-value">
              {{ ['Associate', 'Soldier', 'Capo', 'Underboss', 'Consigliere', 'Boss', 'Godfather'][selectedCampaign?.requiredLevel - 1 || 0] || 'None' }}
            </span>
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
            :loading="isStarting"
            @click="confirmStartCampaign"
          >
            Start Campaign
          </BaseButton>
        </div>
      </template>
    </BaseModal>
  </div>
</template>

<style lang="scss">
.campaigns-view {
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

  .campaigns-tabs {
    display: flex;
    gap: 1px;
    margin-bottom: $spacing-lg;
    background-color: $border-color;
    border-radius: $border-radius-sm;
    overflow: hidden;

    .tab-button {
      flex: 1;
      background-color: $background-lighter;
      border: none;
      padding: $spacing-md;
      color: $text-secondary;
      font-weight: 500;
      cursor: pointer;
      transition: $transition-base;

      &.active {
        background-color: $primary-color;
        color: $text-color;
      }

      &:hover:not(.active) {
        background-color: lighten($background-lighter, 5%);
      }
    }
  }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: $spacing-xl;
    color: $text-secondary;

    .loading-spinner {
      width: 50px;
      height: 50px;
      border: 4px solid rgba($secondary-color, 0.3);
      border-radius: 50%;
      border-top-color: $secondary-color;
      animation: spin 1s linear infinite;
      margin-bottom: $spacing-md;
    }

    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
  }

  .campaigns-grid {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    gap: $spacing-md;

    @include respond-to(sm) {
      grid-template-columns: repeat(2, 1fr);
    }

    @include respond-to(md) {
      grid-template-columns: repeat(3, 1fr);
    }

    .campaign-card {
      overflow: hidden;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-5px);
        box-shadow: 0 10px 20px rgba(0, 0, 0, 0.3);
      }

      &.completed {
        border: 2px solid $gold-color;
      }

      .campaign-image {
        height: 180px;
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

        .campaign-title {
          position: absolute;
          bottom: $spacing-md;
          left: $spacing-md;
          color: white;
          font-size: $font-size-lg;
          font-weight: 600;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
        }

        .completion-badge {
          position: absolute;
          top: $spacing-md;
          right: $spacing-md;
          background-color: $gold-color;
          color: $background-darker;
          padding: 4px 10px;
          border-radius: $border-radius-sm;
          font-weight: 600;
          font-size: $font-size-sm;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
        }
      }

      .campaign-details {
        padding: $spacing-md;

        .campaign-description {
          color: $text-secondary;
          margin-bottom: $spacing-md;
          display: -webkit-box;
          -webkit-line-clamp: 3;
          -webkit-box-orient: vertical;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .progress-bar {
          height: 8px;
          background-color: rgba($background-lighter, 0.5);
          border-radius: 4px;
          overflow: hidden;
          margin-bottom: $spacing-xs;

          .progress-fill {
            height: 100%;
            background-color: $primary-color;
            border-radius: 4px;
          }
        }

        .progress-text {
          font-size: $font-size-sm;
          color: $text-secondary;
          text-align: right;
        }

        .completion-date {
          font-size: $font-size-sm;
          color: $gold-color;
          margin-top: $spacing-md;
        }
      }

      .campaign-footer {
        display: flex;
        justify-content: center;
      }
    }

    .empty-state {
      grid-column: 1 / -1;
      @include flex-column;
      align-items: center;
      justify-content: center;
      gap: $spacing-md;
      padding: $spacing-xl;
      text-align: center;
      color: $text-secondary;
      background-color: $background-card;
      border-radius: $border-radius-md;

      .empty-icon {
        font-size: 48px;
        margin-bottom: $spacing-sm;
      }

      h4 {
        margin: 0;
        color: $text-color;
      }
    }
  }

  .start-campaign-modal {
    @include flex-column;
    gap: $spacing-lg;

    .campaign-preview {
      height: 180px;
      background-size: cover;
      background-position: center;
      border-radius: $border-radius-md;
      position: relative;
      overflow: hidden;

      .preview-overlay {
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

    .campaign-description {
      color: $text-secondary;
      line-height: 1.6;
    }

    .requirements-check {
      background-color: rgba($background-lighter, 0.2);
      border-radius: $border-radius-md;
      padding: $spacing-md;

      h4 {
        margin: 0 0 $spacing-md 0;
        border-bottom: 1px solid $border-color;
        padding-bottom: $spacing-xs;
      }

      .requirement-item {
        display: flex;
        justify-content: space-between;
        margin-bottom: $spacing-xs;

        .requirement-label {
          color: $text-secondary;
        }

        .requirement-value {
          font-weight: 600;
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
