<!-- fe/src/components/campaign/CampaignsList.vue -->

<script setup lang="ts">
import { Campaign } from '@/types/campaign';
import BaseCard from '@/components/ui/BaseCard.vue';

defineProps<{
  campaigns: Campaign[];
}>();

const emit = defineEmits<{
  (e: 'selectCampaign', campaign: Campaign): void;
}>();

function selectCampaign(campaign: Campaign) {
  emit('selectCampaign', campaign);
}
</script>

<template>
  <div class="campaigns-list">
    <div v-if="campaigns.length === 0" class="empty-state">
      <div class="empty-icon">📜</div>
      <h3>No Campaigns Available</h3>
      <p>Check back later for new campaign content.</p>
    </div>

    <div v-else class="campaigns-grid">
      <BaseCard v-for="campaign in campaigns" :key="campaign.id" class="campaign-card" @click="selectCampaign(campaign)">
        <div class="campaign-image" :style="{ backgroundImage: campaign.imageUrl ? `url(${campaign.imageUrl})` : '' }">
          <div v-if="!campaign.imageUrl" class="placeholder-image">
            🏙️
          </div>
        </div>
        <div class="campaign-content">
          <h3 class="campaign-title">{{ campaign.name }}</h3>
          <p class="campaign-description">{{ campaign.description }}</p>
          <div class="campaign-stats">
            <div class="stat">
              <span class="stat-icon">📊</span>
              <span class="stat-value">{{ campaign.chapters.length }} Chapters</span>
            </div>
            <div class="stat">
              <span class="stat-icon">🎯</span>
              <span class="stat-value">{{ campaign.chapters.reduce((acc, chapter) => acc + chapter.missions.length, 0) }} Missions</span>
            </div>
          </div>
        </div>
      </BaseCard>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.campaigns-list {
  .empty-state {
    @include flex-center;
    flex-direction: column;
    padding: $spacing-xl 0;
    text-align: center;

    .empty-icon {
      font-size: 48px;
      margin-bottom: $spacing-md;
    }

    h3 {
      margin-bottom: $spacing-sm;
    }

    p {
      color: $text-secondary;
    }
  }

  .campaigns-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: $spacing-lg;

    .campaign-card {
      cursor: pointer;
      transition: $transition-base;
      overflow: hidden;
      height: 100%;

      &:hover {
        transform: translateY(-5px);
        box-shadow: $shadow-lg;
      }

      .campaign-image {
        height: 160px;
        background-size: cover;
        background-position: center;
        position: relative;
        border-radius: $border-radius-sm $border-radius-sm 0 0;
        overflow: hidden;

        &::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: linear-gradient(to bottom, rgba(0, 0, 0, 0.3) 0%, rgba(0, 0, 0, 0.6) 100%);
        }

        .placeholder-image {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 100%;
          font-size: 48px;
          background-color: $background-darker;
        }
      }

      .campaign-content {
        padding: $spacing-md;

        .campaign-title {
          margin-bottom: $spacing-xs;
          font-size: $font-size-lg;
        }

        .campaign-description {
          color: $text-secondary;
          margin-bottom: $spacing-md;
          font-size: $font-size-sm;
          line-height: 1.5;
        }

        .campaign-stats {
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
    }
  }
}
</style>
