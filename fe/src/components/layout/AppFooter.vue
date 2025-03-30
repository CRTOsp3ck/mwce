// src/components/layout/AppFooter.vue

<template>
    <footer class="app-footer">
        <div class="footer-content">
            <div class="footer-info">
                <p>© {{ currentYear }} Mafia Wars: Criminal Empire. All rights reserved.</p>
            </div>
            <div class="footer-links">
                <a href="#" @click.prevent="openHelp">Help</a>
                <a href="#" @click.prevent="openAbout">About</a>
                <a href="#" @click.prevent="openTerms">Terms</a>
                <a href="#" @click.prevent="openPrivacy">Privacy</a>
            </div>
        </div>
    </footer>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const currentYear = new Date().getFullYear();

// Modal control functions
function openHelp() {
    // Will implement modal system later
    console.log('Open help modal');
}

function openAbout() {
    console.log('Open about modal');
}

function openTerms() {
    console.log('Open terms modal');
}

function openPrivacy() {
    console.log('Open privacy modal');
}
</script>

<style lang="scss">
.app-footer {
    background-color: $background-lighter;
    padding: $spacing-md $spacing-lg;
    border-top: 1px solid $border-color;

    .footer-content {
        @include flex-between;

        @include respond-to(md) {
            max-width: 1200px;
            margin: 0 auto;
        }

        .footer-info {
            p {
                margin: 0;
                font-size: $font-size-sm;
                color: $text-secondary;
            }
        }

        .footer-links {
            display: flex;
            gap: $spacing-lg;

            a {
                font-size: $font-size-sm;
                color: $text-secondary;
                text-decoration: none;

                &:hover {
                    color: $secondary-color;
                    text-decoration: none;
                }
            }
        }
    }
}
</style>