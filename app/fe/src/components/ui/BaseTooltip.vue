// src/components/ui/BaseTooltip.vue

<template>
  <div class="tooltip-wrapper" ref="wrapperRef" @mouseenter="handleMouseEnter" @mouseleave="handleMouseLeave">
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
          <div class="tooltip-arrow" :style="arrowStyle"></div>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { defineProps, ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue';

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
const wrapperRef = ref<HTMLElement | null>(null);
let delayTimeout: number | null = null;

const tooltipStyle = ref({
  top: '0px',
  left: '0px',
  maxWidth: props.maxWidth,
  position: 'fixed' as const,
  zIndex: 9999,
  visibility: 'hidden' as 'hidden' | 'visible'
});

const arrowStyle = ref({});

const handleMouseEnter = () => {
  if (props.delay > 0) {
    delayTimeout = window.setTimeout(() => {
      showTooltip.value = true;
    }, props.delay);
  } else {
    showTooltip.value = true;
  }
};

const handleMouseLeave = () => {
  if (delayTimeout) {
    clearTimeout(delayTimeout);
    delayTimeout = null;
  }
  showTooltip.value = false;
};

const updatePosition = async () => {
  if (!showTooltip.value || !tooltipEl.value || !wrapperRef.value) return;

  // Wait for the tooltip to be rendered in the DOM
  await nextTick();

  const wrapperRect = wrapperRef.value.getBoundingClientRect();
  const tooltipRect = tooltipEl.value.getBoundingClientRect();
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;
  const scrollX = window.scrollX;
  const scrollY = window.scrollY;

  let top = 0;
  let left = 0;
  let arrowLeft = 'auto';
  let arrowTop = 'auto';
  let arrowRight = 'auto';
  let arrowBottom = 'auto';

  // Calculate position based on the specified position prop
  switch (props.position) {
    case 'top':
      top = wrapperRect.top + scrollY - tooltipRect.height - props.offset;
      left = wrapperRect.left + scrollX + (wrapperRect.width - tooltipRect.width) / 2;
      arrowLeft = '50%';
      arrowBottom = '-5px';
      arrowStyle.value = {
        left: arrowLeft,
        bottom: arrowBottom,
        transform: 'translateX(-50%)',
        borderWidth: '5px 5px 0 5px',
        borderColor: `${getComputedStyle(tooltipEl.value).backgroundColor} transparent transparent transparent`
      };
      break;
    case 'bottom':
      top = wrapperRect.bottom + scrollY + props.offset;
      left = wrapperRect.left + scrollX + (wrapperRect.width - tooltipRect.width) / 2;
      arrowLeft = '50%';
      arrowTop = '-5px';
      arrowStyle.value = {
        left: arrowLeft,
        top: arrowTop,
        transform: 'translateX(-50%)',
        borderWidth: '0 5px 5px 5px',
        borderColor: `transparent transparent ${getComputedStyle(tooltipEl.value).backgroundColor} transparent`
      };
      break;
    case 'left':
      top = wrapperRect.top + scrollY + (wrapperRect.height - tooltipRect.height) / 2;
      left = wrapperRect.left + scrollX - tooltipRect.width - props.offset;
      arrowTop = '50%';
      arrowRight = '-5px';
      arrowStyle.value = {
        top: arrowTop,
        right: arrowRight,
        transform: 'translateY(-50%)',
        borderWidth: '5px 0 5px 5px',
        borderColor: `transparent transparent transparent ${getComputedStyle(tooltipEl.value).backgroundColor}`
      };
      break;
    case 'right':
      top = wrapperRect.top + scrollY + (wrapperRect.height - tooltipRect.height) / 2;
      left = wrapperRect.right + scrollX + props.offset;
      arrowTop = '50%';
      arrowLeft = '-5px';
      arrowStyle.value = {
        top: arrowTop,
        left: arrowLeft,
        transform: 'translateY(-50%)',
        borderWidth: '5px 5px 5px 0',
        borderColor: `transparent ${getComputedStyle(tooltipEl.value).backgroundColor} transparent transparent`
      };
      break;
  }

  // Ensure tooltip stays within viewport boundaries
  // Horizontal bounds
  const padding = 10;
  if (left < padding) {
    const diff = padding - left;
    left = padding;

    // Adjust arrow position when tooltip is repositioned
    if (props.position === 'top' || props.position === 'bottom') {
      arrowStyle.value = {
        ...arrowStyle.value,
        left: `${Math.max(10, Math.min(tooltipRect.width / 2 - diff, tooltipRect.width - 10))}px`,
        transform: 'none'
      };
    }
  } else if (left + tooltipRect.width > viewportWidth - padding) {
    const diff = (left + tooltipRect.width) - (viewportWidth - padding);
    left = viewportWidth - tooltipRect.width - padding;

    // Adjust arrow position when tooltip is repositioned
    if (props.position === 'top' || props.position === 'bottom') {
      arrowStyle.value = {
        ...arrowStyle.value,
        left: `${Math.max(10, Math.min(tooltipRect.width / 2 + diff, tooltipRect.width - 10))}px`,
        transform: 'none'
      };
    }
  }

  // Vertical bounds
  if (top < padding) {
    const diff = padding - top;
    top = padding;

    // Adjust arrow position when tooltip is repositioned
    if (props.position === 'left' || props.position === 'right') {
      arrowStyle.value = {
        ...arrowStyle.value,
        top: `${Math.max(10, Math.min(tooltipRect.height / 2 - diff, tooltipRect.height - 10))}px`,
        transform: 'none'
      };
    }
  } else if (top + tooltipRect.height > viewportHeight + scrollY - padding) {
    const diff = (top + tooltipRect.height) - (viewportHeight + scrollY - padding);
    top = viewportHeight + scrollY - tooltipRect.height - padding;

    // Adjust arrow position when tooltip is repositioned
    if (props.position === 'left' || props.position === 'right') {
      arrowStyle.value = {
        ...arrowStyle.value,
        top: `${Math.max(10, Math.min(tooltipRect.height / 2 + diff, tooltipRect.height - 10))}px`,
        transform: 'none'
      };
    }
  }

  // Update the tooltip position
  tooltipStyle.value = {
    ...tooltipStyle.value,
    top: `${top}px`,
    left: `${left}px`,
    visibility: 'visible'
  };
};

// Watch for visibility changes and update position accordingly
watch(showTooltip, async (newVal) => {
  if (newVal) {
    // Reset visibility to hidden initially
    tooltipStyle.value.visibility = 'hidden';
    // Update position after next tick
    await nextTick();
    updatePosition();
  } else {
    // Reset arrow style when hiding
    arrowStyle.value = {};
  }
});

// Update position on window resize or scroll
let resizeTimeout: number | null = null;
const handleResize = () => {
  if (resizeTimeout) clearTimeout(resizeTimeout);
  resizeTimeout = window.setTimeout(updatePosition, 100);
};

const handleScroll = () => {
  if (showTooltip.value) {
    updatePosition();
  }
};

onMounted(() => {
  window.addEventListener('resize', handleResize);
  window.addEventListener('scroll', handleScroll, true);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize);
  window.removeEventListener('scroll', handleScroll, true);
  if (resizeTimeout) clearTimeout(resizeTimeout);
  if (delayTimeout) clearTimeout(delayTimeout);
});
</script>

<style lang="scss">
.tooltip-wrapper {
  position: relative;
  display: inline-block;
}

.tooltip {
  position: fixed !important;
  z-index: 9999;
  padding: $spacing-xs $spacing-sm;
  background-color: rgba($background-darker, 0.95);
  color: $text-color;
  border-radius: $border-radius-sm;
  font-size: $font-size-sm;
  white-space: nowrap;
  pointer-events: none;
  box-shadow: $shadow-md;

  .tooltip-arrow {
    position: absolute;
    width: 0;
    height: 0;
    border-style: solid;
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

// Prevent text wrapping for longer tooltips
.tooltip {
  max-width: v-bind('props.maxWidth');
  white-space: normal;
  word-wrap: break-word;
}
</style>
