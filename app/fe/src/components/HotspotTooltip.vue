<!-- HotspotTooltip.vue -->
<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { Hotspot } from '@/types/territory';

const props = defineProps<{
  hotspot: Hotspot;
  visible: boolean;
  position: { x: number; y: number };
}>();

const tooltipRef = ref<HTMLElement | null>(null);
const adjustedPosition = computed(() => {
  if (!tooltipRef.value) return { left: props.position.x + 'px', top: props.position.y + 'px' };

  // Get viewport dimensions
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;

  // Get tooltip dimensions
  const tooltipWidth = tooltipRef.value.offsetWidth;
  const tooltipHeight = tooltipRef.value.offsetHeight;

  // Calculate positions
  let left = props.position.x + 15; // 15px offset from cursor
  let top = props.position.y + 15;

  // Adjust if tooltip would extend beyond viewport
  if (left + tooltipWidth > viewportWidth - 20) {
    left = props.position.x - tooltipWidth - 15;
  }

  if (top + tooltipHeight > viewportHeight - 20) {
    top = props.position.y - tooltipHeight - 15;
  }

  return {
    left: left + 'px',
    top: top + 'px'
  };
});

// Force recalculation of position when visibility changes
watch(() => props.visible, () => {
  // Small delay to ensure DOM is updated
  setTimeout(() => {
    if (tooltipRef.value) {
      tooltipRef.value.style.left = adjustedPosition.value.left;
      tooltipRef.value.style.top = adjustedPosition.value.top;
    }
  }, 10);
});
</script>

<template>
  <div
    v-if="visible"
    ref="tooltipRef"
    class="hotspot-tooltip"
    :style="{
      left: adjustedPosition.left,
      top: adjustedPosition.top
    }"
  >
    <div class="tooltip-header">
      <div class="hotspot-name">{{ hotspot.name }}</div>
      <div class="hotspot-type">{{ hotspot.type }}</div>
    </div>

    <div class="tooltip-content">
      <div class="tooltip-row">
        <div class="tooltip-icon">📍</div>
        <div class="tooltip-text">{{ hotspot.businessType }}</div>
      </div>

      <div class="tooltip-row" v-if="hotspot.isLegal">
        <div class="tooltip-icon">💵</div>
        <div class="tooltip-text">${{ hotspot.income }}/hr</div>
      </div>

      <div class="tooltip-row" v-if="hotspot.controller">
        <div class="tooltip-icon">👑</div>
        <div class="tooltip-text">{{ hotspot.controllerName || 'Unknown Boss' }}</div>
      </div>

      <div class="tooltip-divider"></div>

      <div class="tooltip-footer">Click for details</div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.hotspot-tooltip {
  position: fixed;
  z-index: 1000;
  width: 220px;
  background: $background-darker;
  border: 1px solid $border-color;
  border-radius: $border-radius-md;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
  pointer-events: none;
  opacity: 0;
  transform: translateY(10px);
  animation: tooltip-fade-in 0.2s forwards;

  // Gold accent border
  &::after {
    content: "";
    position: absolute;
    inset: 0;
    border-radius: $border-radius-md;
    padding: 1px;
    background: linear-gradient(
      45deg,
      rgba($gold-color, 0.3),
      rgba($gold-color, 0.8),
      rgba($gold-color, 0.3)
    );
    mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
    mask-composite: exclude;
    pointer-events: none;
  }

  .tooltip-header {
    padding: $spacing-sm;
    background-color: rgba($background-lighter, 0.2);
    border-bottom: 1px solid rgba($border-color, 0.6);
    border-radius: $border-radius-md $border-radius-md 0 0;

    .hotspot-name {
      font-weight: 600;
      font-size: $font-size-md;
      color: $gold-color;
    }

    .hotspot-type {
      font-size: $font-size-xs;
      color: $text-secondary;
    }
  }

  .tooltip-content {
    padding: $spacing-sm;

    .tooltip-row {
      display: flex;
      align-items: center;
      gap: $spacing-sm;
      margin-bottom: $spacing-xs;

      .tooltip-icon {
        width: 20px;
        height: 20px;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      .tooltip-text {
        font-size: $font-size-sm;
      }
    }

    .tooltip-divider {
      height: 1px;
      background: linear-gradient(
        to right,
        rgba($border-color, 0),
        rgba($border-color, 0.8),
        rgba($border-color, 0)
      );
      margin: $spacing-xs 0;
    }

    .tooltip-footer {
      text-align: center;
      font-size: $font-size-xs;
      color: $text-secondary;
      font-style: italic;
    }
  }
}

@keyframes tooltip-fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
