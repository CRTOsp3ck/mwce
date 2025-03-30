// src/views/MarketView.vue

<template>
    <div class="market-view">
      <div class="page-title">
        <h2>Black Market</h2>
        <p class="subtitle">Buy and sell resources to expand your criminal empire.</p>
      </div>
      
      <div class="market-tabs">
        <button 
          class="tab-button" 
          :class="{ active: activeTab === 'buy' }"
          @click="navigateToTab('buy')"
        >
          Buy Resources
        </button>
        <button 
          class="tab-button" 
          :class="{ active: activeTab === 'sell' }"
          @click="navigateToTab('sell')"
        >
          Sell Resources
        </button>
        <button 
          class="tab-button" 
          :class="{ active: activeTab === 'history' }"
          @click="navigateToTab('history')"
        >
          Transaction History
        </button>
      </div>
      
      <!-- Buy Resources Tab -->
      <div v-if="activeTab === 'buy'" class="market-content">
        <div class="market-grid">
          <BaseCard 
            v-for="listing in marketListings" 
            :key="listing.type" 
            class="market-card"
          >
            <div class="listing-header">
              <h3>{{ formatResourceType(listing.type) }}</h3>
              <div 
                class="price-trend" 
                :class="{ 
                  'trend-up': listing.trend === 'up', 
                  'trend-down': listing.trend === 'down',
                  'trend-stable': listing.trend === 'stable'
                }"
              >
                {{ getTrendIcon(listing.trend) }} {{ listing.trendPercentage }}%
              </div>
            </div>
            
            <div class="listing-details">
              <div class="current-price">
                <div class="price-label">Current Price:</div>
                <div class="price-value">${{ formatNumber(listing.price) }} each</div>
              </div>
              
              <div class="player-resource">
                <div class="resource-label">You currently have:</div>
                <div class="resource-value">{{ getResourceAmount(listing.type) }}</div>
              </div>
              
              <div class="transaction-form">
                <div class="form-group">
                  <label>Quantity to Buy:</label>
                  <div class="quantity-control">
                    <button 
                      class="quantity-btn" 
                      @click="decrementQuantity(listing.type)"
                      :disabled="buyQuantities[listing.type] <= 1"
                    >
                      -
                    </button>
                    <input 
                      type="number" 
                      v-model="buyQuantities[listing.type]" 
                      min="1" 
                      :max="getMaxBuyQuantity(listing)"
                    />
                    <button 
                      class="quantity-btn" 
                      @click="incrementQuantity(listing.type)"
                      :disabled="buyQuantities[listing.type] >= getMaxBuyQuantity(listing)"
                    >
                      +
                    </button>
                  </div>
                </div>
                
                <div class="transaction-summary">
                  <div class="summary-label">Total Cost:</div>
                  <div class="summary-value">${{ formatNumber(calculateTotalCost(listing)) }}</div>
                </div>
                
                <BaseButton 
                  variant="secondary" 
                  class="transaction-btn"
                  :disabled="!canBuyResource(listing) || isLoading"
                  :loading="isBuying === listing.type"
                  @click="buyResource(listing)"
                >
                  Buy {{ formatResourceType(listing.type) }}
                </BaseButton>
                
                <div v-if="getTransactionWarning(listing, 'buy')" class="transaction-warning">
                  {{ getTransactionWarning(listing, 'buy') }}
                </div>
              </div>
            </div>
          </BaseCard>
        </div>
      </div>
      
      <!-- Sell Resources Tab -->
      <div v-else-if="activeTab === 'sell'" class="market-content">
        <div class="market-grid">
          <BaseCard 
            v-for="listing in marketListings" 
            :key="listing.type" 
            class="market-card"
          >
            <div class="listing-header">
              <h3>{{ formatResourceType(listing.type) }}</h3>
              <div 
              class="price-trend" 
              :class="{ 
                'trend-up': listing.trend === 'up', 
                'trend-down': listing.trend === 'down',
                'trend-stable': listing.trend === 'stable'
              }"
            >
              {{ getTrendIcon(listing.trend) }} {{ listing.trendPercentage }}%
            </div>
          </div>
          
          <div class="listing-details">
            <div class="current-price">
              <div class="price-label">Current Price:</div>
              <div class="price-value">${{ formatNumber(listing.price) }} each</div>
            </div>
            
            <div class="player-resource">
              <div class="resource-label">You currently have:</div>
              <div class="resource-value">{{ getResourceAmount(listing.type) }}</div>
            </div>
            
            <div class="transaction-form">
              <div class="form-group">
                <label>Quantity to Sell:</label>
                <div class="quantity-control">
                  <button 
                    class="quantity-btn" 
                    @click="decrementQuantity(listing.type, 'sell')"
                    :disabled="sellQuantities[listing.type] <= 1"
                  >
                    -
                  </button>
                  <input 
                    type="number" 
                    v-model="sellQuantities[listing.type]" 
                    min="1" 
                    :max="getMaxSellQuantity(listing)"
                  />
                  <button 
                    class="quantity-btn" 
                    @click="incrementQuantity(listing.type, 'sell')"
                    :disabled="sellQuantities[listing.type] >= getMaxSellQuantity(listing)"
                  >
                    +
                  </button>
                </div>
              </div>
              
              <div class="transaction-summary">
                <div class="summary-label">Total Value:</div>
                <div class="summary-value">${{ formatNumber(calculateTotalValue(listing)) }}</div>
              </div>
              
              <BaseButton 
                variant="primary" 
                class="transaction-btn"
                :disabled="!canSellResource(listing) || isLoading"
                :loading="isSelling === listing.type"
                @click="sellResource(listing)"
              >
                Sell {{ formatResourceType(listing.type) }}
              </BaseButton>
              
              <div v-if="getTransactionWarning(listing, 'sell')" class="transaction-warning">
                {{ getTransactionWarning(listing, 'sell') }}
              </div>
            </div>
          </div>
        </BaseCard>
      </div>
    </div>
    
    <!-- Transaction History Tab -->
    <div v-else-if="activeTab === 'history'" class="market-content">
      <BaseCard class="history-card">
        <template #header>
          <div class="card-actions">
            <div class="transaction-filter">
              <label>Filter:</label>
              <select v-model="historyFilter">
                <option value="all">All Transactions</option>
                <option value="buy">Buy Only</option>
                <option value="sell">Sell Only</option>
              </select>
            </div>
          </div>
        </template>
        
        <div class="transaction-history">
          <div class="history-header">
            <div class="header-cell">Type</div>
            <div class="header-cell">Resource</div>
            <div class="header-cell">Quantity</div>
            <div class="header-cell">Price</div>
            <div class="header-cell">Total</div>
            <div class="header-cell">Date</div>
          </div>
          
          <div class="transaction-list">
            <div 
              v-for="transaction in filteredTransactions" 
              :key="transaction.id" 
              class="transaction-item"
              :class="{ 
                'buy-transaction': transaction.transactionType === 'buy', 
                'sell-transaction': transaction.transactionType === 'sell' 
              }"
            >
              <div class="transaction-cell">
                {{ transaction.transactionType === 'buy' ? 'Buy' : 'Sell' }}
              </div>
              <div class="transaction-cell">
                {{ formatResourceType(transaction.resourceType) }}
              </div>
              <div class="transaction-cell">
                {{ transaction.quantity }}
              </div>
              <div class="transaction-cell">
                ${{ formatNumber(transaction.price) }}
              </div>
              <div class="transaction-cell">
                ${{ formatNumber(transaction.totalCost) }}
              </div>
              <div class="transaction-cell">
                {{ formatDate(transaction.timestamp) }}
              </div>
            </div>
          </div>
          
          <div v-if="filteredTransactions.length === 0" class="empty-history">
            <p>No transaction history found.</p>
          </div>
        </div>
      </BaseCard>
      
      <BaseCard class="price-history-card">
        <template #header>
          <h3>Price History</h3>
        </template>
        
        <div class="price-history">
          <div class="chart-container">
            <!-- Price history chart will go here -->
            <div class="chart-placeholder">
              <p>Price history chart is under development.</p>
            </div>
          </div>
          
          <div class="chart-legend">
            <div class="legend-item crew">
              <div class="legend-color"></div>
              <div class="legend-label">Crew</div>
            </div>
            <div class="legend-item weapons">
              <div class="legend-color"></div>
              <div class="legend-label">Weapons</div>
            </div>
            <div class="legend-item vehicles">
              <div class="legend-color"></div>
              <div class="legend-label">Vehicles</div>
            </div>
          </div>
        </div>
      </BaseCard>
    </div>
    
    <!-- Transaction Confirmation Modal -->
    <BaseModal 
      v-model="showConfirmModal"
      :title="confirmationTitle"
    >
      <div v-if="selectedListing && transactionType" class="confirmation-content">
        <p class="confirmation-message">
          Are you sure you want to {{ transactionType === 'buy' ? 'buy' : 'sell' }}
          <strong>{{ transactionQuantity }}</strong>
          {{ formatResourceType(selectedListing.type) }}
          for a total of
          <strong>${{ formatNumber(transactionTotal) }}</strong>?
        </p>
        
        <div class="confirmation-details">
          <div class="detail-item">
            <div class="detail-label">Unit Price:</div>
            <div class="detail-value">${{ formatNumber(selectedListing.price) }}</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Quantity:</div>
            <div class="detail-value">{{ transactionQuantity }}</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Total:</div>
            <div class="detail-value">${{ formatNumber(transactionTotal) }}</div>
          </div>
        </div>
        
        <div class="balance-preview">
          <div class="preview-item">
            <div class="preview-label">Current Balance:</div>
            <div class="preview-value">${{ formatNumber(playerMoney) }}</div>
          </div>
          <div class="preview-item">
            <div class="preview-label">After Transaction:</div>
            <div class="preview-value" :class="{ 'negative': newBalance < 0 }">
              ${{ formatNumber(newBalance) }}
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton 
            variant="text" 
            @click="closeConfirmModal"
          >
            Cancel
          </BaseButton>
          <BaseButton 
            :variant="transactionType === 'buy' ? 'secondary' : 'primary'"
            :loading="isConfirming"
            @click="confirmTransaction"
          >
            Confirm {{ transactionType === 'buy' ? 'Purchase' : 'Sale' }}
          </BaseButton>
        </div>
      </template>
    </BaseModal>
    
    <!-- Result Modal -->
    <BaseModal 
      v-model="showResultModal"
      title="Transaction Complete"
    >
      <div class="transaction-result">
        <div class="result-icon">✅</div>
        <div class="result-message">
          {{ resultMessage }}
        </div>
        <div class="result-details">
          <div class="detail-item">
            <div class="detail-label">Type:</div>
            <div class="detail-value">{{ transactionType === 'buy' ? 'Purchase' : 'Sale' }}</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Resource:</div>
            <div class="detail-value">{{ selectedListing ? formatResourceType(selectedListing.type) : '' }}</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Quantity:</div>
            <div class="detail-value">{{ transactionQuantity }}</div>
          </div>
          <div class="detail-item">
            <div class="detail-label">Total:</div>
            <div class="detail-value">${{ formatNumber(transactionTotal) }}</div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <BaseButton @click="closeResultModal">Close</BaseButton>
      </template>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import BaseCard from '@/components/ui/BaseCard.vue';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useMarketStore } from '@/stores/modules/market';
import { 
  MarketListing, 
  ResourceType,
  PriceTrend,
  TransactionType,
  MarketTransaction
} from '@/types/market';

const route = useRoute();
const router = useRouter();

const playerStore = usePlayerStore();
const marketStore = useMarketStore();

// View state
// const activeTab = ref<'buy' | 'sell' | 'history'>('buy');
const activeTab = computed(()=> route.query.tab as string || 'buy')
const isLoading = ref(false);
const isBuying = ref<ResourceType | null>(null);
const isSelling = ref<ResourceType | null>(null);
const historyFilter = ref<'all' | 'buy' | 'sell'>('all');

// Quantity inputs
const buyQuantities = ref<Record<ResourceType, number>>({
  [ResourceType.CREW]: 1,
  [ResourceType.WEAPONS]: 1,
  [ResourceType.VEHICLES]: 1
});

const sellQuantities = ref<Record<ResourceType, number>>({
  [ResourceType.CREW]: 1,
  [ResourceType.WEAPONS]: 1,
  [ResourceType.VEHICLES]: 1
});

// Modals
const showConfirmModal = ref(false);
const showResultModal = ref(false);
const selectedListing = ref<MarketListing | null>(null);
const transactionType = ref<'buy' | 'sell' | null>(null);
const transactionQuantity = ref(0);
const transactionTotal = ref(0);
const isConfirming = ref(false);
const resultMessage = ref('');

// Computed properties
const playerMoney = computed(() => playerStore.playerMoney);
const playerCrew = computed(() => playerStore.playerCrew);
const playerWeapons = computed(() => playerStore.playerWeapons);
const playerVehicles = computed(() => playerStore.playerVehicles);
const maxCrew = computed(() => playerStore.maxCrew);
const maxWeapons = computed(() => playerStore.maxWeapons);
const maxVehicles = computed(() => playerStore.maxVehicles);

const marketListings = computed(() => marketStore.listings);
const transactions = computed(() => marketStore.transactions);

const filteredTransactions = computed(() => {
  if (historyFilter.value === 'all') {
    return transactions.value;
  } else if (historyFilter.value === 'buy') {
    return transactions.value.filter(t => t.transactionType === TransactionType.BUY);
  } else {
    return transactions.value.filter(t => t.transactionType === TransactionType.SELL);
  }
});

const confirmationTitle = computed(() => {
  if (!transactionType.value || !selectedListing.value) return 'Confirm Transaction';
  
  return transactionType.value === 'buy' 
    ? `Buy ${formatResourceType(selectedListing.value.type)}` 
    : `Sell ${formatResourceType(selectedListing.value.type)}`;
});

const newBalance = computed(() => {
  if (!transactionType.value || !transactionTotal.value) return playerMoney.value;
  
  return transactionType.value === 'buy'
    ? playerMoney.value - transactionTotal.value
    : playerMoney.value + transactionTotal.value;
});

// Load data when component is mounted
onMounted(async () => {
  isLoading.value = true;
  
  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }
  
  await marketStore.fetchMarketData();
  
  isLoading.value = false;
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

function formatResourceType(resourceType: ResourceType): string {
  switch (resourceType) {
    case ResourceType.CREW:
      return 'Crew Members';
    case ResourceType.WEAPONS:
      return 'Weapons';
    case ResourceType.VEHICLES:
      return 'Vehicles';
    default:
      return resourceType;
  }
}

function formatDate(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function getTrendIcon(trend: PriceTrend): string {
  switch (trend) {
    case PriceTrend.UP:
      return '↑';
    case PriceTrend.DOWN:
      return '↓';
    case PriceTrend.STABLE:
      return '→';
    default:
      return '';
  }
}

function getResourceAmount(resourceType: ResourceType): number {
  switch (resourceType) {
    case ResourceType.CREW:
      return playerCrew.value;
    case ResourceType.WEAPONS:
      return playerWeapons.value;
    case ResourceType.VEHICLES:
      return playerVehicles.value;
    default:
      return 0;
  }
}

function getResourceMax(resourceType: ResourceType): number {
  switch (resourceType) {
    case ResourceType.CREW:
      return maxCrew.value;
    case ResourceType.WEAPONS:
      return maxWeapons.value;
    case ResourceType.VEHICLES:
      return maxVehicles.value;
    default:
      return 0;
  }
}

function getMaxBuyQuantity(listing: MarketListing): number {
  // Calculate max based on money and resource capacity
  const maxByMoney = Math.floor(playerMoney.value / listing.price);
  const currentAmount = getResourceAmount(listing.type);
  const maxCapacity = getResourceMax(listing.type);
  const maxByCapacity = maxCapacity - currentAmount;
  
  return Math.max(0, Math.min(maxByMoney, maxByCapacity));
}

function getMaxSellQuantity(listing: MarketListing): number {
  return getResourceAmount(listing.type);
}

function calculateTotalCost(listing: MarketListing): number {
  return listing.price * buyQuantities.value[listing.type];
}

function calculateTotalValue(listing: MarketListing): number {
  return listing.price * sellQuantities.value[listing.type];
}

function canBuyResource(listing: MarketListing): boolean {
  const quantity = buyQuantities.value[listing.type];
  if (quantity <= 0) return false;
  
  const totalCost = calculateTotalCost(listing);
  if (totalCost > playerMoney.value) return false;
  
  const currentAmount = getResourceAmount(listing.type);
  const maxCapacity = getResourceMax(listing.type);
  if (currentAmount + quantity > maxCapacity) return false;
  
  return true;
}

function canSellResource(listing: MarketListing): boolean {
  const quantity = sellQuantities.value[listing.type];
  if (quantity <= 0) return false;
  
  const currentAmount = getResourceAmount(listing.type);
  if (quantity > currentAmount) return false;
  
  return true;
}

function getTransactionWarning(listing: MarketListing, action: 'buy' | 'sell'): string {
  if (action === 'buy') {
    const quantity = buyQuantities.value[listing.type];
    if (quantity <= 0) return 'Quantity must be greater than 0';
    
    const totalCost = calculateTotalCost(listing);
    if (totalCost > playerMoney.value) return 'Not enough money';
    
    const currentAmount = getResourceAmount(listing.type);
    const maxCapacity = getResourceMax(listing.type);
    if (currentAmount + quantity > maxCapacity) return 'Exceeds maximum capacity';
    
    return '';
  } else {
    const quantity = sellQuantities.value[listing.type];
    if (quantity <= 0) return 'Quantity must be greater than 0';
    
    const currentAmount = getResourceAmount(listing.type);
    if (quantity > currentAmount) return 'Not enough resources to sell';
    
    return '';
  }
}

function incrementQuantity(resourceType: ResourceType, action: 'buy' | 'sell' = 'buy'): void {
  if (action === 'buy') {
    const maxQuantity = getMaxBuyQuantity(
      marketListings.value.find(l => l.type === resourceType) as MarketListing
    );
    
    if (buyQuantities.value[resourceType] < maxQuantity) {
      buyQuantities.value[resourceType]++;
    }
  } else {
    const maxQuantity = getMaxSellQuantity(
      marketListings.value.find(l => l.type === resourceType) as MarketListing
    );
    
    if (sellQuantities.value[resourceType] < maxQuantity) {
      sellQuantities.value[resourceType]++;
    }
  }
}

function decrementQuantity(resourceType: ResourceType, action: 'buy' | 'sell' = 'buy'): void {
  if (action === 'buy') {
    if (buyQuantities.value[resourceType] > 1) {
      buyQuantities.value[resourceType]--;
    }
  } else {
    if (sellQuantities.value[resourceType] > 1) {
      sellQuantities.value[resourceType]--;
    }
  }
}

// Transaction functions
function buyResource(listing: MarketListing): void {
  if (!canBuyResource(listing)) return;
  
  selectedListing.value = listing;
  transactionType.value = 'buy';
  transactionQuantity.value = buyQuantities.value[listing.type];
  transactionTotal.value = calculateTotalCost(listing);
  
  showConfirmModal.value = true;
}

function sellResource(listing: MarketListing): void {
  if (!canSellResource(listing)) return;
  
  selectedListing.value = listing;
  transactionType.value = 'sell';
  transactionQuantity.value = sellQuantities.value[listing.type];
  transactionTotal.value = calculateTotalValue(listing);
  
  showConfirmModal.value = true;
}

function closeConfirmModal(): void {
  showConfirmModal.value = false;
  selectedListing.value = null;
  transactionType.value = null;
  transactionQuantity.value = 0;
  transactionTotal.value = 0;
}

async function confirmTransaction(): Promise<void> {
  if (!selectedListing.value || !transactionType.value || isConfirming.value) return;
  
  isConfirming.value = true;
  
  try {
    let result;
    
    if (transactionType.value === 'buy') {
      isBuying.value = selectedListing.value.type;
      result = await marketStore.buyResource(
        selectedListing.value.type, 
        transactionQuantity.value
      );
    } else {
      isSelling.value = selectedListing.value.type;
      result = await marketStore.sellResource(
        selectedListing.value.type, 
        transactionQuantity.value
      );
    }
    
    if (result) {
      // Set result message
      resultMessage.value = transactionType.value === 'buy'
        ? `You successfully purchased ${transactionQuantity.value} ${formatResourceType(selectedListing.value.type)} for $${formatNumber(transactionTotal.value)}.`
        : `You successfully sold ${transactionQuantity.value} ${formatResourceType(selectedListing.value.type)} for $${formatNumber(transactionTotal.value)}.`;
      
      // Close confirm modal and show result
      showConfirmModal.value = false;
      showResultModal.value = true;
      
      // Reset quantities
      if (transactionType.value === 'buy') {
        buyQuantities.value[selectedListing.value.type] = 1;
      } else {
        sellQuantities.value[selectedListing.value.type] = 1;
      }
    }
  } catch (error) {
    console.error('Transaction error:', error);
  } finally {
    isConfirming.value = false;
    isBuying.value = null;
    isSelling.value = null;
  }
}

function closeResultModal(): void {
  showResultModal.value = false;
  resultMessage.value = '';
}

function navigateToTab(tab:'buy' | 'sell' | 'history'){
  router.push({ path:'/market', query: { tab }})
}
</script>

<style lang="scss">
.market-view {
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
  
  .market-tabs {
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
  
  .market-content {
    .market-grid {
      display: grid;
      grid-template-columns: repeat(1, 1fr);
      gap: $spacing-lg;
      
      @include respond-to(md) {
        grid-template-columns: repeat(3, 1fr);
      }
      
      .market-card {
        .listing-header {
          @include flex-between;
          margin-bottom: $spacing-md;
          
          h3 {
            margin: 0;
          }
          
          .price-trend {
            font-weight: 600;
            
            &.trend-up {
              color: $success-color;
            }
            
            &.trend-down {
              color: $danger-color;
            }
            
            &.trend-stable {
              color: $warning-color;
            }
          }
        }
        
        .listing-details {
          @include flex-column;
          gap: $spacing-md;
          
          .current-price, .player-resource {
            @include flex-between;
            
            .price-label, .resource-label {
              color: $text-secondary;
            }
            
            .price-value, .resource-value {
              font-weight: 600;
            }
          }
          
          .transaction-form {
            @include flex-column;
            gap: $spacing-md;
            padding-top: $spacing-md;
            border-top: 1px solid $border-color;
            
            .form-group {
              @include flex-column;
              gap: $spacing-xs;
              
              label {
                font-weight: 500;
              }
              
              .quantity-control {
                display: flex;
                
                input {
                  width: 60px;
                  text-align: center;
                  border: 1px solid $border-color;
                  background-color: $background-lighter;
                  color: $text-color;
                  padding: 6px;
                }
                
                .quantity-btn {
                  width: 36px;
                  background-color: $background-lighter;
                  border: 1px solid $border-color;
                  color: $text-color;
                  font-weight: 600;
                  cursor: pointer;
                  
                  &:first-child {
                    border-radius: $border-radius-sm 0 0 $border-radius-sm;
                    border-right: none;
                  }
                  
                  &:last-child {
                    border-radius: 0 $border-radius-sm $border-radius-sm 0;
                    border-left: none;
                  }
                  
                  &:disabled {
                    opacity: 0.5;
                    cursor: not-allowed;
                  }
                  
                  &:hover:not(:disabled) {
                    background-color: lighten($background-lighter, 5%);
                  }
                }
              }
            }
            
            .transaction-summary {
              @include flex-between;
              
              .summary-label {
                color: $text-secondary;
              }
              
              .summary-value {
                font-weight: 600;
                @include gold-accent;
              }
            }
            
            .transaction-btn {
              margin-top: $spacing-sm;
            }
            
            .transaction-warning {
              font-size: $font-size-sm;
              color: $danger-color;
              text-align: center;
            }
          }
        }
      }
    }
    
    .history-card, .price-history-card {
      margin-bottom: $spacing-lg;
      
      .card-actions {
        display: flex;
        gap: $spacing-md;
        
        .transaction-filter {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          
          select {
            background-color: $background-lighter;
            color: $text-color;
            border: 1px solid $border-color;
            border-radius: $border-radius-sm;
            padding: 4px 8px;
          }
        }
      }
      
      .transaction-history {
        .history-header {
          display: grid;
          grid-template-columns: 1fr 1fr 1fr 1fr 1fr 2fr;
          gap: $spacing-sm;
          padding: $spacing-sm;
          background-color: $background-lighter;
          border-radius: $border-radius-sm;
          margin-bottom: $spacing-md;
          
          .header-cell {
            font-weight: 600;
          }
        }
        
        .transaction-list {
          @include flex-column;
          gap: $spacing-xs;
          
          .transaction-item {
            display: grid;
            grid-template-columns: 1fr 1fr 1fr 1fr 1fr 2fr;
            gap: $spacing-sm;
            padding: $spacing-sm;
            border-radius: $border-radius-sm;
            transition: $transition-base;
            
            &:hover {
              background-color: rgba(255, 255, 255, 0.05);
            }
            
            &.buy-transaction {
              border-left: 2px solid $primary-color;
            }
            
            &.sell-transaction {
              border-left: 2px solid $secondary-color;
            }
          }
        }
        
        .empty-history {
          text-align: center;
          padding: $spacing-lg;
          color: $text-secondary;
        }
      }
      
      .price-history {
        .chart-container {
          height: 300px;
          margin-bottom: $spacing-md;
          
          .chart-placeholder {
            @include flex-center;
            height: 100%;
            background-color: $background-lighter;
            border-radius: $border-radius-md;
            color: $text-secondary;
          }
        }
        
        .chart-legend {
          display: flex;
          justify-content: center;
          gap: $spacing-lg;
          
          .legend-item {
            display: flex;
            align-items: center;
            gap: $spacing-xs;
            
            .legend-color {
              width: 16px;
              height: 16px;
              border-radius: 50%;
            }
            
            &.crew .legend-color {
              background-color: #3498db;
            }
            
            &.weapons .legend-color {
              background-color: #e74c3c;
            }
            
            &.vehicles .legend-color {
              background-color: #2ecc71;
            }
          }
        }
      }
    }
  }
  
  .confirmation-content {
    @include flex-column;
    gap: $spacing-lg;
    
    .confirmation-message {
      font-size: $font-size-lg;
      text-align: center;
      margin: 0;
    }
    
    .confirmation-details {
      background-color: rgba(255, 255, 255, 0.05);
      border-radius: $border-radius-sm;
      padding: $spacing-md;
      
      .detail-item {
        @include flex-between;
        margin-bottom: $spacing-xs;
        
        &:last-child {
          margin-bottom: 0;
        }
        
        .detail-label {
          color: $text-secondary;
        }
        
        .detail-value {
          font-weight: 600;
        }
      }
    }
    
    .balance-preview {
      padding-top: $spacing-md;
      border-top: 1px solid $border-color;
      
      .preview-item {
        @include flex-between;
        margin-bottom: $spacing-xs;
        
        &:last-child {
          margin-bottom: 0;
        }
        
        .preview-label {
          color: $text-secondary;
        }
        
        .preview-value {
          font-weight: 600;
          
          &.negative {
            color: $danger-color;
          }
        }
      }
    }
  }
  
  .transaction-result {
    @include flex-column;
    align-items: center;
    text-align: center;
    gap: $spacing-md;
    
    .result-icon {
      font-size: 48px;
    }
    
    .result-message {
      font-size: $font-size-lg;
      font-weight: 500;
    }
    
    .result-details {
      background-color: rgba(255, 255, 255, 0.05);
      border-radius: $border-radius-sm;
      padding: $spacing-md;
      width: 100%;
      
      .detail-item {
        @include flex-between;
        margin-bottom: $spacing-xs;
        
        &:last-child {
          margin-bottom: 0;
        }
        
        .detail-label {
          color: $text-secondary;
        }
        
        .detail-value {
          font-weight: 600;
        }
      }
    }
  }
  
  .modal-footer-actions {
    @include flex-between;
    width: 100%;
  }
}
</style>