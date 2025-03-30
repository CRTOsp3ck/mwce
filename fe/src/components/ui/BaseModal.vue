// src/components/ui/BaseModal.vue
<template>
    <transition name="modal">
        <div v-if="modelValue" class="modal-overlay" @click="closeOnOutsideClick && close()">
            <div class="modal-container" @click.stop>
                <div class="modal-header">
                    <h3 v-if="title" class="modal-title">{{ title }}</h3>
                    <button class="modal-close" @click="close">×</button>
                </div>

                <div class="modal-body">
                    <slot></slot>
                </div>

                <div v-if="$slots.footer" class="modal-footer">
                    <slot name="footer"></slot>
                </div>
                <div v-else-if="showDefaultFooter" class="modal-footer">
                    <BaseButton variant="text" @click="close">Cancel</BaseButton>
                    <BaseButton @click="confirm">Confirm</BaseButton>
                </div>
            </div>
        </div>
    </transition>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue';
import BaseButton from './BaseButton.vue';

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true
    },
    title: {
        type: String,
        default: ''
    },
    closeOnOutsideClick: {
        type: Boolean,
        default: true
    },
    showDefaultFooter: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['update:modelValue', 'confirm']);

const close = () => {
    emit('update:modelValue', false);
};

const confirm = () => {
    emit('confirm');
    close();
};
</script>

<style lang="scss">
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.75);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: $z-index-modal;
    padding: $spacing-md;
}

.modal-container {
    background-color: $background-card;
    border-radius: $border-radius-md;
    box-shadow: $shadow-lg;
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    animation: modal-appear 0.3s ease-out;

    @include respond-to(sm) {
        width: 500px;
    }
}

.modal-header {
    @include flex-between;
    padding: $spacing-md;
    border-bottom: 1px solid $border-color;

    .modal-title {
        margin: 0;
        font-size: $font-size-lg;
        color: $secondary-color;
    }

    .modal-close {
        background: none;
        border: none;
        color: $text-secondary;
        font-size: 24px;
        cursor: pointer;
        padding: 0;
        line-height: 1;

        &:hover {
            color: $text-color;
        }
    }
}

.modal-body {
    padding: $spacing-lg;
    overflow-y: auto;
    max-height: 60vh;
}

.modal-footer {
    @include flex-between;
    padding: $spacing-md;
    border-top: 1px solid $border-color;
}

// Transitions
.modal-enter-active,
.modal-leave-active {
    transition: opacity 0.3s;
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

@keyframes modal-appear {
    from {
        opacity: 0;
        transform: translateY(20px);
    }

    to {
        opacity: 1;
        transform: translateY(0);
    }
}
</style>