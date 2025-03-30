// src/components/ui/BaseCard.vue

<template>
    <div class="base-card" :class="{ 'gold-border': goldBorder }">
        <div v-if="title || $slots.header" class="card-header">
            <h3 v-if="title" class="card-title">{{ title }}</h3>
            <slot name="header"></slot>
        </div>

        <div class="card-body">
            <slot></slot>
        </div>

        <div v-if="$slots.footer" class="card-footer">
            <slot name="footer"></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue';

defineProps({
    title: {
        type: String,
        default: ''
    },
    goldBorder: {
        type: Boolean,
        default: false
    }
});
</script>

<style lang="scss">
.base-card {
    @include card;
    overflow: hidden;
    transition: $transition-base;

    &.gold-border {
        @include gold-border;
    }

    .card-header {
        padding-bottom: $spacing-md;
        margin-bottom: $spacing-md;
        border-bottom: 1px solid $border-color;
        @include flex-between;

        .card-title {
            margin: 0;
            font-size: $font-size-lg;
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
    }
}
</style>