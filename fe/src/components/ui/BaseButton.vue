// src/components/ui/BaseButton.vue

<template>
    <button :class="[
        'base-button',
        `variant-${variant}`,
        { 'is-loading': loading, 'is-block': block, 'is-small': small }
    ]" :disabled="disabled || loading" @click="handleClick">
        <span v-if="loading" class="loading-spinner"></span>
        <span class="button-content" :class="{ 'invisible': loading }">
            <slot></slot>
        </span>
    </button>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue';

const props = defineProps({
    variant: {
        type: String,
        default: 'primary',
        validator: (value: string) => {
            return ['primary', 'secondary', 'danger', 'success', 'warning', 'info', 'outline', 'text'].includes(value);
        }
    },
    loading: {
        type: Boolean,
        default: false
    },
    disabled: {
        type: Boolean,
        default: false
    },
    block: {
        type: Boolean,
        default: false
    },
    small: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['click']);

const handleClick = (event: MouseEvent) => {
    if (!props.disabled && !props.loading) {
        emit('click', event);
    }
};
</script>

<style lang="scss">
.base-button {
    @include button-base;
    position: relative;
    min-width: 120px;

    &.is-block {
        display: block;
        width: 100%;
    }

    &.is-small {
        padding: $spacing-xs $spacing-md;
        font-size: $font-size-sm;
        min-width: auto;
    }

    .button-content {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: $spacing-sm;

        &.invisible {
            visibility: hidden;
        }
    }

    .loading-spinner {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        width: 20px;
        height: 20px;
        border: 2px solid rgba(255, 255, 255, 0.3);
        border-radius: 50%;
        border-top-color: white;
        animation: spin 0.8s linear infinite;
    }

    &.variant-primary {
        background-color: $primary-color;
        color: white;

        &:hover:not(:disabled) {
            background-color: lighten($primary-color, 5%);
        }

        &:active:not(:disabled) {
            background-color: darken($primary-color, 5%);
        }
    }

    &.variant-secondary {
        background-color: $secondary-color;
        color: $background-color;

        &:hover:not(:disabled) {
            background-color: lighten($secondary-color, 5%);
        }

        &:active:not(:disabled) {
            background-color: darken($secondary-color, 5%);
        }

        .loading-spinner {
            border: 2px solid rgba($background-color, 0.3);
            border-top-color: $background-color;
        }
    }

    &.variant-danger {
        background-color: $danger-color;
        color: white;

        &:hover:not(:disabled) {
            background-color: lighten($danger-color, 5%);
        }

        &:active:not(:disabled) {
            background-color: darken($danger-color, 5%);
        }
    }

    &.variant-success {
        background-color: $success-color;
        color: white;

        &:hover:not(:disabled) {
            background-color: lighten($success-color, 5%);
        }

        &:active:not(:disabled) {
            background-color: darken($success-color, 5%);
        }
    }

    &.variant-warning {
        background-color: $warning-color;
        color: $background-color;

        &:hover:not(:disabled) {
            background-color: lighten($warning-color, 5%);
        }

        &:active:not(:disabled) {
            background-color: darken($warning-color, 5%);
        }

        .loading-spinner {
            border: 2px solid rgba($background-color, 0.3);
            border-top-color: $background-color;
        }
    }

    &.variant-info {
        background-color: $info-color;
        color: white;

        &:hover:not(:disabled) {
            background-color: lighten($info-color, 5%);
        }

        &:active:not(:disabled) {
            background-color: darken($info-color, 5%);
        }
    }

    &.variant-outline {
        background-color: transparent;
        border: 1px solid $secondary-color;
        color: $secondary-color;

        &:hover:not(:disabled) {
            background-color: rgba($secondary-color, 0.1);
        }

        &:active:not(:disabled) {
            background-color: rgba($secondary-color, 0.2);
        }

        .loading-spinner {
            border: 2px solid rgba($secondary-color, 0.3);
            border-top-color: $secondary-color;
        }
    }

    &.variant-text {
        background-color: transparent;
        color: $secondary-color;
        padding: $spacing-xs;
        min-width: auto;

        &:hover:not(:disabled) {
            background-color: rgba($secondary-color, 0.1);
        }

        &:active:not(:disabled) {
            background-color: rgba($secondary-color, 0.2);
        }

        .loading-spinner {
            border: 2px solid rgba($secondary-color, 0.3);
            border-top-color: $secondary-color;
        }
    }
}

@keyframes spin {
    to {
        transform: translate(-50%, -50%) rotate(360deg);
    }
}
</style>