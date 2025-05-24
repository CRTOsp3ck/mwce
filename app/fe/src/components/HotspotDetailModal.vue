<!-- HotspotDetailModal.vue -->
<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import { Hotspot, TerritoryActionType } from '@/types/territory';

const props = defineProps<{
  hotspot: Hotspot | null;
  visible: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'open-action-modal', hotspot: Hotspot, action: TerritoryActionType): void;
}>();

// Past controllers (simulated data)
const pastControllers = ref([
  {
    id: 'pc1',
    name: 'Tony "The Shark" Moretti',
    startDate: '2025-01-15',
    endDate: '2025-02-28',
    takenBy: 'Bloody coup by the Rossi Family',
    notableEvents: 'Survived an assassination attempt by the Calabria Brothers'
  },
  {
    id: 'pc2',
    name: 'The Rossi Family',
    startDate: '2025-02-28',
    endDate: '2025-03-21',
    takenBy: 'Police raid',
    notableEvents: 'Expanded operations into counterfeiting'
  },
  {
    id: 'pc3',
    name: 'Independent - Marcel "Lucky" Dupont',
    startDate: '2025-03-30',
    endDate: '2025-04-15',
    takenBy: 'Voluntarily sold control due to heat',
    notableEvents: 'Transformed the business into a money laundering front'
  }
]);

// Lore information (random based on hotspot type)
const loreSections = computed(() => {
  if (!props.hotspot) return [];

  const sections = [];

  // Base on business type
  if (props.hotspot.businessType?.toLowerCase().includes('restaurant') ||
      props.hotspot.businessType?.toLowerCase().includes('cafe')) {
    sections.push({
      title: 'Establishment History',
      content: 'Originally opened in 1947 by an Italian immigrant family, this establishment has been a neighborhood fixture for decades. The basement contains a hidden speakeasy dating back to prohibition days, now used for private meetings between crime figures.'
    });
  } else if (props.hotspot.businessType?.toLowerCase().includes('club') ||
            props.hotspot.businessType?.toLowerCase().includes('bar')) {
    sections.push({
      title: 'Notable Clientele',
      content: 'Frequented by local politicians and corrupt cops looking for a discreet place to conduct business. The VIP room in the back has witnessed countless deals and betrayals over the years.'
    });
  } else if (props.hotspot.businessType?.toLowerCase().includes('hotel')) {
    sections.push({
      title: 'Secret Operations',
      content: 'Behind its respectable façade, select rooms on the fourth floor are reserved for high-stakes poker games. Room 407 is rumored to be soundproofed for "special interrogations."'
    });
  } else if (props.hotspot.businessType?.toLowerCase().includes('shop') ||
            props.hotspot.businessType?.toLowerCase().includes('store')) {
    sections.push({
      title: 'Front Operation',
      content: 'While appearing as a legitimate business, the stockroom contains a sophisticated distribution hub for contraband goods. The owner maintains two sets of books - one for the tax man, one for the real business.'
    });
  } else if (!props.hotspot.isLegal) {
    sections.push({
      title: 'Street Intelligence',
      content: 'This operation sits on the border between two rival territories, making it a frequent flashpoint for turf wars. Local police are paid to look the other way, but federal agents have been spotted in the area recently.'
    });
  } else {
    sections.push({
      title: 'Business Background',
      content: 'Established during the post-war economic boom, this location has changed hands numerous times. The current legitimate operation masks its importance in the criminal underworld as a key hub for information exchange.'
    });
  }

  // Add location-specific lore based on region/district
  if (props.hotspot.cityId) {
    sections.push({
      title: 'Neighborhood Dynamics',
      content: 'Located in a strategic area with access to shipping routes and major highways. Control of this operation provides significant leverage in negotiations with rival families. Local street gangs occasionally attempt to muscle in, but are kept in check by the established power structure.'
    });
  }

  // Add unique features
  sections.push({
    title: 'Unique Features',
    content: 'The property includes a secret entrance through the adjacent alley, offering discreet access during police crackdowns. A sophisticated alarm system installed by a former security expert warns of approaching law enforcement.'
  });

  return sections;
});

// For active/visible tab tracking
const activeTab = ref('overview');

// If hotspot changes, reset active tab
watch(() => props.hotspot, () => {
  activeTab.value = 'overview';
});

// Helper function to format date
function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
}

// Method to open action modal
function openActionModal(action: TerritoryActionType) {
  if (props.hotspot) {
    emit('open-action-modal', props.hotspot, action);
    emit('update:visible', false);
  }
}

// Method to close this modal
function closeModal() {
  emit('update:visible', false);
}
</script>

<template>
  <BaseModal
    :modelValue="visible"
    :title="hotspot?.name || 'Hotspot Details'"
    class="hotspot-detail-modal"
    @update:modelValue="(v) => emit('update:visible', v)"
  >
    <div v-if="hotspot" class="hotspot-detail-content">
      <!-- Header with business type and location -->
      <div class="detail-header">
        <div class="business-badge" :class="{ 'illegal': !hotspot.isLegal }">
          {{ hotspot.isLegal ? 'Legal Business' : 'Illegal Operation' }}
        </div>

        <div class="business-type">{{ hotspot.type }} - {{ hotspot.businessType }}</div>

        <div class="location-line">
          <div class="location-icon">📍</div>
          <div class="location-name">{{ hotspot.location || 'Unknown Location' }}</div>
        </div>
      </div>

      <!-- Navigation tabs -->
      <div class="detail-tabs">
        <button
          class="tab-button"
          :class="{ active: activeTab === 'overview' }"
          @click="activeTab = 'overview'"
        >
          Overview
        </button>
        <button
          class="tab-button"
          :class="{ active: activeTab === 'history' }"
          @click="activeTab = 'history'"
        >
          History
        </button>
        <button
          class="tab-button"
          :class="{ active: activeTab === 'lore' }"
          @click="activeTab = 'lore'"
        >
          Intelligence
        </button>
      </div>

      <!-- Tab content -->
      <div class="tab-content">
        <!-- Overview Tab -->
        <div v-if="activeTab === 'overview'" class="overview-tab">
          <div class="content-columns">
            <!-- Left column: Business details -->
            <div class="detail-column">
              <h4 class="column-title">Business Details</h4>

              <div class="detail-section">
                <div class="detail-item" v-if="hotspot.isLegal">
                  <div class="item-label">Income:</div>
                  <div class="item-value">${{ hotspot.income }}/hr</div>
                </div>

                <div class="detail-item">
                  <div class="item-label">Value:</div>
                  <div class="item-value">${{ (hotspot.income * 50) || 5000 }}</div>
                </div>

                <div class="detail-item" v-if="hotspot.isLegal">
                  <div class="item-label">Current Status:</div>
                  <div class="item-value status" :class="{
                    'owned': hotspot.controller === 'player-id',
                    'rival': hotspot.controller && hotspot.controller !== 'player-id',
                    'unowned': !hotspot.controller
                  }">
                    {{
                      hotspot.controller === 'player-id' ? 'Owned by You' :
                      hotspot.controller ? `Controlled by ${hotspot.controllerName || 'Rival'}` :
                      'Uncontrolled'
                    }}
                  </div>
                </div>

                <div class="detail-item" v-if="hotspot.isLegal && hotspot.controller && hotspot.controller !== 'player-id'">
                  <div class="item-label">Defense Strength:</div>
                  <div class="item-value defense" :class="{
                    'high': hotspot.defenseStrength >= 70,
                    'medium': hotspot.defenseStrength >= 30 && hotspot.defenseStrength < 70,
                    'low': hotspot.defenseStrength < 30
                  }">
                    {{
                      hotspot.defenseStrength >= 70 ? 'Heavily Defended' :
                      hotspot.defenseStrength >= 30 ? 'Moderately Defended' :
                      'Lightly Defended'
                    }}
                  </div>
                </div>

                <div class="detail-item" v-if="hotspot.pendingCollection && hotspot.pendingCollection > 0">
                  <div class="item-label">Pending Collection:</div>
                  <div class="item-value income">${{ hotspot.pendingCollection }}</div>
                </div>
              </div>

              <div class="detail-section" v-if="hotspot.isLegal && hotspot.controller === 'player-id'">
                <h5 class="section-title">Defense Allocation</h5>

                <div class="resource-allocation">
                  <div class="resource-item" v-if="hotspot.crew > 0">
                    <div class="resource-icon">👥</div>
                    <div class="resource-value">{{ hotspot.crew }} Crew</div>
                  </div>

                  <div class="resource-item" v-if="hotspot.weapons > 0">
                    <div class="resource-icon">🔫</div>
                    <div class="resource-value">{{ hotspot.weapons }} Weapons</div>
                  </div>

                  <div class="resource-item" v-if="hotspot.vehicles > 0">
                    <div class="resource-icon">🚗</div>
                    <div class="resource-value">{{ hotspot.vehicles }} Vehicles</div>
                  </div>

                  <div class="resource-item empty" v-if="!hotspot.crew && !hotspot.weapons && !hotspot.vehicles">
                    <div class="resource-icon">⚠️</div>
                    <div class="resource-value">No defenses allocated</div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Right column: Operations -->
            <div class="detail-column">
              <h4 class="column-title">Available Operations</h4>

              <div class="operations-list">
                <!-- Show operations based on business type and player status -->
                <div v-if="!hotspot.isLegal" class="operation-item">
                  <div class="operation-header">
                    <div class="operation-icon">💰</div>
                    <div class="operation-name">Extortion</div>
                  </div>

                  <div class="operation-description">
                    Force this illegal business to pay protection money.
                  </div>

                  <BaseButton
                    variant="danger"
                    small
                    class="operation-button"
                    @click="openActionModal(TerritoryActionType.EXTORTION)"
                  >
                    Extort
                  </BaseButton>
                </div>

                <div v-if="hotspot.isLegal && hotspot.controller !== 'player-id'" class="operation-item">
                  <div class="operation-header">
                    <div class="operation-icon">🏢</div>
                    <div class="operation-name">Takeover</div>
                  </div>

                  <div class="operation-description">
                    {{
                      hotspot.controller ?
                      `Take this business from ${hotspot.controllerName || 'its current owner'}.` :
                      'Take control of this unowned business.'
                    }}
                  </div>

                  <BaseButton
                    :variant="hotspot.controller ? 'danger' : 'primary'"
                    small
                    class="operation-button"
                    @click="openActionModal(TerritoryActionType.TAKEOVER)"
                  >
                    Takeover
                  </BaseButton>
                </div>

                <div v-if="hotspot.isLegal && hotspot.controller === 'player-id' && hotspot.pendingCollection && hotspot.pendingCollection > 0" class="operation-item">
                  <div class="operation-header">
                    <div class="operation-icon">💼</div>
                    <div class="operation-name">Collection</div>
                  </div>

                  <div class="operation-description">
                    Collect ${{ hotspot.pendingCollection }} in pending income.
                  </div>

                  <BaseButton
                    variant="primary"
                    small
                    class="operation-button"
                    @click="openActionModal(TerritoryActionType.COLLECTION)"
                  >
                    Collect
                  </BaseButton>
                </div>

                <div v-if="hotspot.isLegal && hotspot.controller === 'player-id'" class="operation-item">
                  <div class="operation-header">
                    <div class="operation-icon">🛡️</div>
                    <div class="operation-name">Defense</div>
                  </div>

                  <div class="operation-description">
                    Allocate resources to protect this business from takeovers.
                  </div>

                  <BaseButton
                    variant="secondary"
                    small
                    class="operation-button"
                    @click="openActionModal(TerritoryActionType.DEFEND)"
                  >
                    Defend
                  </BaseButton>
                </div>
              </div>
            </div>
          </div>

          <!-- Additional info box -->
          <div class="info-box">
            <div class="info-icon">ℹ️</div>
            <div class="info-content">
              <div class="info-title">Strategic Value</div>
              <div class="info-text">
                {{ hotspot.isLegal ?
                  'Controlling this business provides steady income and increases your influence in the area.' :
                  'This illegal operation can be extorted for quick cash, but generates heat with law enforcement.'
                }}
              </div>
            </div>
          </div>
        </div>

        <!-- History Tab -->
        <div v-if="activeTab === 'history'" class="history-tab">
          <div class="timeline">
            <!-- Current controller -->
            <div class="timeline-item current">
              <div class="timeline-marker"></div>

              <div class="timeline-content">
                <div class="timeline-date">Current Controller</div>

                <div class="controller-card">
                  <div class="controller-icon">👑</div>

                  <div class="controller-info">
                    <div class="controller-name">
                      {{
                        hotspot.controller === 'player-id' ? 'Your Organization' :
                        hotspot.controller ? hotspot.controllerName || 'Unknown Organization' :
                        'Uncontrolled Territory'
                      }}
                    </div>

                    <div class="control-period">
                      Since {{ formatDate(new Date().toISOString()) }}
                    </div>

                    <div class="control-details" v-if="hotspot.controller === 'player-id'">
                      This business is currently under your control.
                    </div>
                    <div class="control-details" v-else-if="hotspot.controller">
                      This business is controlled by a rival organization.
                    </div>
                    <div class="control-details" v-else>
                      This business is currently unclaimed and ripe for takeover.
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Past controllers -->
            <div v-for="(controller, index) in pastControllers" :key="controller.id" class="timeline-item past">
              <div class="timeline-marker"></div>

              <div class="timeline-content">
                <div class="timeline-date">
                  {{ formatDate(controller.startDate) }} - {{ formatDate(controller.endDate) }}
                </div>

                <div class="controller-card">
                  <div class="controller-info">
                    <div class="controller-name">{{ controller.name }}</div>

                    <div class="control-takeover" v-if="controller.takenBy">
                      <span class="takeover-label">Lost control via:</span>
                      <span class="takeover-value">{{ controller.takenBy }}</span>
                    </div>

                    <div class="control-events" v-if="controller.notableEvents">
                      <span class="events-label">Notable events:</span>
                      <span class="events-value">{{ controller.notableEvents }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Origin marker -->
            <div class="timeline-item origin">
              <div class="timeline-marker"></div>

              <div class="timeline-content">
                <div class="timeline-date">Establishment</div>

                <div class="controller-card origin">
                  <div class="origin-text">
                    This establishment was founded in the 1940s.
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Lore/Intelligence Tab -->
        <div v-if="activeTab === 'lore'" class="lore-tab">
          <div class="intel-header">
            <div class="intel-stamp">CONFIDENTIAL</div>
            <div class="intel-title">Intelligence Report</div>
            <div class="intel-date">{{ formatDate(new Date().toISOString()) }}</div>
          </div>

          <div class="lore-sections">
            <div v-for="(section, index) in loreSections" :key="index" class="lore-section">
              <div class="section-header">
                <div class="section-icon">📋</div>
                <div class="section-title">{{ section.title }}</div>
              </div>

              <div class="section-content">
                {{ section.content }}
              </div>
            </div>

            <!-- Local influences section -->
            <div class="lore-section">
              <div class="section-header">
                <div class="section-icon">🌆</div>
                <div class="section-title">Local Influences</div>
              </div>

              <div class="section-content">
                <div class="influence-group">
                  <div class="influence-title">Law Enforcement</div>
                  <div class="influence-bar">
                    <div class="bar-fill" :style="{ width: `${35 + Math.floor(Math.random() * 40)}%` }"></div>
                  </div>
                  <div class="influence-label">{{ ['Minimal', 'Moderate', 'Heavy', 'Extreme'][Math.floor(Math.random() * 4)] }}</div>
                </div>

                <div class="influence-group">
                  <div class="influence-title">Rival Gangs</div>
                  <div class="influence-bar">
                    <div class="bar-fill" :style="{ width: `${35 + Math.floor(Math.random() * 40)}%` }"></div>
                  </div>
                  <div class="influence-label">{{ ['Minimal', 'Moderate', 'Heavy', 'Extreme'][Math.floor(Math.random() * 4)] }}</div>
                </div>

                <div class="influence-group">
                  <div class="influence-title">Civilian Support</div>
                  <div class="influence-bar">
                    <div class="bar-fill" :style="{ width: `${35 + Math.floor(Math.random() * 40)}%` }"></div>
                  </div>
                  <div class="influence-label">{{ ['Low', 'Moderate', 'High', 'Very High'][Math.floor(Math.random() * 4)] }}</div>
                </div>
              </div>
            </div>

            <!-- Informant notes -->
            <div class="lore-section informant">
              <div class="section-header">
                <div class="section-icon">👤</div>
                <div class="section-title">Informant Notes</div>
              </div>

              <div class="section-content informant-notes">
                <div class="note">
                  <div class="note-text">
                    "Word on the street is that the cops are planning a major sweep through this area next week. Could be trouble for anyone caught with their hands dirty."
                  </div>
                  <div class="note-author">- Street Informant</div>
                </div>

                <div class="note">
                  <div class="note-text">
                    "The previous owner had a hidden safe somewhere on the premises. Rumor says it contained evidence that could bring down several high-ranking officials."
                  </div>
                  <div class="note-author">- Former Employee</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <BaseButton variant="primary" @click="closeModal">
        Close
      </BaseButton>
    </template>
  </BaseModal>
</template>

<style lang="scss" scoped>
.hotspot-detail-modal {
  // Modal sizing
  @media (min-width: 768px) {
    :deep(.modal-content) {
      width: 800px;
      max-width: 90vw;
    }
  }

  .hotspot-detail-content {
    @include flex-column;
    gap: $spacing-md;

    // Header styling
    .detail-header {
      @include flex-column;
      align-items: center;
      text-align: center;
      padding-bottom: $spacing-md;
      margin-bottom: $spacing-md;
      border-bottom: 1px solid $border-color;

      .business-badge {
        display: inline-block;
        background-color: $primary-color;
        color: $text-color;
        font-size: $font-size-sm;
        font-weight: 600;
        padding: 3px 12px;
        border-radius: $border-radius-md;
        margin-bottom: $spacing-sm;

        &.illegal {
          background-color: $danger-color;
        }
      }

      .business-type {
        font-size: $font-size-md;
        color: $text-secondary;
        margin-bottom: $spacing-xs;
      }

      .location-line {
        display: flex;
        align-items: center;
        gap: $spacing-xs;
        font-size: $font-size-sm;

        .location-icon {
          color: $text-secondary;
        }

        .location-name {
          font-weight: 500;
        }
      }
    }

    // Tab navigation
    .detail-tabs {
      display: flex;
      background-color: rgba($background-darker, 0.5);
      border-radius: $border-radius-md;
      overflow: hidden;
      margin-bottom: $spacing-md;

      .tab-button {
        flex: 1;
        background: none;
        border: none;
        color: $text-secondary;
        padding: $spacing-sm;
        font-size: $font-size-md;
        cursor: pointer;
        transition: $transition-base;

        &:hover {
          background-color: rgba($background-lighter, 0.1);
        }

        &.active {
          background-color: $background-lighter;
          color: $gold-color;
          font-weight: 500;
        }
      }
    }

    // Tab content
    .tab-content {
      min-height: 300px;

      // Overview tab
      .overview-tab {
        @include flex-column;
        gap: $spacing-md;

        .content-columns {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: $spacing-md;

          @include respond-to(sm-down) {
            grid-template-columns: 1fr;
          }

          .detail-column {
            @include flex-column;
            gap: $spacing-md;

            .column-title {
              margin: 0 0 $spacing-sm 0;
              padding-bottom: $spacing-xs;
              border-bottom: 1px solid rgba($border-color, 0.5);
              color: $gold-color;
            }

            .detail-section {
              @include flex-column;
              gap: $spacing-sm;

              .section-title {
                margin: 0 0 $spacing-sm 0;
                font-size: $font-size-md;
                color: $text-secondary;
              }

              .detail-item {
                display: flex;
                justify-content: space-between;
                align-items: center;

                .item-label {
                  color: $text-secondary;
                  font-size: $font-size-sm;
                }

                .item-value {
                  font-weight: 500;

                  &.status {
                    &.owned {
                      color: $success-color;
                    }

                    &.rival {
                      color: $danger-color;
                    }

                    &.unowned {
                      color: $text-secondary;
                    }
                  }

                  &.defense {
                    &.high {
                      color: $danger-color;
                    }

                    &.medium {
                      color: $warning-color;
                    }

                    &.low {
                      color: $success-color;
                    }
                  }

                  &.income {
                    color: $secondary-color;
                  }
                }
              }

              .resource-allocation {
                display: flex;
                flex-wrap: wrap;
                gap: $spacing-sm;

                .resource-item {
                  background-color: rgba($background-lighter, 0.2);
                  border-radius: $border-radius-sm;
                  padding: $spacing-xs $spacing-sm;
                  display: flex;
                  align-items: center;
                  gap: $spacing-xs;

                  .resource-icon {
                    font-size: $font-size-md;
                  }

                  .resource-value {
                    font-size: $font-size-sm;
                    font-weight: 500;
                  }

                  &.empty {
                    background-color: rgba($danger-color, 0.1);
                    color: $danger-color;
                  }
                }
              }
            }

            .operations-list {
              @include flex-column;
              gap: $spacing-md;

              .operation-item {
                background-color: rgba($background-lighter, 0.1);
                border-radius: $border-radius-md;
                padding: $spacing-md;

                .operation-header {
                  display: flex;
                  align-items: center;
                  gap: $spacing-sm;
                  margin-bottom: $spacing-xs;

                  .operation-icon {
                    font-size: 20px;
                  }

                  .operation-name {
                    font-weight: 600;
                  }
                }

                .operation-description {
                  font-size: $font-size-sm;
                  color: $text-secondary;
                  margin-bottom: $spacing-sm;
                }

                .operation-button {
                  align-self: flex-start;
                }
              }
            }
          }
        }

        .info-box {
          display: flex;
          gap: $spacing-md;
          padding: $spacing-md;
          background-color: rgba($secondary-color, 0.1);
          border-radius: $border-radius-md;

          .info-icon {
            font-size: 24px;
          }

          .info-content {
            .info-title {
              font-weight: 600;
              margin-bottom: $spacing-xs;
            }

            .info-text {
              font-size: $font-size-sm;
              color: $text-secondary;
            }
          }
        }
      }

      // History tab
      .history-tab {
        .timeline {
          @include flex-column;
          gap: 0;
          position: relative;

          &::before {
            content: "";
            position: absolute;
            top: 0;
            bottom: 0;
            left: 16px;
            width: 2px;
            background: linear-gradient(to bottom,
              $gold-color,
              rgba($text-secondary, 0.5),
              rgba($text-secondary, 0.2));
          }

          .timeline-item {
            position: relative;
            padding-left: 45px;
            padding-bottom: $spacing-xl;

            .timeline-marker {
              position: absolute;
              left: 10px;
              top: 0;
              width: 14px;
              height: 14px;
              border-radius: 50%;
              border: 2px solid $background-dark;
            }

            &.current .timeline-marker {
              background-color: $gold-color;
              box-shadow: 0 0 10px rgba($gold-color, 0.7);
            }

            &.past .timeline-marker {
              background-color: $secondary-color;
            }

            &.origin .timeline-marker {
              background-color: $text-secondary;
            }

            .timeline-content {
              .timeline-date {
                font-size: $font-size-sm;
                color: $text-secondary;
                margin-bottom: $spacing-sm;
              }

              .controller-card {
                background-color: rgba($background-lighter, 0.1);
                border-radius: $border-radius-md;
                padding: $spacing-md;
                display: flex;
                gap: $spacing-md;

                .controller-icon {
                  font-size: 24px;
                  color: $gold-color;
                }

                .controller-info {
                  @include flex-column;
                  gap: $spacing-xs;

                  .controller-name {
                    font-weight: 600;
                    font-size: $font-size-md;
                  }

                  .control-period {
                    font-size: $font-size-sm;
                    color: $text-secondary;
                  }

                  .control-details,
                  .control-takeover,
                  .control-events {
                    font-size: $font-size-sm;

                    .takeover-label,
                    .events-label {
                      color: $text-secondary;
                      margin-right: $spacing-xs;
                    }
                  }
                }

                &.origin {
                  border: 1px dashed rgba($text-secondary, 0.3);
                  background-color: transparent;

                  .origin-text {
                    font-style: italic;
                    color: $text-secondary;
                  }
                }
              }
            }
          }
        }
      }

      // Lore/Intelligence tab
      .lore-tab {
        @include flex-column;
        gap: $spacing-md;

        .intel-header {
          text-align: center;
          border-bottom: 1px dashed rgba($danger-color, 0.5);
          padding-bottom: $spacing-md;
          margin-bottom: $spacing-md;

          .intel-stamp {
            display: inline-block;
            color: $danger-color;
            border: 2px solid $danger-color;
            padding: 2px 10px;
            font-weight: 700;
            transform: rotate(-5deg);
            font-size: $font-size-sm;
            margin-bottom: $spacing-xs;
          }

          .intel-title {
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 1px;
            font-size: $font-size-lg;
            margin-bottom: $spacing-xs;
          }

          .intel-date {
            font-size: $font-size-sm;
            color: $text-secondary;
          }
        }

        .lore-sections {
          @include flex-column;
          gap: $spacing-lg;

          .lore-section {
            .section-header {
              display: flex;
              align-items: center;
              gap: $spacing-sm;
              margin-bottom: $spacing-sm;

              .section-icon {
                font-size: 20px;
              }

              .section-title {
                font-weight: 600;
                font-size: $font-size-md;
                color: $gold-color;
              }
            }

            .section-content {
              font-size: $font-size-sm;
              line-height: 1.6;
              color: $text-secondary;
              padding-left: 30px;

              .influence-group {
                display: flex;
                align-items: center;
                margin-bottom: $spacing-sm;

                .influence-title {
                  width: 120px;
                  font-weight: 500;
                  color: $text-color;
                }

                .influence-bar {
                  flex: 1;
                  height: 6px;
                  background-color: rgba($background-lighter, 0.2);
                  border-radius: 3px;
                  overflow: hidden;
                  margin: 0 $spacing-sm;

                  .bar-fill {
                    height: 100%;
                    background-color: $primary-color;
                  }
                }

                .influence-label {
                  width: 80px;
                  text-align: right;
                  font-size: $font-size-xs;
                }
              }
            }

            &.informant {
              .informant-notes {
                .note {
                  background-color: rgba($background-lighter, 0.1);
                  border-radius: $border-radius-md;
                  padding: $spacing-sm;
                  margin-bottom: $spacing-sm;

                  .note-text {
                    font-style: italic;
                    margin-bottom: $spacing-xs;
                    color: $text-color;
                  }

                  .note-author {
                    text-align: right;
                    font-size: $font-size-xs;
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
</style>
