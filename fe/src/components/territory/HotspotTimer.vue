// src/components/territory/HotspotTimer.vue

<template>
  <div class="hotspot-timer" :class="{ 'completed': timeRemaining <= 0 }">
    <div class="timer-icon">⏱️</div>
    <div class="timer-content">
      <div class="timer-label">{{ label }}</div>
      <div class="timer-value">{{ displayTime }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { formatTimeRemaining } from '@/utils/incomeScheduler';

const props = defineProps<{
  timeRemaining: number;
  label?: string;
}>();

const emit = defineEmits<{
  (e: 'timerCompleted'): void;
}>();

// Local copy of the remaining time
const secondsRemaining = ref(props.timeRemaining);

// Timer interval reference
const timerInterval = ref<number | null>(null);

// Formatted time display
const displayTime = computed(() => {
  return formatTimeRemaining(secondsRemaining.value);
});

// Update timer every second
function updateTimer() {
  if (secondsRemaining.value <= 0) {
    // Timer already complete, no need to count down
    return;
  }
  
  secondsRemaining.value -= 1;
  
  // Emit event when timer reaches zero
  if (secondsRemaining.value <= 0) {
    emit('timerCompleted');
  }
}

// Start timer
onMounted(() => {
  // Initialize with prop value
  secondsRemaining.value = props.timeRemaining;
  
  // Start interval
  timerInterval.value = window.setInterval(updateTimer, 1000);
});

// Clean up
onBeforeUnmount(() => {
  if (timerInterval.value) {
    clearInterval(timerInterval.value);
  }
});
</script>

<style lang="scss">
.hotspot-timer {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
  padding: $spacing-xs;
  background-color: rgba($background-darker, 0.3);
  border-radius: $border-radius-sm;
  
  &.completed {
    background-color: rgba($success-color, 0.2);
    
    .timer-value {
      color: $success-color;
      font-weight: 600;
    }
  }
  
  .timer-icon {
    font-size: 16px;
  }
  
  .timer-content {
    .timer-label {
      font-size: $font-size-xs;
      color: $text-secondary;
    }
    
    .timer-value {
      font-size: $font-size-sm;
      font-family: monospace;
    }
  }
}
</style>