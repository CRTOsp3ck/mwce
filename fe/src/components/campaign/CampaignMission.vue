<!-- fe/src/components/campaign/CampaignMission.vue -->

<script setup lang="ts">
import { computed } from 'vue';
import { Mission, Branch } from '@/types/campaign';
import BaseCard from '@/components/ui/BaseCard.vue';

const props = defineProps<{
  mission: Mission;
}>();

const emit = defineEmits<{
  (e: 'selectBranch', mission: Mission, branch: Branch): void;
}>();

function selectBranch(branch: Branch) {
  emit('selectBranch', props.mission, branch);
}

const branches = computed(() => props.mission.branches || []);
</script>

<template>
  <div class="mission-view">
    <div class="mission-header">
      <h3>{{ mission.name }}</h3>
      <p class="description">{{ mission.description }}</p>
    </div>

    <div class="mission-branches">
      <h4>Choose Your Approach</h4>
      <p class="subtitle">Select a branch to pursue. Each branch represents a different approach to this mission.</p>

      <div class="branches-grid">
        <BaseCard v-for="branch in branches" :key="branch.id" class="branch-card" @click="selectBranch(branch)">
          <div class="branch-content">
            <h4 class="branch-title">{{ branch.name }}</h4>
            <p class="branch-description">{{ branch.description }}</p>
            <div class="branch-stats">
              <div class="stat" v-if="branch.operations && branch.operations.length > 0">
                <span class="stat-icon">🎯</span>
                <span class="stat-value">{{ branch.operations.length }} Operations</span>
              </div>
              <div class="stat" v-if="branch.pois && branch.pois.length > 0">
                <span class="stat-icon">📍</span>
                <span class="stat-value">{{ branch.pois.length }} Points of Interest</span>
              </div>
            </div>
          </div>
        </BaseCard>
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

    .branches-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: $spacing-md;

      .branch-card {
        cursor: pointer;
        transition: $transition-base;
        height: 100%;

        &:hover {
          transform: translateY(-5px);
          box-shadow: $shadow-lg;
        }

        .branch-content {
          padding: $spacing-md;

          .branch-title {
            margin-bottom: $spacing-xs;
            font-weight: 600;
          }

          .branch-description {
            color: $text-secondary;
            margin-bottom: $spacing-md;
            font-size: $font-size-sm;
            line-height: 1.5;
          }

          .branch-stats {
            display: flex;
            gap: $spacing-md;

            .stat {
              display: flex;
              align-items: center;
              gap: $spacing-xs;
              font-size: $font-size-sm;
              color: $text-secondary;
              background-color: rgba($background-lighter, 0.2);
              padding: 4px 8px;
              border-radius: $border-radius-sm;
            }
          }
        }
      }
    }
  }
}
</style>
