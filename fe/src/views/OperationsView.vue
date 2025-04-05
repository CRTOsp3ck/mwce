// src/views/OperationsView.vue

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useOperationsStore } from '@/stores/modules/operations';
import { 
  Operation, 
  OperationType, 
  OperationStatus, 
  OperationAttempt,
  OperationResources
} from '@/types/operations';
import { PlayerTitle } from '@/types/player';

const route = useRoute();
const router = useRouter();
const playerStore = usePlayerStore();
const operationsStore = useOperationsStore();

// View state
// const activeTab = ref<'available' | 'in-progress' | 'completed'>('available');
const activeTab = computed(()=> route.query.tab as string || 'available')
const typeFilter = ref<'all' | 'basic' | 'special'>('all');
const searchQuery = ref('');
const isLoading = ref(false);

// Modals
const showStartModal = ref(false);
const showCancelModal = ref(false);
const selectedOperation = ref<Operation | null>(null);
const selectedOperationAttempt = ref<OperationAttempt | null>(null);
const isStartingOperation = ref(false);
const isCancellingOperation = ref(false);

// Timer for checking operation status
let statusCheckTimer: number | null = null;

// Computed properties
const playerMoney = computed(() => playerStore.playerMoney);
const playerCrew = computed(() => playerStore.playerCrew);
const playerWeapons = computed(() => playerStore.playerWeapons);
const playerVehicles = computed(() => playerStore.playerVehicles);
const playerInfluence = computed(() => playerStore.playerInfluence);
const playerHeat = computed(() => playerStore.playerHeat);
const playerTitle = computed(() => playerStore.playerTitle);

const availableOperations = computed(() => operationsStore.availableOperations);
const inProgressOperations = computed(() => operationsStore.inProgressOperations);
const completedOperations = computed(() => operationsStore.completedOperations);

const filteredOperations = computed(() => {
  let result = [...availableOperations.value];
  
  // Apply type filter
  if (typeFilter.value === 'basic') {
    result = result.filter(op => !op.isSpecial);
  } else if (typeFilter.value === 'special') {
    result = result.filter(op => op.isSpecial);
  }
  
  // Apply search filter
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(op => 
      op.name.toLowerCase().includes(query) || 
      op.description.toLowerCase().includes(query) ||
      formatOperationType(op.type).toLowerCase().includes(query)
    );
  }
  
  return result;
});

const canStartSelectedOperation = computed(() => {
  if (!selectedOperation.value) return false;
  
  return canStartOperation(selectedOperation.value);
});

const confirmationWarning = computed(() => {
  if (!selectedOperation.value) return '';
  
  if (selectedOperation.value.resources.crew > playerCrew.value) {
    return 'You don\'t have enough crew members for this operation.';
  }
  
  if (selectedOperation.value.resources.weapons > playerWeapons.value) {
    return 'You don\'t have enough weapons for this operation.';
  }
  
  if (selectedOperation.value.resources.vehicles > playerVehicles.value) {
    return 'You don\'t have enough vehicles for this operation.';
  }
  
  if (selectedOperation.value.resources.money && selectedOperation.value.resources.money > playerMoney.value) {
    return 'You don\'t have enough money for this operation.';
  }
  
  if (selectedOperation.value.requirements.minInfluence && selectedOperation.value.requirements.minInfluence > playerInfluence.value) {
    return 'Your influence is too low for this operation.';
  }
  
  if (selectedOperation.value.requirements.maxHeat && selectedOperation.value.requirements.maxHeat < playerHeat.value) {
    return 'Your heat level is too high for this operation.';
  }
  
  if (selectedOperation.value.requirements.minTitle && !meetsMinimumTitle(selectedOperation.value.requirements.minTitle)) {
    return `You need to be at least a ${selectedOperation.value.requirements.minTitle} to start this operation.`;
  }
  
  return '';
});

// Load data when component is mounted
onMounted(async () => {
  isLoading.value = true;
  
  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }
  
  if (operationsStore.availableOperations.length === 0) {
    await operationsStore.fetchAvailableOperations();
  }
  
  await operationsStore.fetchPlayerOperations();
  
  isLoading.value = false;
  
  // Check if there's an operation ID in the query params
  const operationId = route.query.operation as string;
  if (operationId) {
    const operation = availableOperations.value.find(op => op.id === operationId);
    if (operation) {
      startOperation(operation);
    }
  }
  
  // Set up timer to check operation status
  statusCheckTimer = window.setInterval(() => {
    operationsStore.checkOperationStatus();
  }, 10000); // Check every 10 seconds
});

// Clean up timer on component unmount
onBeforeUnmount(() => {
  if (statusCheckTimer) {
    clearInterval(statusCheckTimer);
  }
});

// Helper functions
function formatNumber(value: number): string {
  if (value >= 1000000) {
    return (value / 1000000).toFixed(1) + 'M';
  } else if (value >= 1000) {
    return (value / 1000).toFixed(1) + 'K';
  }
  return value.toString();
}

function formatOperationType(type: OperationType): string {
  switch (type) {
    case OperationType.CARJACKING:
      return 'Carjacking';
    case OperationType.GOODS_SMUGGLING:
      return 'Goods Smuggling';
    case OperationType.DRUG_TRAFFICKING:
      return 'Drug Trafficking';
    case OperationType.OFFICIAL_BRIBING:
      return 'Official Bribing';
    case OperationType.INTELLIGENCE_GATHERING:
      return 'Intelligence Gathering';
    case OperationType.CREW_RECRUITMENT:
      return 'Crew Recruitment';
    default:
      return type;
  }
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
}

function formatDate(dateString: string | undefined): string {
  if (!dateString) return 'Unknown';
  
  const date = new Date(dateString);
  const now = new Date();
  const diff = Math.floor((now.getTime() - date.getTime()) / 60000); // difference in minutes
  
  if (diff < 1) return 'Just now';
  if (diff < 60) return `${diff} min ago`;
  if (diff < 1440) return `${Math.floor(diff / 60)} hours ago`;
  
  return date.toLocaleDateString();
}

function hasRequirements(operation: Operation): boolean {
  return !!(
    operation.requirements.minInfluence ||
    operation.requirements.maxHeat ||
    operation.requirements.minTitle
  );
}

function meetsMinimumTitle(requiredTitle: string): boolean {
  const titleRanks = {
    [PlayerTitle.ASSOCIATE]: 1,
    [PlayerTitle.SOLDIER]: 2,
    [PlayerTitle.CAPO]: 3,
    [PlayerTitle.UNDERBOSS]: 4,
    [PlayerTitle.CONSIGLIERE]: 5,
    [PlayerTitle.BOSS]: 6,
    [PlayerTitle.GODFATHER]: 7
  };
  
  const requiredRank = titleRanks[requiredTitle as PlayerTitle] || 0;
  const playerRank = titleRanks[playerTitle.value] || 0;
  
  return playerRank >= requiredRank;
}

function canStartOperation(operation: Operation): boolean {
  // Check if player has enough resources
  if (operation.resources.crew > playerCrew.value) return false;
  if (operation.resources.weapons > playerWeapons.value) return false;
  if (operation.resources.vehicles > playerVehicles.value) return false;
  if (operation.resources.money && operation.resources.money > playerMoney.value) return false;
  
  // Check if player meets requirements
  if (operation.requirements.minInfluence && operation.requirements.minInfluence > playerInfluence.value) return false;
  if (operation.requirements.maxHeat && operation.requirements.maxHeat < playerHeat.value) return false;
  if (operation.requirements.minTitle && !meetsMinimumTitle(operation.requirements.minTitle)) return false;
  
  return true;
}

function getOperationWarning(operation: Operation): string {
  if (operation.resources.crew > playerCrew.value) {
    return 'Not enough crew';
  }
  
  if (operation.resources.weapons > playerWeapons.value) {
    return 'Not enough weapons';
  }
  
  if (operation.resources.vehicles > playerVehicles.value) {
    return 'Not enough vehicles';
  }
  
  if (operation.resources.money && operation.resources.money > playerMoney.value) {
    return 'Not enough money';
  }
  
  if (operation.requirements.minInfluence && operation.requirements.minInfluence > playerInfluence.value) {
    return `Requires ${operation.requirements.minInfluence} influence`;
  }
  
  if (operation.requirements.maxHeat && operation.requirements.maxHeat < playerHeat.value) {
    return `Heat too high (max ${operation.requirements.maxHeat})`;
  }
  
  if (operation.requirements.minTitle && !meetsMinimumTitle(operation.requirements.minTitle)) {
    return `Requires ${operation.requirements.minTitle} rank`;
  }
  
  return '';
}

function getOperationName(operationAttempt: OperationAttempt): string {
  const operation = availableOperations.value.find(op => op.id === operationAttempt.operationId);
  return operation ? operation.name : 'Unknown Operation';
}

function getOperationType(operationAttempt: OperationAttempt): string {
  const operation = availableOperations.value.find(op => op.id === operationAttempt.operationId);
  return operation ? formatOperationType(operation.type) : 'Unknown Type';
}

function getTimeRemaining(operationAttempt: OperationAttempt): string {
  if (!operationAttempt || operationAttempt.status !== OperationStatus.IN_PROGRESS) {
    return 'Completed';
  }
  
  const operation = availableOperations.value.find(op => op.id === operationAttempt.operationId);
  if (!operation) return 'Unknown';
  
  const startTime = new Date(operationAttempt.timestamp);
  const endTime = new Date(startTime.getTime() + (operation.duration * 1000));
  const now = new Date();
  
  if (now >= endTime) {
    return 'Ready to collect';
  }
  
  const remainingSeconds = Math.floor((endTime.getTime() - now.getTime()) / 1000);
  return formatDuration(remainingSeconds);
}

function getEstimatedCompletion(operationAttempt: OperationAttempt): string {
  if (!operationAttempt || operationAttempt.status !== OperationStatus.IN_PROGRESS) {
    return 'Completed';
  }
  
  const operation = availableOperations.value.find(op => op.id === operationAttempt.operationId);
  if (!operation) return 'Unknown';
  
  const startTime = new Date(operationAttempt.timestamp);
  const endTime = new Date(startTime.getTime() + (operation.duration * 1000));
  
  const hours = endTime.getHours().toString().padStart(2, '0');
  const minutes = endTime.getMinutes().toString().padStart(2, '0');
  
  return `${hours}:${minutes}`;
}

function getProgressPercentage(operationAttempt: OperationAttempt): number {
  if (!operationAttempt || operationAttempt.status !== OperationStatus.IN_PROGRESS) {
    return 100;
  }
  
  const operation = availableOperations.value.find(op => op.id === operationAttempt.operationId);
  if (!operation) return 0;
  
  const startTime = new Date(operationAttempt.timestamp);
  const endTime = new Date(startTime.getTime() + (operation.duration * 1000));
  const now = new Date();
  
  if (now >= endTime) {
    return 100;
  }
  
  const totalDuration = operation.duration * 1000;
  const elapsed = now.getTime() - startTime.getTime();
  
  return Math.floor((elapsed / totalDuration) * 100);
}

function isSuccessfulOperation(operation: OperationAttempt): boolean {
  return operation.status === OperationStatus.COMPLETED && operation.result && operation.result.success;
}

function isFailedOperation(operation: OperationAttempt): boolean {
  return operation.status === OperationStatus.FAILED || (operation.result && !operation.result.success);
}

// Action functions
function resetFilters() {
  typeFilter.value = 'all';
  searchQuery.value = '';
}

function startOperation(operation: Operation) {
  selectedOperation.value = operation;
  showStartModal.value = true;
}

function closeStartModal() {
  showStartModal.value = false;
  selectedOperation.value = null;
}

async function confirmStartOperation() {
  if (!selectedOperation.value || isStartingOperation.value) return;
  
  isStartingOperation.value = true;
  
  try {
    await operationsStore.startOperation(
      selectedOperation.value.id, 
      selectedOperation.value.resources
    );
    
    // Switch to in-progress tab
    activeTab.value = 'in-progress';
    
    // Close modal
    showStartModal.value = false;
    selectedOperation.value = null;
  } catch (error) {
    console.error('Error starting operation:', error);
  } finally {
    isStartingOperation.value = false;
  }
}

function cancelOperation(operation: OperationAttempt) {
  selectedOperationAttempt.value = operation;
  showCancelModal.value = true;
}

function closeCancelModal() {
  showCancelModal.value = false;
  selectedOperationAttempt.value = null;
}

async function confirmCancelOperation() {
  if (!selectedOperationAttempt.value || isCancellingOperation.value) return;
  
  isCancellingOperation.value = true;
  
  try {
    await operationsStore.cancelOperation(selectedOperationAttempt.value.id);
    
    // Close modal
    showCancelModal.value = false;
    selectedOperationAttempt.value = null;
  } catch (error) {
    console.error('Error cancelling operation:', error);
  } finally {
    isCancellingOperation.value = false;
  }
}

function navigateToTab(tab:'available' | 'in-progress' | 'completed'){
  router.push({ path:'/operations', query: { tab }})
}
</script>

<template>
    <div class="operations-view">
      <div class="page-title">
        <h2>Operations</h2>
        <p class="subtitle">Complete operations to gain resources, respect, and influence across the city.</p>
      </div>
      
      <div class="operations-tabs">
        <button 
          class="tab-button" 
          :class="{ active: activeTab === 'available' }"
          @click="navigateToTab('available')"
        >
          Available Operations
        </button>
        <button 
          class="tab-button" 
          :class="{ active: activeTab === 'in-progress' }"
          @click="navigateToTab('in-progress')"
        >
          In Progress ({{ inProgressOperations.length }})
        </button>
        <button 
          class="tab-button" 
          :class="{ active: activeTab === 'completed' }"
          @click="navigateToTab('completed')"
        >
          Completed
        </button>
      </div>
      
      <div class="operations-filters" v-if="activeTab === 'available'">
        <div class="filter">
          <label>Type:</label>
          <select v-model="typeFilter">
            <option value="all">All Types</option>
            <option value="basic">Basic Operations</option>
            <option value="special">Special Operations</option>
          </select>
        </div>
        
        <div class="search">
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="Search operations..." 
          />
        </div>
      </div>
      
      <!-- Available Operations Tab -->
      <div v-if="activeTab === 'available'" class="operations-list available-operations">
        <BaseCard 
          v-for="operation in filteredOperations" 
          :key="operation.id" 
          class="operation-card"
          :class="{ 'special-operation': operation.isSpecial }"
        >
          <template #header>
            <div class="operation-badge" v-if="operation.isSpecial">
              Special
            </div>
          </template>
          
          <div class="operation-header">
            <h3>{{ operation.name }}</h3>
            <div class="operation-type">{{ formatOperationType(operation.type) }}</div>
          </div>
          
          <div class="operation-details">
            <p class="description">{{ operation.description }}</p>
            
            <div class="requirements" v-if="hasRequirements(operation)">
              <h4>Requirements</h4>
              <ul class="requirements-list">
                <li v-if="operation.requirements.minInfluence">
                Minimum Influence: {{ operation.requirements.minInfluence }}
                <span 
                  class="requirement-status" 
                  :class="{ met: playerInfluence >= operation.requirements.minInfluence }"
                >
                  {{ playerInfluence >= operation.requirements.minInfluence ? '✓' : '✗' }}
                </span>
              </li>
              <li v-if="operation.requirements.maxHeat">
                Maximum Heat: {{ operation.requirements.maxHeat }}
                <span 
                  class="requirement-status" 
                  :class="{ met: playerHeat <= operation.requirements.maxHeat }"
                >
                  {{ playerHeat <= operation.requirements.maxHeat ? '✓' : '✗' }}
                </span>
              </li>
              <li v-if="operation.requirements.minTitle">
                Minimum Title: {{ operation.requirements.minTitle }}
                <span 
                  class="requirement-status" 
                  :class="{ met: meetsMinimumTitle(operation.requirements.minTitle) }"
                >
                  {{ meetsMinimumTitle(operation.requirements.minTitle) ? '✓' : '✗' }}
                </span>
              </li>
            </ul>
          </div>
          
          <div class="operation-resources">
            <h4>Required Resources</h4>
            <div class="resources-grid">
              <div class="resource" v-if="operation.resources.crew > 0">
                <div class="resource-icon">👥</div>
                <div class="resource-details">
                  <div class="resource-name">Crew</div>
                  <div class="resource-value">
                    {{ operation.resources.crew }}
                    <span 
                      class="resource-status" 
                      :class="{ shortage: playerCrew < operation.resources.crew }"
                    >
                      ({{ playerCrew }} available)
                    </span>
                  </div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.weapons > 0">
                <div class="resource-icon">🔫</div>
                <div class="resource-details">
                  <div class="resource-name">Weapons</div>
                  <div class="resource-value">
                    {{ operation.resources.weapons }}
                    <span 
                      class="resource-status" 
                      :class="{ shortage: playerWeapons < operation.resources.weapons }"
                    >
                      ({{ playerWeapons }} available)
                    </span>
                  </div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.vehicles > 0">
                <div class="resource-icon">🚗</div>
                <div class="resource-details">
                  <div class="resource-name">Vehicles</div>
                  <div class="resource-value">
                    {{ operation.resources.vehicles }}
                    <span 
                      class="resource-status" 
                      :class="{ shortage: playerVehicles < operation.resources.vehicles }"
                    >
                      ({{ playerVehicles }} available)
                    </span>
                  </div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.money">
                <div class="resource-icon">💵</div>
                <div class="resource-details">
                  <div class="resource-name">Money</div>
                  <div class="resource-value">
                    ${{ formatNumber(operation.resources.money) }}
                    <span 
                      class="resource-status" 
                      :class="{ shortage: playerMoney < operation.resources.money }"
                    >
                      (${{ formatNumber(playerMoney) }} available)
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <div class="operation-details-grid">
            <div class="detail-column">
              <h4>Rewards</h4>
              <ul class="rewards-list">
                <li v-if="operation.rewards.money">
                  <span class="reward-icon">💰</span>
                  <span class="reward-text">${{ formatNumber(operation.rewards.money) }}</span>
                </li>
                <li v-if="operation.rewards.crew">
                  <span class="reward-icon">👥</span>
                  <span class="reward-text">{{ operation.rewards.crew }} Crew Members</span>
                </li>
                <li v-if="operation.rewards.weapons">
                  <span class="reward-icon">🔫</span>
                  <span class="reward-text">{{ operation.rewards.weapons }} Weapons</span>
                </li>
                <li v-if="operation.rewards.vehicles">
                  <span class="reward-icon">🚗</span>
                  <span class="reward-text">{{ operation.rewards.vehicles }} Vehicles</span>
                </li>
                <li v-if="operation.rewards.respect">
                  <span class="reward-icon">👊</span>
                  <span class="reward-text">{{ operation.rewards.respect }} Respect</span>
                </li>
                <li v-if="operation.rewards.influence">
                  <span class="reward-icon">🏛️</span>
                  <span class="reward-text">{{ operation.rewards.influence }} Influence</span>
                </li>
                <li v-if="operation.rewards.heatReduction">
                  <span class="reward-icon">❄️</span>
                  <span class="reward-text">{{ operation.rewards.heatReduction }} Heat Reduction</span>
                </li>
              </ul>
            </div>
            
            <div class="detail-column">
              <h4>Risks</h4>
              <ul class="risks-list">
                <li v-if="operation.risks.crewLoss">
                  <span class="risk-icon">👥</span>
                  <span class="risk-text">Up to {{ operation.risks.crewLoss }} Crew Loss</span>
                </li>
                <li v-if="operation.risks.weaponsLoss">
                  <span class="risk-icon">🔫</span>
                  <span class="risk-text">Up to {{ operation.risks.weaponsLoss }} Weapons Loss</span>
                </li>
                <li v-if="operation.risks.vehiclesLoss">
                  <span class="risk-icon">🚗</span>
                  <span class="risk-text">Up to {{ operation.risks.vehiclesLoss }} Vehicles Loss</span>
                </li>
                <li v-if="operation.risks.moneyLoss">
                  <span class="risk-icon">💰</span>
                  <span class="risk-text">Up to ${{ formatNumber(operation.risks.moneyLoss) }} Loss</span>
                </li>
                <li v-if="operation.risks.heatIncrease">
                  <span class="risk-icon">🔥</span>
                  <span class="risk-text">{{ operation.risks.heatIncrease }} Heat Increase</span>
                </li>
                <li v-if="operation.risks.respectLoss">
                  <span class="risk-icon">👊</span>
                  <span class="risk-text">{{ operation.risks.respectLoss }} Respect Loss</span>
                </li>
              </ul>
            </div>
          </div>
          
          <div class="operation-stats">
            <div class="stat">
              <div class="stat-label">Success Rate:</div>
              <div class="stat-value">{{ operation.successRate }}%</div>
            </div>
            <div class="stat">
              <div class="stat-label">Duration:</div>
              <div class="stat-value">{{ formatDuration(operation.duration) }}</div>
            </div>
            <div class="stat">
              <div class="stat-label">Available Until:</div>
              <div class="stat-value">{{ formatDate(operation.availableUntil) }}</div>
            </div>
          </div>
        </div>
        
        <template #footer>
          <div class="operation-footer">
            <BaseButton
              :disabled="!canStartOperation(operation)"
              :variant="operation.isSpecial ? 'primary' : 'secondary'"
              @click="startOperation(operation)"
            >
              Start Operation
            </BaseButton>
            <div class="operation-warning" v-if="getOperationWarning(operation)">
              {{ getOperationWarning(operation) }}
            </div>
          </div>
        </template>
      </BaseCard>
      
      <div v-if="filteredOperations.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <p>No operations found matching your criteria.</p>
        <BaseButton @click="resetFilters" variant="outline">Reset Filters</BaseButton>
      </div>
    </div>
    
    <!-- In Progress Operations Tab -->
    <div v-else-if="activeTab === 'in-progress'" class="operations-list in-progress-operations">
      <BaseCard 
        v-for="operation in inProgressOperations" 
        :key="operation.id" 
        class="operation-card in-progress"
      >
        <div class="operation-header">
          <h3>{{ getOperationName(operation) }}</h3>
          <div class="operation-type">{{ getOperationType(operation) }}</div>
        </div>
        
        <div class="operation-details">
          <div class="progress-tracker">
            <div class="progress-info">
              <div class="time-remaining">
                {{ getTimeRemaining(operation) }} remaining
              </div>
              <div class="completion-time">
                Estimated completion: {{ getEstimatedCompletion(operation) }}
              </div>
            </div>
            <div class="progress-bar">
              <div 
                class="progress-fill" 
                :style="{ width: `${getProgressPercentage(operation)}%` }"
              ></div>
            </div>
          </div>
          
          <div class="resources-committed">
            <h4>Resources Committed</h4>
            <div class="resources-grid">
              <div class="resource" v-if="operation.resources.crew > 0">
                <div class="resource-icon">👥</div>
                <div class="resource-details">
                  <div class="resource-name">Crew</div>
                  <div class="resource-value">{{ operation.resources.crew }}</div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.weapons > 0">
                <div class="resource-icon">🔫</div>
                <div class="resource-details">
                  <div class="resource-name">Weapons</div>
                  <div class="resource-value">{{ operation.resources.weapons }}</div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.vehicles > 0">
                <div class="resource-icon">🚗</div>
                <div class="resource-details">
                  <div class="resource-name">Vehicles</div>
                  <div class="resource-value">{{ operation.resources.vehicles }}</div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.money">
                <div class="resource-icon">💵</div>
                <div class="resource-details">
                  <div class="resource-name">Money</div>
                  <div class="resource-value">${{ formatNumber(operation.resources.money) }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <template #footer>
          <div class="operation-footer">
            <BaseButton
              variant="danger"
              @click="cancelOperation(operation)"
            >
              Cancel Operation
            </BaseButton>
          </div>
        </template>
      </BaseCard>
      
      <div v-if="inProgressOperations.length === 0" class="empty-state">
        <div class="empty-icon">🕒</div>
        <p>No operations in progress.</p>
        <BaseButton @click="activeTab = 'available'" variant="outline">Start an Operation</BaseButton>
      </div>
    </div>
    
    <!-- Completed Operations Tab -->
    <div v-else-if="activeTab === 'completed'" class="operations-list completed-operations">
      <BaseCard 
        v-for="operation in completedOperations" 
        :key="operation.id" 
        class="operation-card"
        :class="{ 
          'success': isSuccessfulOperation(operation), 
          'failure': isFailedOperation(operation) 
        }"
      >
        <div class="operation-header">
          <h3>{{ getOperationName(operation) }}</h3>
          <div class="operation-type">{{ getOperationType(operation) }}</div>
        </div>
        
        <div class="operation-details">
          <div class="completion-result">
            <div class="result-status" :class="{ 
              'success': isSuccessfulOperation(operation), 
              'failure': isFailedOperation(operation) 
            }">
              {{ isSuccessfulOperation(operation) ? 'Success!' : 'Failed!' }}
            </div>
            <div class="completion-time">
              Completed: {{ formatDate(operation.completionTime) }}
            </div>
          </div>
          
          <div class="operation-result" v-if="operation.result">
            <div class="result-message">{{ operation.result.message }}</div>
            
            <div class="result-details">
              <div class="result-column">
                <h4>Gains</h4>
                <ul class="result-list gains">
                  <li v-if="operation.result.moneyGained">
                    <span class="result-icon">💰</span>
                    <span class="result-text">${{ formatNumber(operation.result.moneyGained) }}</span>
                  </li>
                  <li v-if="operation.result.crewGained">
                    <span class="result-icon">👥</span>
                    <span class="result-text">{{ operation.result.crewGained }} Crew</span>
                  </li>
                  <li v-if="operation.result.weaponsGained">
                    <span class="result-icon">🔫</span>
                    <span class="result-text">{{ operation.result.weaponsGained }} Weapons</span>
                  </li>
                  <li v-if="operation.result.vehiclesGained">
                    <span class="result-icon">🚗</span>
                    <span class="result-text">{{ operation.result.vehiclesGained }} Vehicles</span>
                  </li>
                  <li v-if="operation.result.respectGained">
                    <span class="result-icon">👊</span>
                    <span class="result-text">{{ operation.result.respectGained }} Respect</span>
                  </li>
                  <li v-if="operation.result.influenceGained">
                    <span class="result-icon">🏛️</span>
                    <span class="result-text">{{ operation.result.influenceGained }} Influence</span>
                  </li>
                  <li v-if="operation.result.heatReduced">
                    <span class="result-icon">❄️</span>
                    <span class="result-text">{{ operation.result.heatReduced }} Heat Reduced</span>
                  </li>
                </ul>
              </div>
              
              <div class="result-column">
                <h4>Losses</h4>
                <ul class="result-list losses">
                  <li v-if="operation.result.moneyLost">
                    <span class="result-icon">💰</span>
                    <span class="result-text">${{ formatNumber(operation.result.moneyLost) }}</span>
                  </li>
                  <li v-if="operation.result.crewLost">
                    <span class="result-icon">👥</span>
                    <span class="result-text">{{ operation.result.crewLost }} Crew</span>
                  </li>
                  <li v-if="operation.result.weaponsLost">
                    <span class="result-icon">🔫</span>
                    <span class="result-text">{{ operation.result.weaponsLost }} Weapons</span>
                  </li>
                  <li v-if="operation.result.vehiclesLost">
                    <span class="result-icon">🚗</span>
                    <span class="result-text">{{ operation.result.vehiclesLost }} Vehicles</span>
                  </li>
                  <li v-if="operation.result.respectLost">
                    <span class="result-icon">👊</span>
                    <span class="result-text">{{ operation.result.respectLost }} Respect</span>
                  </li>
                  <li v-if="operation.result.heatGenerated">
                    <span class="result-icon">🔥</span>
                    <span class="result-text">{{ operation.result.heatGenerated }} Heat</span>
                  </li>
                </ul>
              </div>
            </div>
          </div>
          
          <div class="resources-committed">
            <h4>Resources Committed</h4>
            <div class="resources-grid">
              <div class="resource" v-if="operation.resources.crew > 0">
                <div class="resource-icon">👥</div>
                <div class="resource-details">
                  <div class="resource-name">Crew</div>
                  <div class="resource-value">{{ operation.resources.crew }}</div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.weapons > 0">
                <div class="resource-icon">🔫</div>
                <div class="resource-details">
                  <div class="resource-name">Weapons</div>
                  <div class="resource-value">{{ operation.resources.weapons }}</div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.vehicles > 0">
                <div class="resource-icon">🚗</div>
                <div class="resource-details">
                  <div class="resource-name">Vehicles</div>
                  <div class="resource-value">{{ operation.resources.vehicles }}</div>
                </div>
              </div>
              
              <div class="resource" v-if="operation.resources.money">
                <div class="resource-icon">💵</div>
                <div class="resource-details">
                  <div class="resource-name">Money</div>
                  <div class="resource-value">${{ formatNumber(operation.resources.money) }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </BaseCard>
      
      <div v-if="completedOperations.length === 0" class="empty-state">
        <div class="empty-icon">📝</div>
        <p>No completed operations yet.</p>
        <BaseButton @click="activeTab = 'available'" variant="outline">Start an Operation</BaseButton>
      </div>
    </div>
    
    <!-- Start Operation Modal -->
    <BaseModal 
      v-model="showStartModal"
      :title="selectedOperation ? `Start Operation: ${selectedOperation.name}` : 'Start Operation'"
    >
      <div v-if="selectedOperation" class="start-operation-modal">
        <div class="operation-summary">
          <h3>{{ selectedOperation.name }}</h3>
          <div class="summary-type">{{ formatOperationType(selectedOperation.type) }}</div>
          <p class="summary-description">{{ selectedOperation.description }}</p>
          
          <div class="summary-stats">
            <div class="stat">
              <div class="stat-label">Success Rate:</div>
              <div class="stat-value">{{ selectedOperation.successRate }}%</div>
            </div>
            <div class="stat">
              <div class="stat-label">Duration:</div>
              <div class="stat-value">{{ formatDuration(selectedOperation.duration) }}</div>
            </div>
          </div>
        </div>
        
        <div class="resource-allocation">
          <h4>Resources Required</h4>
          
          <div class="resources-grid">
            <div class="resource" v-if="selectedOperation.resources.crew > 0">
              <div class="resource-icon">👥</div>
              <div class="resource-details">
                <div class="resource-name">Crew</div>
                <div class="resource-value">
                  {{ selectedOperation.resources.crew }}
                  <span 
                    class="resource-status" 
                    :class="{ shortage: playerCrew < selectedOperation.resources.crew }"
                  >
                    ({{ playerCrew }} available)
                  </span>
                </div>
              </div>
            </div>
            
            <div class="resource" v-if="selectedOperation.resources.weapons > 0">
              <div class="resource-icon">🔫</div>
              <div class="resource-details">
                <div class="resource-name">Weapons</div>
                <div class="resource-value">
                  {{ selectedOperation.resources.weapons }}
                  <span 
                    class="resource-status" 
                    :class="{ shortage: playerWeapons < selectedOperation.resources.weapons }"
                  >
                    ({{ playerWeapons }} available)
                  </span>
                </div>
              </div>
            </div>
            
            <div class="resource" v-if="selectedOperation.resources.vehicles > 0">
              <div class="resource-icon">🚗</div>
              <div class="resource-details">
                <div class="resource-name">Vehicles</div>
                <div class="resource-value">
                  {{ selectedOperation.resources.vehicles }}
                  <span 
                    class="resource-status" 
                    :class="{ shortage: playerVehicles < selectedOperation.resources.vehicles }"
                  >
                    ({{ playerVehicles }} available)
                  </span>
                </div>
              </div>
            </div>
            
            <div class="resource" v-if="selectedOperation.resources.money">
              <div class="resource-icon">💵</div>
              <div class="resource-details">
                <div class="resource-name">Money</div>
                <div class="resource-value">
                  ${{ formatNumber(selectedOperation.resources.money) }}
                  <span 
                    class="resource-status" 
                    :class="{ shortage: playerMoney < selectedOperation.resources.money }"
                  >
                    (${{ formatNumber(playerMoney) }} available)
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="confirmation-warning" v-if="confirmationWarning">
          <div class="warning-icon">⚠️</div>
          <div class="warning-text">{{ confirmationWarning }}</div>
        </div>
      </div>
      
      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton 
            variant="text" 
            @click="closeStartModal"
          >
            Cancel
          </BaseButton>
          <BaseButton 
            :disabled="!canStartSelectedOperation"
            :loading="isStartingOperation"
            @click="confirmStartOperation"
          >
            Start Operation
          </BaseButton>
        </div>
      </template>
    </BaseModal>
    
    <!-- Cancel Operation Modal -->
    <BaseModal 
      v-model="showCancelModal"
      title="Cancel Operation"
    >
      <div class="cancel-operation-modal">
        <p>Are you sure you want to cancel this operation?</p>
        <p>Your committed resources will be returned, but any progress will be lost.</p>
      </div>
      
      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton 
            variant="text" 
            @click="closeCancelModal"
          >
            No, Continue Operation
          </BaseButton>
          <BaseButton 
            variant="danger"
            :loading="isCancellingOperation"
            @click="confirmCancelOperation"
          >
            Yes, Cancel Operation
          </BaseButton>
        </div>
      </template>
    </BaseModal>
  </div>
</template>

<style lang="scss">
.operations-view {
  .page-title {
    margin-bottom: $spacing-xl;
    
    h2 {
      @include gold-accent;
      margin-bottom: $spacing-xs;
    }
    
    .subtitle {
      color: $text-secondary;
    }
  }
  
  .operations-tabs {
    display: flex;
    gap: 1px;
    margin-bottom: $spacing-lg;
    background-color: $border-color;
    border-radius: $border-radius-sm;
    overflow: hidden;
    
    .tab-button {
      flex: 1;
      background-color: $background-lighter;
      border: none;
      padding: $spacing-md;
      color: $text-secondary;
      font-weight: 500;
      cursor: pointer;
      transition: $transition-base;
      
      &.active {
        background-color: $primary-color;
        color: $text-color;
      }
      
      &:hover:not(.active) {
        background-color: lighten($background-lighter, 5%);
      }
    }
  }
  
  .operations-filters {
    @include flex-between;
    margin-bottom: $spacing-lg;
    padding-bottom: $spacing-md;
    border-bottom: 1px solid $border-color;
    
    .filter {
      display: flex;
      align-items: center;
      gap: $spacing-sm;
      
      label {
        font-weight: 500;
      }
      
      select {
        background-color: $background-lighter;
        color: $text-color;
        border: 1px solid $border-color;
        border-radius: $border-radius-sm;
        padding: 6px 12px;
      }
    }
    
    .search {
      input {
        background-color: $background-lighter;
        color: $text-color;
        border: 1px solid $border-color;
        border-radius: $border-radius-sm;
        padding: 8px 12px;
        width: 250px;
        
        &::placeholder {
          color: $text-secondary;
        }
      }
    }
  }
  
  .operations-list {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    gap: $spacing-md;
    
    @include respond-to(lg) {
      grid-template-columns: repeat(2, 1fr);
    }
    
    .operation-card {
      position: relative;
      overflow: hidden;
      
      &.special-operation {
        @include gold-border;
      }
      
      &.in-progress {
        border-left: 4px solid $info-color;
      }
      
      &.success {
        border-left: 4px solid $success-color;
      }
      
      &.failure {
        border-left: 4px solid $danger-color;
      }
      
      .operation-badge {
        background-color: $secondary-color;
        color: $background-color;
        font-size: $font-size-sm;
        font-weight: 600;
        padding: 2px 8px;
        border-radius: $border-radius-sm;
      }
      
      .operation-header {
        margin-bottom: $spacing-md;
        
        h3 {
          margin: 0 0 $spacing-xs 0;
          font-size: $font-size-lg;
        }
        
        .operation-type {
          color: $text-secondary;
          font-size: $font-size-sm;
        }
      }
      
      .operation-details {
        @include flex-column;
        gap: $spacing-md;
        
        .description {
          color: $text-secondary;
          margin: 0;
        }
        
        .requirements {
          .requirements-list {
            list-style: none;
            padding: 0;
            margin: $spacing-sm 0 0 0;
            
            li {
              display: flex;
              justify-content: space-between;
              margin-bottom: $spacing-xs;
              
              .requirement-status {
                font-weight: 600;
                
                &.met {
                  color: $success-color;
                }
                
                &:not(.met) {
                  color: $danger-color;
                }
              }
            }
          }
        }
        
        .operation-resources, .resources-committed {
          h4 {
            margin: 0 0 $spacing-sm 0;
          }
          
          .resources-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: $spacing-md;
            
            .resource {
              display: flex;
              align-items: center;
              gap: $spacing-sm;
              
              .resource-icon {
                font-size: 24px;
              }
              
              .resource-details {
                .resource-name {
                  font-weight: 500;
                }
                
                .resource-value {
                  font-size: $font-size-sm;
                  color: $text-secondary;
                  
                  .resource-status {
                    &.shortage {
                      color: $danger-color;
                    }
                  }
                }
              }
            }
          }
        }
        
        .operation-details-grid {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: $spacing-lg;
          
          .detail-column {
            h4 {
              margin: 0 0 $spacing-sm 0;
            }
            
            .rewards-list, .risks-list {
              list-style: none;
              padding: 0;
              margin: 0;
              
              li {
                display: flex;
                align-items: center;
                gap: $spacing-sm;
                margin-bottom: $spacing-xs;
                
                .reward-icon, .risk-icon {
                  flex-shrink: 0;
                }
                
                .reward-text {
                  color: $success-color;
                }
                
                .risk-text {
                  color: $danger-color;
                }
              }
            }
          }
        }
        
        .operation-stats {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: $spacing-md;
          padding-top: $spacing-md;
          border-top: 1px solid $border-color;
          
          .stat {
            .stat-label {
              font-size: $font-size-sm;
              color: $text-secondary;
              margin-bottom: 4px;
            }
            
            .stat-value {
              font-weight: 600;
            }
          }
        }
        
        .progress-tracker {
          background-color: rgba(255, 255, 255, 0.05);
          border-radius: $border-radius-sm;
          padding: $spacing-md;
          
          .progress-info {
            @include flex-between;
            margin-bottom: $spacing-sm;
            
            .time-remaining {
              font-weight: 600;
              color: $info-color;
            }
            
            .completion-time {
              font-size: $font-size-sm;
              color: $text-secondary;
            }
          }
          
          .progress-bar {
            height: 10px;
            background-color: rgba(255, 255, 255, 0.1);
            border-radius: $border-radius-sm;
            overflow: hidden;
            
            .progress-fill {
              height: 100%;
              background-color: $info-color;
              width: 0%;
              transition: width 0.5s ease;
            }
          }
        }
        
        .completion-result {
          text-align: center;
          margin-bottom: $spacing-md;
          
          .result-status {
            font-size: $font-size-xl;
            font-weight: 700;
            margin-bottom: $spacing-xs;
            
            &.success {
              color: $success-color;
            }
            
            &.failure {
              color: $danger-color;
            }
          }
          
          .completion-time {
            color: $text-secondary;
            font-size: $font-size-sm;
          }
        }
        
        .operation-result {
          background-color: rgba(255, 255, 255, 0.05);
          border-radius: $border-radius-sm;
          padding: $spacing-md;
          margin-bottom: $spacing-md;
          
          .result-message {
            margin-bottom: $spacing-md;
            font-style: italic;
          }
          
          .result-details {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: $spacing-md;
            
            .result-column {
              h4 {
                margin: 0 0 $spacing-sm 0;
              }
              
              .result-list {
                list-style: none;
                padding: 0;
                margin: 0;
                
                li {
                  display: flex;
                  align-items: center;
                  gap: $spacing-sm;
                  margin-bottom: $spacing-xs;
                  
                  .result-icon {
                    flex-shrink: 0;
                  }
                }
                
                &.gains .result-text {
                  color: $success-color;
                }
                
                &.losses .result-text {
                  color: $danger-color;
                }
              }
            }
          }
        }
      }
      
      .operation-footer {
        @include flex-column;
        gap: $spacing-sm;
        
        .operation-warning {
          font-size: $font-size-sm;
          color: $warning-color;
          text-align: center;
        }
      }
    }
    
    .empty-state {
      grid-column: 1 / -1;
      @include flex-column;
      align-items: center;
      justify-content: center;
      gap: $spacing-md;
      padding: $spacing-xl;
      text-align: center;
      color: $text-secondary;
      background-color: $background-card;
      border-radius: $border-radius-md;
      
      .empty-icon {
        font-size: 48px;
        margin-bottom: $spacing-md;
      }
    }
  }
  
  .start-operation-modal {
    @include flex-column;
    gap: $spacing-lg;
    
    .operation-summary {
      @include flex-column;
      gap: $spacing-sm;
      
      h3 {
        margin: 0;
        @include gold-accent;
      }
      
      .summary-type {
        color: $text-secondary;
        font-size: $font-size-sm;
      }
      
      .summary-description {
        margin: $spacing-md 0;
      }
      
      .summary-stats {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: $spacing-md;
        margin-top: $spacing-sm;
        
        .stat {
          .stat-label {
            font-size: $font-size-sm;
            color: $text-secondary;
            margin-bottom: 4px;
          }
          
          .stat-value {
            font-weight: 600;
          }
        }
      }
    }
    
    .resource-allocation {
      h4 {
        margin: 0 0 $spacing-md 0;
      }
      
      .resources-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: $spacing-md;
        
        .resource {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          
          .resource-icon {
            font-size: 24px;
          }
          
          .resource-details {
            flex: 1;
            
            .resource-name {
              font-weight: 500;
            }
            
            .resource-value {
              font-size: $font-size-sm;
              color: $text-secondary;
              
              .resource-status {
                &.shortage {
                  color: $danger-color;
                }
              }
            }
          }
        }
      }
    }
    
    .confirmation-warning {
      display: flex;
      gap: $spacing-sm;
      padding: $spacing-md;
      background-color: rgba($warning-color, 0.1);
      border-left: 3px solid $warning-color;
      border-radius: $border-radius-sm;
      
      .warning-text {
        font-size: $font-size-sm;
      }
    }
  }
  
  .cancel-operation-modal {
    text-align: center;
    
    p:first-child {
      font-size: $font-size-lg;
      margin-bottom: $spacing-md;
    }
    
    p:last-child {
      color: $text-secondary;
    }
  }
  
  .modal-footer-actions {
    @include flex-between;
    width: 100%;
  }
}
</style>