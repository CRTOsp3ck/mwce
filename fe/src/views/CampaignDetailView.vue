<!-- src/views/CampaignDetailView.vue -->

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import { useCampaignStore } from '@/stores/modules/campaign';
import { usePlayerStore } from '@/stores/modules/player';
import { Chapter, Mission, MissionStatus } from '@/types/campaign';

const route = useRoute();
const router = useRouter();
const campaignStore = useCampaignStore();
const playerStore = usePlayerStore();

const isLoading = ref(true);
const selectedChapter = ref<Chapter | null>(null);

// Computed properties
const campaign = computed(() => campaignStore.currentCampaign);
const chapters = computed(() => campaign.value?.chapters || []);
const currentChapter = computed(() => campaignStore.currentChapter);
const missions = computed(() => currentChapter.value?.missions || []);
const campaignProgress = computed(() => {
  if (!campaign.value) return null;
  return campaignStore.campaignProgress[campaign.value.id];
});

// Methods
function selectChapter(chapter: Chapter) {
  if (chapter.isLocked) return;
  selectedChapter.value = chapter;
  campaignStore.fetchChapter(chapter.id);
}

function handleMissionClick(mission: Mission) {
  if (mission.isLocked) return;
  router.push(`/campaigns/missions/${mission.id}`);
}

function getMissionStatus(mission: Mission) {
  return campaignStore.getMissionStatus(mission.id);
}

function formatMissionStatus(status: MissionStatus) {
  switch (status) {
    case MissionStatus.COMPLETED:
      return 'Completed';
    case MissionStatus.IN_PROGRESS:
      return 'In Progress';
    case MissionStatus.NOT_STARTED:
      return 'Not Started';
    case MissionStatus.FAILED:
      return 'Failed';
    default:
      return 'Unknown';
  }
}

function getMissionStatusClass(status: MissionStatus) {
  switch (status) {
    case MissionStatus.COMPLETED:
      return 'status-completed';
    case MissionStatus.IN_PROGRESS:
      return 'status-progress';
    case MissionStatus.NOT_STARTED:
      return 'status-not-started';
    case MissionStatus.FAILED:
      return 'status-failed';
    default:
      return '';
  }
}

function getChapterProgress(chapter: Chapter): number {
  if (!chapter.missions || chapter.missions.length === 0) {
    return 0;
  }

  const completedCount = chapter.missions.filter(
    m => campaignStore.getMissionStatus(m.id) === MissionStatus.COMPLETED
  ).length;

  return Math.floor((completedCount / chapter.missions.length) * 100);
}

function isCurrentMission(mission: Mission): boolean {
  return !!campaignProgress.value && campaignProgress.value.currentMissionId === mission.id;
}

// On component mount
onMounted(async () => {
  isLoading.value = true;

  try {
    const campaignId = route.params.id as string;
    await campaignStore.fetchCampaignDetail(campaignId);

    if (currentChapter.value) {
      selectedChapter.value = currentChapter.value;
    }
  } catch (error) {
    console.error('Error loading campaign:', error);
  } finally {
    isLoading.value = false;
  }
});
</script>

<template>
  <div class="campaign-detail-view">
    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>Loading campaign...</p>
    </div>

    <!-- Campaign Content -->
    <div v-else-if="campaign" class="campaign-content">
      <!-- Campaign Header -->
      <div
        class="campaign-header"
        :style="{ backgroundImage: `url(${campaign.imageUrl || '/images/campaigns/default.jpg'})` }"
      >
        <div class="header-overlay">
          <div class="header-content">
            <h2 class="campaign-title">{{ campaign.title }}</h2>
            <p class="campaign-description">{{ campaign.description }}</p>
          </div>
        </div>
      </div>

      <!-- Chapter Selection -->
      <div class="chapters-selector">
        <h3 class="section-title">Chapters</h3>

        <div class="chapters-list">
          <div
            v-for="chapter in chapters"
            :key="chapter.id"
            class="chapter-item"
            :class="{
              'selected': selectedChapter?.id === chapter.id,
              'locked': chapter.isLocked,
              'completed': getChapterProgress(chapter) === 100
            }"
            @click="selectChapter(chapter)"
          >
            <div class="chapter-name">{{ chapter.title }}</div>
            <div class="chapter-progress">
              <div
                class="progress-bar"
                :style="{ width: `${getChapterProgress(chapter)}%` }"
              ></div>
            </div>
            <div class="lock-icon" v-if="chapter.isLocked">🔒</div>
            <div class="complete-icon" v-else-if="getChapterProgress(chapter) === 100">✅</div>
            <div class="progress-percent" v-else>{{ getChapterProgress(chapter) }}%</div>
          </div>
        </div>
      </div>

      <!-- Missions List -->
      <div class="missions-section" v-if="currentChapter">
        <h3 class="section-title">{{ currentChapter.title }} Missions</h3>
        <p class="chapter-description">{{ currentChapter.description }}</p>

        <div class="missions-list">
          <BaseCard
            v-for="mission in missions"
            :key="mission.id"
            class="mission-card"
            :class="{
              'locked': mission.isLocked,
              'current': isCurrentMission(mission),
              'completed': getMissionStatus(mission) === MissionStatus.COMPLETED
            }"
            @click="handleMissionClick(mission)"
          >
            <div class="mission-header">
              <h4 class="mission-title">{{ mission.title }}</h4>
              <div
                class="mission-status"
                :class="getMissionStatusClass(getMissionStatus(mission))"
              >
                {{ formatMissionStatus(getMissionStatus(mission)) }}
              </div>
            </div>

            <div class="mission-details">
              <p class="mission-description">{{ mission.description }}</p>

              <div class="mission-type">
                <span class="type-label">Type:</span>
                <span class="type-value">{{ mission.missionType.charAt(0).toUpperCase() + mission.missionType.slice(1) }}</span>
              </div>
            </div>

            <div class="mission-requirements" v-if="mission.requirements">
              <h5>Requirements</h5>
              <div class="requirements-list">
                <div class="requirement-item" v-if="mission.requirements.money">
                  <span class="req-icon">💰</span>
                  <span class="req-text">${{ mission.requirements.money }}</span>
                </div>
                <div class="requirement-item" v-if="mission.requirements.crew">
                  <span class="req-icon">👥</span>
                  <span class="req-text">{{ mission.requirements.crew }} Crew</span>
                </div>
                <div class="requirement-item" v-if="mission.requirements.weapons">
                  <span class="req-icon">🔫</span>
                  <span class="req-text">{{ mission.requirements.weapons }} Weapons</span>
                </div>
                <div class="requirement-item" v-if="mission.requirements.vehicles">
                  <span class="req-icon">🚗</span>
                  <span class="req-text">{{ mission.requirements.vehicles }} Vehicles</span>
                </div>
              </div>
            </div>

            <div class="lock-overlay" v-if="mission.isLocked">
              <div class="lock-icon">🔒</div>
              <div class="lock-text">Complete previous missions to unlock</div>
            </div>

            <div class="current-mission-badge" v-if="isCurrentMission(mission)">
              Current Mission
            </div>
          </BaseCard>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else class="error-state">
      <div class="error-icon">❌</div>
      <h3>Failed to Load Campaign</h3>
      <p>There was an error loading the campaign. Please try again later.</p>
      <BaseButton @click="router.push('/campaigns')">Back to Campaigns</BaseButton>
    </div>
  </div>
</template>

<style lang="scss">
.campaign-detail-view {
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

  .campaign-content {
    .campaign-header {
      height: 300px;
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
        .campaign-title {
          @include gold-accent;
          font-size: $font-size-xxl;
          margin-bottom: $spacing-sm;
        }

        .campaign-description {
          color: $text-secondary;
          max-width: 800px;
          line-height: 1.6;
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
        width: 60px;
        height: 3px;
        background-color: $primary-color;
      }
    }

    .chapters-selector {
      margin-bottom: $spacing-xl;

      .chapters-list {
        display: flex;
        gap: $spacing-md;
        overflow-x: auto;
        padding-bottom: $spacing-md;
        -webkit-overflow-scrolling: touch;

        /* Hide scrollbar for Chrome, Safari and Opera */
        &::-webkit-scrollbar {
          display: none;
        }

        /* IE and Edge */
        -ms-overflow-style: none;
        /* Firefox */
        scrollbar-width: none;

        .chapter-item {
          min-width: 200px;
          background-color: $background-card;
          border-radius: $border-radius-md;
          padding: $spacing-md;
          cursor: pointer;
          transition: all 0.3s ease;
          position: relative;
          border: 1px solid $border-color;

          &:hover:not(.locked) {
            transform: translateY(-5px);
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
          }

          &.selected {
            border: 2px solid $secondary-color;
            box-shadow: 0 0 10px rgba($secondary-color, 0.3);
          }

          &.locked {
            opacity: 0.7;
            cursor: not-allowed;
          }

          &.completed {
            border-color: $success-color;
          }

          .chapter-name {
            font-weight: 600;
            margin-bottom: $spacing-sm;
            padding-right: 24px; /* Space for icons */
          }

          .chapter-progress {
            height: 6px;
            background-color: rgba($background-lighter, 0.5);
            border-radius: 3px;
            overflow: hidden;
            margin-top: $spacing-sm;

            .progress-bar {
              height: 100%;
              background-color: $success-color;
              border-radius: 3px;
            }
          }

          .lock-icon, .complete-icon, .progress-percent {
            position: absolute;
            top: $spacing-md;
            right: $spacing-md;
            font-size: $font-size-md;
          }

          .progress-percent {
            font-size: $font-size-sm;
            color: $text-secondary;
          }
        }
      }
    }

    .missions-section {
      .chapter-description {
        color: $text-secondary;
        margin-bottom: $spacing-lg;
      }

      .missions-list {
        display: grid;
        grid-template-columns: repeat(1, 1fr);
        gap: $spacing-md;

        @include respond-to(md) {
          grid-template-columns: repeat(2, 1fr);
        }

        .mission-card {
          position: relative;
          cursor: pointer;
          transition: all 0.3s ease;
          overflow: hidden;

          &:hover:not(.locked) {
            transform: translateY(-5px);
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
          }

          &.current {
            border: 2px solid $primary-color;
          }

          &.completed {
            border-left: 4px solid $success-color;
          }

          &.locked {
            opacity: 0.7;
            cursor: not-allowed;
          }

          .mission-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: $spacing-md;

            .mission-title {
              margin: 0;
              font-size: $font-size-lg;
            }

            .mission-status {
              font-size: $font-size-sm;
              padding: 4px 8px;
              border-radius: $border-radius-sm;
              font-weight: 600;

              &.status-completed {
                background-color: rgba($success-color, 0.2);
                color: $success-color;
              }

              &.status-progress {
                background-color: rgba($info-color, 0.2);
                color: $info-color;
              }

              &.status-not-started {
                background-color: rgba($text-secondary, 0.2);
                color: $text-secondary;
              }

              &.status-failed {
                background-color: rgba($danger-color, 0.2);
                color: $danger-color;
              }
            }
          }

          .mission-details {
            margin-bottom: $spacing-md;

            .mission-description {
              color: $text-secondary;
              margin-bottom: $spacing-md;
              display: -webkit-box;
              -webkit-line-clamp: 3;
              -webkit-box-orient: vertical;
              overflow: hidden;
              text-overflow: ellipsis;
            }

            .mission-type {
              font-size: $font-size-sm;

              .type-label {
                color: $text-secondary;
                margin-right: $spacing-xs;
              }

              .type-value {
                font-weight: 600;
              }
            }
          }

          .mission-requirements {
            h5 {
              margin: 0 0 $spacing-xs 0;
              font-size: $font-size-md;
            }

            .requirements-list {
              display: flex;
              flex-wrap: wrap;
              gap: $spacing-sm;

              .requirement-item {
                background-color: rgba($background-lighter, 0.3);
                border-radius: $border-radius-sm;
                padding: 4px 8px;
                display: flex;
                align-items: center;
                gap: 4px;
                font-size: $font-size-sm;

                .req-icon {
                  font-size: 14px;
                }
              }
            }
          }

          .lock-overlay {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: rgba(0, 0, 0, 0.7);
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: $spacing-sm;

            .lock-icon {
              font-size: 32px;
            }

            .lock-text {
              font-size: $font-size-sm;
              color: $text-secondary;
              text-align: center;
              padding: 0 $spacing-md;
            }
          }

          .current-mission-badge {
            position: absolute;
            top: $spacing-md;
            right: $spacing-md;
            background-color: $primary-color;
            color: white;
            padding: 4px 10px;
            border-radius: $border-radius-sm;
            font-size: $font-size-xs;
            font-weight: 600;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
          }
        }
      }
    }
  }
}
</style>
