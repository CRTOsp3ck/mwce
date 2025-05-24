// src/components/ui/BaseNotification.vue
<template>
    <transition name="notification">
        <div v-if="visible" class="notification"
            :class="[`notification-${type}`, { 'notification-with-close': dismissible }]">
            <div class="notification-icon" v-if="icon">{{ icon }}</div>
            <div class="notification-content">
                <div v-if="title" class="notification-title">{{ title }}</div>
                <div class="notification-message">
                    <slot>{{ message }}</slot>
                </div>
            </div>
            <button v-if="dismissible" class="notification-close" @click="dismiss">×</button>
        </div>
    </transition>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, ref, onMounted, onUnmounted } from 'vue';

const props = defineProps({
    type: {
        type: String,
        default: 'info',
        validator: (value: string) => {
            return ['info', 'success', 'warning', 'danger'].includes(value);
        }
    },
    title: {
        type: String,
        default: ''
    },
    message: {
        type: String,
        default: ''
    },
    icon: {
        type: String,
        default: ''
    },
    dismissible: {
        type: Boolean,
        default: true
    },
    duration: {
        type: Number,
        default: 5000 // 5 seconds
    },
    autoClose: {
        type: Boolean,
        default: true
    }
});

const emit = defineEmits(['close']);
const visible = ref(true);
let timer: number | null = null;

// Set up auto-close timer
onMounted(() => {
    if (props.autoClose && props.duration > 0) {
        timer = window.setTimeout(() => {
            dismiss();
        }, props.duration);
    }
});

// Clean up timer if component is unmounted
onUnmounted(() => {
    if (timer) {
        clearTimeout(timer);
    }
});

// Dismiss the notification
const dismiss = () => {
    visible.value = false;
    emit('close');
};
</script>

<style lang="scss">
.notification {
    display: flex;
    align-items: flex-start;
    padding: $spacing-md;
    margin-bottom: $spacing-md;
    border-radius: $border-radius-md;
    box-shadow: $shadow-md;
    max-width: 400px;
    position: relative;

    &.notification-with-close {
        padding-right: 36px;
    }

    .notification-icon {
        margin-right: $spacing-md;
        font-size: 24px;
    }

    .notification-content {
        flex: 1;

        .notification-title {
            font-weight: 600;
            margin-bottom: $spacing-xs;
        }

        .notification-message {
            color: $text-color;
        }
    }

    .notification-close {
        position: absolute;
        top: $spacing-sm;
        right: $spacing-sm;
        background: none;
        border: none;
        color: $text-secondary;
        font-size: 18px;
        cursor: pointer;
        padding: 0;
        line-height: 1;

        &:hover {
            color: $text-color;
        }
    }

    // Notification types
    &.notification-info {
        background-color: rgba($info-color, 0.15);
        border-left: 4px solid $info-color;

        .notification-title {
            color: $info-color;
        }
    }

    &.notification-success {
        background-color: rgba($success-color, 0.15);
        border-left: 4px solid $success-color;

        .notification-title {
            color: $success-color;
        }
    }

    &.notification-warning {
        background-color: rgba($warning-color, 0.15);
        border-left: 4px solid $warning-color;

        .notification-title {
            color: $warning-color;
        }
    }

    &.notification-danger {
        background-color: rgba($danger-color, 0.15);
        border-left: 4px solid $danger-color;

        .notification-title {
            color: $danger-color;
        }
    }
}

// Transitions
.notification-enter-active,
.notification-leave-active {
    transition: all 0.3s ease;
}

.notification-enter-from,
.notification-leave-to {
    opacity: 0;
    transform: translateX(30px);
}
</style>