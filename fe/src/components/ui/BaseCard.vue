// src/components/ui/BaseCard.vue

<template>
    <div class="base-card"
         :class="{
            'gold-border': goldBorder,
            'mobile-collapsible': collapsible,
            'is-collapsed': isCollapsed && collapsible
         }">
        <div v-if="title || $slots.header" class="card-header" @click="toggleCollapse" :class="{ 'collapsible': collapsible }">
            <h3 v-if="title" class="card-title">{{ title }}</h3>
            <slot name="header"></slot>
            <div v-if="collapsible" class="collapse-toggle">
                {{ isCollapsed ? '▼' : '▲' }}
            </div>
        </div>

        <div class="card-body" :class="{ 'collapsible-content': collapsible }" v-show="!isCollapsed || !collapsible">
            <slot></slot>
        </div>

        <div v-if="$slots.footer" class="card-footer" v-show="!isCollapsed || !collapsible">
            <slot name="footer"></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
import { defineProps, ref } from 'vue';

const props = defineProps({
    title: {
        type: String,
        default: ''
    },
    goldBorder: {
        type: Boolean,
        default: false
    },
    collapsible: {
        type: Boolean,
        default: false
    },
    defaultCollapsed: {
        type: Boolean,
        default: false
    }
});

// Collapse state for mobile cards
const isCollapsed = ref(props.defaultCollapsed);

// Toggle collapse state
function toggleCollapse() {
    if (props.collapsible) {
        isCollapsed.value = !isCollapsed.value;
    }
}
</script>

<style lang="scss">
.base-card {
    @include card;
    overflow: hidden;
    transition: $transition-base;

    &.gold-border {
        @include gold-border;
    }

    // Mobile-optimized card styling
    &.mobile-collapsible {
        .card-header {
            cursor: pointer;
            position: relative;

            &:hover {
                background-color: rgba(255, 255, 255, 0.05);
            }

            &.collapsible {
                padding-right: 30px; // Space for toggle icon
            }

            .collapse-toggle {
                position: absolute;
                right: $spacing-sm;
                top: 50%;
                transform: translateY(-50%);
                font-size: 12px;
                color: $text-secondary;
                transition: transform 0.3s ease;
            }
        }

        &.is-collapsed {
            .collapse-toggle {
                transform: translateY(-50%) rotate(180deg);
            }
        }

        .collapsible-content {
            transition: max-height 0.3s ease, opacity 0.3s ease;
        }
    }

    .card-header {
        padding-bottom: $spacing-md;
        margin-bottom: $spacing-md;
        border-bottom: 1px solid $border-color;
        @include flex-between;

        // More compact on mobile
        @media (max-width: $breakpoint-md) {
            padding-bottom: $spacing-sm;
            margin-bottom: $spacing-sm;
        }

        .card-title {
            margin: 0;
            font-size: $font-size-lg;

            // Slightly smaller on mobile
            @media (max-width: $breakpoint-md) {
                font-size: $font-size-md;
            }
        }
    }

    .card-body {
        flex: 1;
    }

    .card-footer {
        padding-top: $spacing-md;
        margin-top: $spacing-md;
        border-top: 1px solid $border-color;
        @include flex-between;

        // More compact on mobile
        @media (max-width: $breakpoint-md) {
            padding-top: $spacing-sm;
            margin-top: $spacing-sm;
        }
    }

    // Add game-like animations and effects on mobile
    @media (max-width: $breakpoint-md) {
        &:active {
            transform: scale(0.99);
        }
    }
}
</style>
