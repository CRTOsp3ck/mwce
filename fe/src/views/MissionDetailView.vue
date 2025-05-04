<!-- src/views/MissionDetailView.vue -->

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import MissionPOIList from '@/components/campaign/MissionPOIList.vue';
import MissionOperationList from '@/components/campaign/MissionOperationList.vue';
import ChoiceProgressTracker from '@/components/campaign/ChoiceProgressTracker.vue';
import { useCampaignStore } from '@/stores/modules/campaign';
import { usePlayerStore } from '@/stores/modules/player';
import { MissionChoice, MissionCompleteResult, MissionStatus } from '@/types/campaign';

const route = useRoute();
const router = useRouter();
const campaignStore = useCampaignStore();
const playerStore = usePlayerStore();

// State
const isLoading = ref(true);
const showChoiceModal = ref(false);
const showResultModal = ref(false);
const selectedChoice = ref<MissionChoice | null>(null);
const missionResult = ref<MissionCompleteResult | null>(null);
const isStarting = ref(false);
const isCompleting = ref(false);

// Computed properties
const mission = computed(() => campaignStore.currentMission);
const missionProgress = computed(() => {
  if (!mission.value) return null;
  return campaignStore.missionProgress[mission.value.id];
});

const missionStatus = computed(() => {
  return missionProgress.value?.status || MissionStatus.NOT_STARTED;
});

const canStartMission = computed(() => {
  return missionStatus.value === MissionStatus.NOT_STARTED;
});

const canCompleteMission = computed(() => {
  return missionStatus.value === MissionStatus.IN_PROGRESS;
});

const hasChoices = computed(() => {
  return mission.value?.choices && mission.value.choices.length > 0;
});

const missionComplete = computed(() => {
  return missionStatus.value === MissionStatus.COMPLETED;
});

const activePOIs = computed(() => {
  if (!mission.value) return [];
  return campaignStore.getPOIsByMission(mission.value.id);
});

const activeMissionOperations = computed(() => {
  if (!mission.value) return [];
  return campaignStore.getOperationsByMission(mission.value.id);
});

const activeChoice = computed(() => {
  if (!missionProgress.value || !missionProgress.value.currentActiveChoice) return null;

  // Find the choice from the mission
  if (!mission.value || !mission.value.choices) return null;

  return mission.value.choices.find(c => c.id === missionProgress.value.currentActiveChoice);
});

const activeConditions = computed(() => {
  if (!activeChoice.value) return [];
  return campaignStore.choiceProgress[activeChoice.value.id] || [];
});

// Player resources for requirement checks
const playerMoney = computed(() => playerStore.playerMoney);
const playerCrew = computed(() => playerStore.playerCrew);
const playerWeapons = computed(() => playerStore.playerWeapons);
const playerVehicles = computed(() => playerStore.playerVehicles);
const playerRespect = computed(() => playerStore.playerRespect);
const playerInfluence = computed(() => playerStore.playerInfluence);
const playerHeat = computed(() => playerStore.playerHeat);

// Methods
async function handleStartMission() {
  if (isStarting.value || !canStartMission.value) return;

  isStarting.value = true;

  try {
    const result = await campaignStore.startMission(mission.value!.id);

    if (!result.success) {
      console.error('Failed to start mission:', result.message);
    }
  } catch (error) {
    console.error('Error starting mission:', error);
  } finally {
    isStarting.value = false;
  }
}

function handleCompleteMission() {
  if (!canCompleteMission.value) return;

  if (hasChoices.value) {
    // Show choice modal if the mission has choices
    showChoiceModal.value = true;
  } else {
    // Complete mission directly if no choices
    completeMission();
  }
}

function handleChoiceSelection(choice: MissionChoice) {
  selectedChoice.value = choice;
}

function closeChoiceModal() {
  showChoiceModal.value = false;
  selectedChoice.value = null;
}

function closeResultModal() {
  showResultModal.value = false;

  // If there's a next mission, go to it
  if (missionResult.value?.nextMission) {
    router.push(`/campaigns/missions/${missionResult.value.nextMission.id}`);
  } else {
    // Otherwise, go back to campaign detail
    if (mission.value) {
      const chapter = campaignStore.currentChapter;
      if (chapter) {
        router.push(`/campaigns/${chapter.campaignId}`);
      } else {
        router.push('/campaigns');
      }
    } else {
      router.push('/campaigns');
    }
  }
}

async function confirmChoice() {
  showChoiceModal.value = false;
  await completeMission(selectedChoice.value?.id);
  selectedChoice.value = null;
}

async function completeMission(choiceId?: string) {
  if (isCompleting.value || !canCompleteMission.value) return;

  isCompleting.value = true;

  try {
    const result = await campaignStore.completeMission(mission.value!.id, choiceId);

    if (result.success) {
      missionResult.value = result.result;
      showResultModal.value = true;
    } else {
      console.error('Failed to complete mission:', result.message);
    }
  } catch (error) {
    console.error('Error completing mission:', error);
  } finally {
    isCompleting.value = false;
  }
}

async function handleCompletePOI(poiId: string) {
  await campaignStore.completePOI(poiId);
}

async function handleStartOperation(operationId: string) {
  await campaignStore.startMissionOperation(operationId);
}

async function handleCompleteOperation(operationId: string) {
  await campaignStore.completeMissionOperation(operationId);
}

function checkRequirement(value: number | undefined, playerValue: number): string {
  if (!value) return 'met';
  return playerValue >= value ? 'met' : 'not-met';
}

function goBackToCampaign() {
  const chapter = campaignStore.currentChapter;
  if (chapter) {
    router.push(`/campaigns/${chapter.campaignId}`);
  } else {
    router.push('/campaigns');
  }
}

// Format numbers with k/M for larger values
function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M';
  } else if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K';
  }
  return value.toString();
}

function hasRewards(rewards: any): boolean {
  return !!(
    rewards.money ||
    rewards.crew ||
    rewards.weapons ||
    rewards.vehicles ||
    rewards.respect ||
    rewards.influence ||
    rewards.heatReduction
  );
}

// On component mount
onMounted(async () => {
  isLoading.value = true;

  try {
    const missionId = route.params.id as string;
    await campaignStore.fetchMission(missionId);

    // Also fetch active POIs and operations
    await campaignStore.fetchActivePOIs();
    await campaignStore.fetchActiveMissionOperations();
  } catch (error) {
    console.error('Error loading mission:', error);
  } finally {
    isLoading.value = false;
  }
});

</script>

<template>
  <div class="mission-detail-view">
    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading mission...</p>
    </div>

    <!-- Mission Content -->
    <div v-else-if="mission" class="mission-content">
      <div class="back-button">
        <BaseButton variant="text" @click="goBackToCampaign">
          ← Back to Campaign
        </BaseButton>
      </div>

      <!-- Mission Header -->
      <div
        class="mission-header"
        :style="{ backgroundImage: `url(${mission.imageUrl || '/images/missions/default.jpg'})` }"
      >
        <div class="header-overlay">
          <div class="header-content">
            <div class="mission-badge">{{ mission.missionType.toUpperCase() }}</div>
            <h2 class="mission-title">{{ mission.title }}</h2>
            <div
              class="mission-status"
              :class="{
                'completed': missionStatus === MissionStatus.COMPLETED,
                'in-progress': missionStatus === MissionStatus.IN_PROGRESS,
                'not-started': missionStatus === MissionStatus.NOT_STARTED,
                'failed': missionStatus === MissionStatus.FAILED
              }"
            >
              {{ missionStatus.replace('_', ' ').toUpperCase() }}
            </div>
          </div>
        </div>
      </div>

      <!-- Mission Details -->
      <div class="mission-details-section">
        <div class="mission-description">
          {{ mission.description }}
        </div>

        <div class="mission-narrative">
          <div class="narrative-content">
            {{ mission.narrative }}
          </div>
        </div>
      </div>

      <!-- Active Choice Progress -->
      <div v-if="activeChoice" class="active-choice-section">
        <h3>Current Progress</h3>
        <ChoiceProgressTracker
          :choice="activeChoice"
          :conditions="activeConditions"
        />
      </div>

      <!-- Points of Interest -->
      <MissionPOIList
        v-if="activePOIs.length > 0"
        :pois="activePOIs"
        @complete="handleCompletePOI"
      />

      <!-- Mission Operations -->
      <MissionOperationList
        v-if="activeMissionOperations.length > 0"
        :operations="activeMissionOperations"
        @start="handleStartOperation"
        @complete="handleCompleteOperation"
      />

      <!-- Mission Requirements & Rewards -->
      <div class="mission-data-grid">
        <div class="requirements-section">
          <h3>Requirements</h3>

          <div class="requirements-list">
            <div
              v-if="mission.requirements.money"
              class="requirement-item"
              :class="checkRequirement(mission.requirements.money, playerMoney)"
            >
              <div class="req-icon">💰</div>
              <div class="req-details">
                <div class="req-name">Money</div>
                <div class="req-value">
                  ${{ formatNumber(mission.requirements.money) }}
                  <span class="player-value">(You: ${{ formatNumber(playerMoney) }})</span>
                </div>
              </div>
            </div>

            <div
              v-if="mission.requirements.crew"
              class="requirement-item"
              :class="checkRequirement(mission.requirements.crew, playerCrew)"
            >
              <div class="req-icon">👥</div>
              <div class="req-details">
                <div class="req-name">Crew</div>
                <div class="req-value">
                  {{ mission.requirements.crew }}
                  <span class="player-value">(You: {{ playerCrew }})</span>
                </div>
              </div>
            </div>

            <div
              v-if="mission.requirements.weapons"
              class="requirement-item"
              :class="checkRequirement(mission.requirements.weapons, playerWeapons)"
            >
              <div class="req-icon">🔫</div>
              <div class="req-details">
                <div class="req-name">Weapons</div>
                <div class="req-value">
                  {{ mission.requirements.weapons }}
                  <span class="player-value">(You: {{ playerWeapons }})</span>
                </div>
              </div>
            </div>

            <div
              v-if="mission.requirements.vehicles"
              class="requirement-item"
              :class="checkRequirement(mission.requirements.vehicles, playerVehicles)"
            >
              <div class="req-icon">🚗</div>
              <div class="req-details">
                <div class="req-name">Vehicles</div>
                <div class="req-value">
                  {{ mission.requirements.vehicles }}
                  <span class="player-value">(You: {{ playerVehicles }})</span>
                </div>
              </div>
            </div>

            <div
              v-if="mission.requirements.respect"
              class="requirement-item"
              :class="checkRequirement(mission.requirements.respect, playerRespect)"
            >
              <div class="req-icon">👊</div>
              <div class="req-details">
                <div class="req-name">Respect</div>
                <div class="req-value">
                  {{ mission.requirements.respect }}
                  <span class="player-value">(You: {{ playerRespect }})</span>
                </div>
              </div>
            </div>

            <div
              v-if="mission.requirements.influence"
              class="requirement-item"
              :class="checkRequirement(mission.requirements.influence, playerInfluence)"
            >
              <div class="req-icon">🏛️</div>
              <div class="req-details">
                <div class="req-name">Influence</div>
                <div class="req-value">
                  {{ mission.requirements.influence }}
                  <span class="player-value">(You: {{ playerInfluence }})</span>
                </div>
              </div>
            </div>

            <div
              v-if="mission.requirements.maxHeat"
              class="requirement-item"
              :class="playerHeat <= (mission.requirements.maxHeat || 0) ? 'met' : 'not-met'"
            >
              <div class="req-icon">🔥</div>
              <div class="req-details">
                <div class="req-name">Max Heat</div>
                <div class="req-value">
                  {{ mission.requirements.maxHeat }}
                  <span class="player-value">(You: {{ playerHeat }})</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="rewards-section">
          <h3>Rewards</h3>

          <div class="rewards-list">
            <div v-if="mission.rewards.money" class="reward-item">
              <div class="reward-icon">💰</div>
              <div class="reward-details">
                <div class="reward-name">Money</div>
                <div class="reward-value">${{ formatNumber(mission.rewards.money) }}</div>
              </div>
            </div>

            <div v-if="mission.rewards.crew" class="reward-item">
              <div class="reward-icon">👥</div>
              <div class="reward-details">
                <div class="reward-name">Crew</div>
                <div class="reward-value">{{ mission.rewards.crew }}</div>
              </div>
            </div>

            <div v-if="mission.rewards.weapons" class="reward-item">
              <div class="reward-icon">🔫</div>
              <div class="reward-details">
                <div class="reward-name">Weapons</div>
                <div class="reward-value">{{ mission.rewards.weapons }}</div>
              </div>
            </div>

            <div v-if="mission.rewards.vehicles" class="reward-item">
              <div class="reward-icon">🚗</div>
              <div class="reward-details">
                <div class="reward-name">Vehicles</div>
                <div class="reward-value">{{ mission.rewards.vehicles }}</div>
              </div>
            </div>

            <div v-if="mission.rewards.respect" class="reward-item">
              <div class="reward-icon">👊</div>
              <div class="reward-details">
                <div class="reward-name">Respect</div>
                <div class="reward-value">{{ mission.rewards.respect }}</div>
              </div>
            </div>

            <div v-if="mission.rewards.influence" class="reward-item">
              <div class="reward-icon">🏛️</div>
              <div class="reward-details">
                <div class="reward-name">Influence</div>
                <div class="reward-value">{{ mission.rewards.influence }}</div>
              </div>
            </div>

            <div v-if="mission.rewards.heatReduction" class="reward-item">
              <div class="reward-icon">❄️</div>
              <div class="reward-details">
                <div class="reward-name">Heat Reduction</div>
                <div class="reward-value">{{ mission.rewards.heatReduction }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Mission Actions -->
      <div class="mission-actions">
        <BaseButton
          v-if="canStartMission"
          @click="handleStartMission"
          :loading="isStarting"
        >
          Start Mission
        </BaseButton>

        <BaseButton
          v-else-if="canCompleteMission"
          @click="handleCompleteMission"
          :loading="isCompleting"
          variant="primary"
        >
          Complete Mission
        </BaseButton>

        <BaseButton
          v-else-if="missionComplete"
          @click="goBackToCampaign"
          variant="secondary"
        >
          Back to Campaign
        </BaseButton>
      </div>
    </div>

    <!-- Error State -->
    <div v-else class="error-state">
      <div class="error-icon">❌</div>
      <h3>Failed to Load Mission</h3>
      <p>There was an error loading the mission. Please try again later.</p>
      <BaseButton @click="router.push('/campaigns')">Back to Campaigns</BaseButton>
    </div>

    <!-- Choice Modal -->
    <BaseModal
      v-model="showChoiceModal"
      title="Choose Your Path"
      :close-on-click-outside="false"
    >
      <div class="choices-modal-content">
        <p class="choices-intro">Your decision will affect how the story unfolds. Choose wisely.</p>

        <div class="choices-list">
          <div
            v-for="choice in mission?.choices"
            :key="choice.id"
            class="choice-item"
            :class="{ selected: selectedChoice?.id === choice.id }"
            @click="handleChoiceSelection(choice)"
          >
            <div class="choice-text">{{ choice.text }}</div>

            <div class="choice-rewards" v-if="hasRewards(choice.rewards)">
              <h4>Additional Rewards</h4>

              <div class="rewards-mini-list">
                <div v-if="choice.rewards.money" class="mini-reward">
                  <span class="mini-icon">💰</span> {{ formatNumber(choice.rewards.money) }}
                </div>
                <div v-if="choice.rewards.crew" class="mini-reward">
                  <span class="mini-icon">👥</span> {{ choice.rewards.crew }}
                </div>
                <div v-if="choice.rewards.weapons" class="mini-reward">
                  <span class="mini-icon">🔫</span> {{ choice.rewards.weapons }}
                </div>
                <div v-if="choice.rewards.vehicles" class="mini-reward">
                  <span class="mini-icon">🚗</span> {{ choice.rewards.vehicles }}
                </div>
                <div v-if="choice.rewards.respect" class="mini-reward">
                  <span class="mini-icon">👊</span> {{ choice.rewards.respect }}
                </div>
                <div v-if="choice.rewards.influence" class="mini-reward">
                  <span class="mini-icon">🏛️</span> {{ choice.rewards.influence }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton
            variant="text"
            @click="closeChoiceModal"
          >
            Cancel
          </BaseButton>
          <BaseButton
            :disabled="!selectedChoice"
            @click="confirmChoice"
          >
            Confirm Choice
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- Result Modal -->
    <BaseModal
      v-model="showResultModal"
      :title="missionResult?.success ? 'Mission Successful' : 'Mission Failed'"
      :close-on-click-outside="false"
    >
    <div class="result-modal-content">
        <div
          class="result-icon"
          :class="{ 'success': missionResult?.success, 'failure': !missionResult?.success }"
        >
          {{ missionResult?.success ? '✅' : '❌' }}
        </div>

        <div class="result-message">{{ missionResult?.message }}</div>

        <div class="result-rewards" v-if="missionResult?.rewards">
          <h4>Rewards Received</h4>

          <div class="rewards-grid">
            <div v-if="missionResult.rewards.money" class="reward-item">
              <div class="reward-icon">💰</div>
              <div class="reward-details">
                <div class="reward-value">+${{ formatNumber(missionResult.rewards.money) }}</div>
              </div>
            </div>

            <div v-if="missionResult.rewards.crew" class="reward-item">
              <div class="reward-icon">👥</div>
              <div class="reward-details">
                <div class="reward-value">+{{ missionResult.rewards.crew }} Crew</div>
              </div>
            </div>

            <div v-if="missionResult.rewards.weapons" class="reward-item">
              <div class="reward-icon">🔫</div>
              <div class="reward-details">
                <div class="reward-value">+{{ missionResult.rewards.weapons }} Weapons</div>
              </div>
            </div>

            <div v-if="missionResult.rewards.vehicles" class="reward-item">
              <div class="reward-icon">🚗</div>
              <div class="reward-details">
                <div class="reward-value">+{{ missionResult.rewards.vehicles }} Vehicles</div>
              </div>
            </div>

            <div v-if="missionResult.rewards.respect" class="reward-item">
              <div class="reward-icon">👊</div>
              <div class="reward-details">
                <div class="reward-value">+{{ missionResult.rewards.respect }} Respect</div>
              </div>
            </div>

            <div v-if="missionResult.rewards.influence" class="reward-item">
              <div class="reward-icon">🏛️</div>
              <div class="reward-details">
                <div class="reward-value">+{{ missionResult.rewards.influence }} Influence</div>
              </div>
            </div>

            <div v-if="missionResult.rewards.heatReduction" class="reward-item">
              <div class="reward-icon">❄️</div>
              <div class="reward-details">
                <div class="reward-value">-{{ missionResult.rewards.heatReduction }} Heat</div>
              </div>
            </div>
          </div>
        </div>

        <div class="next-mission" v-if="missionResult?.nextMission">
          <h4>Next Mission</h4>
          <div class="next-mission-card">
            <div class="next-mission-title">{{ missionResult.nextMission.title }}</div>
            <div class="next-mission-desc">{{ missionResult.nextMission.description }}</div>
          </div>
        </div>
      </div>

      <template #footer>
        <BaseButton @click="closeResultModal">
          Continue
        </BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<style lang="scss">
.mission-detail-view {
  .loading-state, .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: $spacing-xl;
    color: $text-secondary;
    text-align: center;

    .loading-spinner {
      width: 50px;
      height: 50px;
      border: 4px solid rgba($secondary-color, 0.3);
      border-radius: 50%;
      border-top-color: $secondary-color;
      animation: spin 1s linear infinite;
      margin-bottom: $spacing-md;
    }

    .error-icon {
      font-size: 48px;
      margin-bottom: $spacing-md;
      color: $danger-color;
    }

    h3 {
      color: $text-color;
      margin-bottom: $spacing-sm;
    }

    p {
      margin-bottom: $spacing-lg;
    }

    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
  }

  .mission-content {
    .back-button {
      margin-bottom: $spacing-md;
    }

    .mission-header {
      height: 250px;
      background-size: cover;
      background-position: center;
      border-radius: $border-radius-md;
      margin-bottom: $spacing-xl;
      position: relative;
      overflow: hidden;

      .header-overlay {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: linear-gradient(to top, rgba(0, 0, 0, 0.9), rgba(0, 0, 0, 0.4));
        display: flex;
        flex-direction: column;
        justify-content: flex-end;
        padding: $spacing-xl;
      }

      .header-content {
        .mission-badge {
          display: inline-block;
          background-color: $primary-color;
          color: white;
          padding: 4px 10px;
          border-radius: $border-radius-sm;
          font-size: $font-size-xs;
          font-weight: 600;
          margin-bottom: $spacing-sm;
        }

        .mission-title {
          @include gold-accent;
          font-size: $font-size-xl;
          margin-bottom: $spacing-sm;
        }

        .mission-status {
          display: inline-block;
          padding: 4px 12px;
          border-radius: $border-radius-sm;
          font-size: $font-size-sm;
          font-weight: 600;

          &.completed {
            background-color: $success-color;
            color: white;
          }

          &.in-progress {
            background-color: $info-color;
            color: white;
          }

          &.not-started {
            background-color: $text-secondary;
            color: $background-darker;
          }

          &.failed {
            background-color: $danger-color;
            color: white;
          }
        }
      }
    }

    .mission-details-section {
      margin-bottom: $spacing-xl;

      .mission-description {
        font-size: $font-size-lg;
        margin-bottom: $spacing-md;
        color: $text-color;
        line-height: 1.5;
      }

      .mission-narrative {
        background-color: rgba($background-lighter, 0.2);
        border-radius: $border-radius-md;
        padding: $spacing-lg;
        border-left: 4px solid $secondary-color;

        .narrative-content {
          font-style: italic;
          color: $text-secondary;
          line-height: 1.6;
          white-space: pre-line; // Preserve line breaks
        }
      }
    }

    .mission-data-grid {
      display: grid;
      grid-template-columns: 1fr;
      gap: $spacing-xl;
      margin-bottom: $spacing-xl;

      @include respond-to(md) {
        grid-template-columns: 1fr 1fr;
      }

      .requirements-section, .rewards-section {
        h3 {
          margin-bottom: $spacing-md;
          position: relative;

          &:after {
            content: '';
            position: absolute;
            bottom: -8px;
            left: 0;
            width: 40px;
            height: 2px;
            background-color: $primary-color;
          }
        }
      }

      .requirements-list, .rewards-list {
        display: flex;
        flex-direction: column;
        gap: $spacing-sm;

        .requirement-item, .reward-item {
          display: flex;
          gap: $spacing-md;
          padding: $spacing-sm;
          border-radius: $border-radius-sm;
          background-color: rgba($background-lighter, 0.2);

          &.met {
            border-left: 3px solid $success-color;
          }

          &.not-met {
            border-left: 3px solid $danger-color;
          }

          .req-icon, .reward-icon {
            font-size: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
          }

          .req-details, .reward-details {
            flex: 1;

            .req-name, .reward-name {
              font-size: $font-size-sm;
              color: $text-secondary;
            }

            .req-value, .reward-value {
              font-weight: 600;

              .player-value {
                font-weight: normal;
                color: $text-secondary;
                font-size: $font-size-sm;
                margin-left: $spacing-xs;
              }
            }
          }
        }
      }
    }

    .mission-actions {
      display: flex;
      justify-content: center;
      margin-bottom: $spacing-xl;
    }
  }

  .choices-modal-content {
    .choices-intro {
      text-align: center;
      margin-bottom: $spacing-lg;
      color: $text-secondary;
    }

    .choices-list {
      display: flex;
      flex-direction: column;
      gap: $spacing-md;

      .choice-item {
        background-color: $background-card;
        border-radius: $border-radius-md;
        padding: $spacing-md;
        cursor: pointer;
        transition: all 0.3s ease;
        border: 1px solid $border-color;

        &:hover {
          transform: translateY(-3px);
          box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
        }

        &.selected {
          border: 2px solid $secondary-color;
          box-shadow: 0 0 10px rgba($secondary-color, 0.3);
        }

        .choice-text {
          font-weight: 600;
          margin-bottom: $spacing-md;
          line-height: 1.5;
        }

        .choice-rewards {
          h4 {
            font-size: $font-size-sm;
            margin-bottom: $spacing-xs;
            color: $text-secondary;
          }

          .rewards-mini-list {
            display: flex;
            flex-wrap: wrap;
            gap: $spacing-sm;

            .mini-reward {
              font-size: $font-size-xs;
              background-color: rgba($success-color, 0.1);
              color: $success-color;
              padding: 2px 8px;
              border-radius: 20px;
              display: flex;
              align-items: center;
              gap: 4px;

              .mini-icon {
                font-size: 12px;
              }
            }
          }
        }
      }
    }
  }

  .result-modal-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: $spacing-lg;

    .result-icon {
      font-size: 48px;
      margin-bottom: $spacing-sm;

      &.success {
        color: $success-color;
      }

      &.failure {
        color: $danger-color;
      }
    }

    .result-message {
      font-size: $font-size-lg;
      color: $text-color;
      line-height: 1.5;
    }

    .result-rewards {
      width: 100%;
      background-color: rgba($success-color, 0.1);
      border-radius: $border-radius-md;
      padding: $spacing-md;

      h4 {
        text-align: center;
        margin-bottom: $spacing-md;
        color: $success-color;
      }

      .rewards-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: $spacing-sm;

        @include respond-to(md) {
          grid-template-columns: repeat(3, 1fr);
        }

        .reward-item {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          background-color: rgba($background-darker, 0.2);
          padding: $spacing-sm;
          border-radius: $border-radius-sm;

          .reward-icon {
            font-size: 20px;
          }

          .reward-details {
            .reward-value {
              font-weight: 600;
              color: $success-color;
            }
          }
        }
      }
    }

    .next-mission {
      width: 100%;

      h4 {
        text-align: center;
        margin-bottom: $spacing-md;
      }

      .next-mission-card {
        background-color: $background-card;
        border-radius: $border-radius-md;
        padding: $spacing-md;
        border-left: 4px solid $primary-color;

        .next-mission-title {
          font-weight: 600;
          margin-bottom: $spacing-xs;
        }

        .next-mission-desc {
          font-size: $font-size-sm;
          color: $text-secondary;
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
