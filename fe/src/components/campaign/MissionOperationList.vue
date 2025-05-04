<!-- File: fe/src/components/campaign/MissionOperationList.vue -->

<script setup lang="ts">
import { ref, computed } from 'vue';
import { MissionOperation } from '@/types/campaign';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';

const props = defineProps<{
  operations: MissionOperation[];
}>();

const emit = defineEmits<{
  (e: 'start', operationId: string): void;
  (e: 'complete', operationId: string): void;
}>();

const activeOperations = computed(() => props.operations.filter(op => op.isActive && !op.isCompleted));
const completedOperations = computed(() => props.operations.filter(op => op.isCompleted));
const inactiveOperations = computed(() => props.operations.filter(op => !op.isActive && !op.isCompleted));

function handleStartOperation(operationId: string) {
  emit('start', operationId);
}

function handleCompleteOperation(operationId: string) {
  emit('complete', operationId);
}

function formatOperationType(type: string): string {
  // Convert camelCase or snake_case to Title Case
  return type
    .replace(/_/g, ' ')
    .replace(/([A-Z])/g, ' $1')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}
</script>

<template>
  <div class="mission-operation-list">
    <h3 class="section-title" v-if="activeOperations.length > 0 || inactiveOperations.length > 0">
      Mission Operations
    </h3>

    <div class="operations-grid" v-if="activeOperations.length > 0">
      <BaseCard v-for="operation in activeOperations" :key="operation.id" class="operation-card active">
        <div class="operation-header">
          <h4 class="operation-name">{{ operation.name }}</h4>
          <div class="operation-type">{{ formatOperationType(operation.operationType) }}</div>
        </div>

        <p class="operation-description">{{ operation.description }}</p>

        <div class="operation-details">
          <div class="detail-item">
            <span class="detail-label">Duration:</span>
            <span class="detail-value">{{ formatDuration(operation.duration) }}</span>
          </div>

          <div class="detail-item">
            <span class="detail-label">Success Rate:</span>
            <span class="detail-value">{{ operation.successRate }}%</span>
          </div>
        </div>

        <div class="operation-resources">
          <div class="resource-item" v-if="operation.resources.crew > 0">
            <span class="resource-icon">👥</span>
            <span class="resource-value">{{ operation.resources.crew }}</span>
          </div>

          <div class="resource-item" v-if="operation.resources.weapons > 0">
            <span class="resource-icon">🔫</span>
            <span class="resource-value">{{ operation.resources.weapons }}</span>
          </div>

          <div class="resource-item" v-if="operation.resources.vehicles > 0">
            <span class="resource-icon">🚗</span>
            <span class="resource-value">{{ operation.resources.vehicles }}</span>
          </div>

          <div class="resource-item" v-if="operation.resources.money > 0">
            <span class="resource-icon">💰</span>
            <span class="resource-value">${{ operation.resources.money }}</span>
          </div>
        </div>

        <div class="operation-actions">
          <BaseButton @click="handleCompleteOperation(operation.id)">
            Complete Operation
          </BaseButton>
        </div>
      </BaseCard>
    </div>

    <div class="operations-grid" v-if="inactiveOperations.length > 0">
      <BaseCard v-for="operation in inactiveOperations" :key="operation.id" class="operation-card inactive">
        <div class="operation-header">
          <h4 class="operation-name">{{ operation.name }}</h4>
          <div class="operation-type">{{ formatOperationType(operation.operationType) }}</div>
        </div>

        <p class="operation-description">{{ operation.description }}</p>

        <div class="operation-details">
          <div class="detail-item">
            <span class="detail-label">Duration:</span>
            <span class="detail-value">{{ formatDuration(operation.duration) }}</span>
          </div>

          <div class="detail-item">
            <span class="detail-label">Success Rate:</span>
            <span class="detail-value">{{ operation.successRate }}%</span>
          </div>
        </div>

        <div class="operation-resources">
          <div class="resource-item" v-if="operation.resources.crew > 0">
            <span class="resource-icon">👥</span>
            <span class="resource-value">{{ operation.resources.crew }}</span>
          </div>

          <div class="resource-item" v-if="operation.resources.weapons > 0">
            <span class="resource-icon">🔫</span>
            <span class="resource-value">{{ operation.resources.weapons }}</span>
          </div>

          <div class="resource-item" v-if="operation.resources.vehicles > 0">
            <span class="resource-icon">🚗</span>
            <span class="resource-value">{{ operation.resources.vehicles }}</span>
          </div>

          <div class="resource-item" v-if="operation.resources.money > 0">
            <span class="resource-icon">💰</span>
            <span class="resource-value">${{ operation.resources.money }}</span>
          </div>
        </div>

        <div class="operation-actions">
          <BaseButton @click="handleStartOperation(operation.id)" variant="secondary">
            Start Operation
          </BaseButton>
        </div>
      </BaseCard>
    </div>

    <div class="completed-operations" v-if="completedOperations.length > 0">
      <h4 class="completed-title">Completed Operations ({{ completedOperations.length }})</h4>

      <div class="completed-list">
        <div v-for="operation in completedOperations" :key="operation.id" class="completed-operation">
          <span class="check-icon">✓</span>
          <span class="operation-name">{{ operation.name }}</span>
        </div>
      </div>
    </div>

    <div class="empty-state" v-if="activeOperations.length === 0 && inactiveOperations.length === 0 && completedOperations.length === 0">
      <p>No operations available for this mission.</p>
    </div>
  </div>
</template>

<style lang="scss">
.mission-operation-list {
  margin-bottom: $spacing-xl;

  .section-title {
    margin-bottom: $spacing-md;
  }

  .operations-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: $spacing-md;
    margin-bottom: $spacing-lg;

    @include respond-to(md) {
      grid-template-columns: repeat(2, 1fr);
    }

    .operation-card {
      .operation-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: $spacing-sm;

        .operation-name {
          margin: 0;
          font-size: $font-size-lg;
        }

        .operation-type {
          font-size: $font-size-xs;
          background-color: rgba($primary-color, 0.2);
          color: $primary-color;
          padding: 2px 8px;
          border-radius: $border-radius-sm;
        }
      }

      .operation-description {
        color: $text-secondary;
        margin-bottom: $spacing-md;
      }

      .operation-details {
        display: flex;
        flex-wrap: wrap;
        gap: $spacing-md;
        margin-bottom: $spacing-md;

        .detail-item {
          .detail-label {
            font-size: $font-size-sm;
            color: $text-secondary;
            margin-right: $spacing-xs;
          }

          .detail-value {
            font-weight: 600;
          }
        }
      }

      .operation-resources {
        display: flex;
        gap: $spacing-sm;
        margin-bottom: $spacing-md;

        .resource-item {
          background-color: rgba($background-lighter, 0.3);
          padding: 4px 8px;
          border-radius: $border-radius-sm;
          display: flex;
          align-items: center;
          gap: 4px;
          font-size: $font-size-sm;

          .resource-icon {
            font-size: 14px;
          }
        }
      }

      .operation-actions {
        display: flex;
        justify-content: flex-end;
      }

      &.active {
        border-left: 3px solid $primary-color;
      }

      &.inactive {
        opacity: 0.8;
      }
    }
  }

  .completed-operations {
    .completed-title {
      font-size: $font-size-md;
      color: $text-secondary;
      margin-bottom: $spacing-sm;
    }

    .completed-list {
      display: flex;
      flex-wrap: wrap;
      gap: $spacing-sm;

      .completed-operation {
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
