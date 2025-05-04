<!-- File: fe/src/components/campaign/MissionPOIList.vue -->

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { POI } from '@/types/campaign';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';

const props = defineProps<{
  pois: POI[];
}>();

const emit = defineEmits<{
  (e: 'complete', poiId: string): void;
}>();

const activePOIs = computed(() => props.pois.filter(poi => poi.isActive && !poi.isCompleted));
const completedPOIs = computed(() => props.pois.filter(poi => poi.isCompleted));

function handleCompletePOI(poiId: string) {
  emit('complete', poiId);
}

function formatLocationType(type: string): string {
  switch (type) {
    case 'hotspot':
      return 'Business';
    case 'region':
      return 'Region';
    case 'district':
      return 'District';
    case 'city':
      return 'City';
    default:
      return type.charAt(0).toUpperCase() + type.slice(1);
  }
}
</script>

<template>
  <div class="mission-poi-list">
    <h3 class="section-title" v-if="activePOIs.length > 0">Points of Interest</h3>

    <div class="pois-grid" v-if="activePOIs.length > 0">
      <BaseCard v-for="poi in activePOIs" :key="poi.id" class="poi-card">
        <div class="poi-header">
          <h4 class="poi-name">{{ poi.name }}</h4>
          <div class="poi-type">{{ formatLocationType(poi.locationType) }}</div>
        </div>

        <p class="poi-description">{{ poi.description }}</p>

        <div class="poi-actions">
          <BaseButton @click="handleCompletePOI(poi.id)">
            Mark as Visited
          </BaseButton>
        </div>
      </BaseCard>
    </div>

    <div class="completed-pois" v-if="completedPOIs.length > 0">
      <h4 class="completed-title">Completed ({{ completedPOIs.length }})</h4>

      <div class="completed-list">
        <div v-for="poi in completedPOIs" :key="poi.id" class="completed-poi">
          <span class="check-icon">✓</span>
          <span class="poi-name">{{ poi.name }}</span>
        </div>
      </div>
    </div>

    <div class="empty-state" v-if="activePOIs.length === 0 && completedPOIs.length === 0">
      <p>No points of interest available for this mission.</p>
    </div>
  </div>
</template>

<style lang="scss">
.mission-poi-list {
  margin-bottom: $spacing-xl;

  .section-title {
    margin-bottom: $spacing-md;
  }

  .pois-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: $spacing-md;

    @include respond-to(md) {
      grid-template-columns: repeat(2, 1fr);
    }

    .poi-card {
      .poi-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: $spacing-sm;

        .poi-name {
          margin: 0;
          font-size: $font-size-lg;
        }

        .poi-type {
          font-size: $font-size-xs;
          background-color: rgba($info-color, 0.2);
          color: $info-color;
          padding: 2px 8px;
          border-radius: $border-radius-sm;
        }
      }

      .poi-description {
        color: $text-secondary;
        margin-bottom: $spacing-md;
      }

      .poi-actions {
        display: flex;
        justify-content: flex-end;
      }
    }
  }

  .completed-pois {
    margin-top: $spacing-lg;

    .completed-title {
      font-size: $font-size-md;
      color: $text-secondary;
      margin-bottom: $spacing-sm;
    }

    .completed-list {
      display: flex;
      flex-wrap: wrap;
      gap: $spacing-sm;

      .completed-poi {
        background-color: rgba($success-color, 0.1);
        color: $success-color;
        padding: 4px 8px;
        border-radius: $border-radius-sm;
        font-size: $font-size-sm;
        display: flex;
        align-items: center;

        .check-icon {
          margin-right: 4px;
        }
      }
    }
  }

  .empty-state {
    padding: $spacing-md;
    text-align: center;
    color: $text-secondary;
    background-color: rgba($background-lighter, 0.2);
    border-radius: $border-radius-md;
  }
}
</style>
