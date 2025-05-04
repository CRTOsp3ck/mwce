<!-- File: fe/src/components/campaign/ChoiceProgressTracker.vue -->

<script setup lang="ts">
import { computed } from 'vue';
import { CompletionCondition } from '@/types/campaign';

const props = defineProps<{
  choice: {
    id: string;
    text: string;
    sequentialOrder: boolean;
  };
  conditions: CompletionCondition[];
}>();

const progress = computed(() => {
  if (!props.conditions || props.conditions.length === 0) return 0;

  const completedCount = props.conditions.filter(c => c.isCompleted).length;
  return Math.floor((completedCount / props.conditions.length) * 100);
});

const nextCondition = computed(() => {
  if (!props.choice.sequentialOrder) return null;

  // Find the first incomplete condition in order
  return props.conditions
    .sort((a, b) => a.orderIndex - b.orderIndex)
    .find(c => !c.isCompleted);
});

function formatConditionType(type: string, value: string): string {
  switch (type) {
    case 'travel':
      return `Travel to ${value}`;
    case 'territory':
      return `Territory action: ${value.replace('_', ' ')}`;
    case 'operation':
      return `Complete operation: ${value}`;
    default:
      return `${type}: ${value}`;
  }
}
</script>

<template>
  <div class="choice-progress-tracker">
    <div class="choice-info">
      <h4 class="choice-text">{{ choice.text }}</h4>

      <div class="progress-container">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: `${progress}%` }"></div>
        </div>
        <div class="progress-text">{{ progress }}% Complete</div>
      </div>
    </div>

    <div class="conditions-list">
      <h5 class="conditions-title">Requirements:</h5>

      <div
        v-for="condition in conditions.sort((a, b) => a.orderIndex - b.orderIndex)"
        :key="condition.id"
        class="condition-item"
        :class="{
          'completed': condition.isCompleted,
          'next': nextCondition && condition.id === nextCondition.id
        }"
      >
        <div class="condition-status">
          <span v-if="condition.isCompleted" class="status-icon completed">✓</span>
          <span v-else-if="nextCondition && condition.id === nextCondition.id" class="status-icon next">→</span>
          <span v-else class="status-icon pending">○</span>
        </div>

        <div class="condition-text">
          {{ formatConditionType(condition.type, condition.requiredValue) }}
        </div>
      </div>
    </div>

    <div v-if="nextCondition && choice.sequentialOrder" class="next-step">
      <div class="next-step-label">Next Step:</div>
      <div class="next-step-text">{{ formatConditionType(nextCondition.type, nextCondition.requiredValue) }}</div>
    </div>
  </div>
</template>

<style lang="scss">
.choice-progress-tracker {
  background-color: rgba($background-lighter, 0.2);
  border-radius: $border-radius-md;
  padding: $spacing-md;
  margin-bottom: $spacing-xl;

  .choice-info {
    margin-bottom: $spacing-md;

    .choice-text {
      margin: 0 0 $spacing-sm 0;
      font-size: $font-size-md;
    }

    .progress-container {
      .progress-bar {
        height: 8px;
        background-color: rgba($background-lighter, 0.3);
        border-radius: 4px;
        overflow: hidden;
        margin-bottom: 4px;

        .progress-fill {
          height: 100%;
          background-color: $success-color;
          border-radius: 4px;
        }
      }

      .progress-text {
        font-size: $font-size-xs;
        color: $text-secondary;
        text-align: right;
      }
    }
  }

  .conditions-list {
    margin-bottom: $spacing-md;

    .conditions-title {
      margin: 0 0 $spacing-sm 0;
      font-size: $font-size-sm;
      color: $text-secondary;
    }

    .condition-item {
      display: flex;
      align-items: flex-start;
      gap: $spacing-sm;
      margin-bottom: $spacing-xs;
      padding: 4px 0;

      .condition-status {
        .status-icon {
          display: inline-flex;
          align-items: center;
          justify-content: center;
          width: 20px;
          height: 20px;
          border-radius: 50%;

          &.completed {
            background-color: $success-color;
            color: white;
          }

          &.next {
            background-color: $primary-color;
            color: white;
          }

          &.pending {
            border: 1px solid $text-secondary;
            color: $text-secondary;
          }
        }
      }

      .condition-text {
        font-size: $font-size-sm;
      }

      &.completed {
        .condition-text {
          text-decoration: line-through;
          color: $text-secondary;
        }
      }

      &.next {
        background-color: rgba($primary-color, 0.1);
        border-radius: $border-radius-sm;
        padding: 4px 8px;
      }
    }
  }

  .next-step {
    background-color: rgba($primary-color, 0.1);
    border-left: 3px solid $primary-color;
    padding: $spacing-sm;
    border-radius: $border-radius-sm;

    .next-step-label {
      font-size: $font-size-xs;
      color: $text-secondary;
      margin-bottom: 2px;
    }

    .next-step-text {
      font-weight: 600;
      color: $primary-color;
    }
  }
}
</style>
