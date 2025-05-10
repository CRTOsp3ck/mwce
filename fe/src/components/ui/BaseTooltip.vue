// src/components/ui/BaseTooltip.vue (Updated)
<template>
  <div class="tooltip-wrapper" @mouseenter="showTooltip = true" @mouseleave="showTooltip = false">
    <slot></slot>
    <teleport to="body">
      <transition name="tooltip-fade">
        <div
          v-if="showTooltip"
          :class="['tooltip', position]"
          :style="tooltipStyle"
          ref="tooltipEl"
        >
          <div class="tooltip-content">
            <slot name="tooltip">{{ text }}</slot>
          </div>
          <div class="tooltip-arrow"></div>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { defineProps, ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';

const props = defineProps({
  text: {
    type: String,
    default: ''
  },
  position: {
    type: String,
    default: 'top',
    validator: (value: string) => {
      return ['top', 'bottom', 'left', 'right'].includes(value);
    }
  },
  delay: {
    type: Number,
    default: 0
  },
  offset: {
    type: Number,
    default: 8
  },
  maxWidth: {
    type: String,
    default: '200px'
  }
});

const showTooltip = ref(false);
const tooltipEl = ref<HTMLElement | null>(null);
const wrapperEl = ref<HTMLElement | null>(null);

const tooltipStyle = ref({
  top: '0px',
  left: '0px',
  maxWidth: props.maxWidth,
  position: 'fixed' as const
});

const updatePosition = () => {
  if (!showTooltip.value || !tooltipEl.value) return;

  const wrapper = document.querySelector('.tooltip-wrapper');
  if (!wrapper) return;

  const wrapperRect = wrapper.getBoundingClientRect();
  const tooltipRect = tooltipEl.value.getBoundingClientRect();

  let top = 0;
  let left = 0;

  switch (props.position) {
    case 'top':
      top = wrapperRect.top - tooltipRect.height - props.offset;
      left = wrapperRect.left + (wrapperRect.width - tooltipRect.width) / 2;
      break;
    case 'bottom':
      top = wrapperRect.bottom + props.offset;
      left = wrapperRect.left + (wrapperRect.width - tooltipRect.width) / 2;
      break;
    case 'left':
      top = wrapperRect.top + (wrapperRect.height - tooltipRect.height) / 2;
      left = wrapperRect.left - tooltipRect.width - props.offset;
      break;
    case 'right':
      top = wrapperRect.top + (wrapperRect.height - tooltipRect.height) / 2;
      left = wrapperRect.right + props.offset;
      break;
  }

  // Ensure tooltip stays within viewport
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;

  // Horizontal bounds
  if (left < 5) {
    left = 5;
  } else if (left + tooltipRect.width > viewportWidth - 5) {
    left = viewportWidth - tooltipRect.width - 5;
  }

  // Vertical bounds
  if (top < 5) {
    top = 5;
  } else if (top + tooltipRect.height > viewportHeight - 5) {
    top = viewportHeight - tooltipRect.height - 5;
  }

  tooltipStyle.value = {
    ...tooltipStyle.value,
    top: `${top}px`,
    left: `${left}px`
  };
};

// Watch for visibility changes
watch(showTooltip, (newVal) => {
  if (newVal) {
    // Use nextTick to ensure DOM is updated
    setTimeout(updatePosition, 0);
  }
});

// Update position on window resize
let resizeTimeout: number | null = null;
const handleResize = () => {
  if (resizeTimeout) clearTimeout(resizeTimeout);
  resizeTimeout = window.setTimeout(updatePosition, 100);
};

onMounted(() => {
  window.addEventListener('resize', handleResize);
  window.addEventListener('scroll', updatePosition);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize);
  window.removeEventListener('scroll', updatePosition);
  if (resizeTimeout) clearTimeout(resizeTimeout);
});
</script>

<style lang="scss">
.tooltip-wrapper {
  position: relative;
  display: inline-block;
}

.tooltip {
  position: fixed !important; // Force fixed positioning
  z-index: $z-index-tooltip;
  padding: $spacing-xs $spacing-sm;
  background-color: rgba($background-darker, 0.95);
  color: $text-color;
  border-radius: $border-radius-sm;
  font-size: $font-size-sm;
  white-space: nowrap;
  pointer-events: none;
  box-shadow: $shadow-md;

  .tooltip-arrow {
    display: none; // Simplified to avoid complex positioning
  }

  .tooltip-content {
    position: relative;
    z-index: 1;
  }
}

.tooltip-fade-enter-active,
.tooltip-fade-leave-active {
  transition: opacity 0.2s ease;
}

.tooltip-fade-enter-from,
.tooltip-fade-leave-to {
  opacity: 0;
}
</style>
