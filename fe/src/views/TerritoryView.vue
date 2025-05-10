<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import BaseButton from '@/components/ui/BaseButton.vue';
import BaseModal from '@/components/ui/BaseModal.vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import {
  Hotspot,
  TerritoryActionType,
  ActionResources,
  ActionResult,
  TerritoryAction
} from '@/types/territory';

import HotspotTooltip from '@/components/HotspotTooltip.vue';
import HotspotDetailModal from '@/components/HotspotDetailModal.vue';

const route = useRoute();
const router = useRouter();

const playerStore = usePlayerStore();
const territoryStore = useTerritoryStore();

// View state
const activeTab = computed(() => route.query.tab as string || 'empire');
const viewMode = ref<'grid' | 'map'>('grid');
const isLoading = ref(false);
const isCollecting = ref(false);

// Filters
const selectedRegionId = ref<string | null>(null);
const selectedDistrictId = ref<string | null>(null);
const selectedCityId = ref<string | null>(null);
const businessFilter = ref<'all' | 'legal' | 'illegal'>('all');
const empireSortBy = ref<'name' | 'income' | 'pending' | 'defense' | 'region'>('income');

// Action modal
const showActionModal = ref(false);
const selectedHotspot = ref<Hotspot | null>(null);
const selectedAction = ref<TerritoryActionType | null>(null);
const actionResources = ref<ActionResources>({
  crew: 0,
  weapons: 0,
  vehicles: 0
});
const isPerformingAction = ref(false);

// Result modal
const showResultModal = ref(false);
const actionResult = ref<ActionResult | null>(null);
const actionSuccess = ref(false);

// For tooltip functionality
const tooltipVisible = ref(false);
const tooltipPosition = ref({ x: 0, y: 0 });
const tooltipHotspot = ref<Hotspot | null>(null);

// For detail modal functionality
const showDetailModal = ref(false);
const detailModalHotspot = ref<Hotspot | null>(null);

// Computed properties
const regions = computed(() => territoryStore.regions);
const districts = computed(() => territoryStore.districts);
const cities = computed(() => territoryStore.cities);
const hotspots = computed(() => territoryStore.hotspots);
const recentActions = computed(() => territoryStore.recentActions);

const filteredDistricts = computed(() => {
  if (!selectedRegionId.value) return [];
  return districts.value.filter(d => d.regionId === selectedRegionId.value);
});

const filteredCities = computed(() => {
  if (!selectedDistrictId.value) return [];
  return cities.value.filter(c => c.districtId === selectedDistrictId.value);
});

const displayedHotspots = computed(() => {
  let result = [...territoryStore.filteredHotspots];

  // Apply business type filter
  if (businessFilter.value === 'legal') {
    result = result.filter(h => h.isLegal);
  } else if (businessFilter.value === 'illegal') {
    result = result.filter(h => !h.isLegal);
  }

  return result;
});

// Allocation preset definitions
const allocationPresets = computed(() => [
  {
    name: 'None',
    allocation: { crew: 0, weapons: 0, vehicles: 0 }
  },
  {
    name: 'Minimal',
    allocation: {
      crew: Math.min(1, availableCrew.value),
      weapons: 0,
      vehicles: 0
    }
  },
  {
    name: 'Balanced',
    allocation: {
      crew: Math.floor(availableCrew.value * 0.3),
      weapons: Math.floor(availableWeapons.value * 0.3),
      vehicles: Math.floor(availableVehicles.value * 0.3)
    }
  },
  {
    name: 'Aggressive',
    allocation: {
      crew: Math.floor(availableCrew.value * 0.5),
      weapons: Math.floor(availableWeapons.value * 0.7),
      vehicles: Math.floor(availableVehicles.value * 0.4)
    }
  },
  {
    name: 'All In',
    allocation: {
      crew: availableCrew.value,
      weapons: availableWeapons.value,
      vehicles: availableVehicles.value
    }
  }
]);

// Controlled hotspots
const controlledHotspots = computed(() => {
  return territoryStore.controlledHotspots;
});

const sortedControlledHotspots = computed(() => {
  const result = [...controlledHotspots.value];

  switch (empireSortBy.value) {
    case 'name':
      return result.sort((a, b) => a.name.localeCompare(b.name));
    case 'income':
      return result.sort((a, b) => b.income - a.income);
    case 'pending':
      return result.sort((a, b) => b.pendingCollection - a.pendingCollection);
    case 'defense':
      return result.sort((a, b) => b.defenseStrength - a.defenseStrength);
    case 'region':
      return result.sort((a, b) => {
        const cityA = cities.value.find(c => c.id === a.cityId);
        const cityB = cities.value.find(c => c.id === b.cityId);

        if (!cityA || !cityB) return 0;

        const districtA = districts.value.find(d => d.id === cityA.districtId);
        const districtB = districts.value.find(d => d.id === cityB.districtId);

        if (!districtA || !districtB) return 0;

        const regionA = regions.value.find(r => r.id === districtA.regionId);
        const regionB = regions.value.find(r => r.id === districtB.regionId);

        if (!regionA || !regionB) return 0;

        return regionA.name.localeCompare(regionB.name);
      });
    default:
      return result;
  }
});

const collectableBusinesses = computed(() => {
  return controlledHotspots.value.filter(h => h.pendingCollection > 0);
});

const hasCollectableBusiness = computed(() => {
  return collectableBusinesses.value.length > 0;
});

// Regional distribution
const regionsWithControlledHotspots = computed(() => {
  const result = [];

  for (const region of regions.value) {
    const districtsInRegion = districts.value.filter(d => d.regionId === region.id);
    const districtIds = districtsInRegion.map(d => d.id);

    const citiesInRegion = cities.value.filter(c => districtIds.includes(c.districtId));
    const cityIds = citiesInRegion.map(c => c.id);

    const hotspotsInRegion = hotspots.value.filter(h => cityIds.includes(h.cityId));
    const legalBusinessesInRegion = hotspotsInRegion.filter(h => h.isLegal);
    const controlledHotspotsInRegion = legalBusinessesInRegion.filter(h => isPlayerControlled(h));

    if (legalBusinessesInRegion.length > 0) {
      result.push({
        id: region.id,
        name: region.name,
        controlled: controlledHotspotsInRegion.length,
        total: legalBusinessesInRegion.length,
        controlPercentage: Math.round((controlledHotspotsInRegion.length / legalBusinessesInRegion.length) * 100)
      });
    }
  }

  return result.sort((a, b) => b.controlPercentage - a.controlPercentage);
});

// Player resources
const availableCrew = computed(() => playerStore.playerCrew);
const availableWeapons = computed(() => playerStore.playerWeapons);
const availableVehicles = computed(() => playerStore.playerVehicles);

// Computed properties for modals
const actionModalTitle = computed(() => {
  if (!selectedHotspot.value) return 'Territory Action';

  switch (selectedAction.value) {
    case TerritoryActionType.EXTORTION:
      return 'Extortion Operation';
    case TerritoryActionType.TAKEOVER:
      return 'Business Takeover';
    case TerritoryActionType.COLLECTION:
      return 'Resource Collection';
    case TerritoryActionType.DEFEND:
      return 'Defense Allocation';
    default:
      return `${selectedHotspot.value.name}`;
  }
});

const resultModalTitle = computed(() => {
  return actionSuccess.value ? 'Operation Successful' : 'Operation Failed';
});

// Success chance calculation - Updated to better match backend calculation
const successChance = computed(() => {
  if (!selectedHotspot.value || !selectedAction.value) return 0;

  // Calculate player strength based on the resources allocated
  const playerStrength = (actionResources.value.crew * 10) +
    (actionResources.value.weapons * 15) +
    (actionResources.value.vehicles * 20);

  // Base chance depends on action type
  let baseChance = 50;
  let opponentStrength = 0;

  switch (selectedAction.value) {
    case TerritoryActionType.EXTORTION:
      baseChance = 70;
      break;
    case TerritoryActionType.TAKEOVER:
      // Harder if already controlled - factor in defense strength
      if (selectedHotspot.value.controller) {
        baseChance = 50;
        opponentStrength = selectedHotspot.value.defenseStrength;
      } else {
        baseChance = 75;
      }
      break;
    case TerritoryActionType.COLLECTION:
      // Base chance high, but decreases with higher collection amounts
      const pendingAmount = selectedHotspot.value.pendingCollection;
      baseChance = Math.max(50, 95 - Math.floor(pendingAmount / 1000));
      break;
    case TerritoryActionType.DEFEND:
      baseChance = 100; // Always succeeds
      break;
  }

  // Adjust for player strength vs opponent strength
  if (playerStrength > 0) {
    if (opponentStrength > 0) {
      // For actions against opponents, compare strengths
      const strengthRatio = playerStrength / opponentStrength;

      if (strengthRatio >= 2.0) {
        baseChance += 20; // Major advantage
      } else if (strengthRatio >= 1.5) {
        baseChance += 15; // Significant advantage
      } else if (strengthRatio >= 1.0) {
        baseChance += 10; // Slight advantage
      } else if (strengthRatio >= 0.75) {
        baseChance += 5; // Nearly even
      } else if (strengthRatio >= 0.5) {
        baseChance -= 5; // Disadvantage
      } else if (strengthRatio >= 0.25) {
        baseChance -= 10; // Major disadvantage
      } else {
        baseChance -= 20; // Severe disadvantage
      }
    } else {
      // For actions without opposition, add based on allocated strength
      if (playerStrength >= 100) {
        baseChance += 20;
      } else if (playerStrength >= 75) {
        baseChance += 15;
      } else if (playerStrength >= 50) {
        baseChance += 10;
      } else if (playerStrength >= 25) {
        baseChance += 5;
      }
    }
  }

  // Cap between 5% and 95%
  return Math.max(5, Math.min(95, baseChance));
});

const actionWarning = computed(() => {
  if (!selectedAction.value || !selectedHotspot.value) return '';

  if (successChance.value < 30) {
    return 'This operation has a very low chance of success. Consider allocating more resources.';
  }

  switch (selectedAction.value) {
    case TerritoryActionType.TAKEOVER:
      if (selectedHotspot.value.controller) {
        return 'This business is controlled by a rival. Takeover attempts may lead to retaliation.';
      }
      break;
    case TerritoryActionType.EXTORTION:
      return 'Extortion generates heat, which can attract law enforcement attention.';
    case TerritoryActionType.COLLECTION:
      if (selectedHotspot.value.pendingCollection > 5000) {
        return 'Large collection amounts may attract unwanted attention and generate heat.';
      }
      break;
  }

  return '';
});

const canPerformAction = computed(() => {
  return selectedHotspot.value !== null &&
    selectedAction.value !== null &&
    (actionResources.value.crew > 0 ||
      actionResources.value.weapons > 0 ||
      actionResources.value.vehicles > 0);
});

// This computed property ensures reactivity with the timer in the territory store
const timerRefreshCounter = computed(() => territoryStore.timerRefreshCounter);

// Helper function to use the centralized timer from the store
function formatTimeRemaining(hotspotId: string): string {
  // Access the refresh counter to ensure reactivity
  // This is crucial for the timer to update in real-time
  const _ = timerRefreshCounter.value;
  return territoryStore.getTimeRemaining(hotspotId);
}

// Function to check if income is coming soon
function isIncomeSoon(hotspotId: string): boolean {
  // Access the refresh counter to ensure reactivity
  const _ = timerRefreshCounter.value;
  return territoryStore.isIncomeSoon(hotspotId);
}

// Watch for filter changes
watch(selectedRegionId, (newValue) => {
  territoryStore.selectRegion(newValue);
  selectedDistrictId.value = null;
  selectedCityId.value = null;
});

watch(selectedDistrictId, (newValue) => {
  territoryStore.selectDistrict(newValue);
  selectedCityId.value = null;
});

watch(selectedCityId, (newValue) => {
  territoryStore.selectCity(newValue);
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

function isPlayerControlled(hotspot: Hotspot): boolean {
  return hotspot.controller === playerStore.profile?.id;
}

function isRivalControlled(hotspot: Hotspot): boolean {
  return hotspot.controller !== null && hotspot.controller !== playerStore.profile?.id;
}

function getHotspotStatus(hotspot: Hotspot): string {
  if (!hotspot.isLegal) {
    return 'Illegal Business';
  }

  if (isPlayerControlled(hotspot)) {
    return 'Controlled by You';
  }

  if (hotspot.controller) {
    return `Controlled by ${hotspot.controllerName || 'Rival'}`;
  }

  return 'Uncontrolled';
}

function getHotspotLocation(hotspot: Hotspot, detailed: boolean = false): string {
  const city = cities.value.find(c => c.id === hotspot.cityId);
  if (!city) return 'Unknown';

  if (!detailed) return city.name;

  const district = districts.value.find(d => d.id === city.districtId);
  if (!district) return city.name;

  const region = regions.value.find(r => r.id === district.regionId);
  if (!region) return `${city.name}, ${district.name}`;

  return `${city.name}, ${district.name}, ${region.name}`;
}

function getHotspotName(hotspotId: string): string {
  const hotspot = hotspots.value.find(h => h.id === hotspotId);
  return hotspot ? hotspot.name : 'Unknown Business';
}

function getDefenseClass(defense: number): string {
  if (defense >= 80) return 'high';
  if (defense >= 40) return 'medium';
  return 'low';
}

function getDefenseLabel(defense: number): string {
  if (defense >= 80) return 'Strong';
  if (defense >= 40) return 'Medium';
  return 'Weak';
}

function getSuccessChanceClass(chance: number): string {
  if (chance >= 70) return 'high';
  if (chance >= 40) return 'medium';
  return 'low';
}

function getActionTypeLabel(actionType: TerritoryActionType | null): string {
  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      return 'Extortion';
    case TerritoryActionType.TAKEOVER:
      return 'Takeover';
    case TerritoryActionType.COLLECTION:
      return 'Collection';
    case TerritoryActionType.DEFEND:
      return 'Defense';
    default:
      return 'Action';
  }
}

function getActionIcon(actionType: TerritoryActionType | null): string {
  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      return '💰';
    case TerritoryActionType.TAKEOVER:
      return '🏢';
    case TerritoryActionType.COLLECTION:
      return '💼';
    case TerritoryActionType.DEFEND:
      return '🛡️';
    default:
      return '❓';
  }
}

function getActionButtonLabel(actionType: TerritoryActionType | null): string {
  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      return 'Extort';
    case TerritoryActionType.TAKEOVER:
      return 'Take Over';
    case TerritoryActionType.COLLECTION:
      return 'Collect';
    case TerritoryActionType.DEFEND:
      return 'Defend';
    default:
      return 'Execute';
  }
}

function getActionDescription(actionType: TerritoryActionType | null, hotspot: Hotspot | null): string {
  if (!actionType || !hotspot) return '';

  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      return 'Extort money from this illegal business. This will generate income but increases heat, potentially attracting law enforcement attention.';

    case TerritoryActionType.TAKEOVER:
      if (hotspot.controller) {
        return `Attempt to take control of this business from ${isRivalControlled(hotspot) ? hotspot.controllerName || 'a rival' : 'its current owner'}. Higher resource allocation increases your chance of success.`;
      } else {
        return 'Take control of this unowned business. Even uncontrolled businesses require some resources to take over.';
      }

    case TerritoryActionType.COLLECTION:
      return `Collect $${formatNumber(hotspot.pendingCollection)} in accumulated income. Larger collections may require more resources and could attract attention.`;

    case TerritoryActionType.DEFEND:
      return 'Allocate resources to strengthen this business against rival takeover attempts. Resources assigned here are at risk if rivals attempt a takeover.';

    default:
      return '';
  }
}

// Updated to reflect more realistic backend calculations
function getPotentialReward(actionType: TerritoryActionType | null, hotspot: Hotspot | null): number {
  if (!actionType || !hotspot) return 0;

  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      // Extortion rewards based on resources committed and base value
      const baseGain = 500 + (Math.floor(Math.random() * 11) * 100); // $500-$1500 base
      const resourceMultiplier = 1.0 + (
        (actionResources.value.crew +
          actionResources.value.weapons * 2 +
          actionResources.value.vehicles * 3) / 20.0
      );
      return Math.round(baseGain * resourceMultiplier);

    case TerritoryActionType.COLLECTION:
      return hotspot.pendingCollection;

    default:
      return 0;
  }
}

// Updated to reflect backend heat generation values
function getPotentialHeat(actionType: TerritoryActionType | null): number {
  if (!actionType) return 0;

  switch (actionType) {
    case TerritoryActionType.EXTORTION:
      return 5 + Math.floor(Math.random() * 6); // 5-10 heat
    case TerritoryActionType.TAKEOVER:
      return 3 + Math.floor(Math.random() * 5); // 3-7 heat
    case TerritoryActionType.COLLECTION:
      return 1 + Math.floor(Math.random() * 3); // 1-3 heat
    default:
      return 0;
  }
}

function formatTimeAgo(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.round(diffMs / 1000);
  const diffMin = Math.round(diffSec / 60);
  const diffHour = Math.round(diffMin / 60);
  const diffDay = Math.round(diffHour / 24);

  if (diffSec < 60) {
    return 'just now';
  } else if (diffMin < 60) {
    return `${diffMin} minute${diffMin > 1 ? 's' : ''} ago`;
  } else if (diffHour < 24) {
    return `${diffHour} hour${diffHour > 1 ? 's' : ''} ago`;
  } else {
    return `${diffDay} day${diffDay > 1 ? 's' : ''} ago`;
  }
}

// Helper methods for resource allocation UI
function getResourceIcon(resource: 'crew' | 'weapons' | 'vehicles'): string {
  const icons = {
    crew: '👥',
    weapons: '🔫',
    vehicles: '🚗'
  };
  return icons[resource];
}

function getResourceLabel(resource: 'crew' | 'weapons' | 'vehicles'): string {
  return resource.charAt(0).toUpperCase() + resource.slice(1);
}

function getAvailableResource(resource: 'crew' | 'weapons' | 'vehicles'): number {
  const availableResources = {
    crew: availableCrew.value,
    weapons: availableWeapons.value,
    vehicles: availableVehicles.value
  };
  return availableResources[resource];
}

// Methods to handle resource allocation UI interactions
function applyAllocationPreset(preset: { name: string, allocation: ActionResources }): void {
  actionResources.value = { ...preset.allocation };
}

function adjustResource(resource: 'crew' | 'weapons' | 'vehicles', amount: number): void {
  const availableResource = getAvailableResource(resource);

  let newValue = actionResources.value[resource] + amount;
  newValue = Math.max(0, Math.min(availableResource, newValue));

  actionResources.value[resource] = newValue;
}

function validateResourceInput(resourceType: 'crew' | 'weapons' | 'vehicles'): void {
  const maxValue = getAvailableResource(resourceType);
  const currentValue = actionResources.value[resourceType];

  // Ensure value is not negative
  if (currentValue < 0) {
    actionResources.value[resourceType] = 0;
  }
  // Ensure value does not exceed maximum
  else if (currentValue > maxValue) {
    actionResources.value[resourceType] = maxValue;
  }
}

// Action functions
function selectHotspot(hotspot: Hotspot) {
  territoryStore.selectHotspot(hotspot.id);
  selectedHotspot.value = hotspot;
}

function openActionModal(hotspot: Hotspot, action: TerritoryActionType) {
  selectHotspot(hotspot);
  selectedAction.value = action;

  // Set default resource allocation based on action type - optimized based on backend expectations
  switch (action) {
    case TerritoryActionType.EXTORTION:
      actionResources.value = {
        crew: Math.min(2, availableCrew.value),
        weapons: Math.min(1, availableWeapons.value),
        vehicles: Math.min(1, availableVehicles.value)
      };
      break;
    case TerritoryActionType.TAKEOVER:
      // For takeovers, more resources are needed, especially against defended businesses
      const defenseStrength = hotspot.defenseStrength || 0;
      const recommendedCrew = defenseStrength > 0 ? Math.min(3, availableCrew.value) : Math.min(2, availableCrew.value);
      const recommendedWeapons = defenseStrength > 0 ? Math.min(2, availableWeapons.value) : Math.min(1, availableWeapons.value);
      const recommendedVehicles = defenseStrength > 0 ? Math.min(1, availableVehicles.value) : 0;

      actionResources.value = {
        crew: recommendedCrew,
        weapons: recommendedWeapons,
        vehicles: recommendedVehicles
      };
      break;
    case TerritoryActionType.COLLECTION:
      // For collections, resource needs increase with amount being collected
      const pendingAmount = hotspot.pendingCollection;
      if (pendingAmount > 5000) {
        actionResources.value = {
          crew: Math.min(2, availableCrew.value),
          weapons: Math.min(1, availableWeapons.value),
          vehicles: 0
        };
      } else {
        actionResources.value = {
          crew: Math.min(1, availableCrew.value),
          weapons: 0,
          vehicles: 0
        };
      }
      break;
    case TerritoryActionType.DEFEND:
      actionResources.value = {
        crew: Math.min(2, availableCrew.value),
        weapons: Math.min(1, availableWeapons.value),
        vehicles: 0
      };
      break;
  }

  showActionModal.value = true;
}

function closeActionModal() {
  showActionModal.value = false;
  selectedAction.value = null;
  actionResources.value = { crew: 0, weapons: 0, vehicles: 0 };
}

async function performAction() {
  if (!selectedHotspot.value || !selectedAction.value || isPerformingAction.value) return;

  isPerformingAction.value = true;

  try {
    const actResult = await territoryStore.performTerritoryAction(
      selectedAction.value,
      selectedHotspot.value.id,
      actionResources.value
    );

    console.log('Result value:', actResult)

    if (actResult) {
      actionResult.value = actResult;
      actionSuccess.value = actResult.success;

      // Close action modal and show result
      showActionModal.value = false;
      showResultModal.value = true;

      console.log('Perform territory action result', actionResult.value);
    }
  } catch (error) {
    console.error('Error performing territory action:', error);
  } finally {
    isPerformingAction.value = false;
  }
}

function closeResultModal() {
  showResultModal.value = false;

  // Reset action resources
  actionResources.value = { crew: 0, weapons: 0, vehicles: 0 };

  // Refresh data
  territoryStore.updateFilteredHotspots();
}

async function collectAllPending() {
  if (isCollecting.value || collectableBusinesses.value.length === 0) return;

  isCollecting.value = true;

  try {
    const result = await territoryStore.collectAllHotspotIncome();

    if (result) {
      // Show a success notification or message
      actionResult.value = {
        success: true,
        moneyGained: result.collectionResult.collectedAmount,
        message: `Successfully collected $${formatNumber(result.collectionResult.collectedAmount)} from all your businesses.`
      };

      actionSuccess.value = true;
      showResultModal.value = true;
    }
  } catch (error) {
    console.error('Error collecting all pending resources:', error);
  } finally {
    isCollecting.value = false;
  }
}

function resetFilters() {
  selectedRegionId.value = null;
  selectedDistrictId.value = null;
  selectedCityId.value = null;
  businessFilter.value = 'all';
  territoryStore.selectRegion(null);
}

// Filter change handlers
function onRegionChange() {
  territoryStore.selectRegion(selectedRegionId.value);
}

function onDistrictChange() {
  territoryStore.selectDistrict(selectedDistrictId.value);
}

function onCityChange() {
  territoryStore.selectCity(selectedCityId.value);
}

function navigateToTab(tab: 'empire' | 'explore' | 'recent') {
  router.push({ path: '/territory', query: { tab } });
}

// State for tracking which hotspot is being collected
const collectingHotspotId = ref<string | null>(null);

// Function to collect income from a specific hotspot
async function collectHotspotIncome(hotspotId: string) {
  if (collectingHotspotId.value === hotspotId) return;

  collectingHotspotId.value = hotspotId;

  try {
    const result = await territoryStore.collectHotspotIncome(hotspotId);

    if (result) {
      // Check the structure of the result and safely extract the collected amount
      let collectedAmount = 0;
      if (result.collectionResult && typeof result.collectionResult.collectedAmount === 'number') {
        collectedAmount = result.collectionResult.collectedAmount;
      }

      // Show result modal
      actionResult.value = {
        success: true,
        moneyGained: collectedAmount,
        message: result.gameMessage?.message || `Successfully collected ${formatNumber(collectedAmount)}`
      };

      actionSuccess.value = true;
      showResultModal.value = true;
    }
  } catch (error) {
    console.error('Error collecting hotspot income:', error);
  } finally {
    collectingHotspotId.value = null;
  }
}

// Tooltip methods
function showTooltip(hotspot: Hotspot, event: MouseEvent) {
  // Don't show tooltip when in modal mode
  if (showActionModal.value || showResultModal.value || showDetailModal.value) return;

  tooltipHotspot.value = hotspot;

  // Position tooltip near the info icon but with slight offset
  tooltipPosition.value = {
    x: event.clientX + 10,
    y: event.clientY + 10
  };

  tooltipVisible.value = true;
}

function hideTooltip() {
  tooltipVisible.value = false;
}

// Detail modal methods
function openDetailModal(hotspot: Hotspot) {
  hideTooltip();
  detailModalHotspot.value = hotspot;
  showDetailModal.value = true;
}

// Method to forward action request from detail modal to action modal
function handleOpenActionModal(hotspot: Hotspot, action: TerritoryActionType) {
  openActionModal(hotspot, action);
}

// OnMounted hook
onMounted(async () => {
  isLoading.value = true;

  if (!playerStore.profile) {
    await playerStore.fetchProfile();
  }

  if (territoryStore.regions.length === 0) {
    await territoryStore.fetchTerritoryData();
  } else {
    // If data already exists, ensure all income timers are properly set
    territoryStore.ensureAllIncomeTimes();

    // Make sure timer is started (this is crucial)
    territoryStore.startIncomeTimer();
  }

  if (territoryStore.recentActions.length === 0) {
    await territoryStore.fetchRecentActions();
  }

  isLoading.value = false;

  // Log a message to confirm initialization
  console.log('TerritoryView mounted and initialized. Timer active:', !!territoryStore.incomeTimerInterval);
});

// Set up cleanup on component unmount
onBeforeUnmount(() => {
  console.log('TerritoryView unmounting, stopping timer');
  // Stop the income timer in the store
  territoryStore.stopIncomeTimer();
});
</script>

<template>
  <div class="territory-view">
    <div class="page-header">
      <div class="page-title">
        <h2>Territory Control</h2>
        <p class="subtitle">Expand your criminal empire through territory dominance.</p>
      </div>

      <div class="territory-stats">
        <div class="stat-card">
          <div class="stat-icon">🏢</div>
          <div class="stat-content">
            <div class="stat-value">{{ playerStore.controlledHotspots }}</div>
            <div class="stat-label">Controlled</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon">💵</div>
          <div class="stat-content">
            <div class="stat-value">${{ formatNumber(playerStore.hourlyRevenue) }}</div>
            <div class="stat-label">Hourly Income</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon">💰</div>
          <div class="stat-content">
            <div class="stat-value">${{ formatNumber(playerStore.pendingCollections) }}</div>
            <div class="stat-label">Pending</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="territory-tabs">
      <button class="tab-btn" :class="{ active: activeTab === 'empire' }" @click="navigateToTab('empire')">
        <span class="tab-icon">👑</span>
        <span class="tab-text">My Empire</span>
      </button>
      <button class="tab-btn" :class="{ active: activeTab === 'explore' }" @click="navigateToTab('explore')">
        <span class="tab-icon">🔍</span>
        <span class="tab-text">Explore Territory</span>
      </button>
      <button class="tab-btn" :class="{ active: activeTab === 'recent' }" @click="navigateToTab('recent')">
        <span class="tab-icon">📜</span>
        <span class="tab-text">Recent Activity</span>
      </button>
    </div>

    <!-- My Empire Tab Content -->
    <div v-if="activeTab === 'empire'" class="tab-content empire-tab">
      <div class="empire-header">
        <h3>Your Controlled Territories</h3>

        <div class="empire-actions">
          <BaseButton v-if="hasCollectableBusiness" @click="collectAllPending" :loading="isCollecting"
            variant="secondary">
            <span class="btn-icon">💼</span>
            Collect All ({{ collectableBusinesses.length }})
          </BaseButton>
        </div>
      </div>

      <!-- Section Separator -->
      <div class="section-separator"></div>

      <!-- Controlled hotspots grid -->
      <div class="hotspots-section">
        <div class="section-header">
          <h4>Your Businesses</h4>

          <div class="section-filters">
            <div class="filter-group">
              <label>Sort by:</label>
              <select v-model="empireSortBy">
                <option value="name">Name</option>
                <option value="income">Income</option>
                <option value="pending">Pending Collections</option>
                <option value="defense">Defense Strength</option>
                <option value="region">Region</option>
              </select>
            </div>
          </div>
        </div>

        <div class="hotspots-grid empire-grid">
          <div v-for="hotspot in sortedControlledHotspots" :key="hotspot.id" class="hotspot-card controlled"
            :class="{ 'has-pending': hotspot.pendingCollection > 0 }" @click="openDetailModal(hotspot)">

            <div class="card-badge" v-if="hotspot.pendingCollection > 0">
              <span class="badge-icon">💰</span>
              ${{ formatNumber(hotspot.pendingCollection) }}
            </div>

            <div class="hotspot-header">
              <div class="hotspot-title-area">
                <h3>{{ hotspot.name }}</h3>
                <div class="hotspot-type">{{ hotspot.type }}</div>
              </div>
              <div class="hotspot-info-icon" @mouseover.stop="showTooltip(hotspot, $event)"
                @mouseleave.stop="hideTooltip">
                <span class="info-icon">ℹ️</span>
              </div>
            </div>

            <div class="hotspot-details">
              <div class="detail-row">
                <div class="detail-item">
                  <span class="detail-label">Location:</span>
                  <span class="detail-value">{{ getHotspotLocation(hotspot) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">Business:</span>
                  <span class="detail-value">{{ hotspot.businessType }}</span>
                </div>
              </div>

              <div class="detail-row">
                <div class="detail-item">
                  <span class="detail-label">Income:</span>
                  <span class="detail-value">${{ formatNumber(hotspot.income) }}/hr</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">Defense:</span>
                  <span class="detail-value defense" :class="getDefenseClass(hotspot.defenseStrength)">
                    {{ getDefenseLabel(hotspot.defenseStrength) }}
                  </span>
                </div>
              </div>

              <!-- Updated next income display using store getters -->
              <div class="detail-row income-timer">
                <div class="detail-item">
                  <span class="detail-label">Next Income:</span>
                  <span class="detail-value countdown" :class="{ 'soon': isIncomeSoon(hotspot.id) }">
                    {{ formatTimeRemaining(hotspot.id) }}
                  </span>
                </div>
                <div class="detail-item" v-if="hotspot.pendingCollection > 0">
                  <span class="detail-label">Pending:</span>
                  <span class="detail-value income">${{ formatNumber(hotspot.pendingCollection) }}</span>
                </div>
              </div>

              <div class="detail-row defense-allocation">
                <div class="resource-allocation">
                  <div class="resource-item" v-if="hotspot.crew > 0">
                    <span class="resource-icon">👥</span>
                    <span class="resource-value">{{ hotspot.crew }}</span>
                  </div>
                  <div class="resource-item" v-if="hotspot.weapons > 0">
                    <span class="resource-icon">🔫</span>
                    <span class="resource-value">{{ hotspot.weapons }}</span>
                  </div>
                  <div class="resource-item" v-if="hotspot.vehicles > 0">
                    <span class="resource-icon">🚗</span>
                    <span class="resource-value">{{ hotspot.vehicles }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="hotspot-footer">
              <BaseButton v-if="hotspot.pendingCollection > 0" variant="primary" small
                @click.stop="collectHotspotIncome(hotspot.id)" :loading="collectingHotspotId === hotspot.id">
                Collect
              </BaseButton>
              <BaseButton variant="secondary" small @click.stop="openActionModal(hotspot, TerritoryActionType.DEFEND)">
                Defend
              </BaseButton>
            </div>
          </div>

          <div v-if="controlledHotspots.length === 0" class="empty-state">
            <div class="empty-icon">🏙️</div>
            <h4>No Controlled Businesses</h4>
            <p>Your criminal empire needs a foundation. Start by taking over a business in the Explore tab.</p>
            <BaseButton @click="activeTab = 'explore'">Explore Territory</BaseButton>
          </div>
        </div>
      </div>

      <!-- Section Separator -->
      <div class="section-separator">
        <!-- <span class="separator-text">Regional Influence</span> -->
      </div>

      <!-- Region distribution chart -->
      <div class="empire-regions-overview">
        <div class="overview-header">
          <h4>Regional Distribution</h4>
          <span class="help-text">Your territorial influence across the city</span>
        </div>
        <div class="regions-chart">
          <div v-for="region in regionsWithControlledHotspots" :key="region.id" class="region-bar">
            <div class="region-name">{{ region.name }}</div>
            <div class="bar-wrapper">
              <div class="bar-fill" :style="{ width: `${region.controlPercentage}%` }"
                :class="{ 'powerful': region.controlPercentage > 60 }"></div>
              <span class="bar-value">{{ region.controlled }}/{{ region.total }}</span>
            </div>
          </div>

          <div v-if="regionsWithControlledHotspots.length === 0" class="no-regions">
            <p>You don't control any territories yet.</p>
            <p>Explore the city to find businesses to take over.</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Explore Tab Content -->
    <div v-else-if="activeTab === 'explore'" class="tab-content explore-tab">
      <div class="explore-header">
        <div class="filters">
          <div class="filter-group region-filter">
            <label>Region:</label>
            <select v-model="selectedRegionId" @change="onRegionChange">
              <option :value="null">All Regions</option>
              <option v-for="region in regions" :key="region.id" :value="region.id">
                {{ region.name }}
              </option>
            </select>
          </div>

          <div class="filter-group">
            <label>District:</label>
            <select v-model="selectedDistrictId" @change="onDistrictChange" :disabled="!selectedRegionId">
              <option :value="null">All Districts</option>
              <option v-for="district in filteredDistricts" :key="district.id" :value="district.id">
                {{ district.name }}
              </option>
            </select>
          </div>

          <div class="filter-group">
            <label>City:</label>
            <select v-model="selectedCityId" @change="onCityChange" :disabled="!selectedDistrictId">
              <option :value="null">All Cities</option>
              <option v-for="city in filteredCities" :key="city.id" :value="city.id">
                {{ city.name }}
              </option>
            </select>
          </div>
        </div>

        <div class="filter-tabs">
          <button class="filter-tab" :class="{ active: businessFilter === 'all' }" @click="businessFilter = 'all'">
            All Businesses
          </button>
          <button class="filter-tab" :class="{ active: businessFilter === 'legal' }" @click="businessFilter = 'legal'">
            Legal Businesses
          </button>
          <button class="filter-tab" :class="{ active: businessFilter === 'illegal' }"
            @click="businessFilter = 'illegal'">
            Illegal Operations
          </button>
        </div>

        <div class="view-toggle">
          <button class="toggle-btn" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'">
            <span class="toggle-icon">▧</span>
            Grid
          </button>
          <button class="toggle-btn" :class="{ active: viewMode === 'map' }" @click="viewMode = 'map'">
            <span class="toggle-icon">🗺️</span>
            Map
          </button>
        </div>
      </div>

      <div class="territory-content">
        <!-- Grid View -->
        <div v-if="viewMode === 'grid'" class="hotspots-grid explore-grid">
          <div v-for="hotspot in displayedHotspots" :key="hotspot.id" class="hotspot-card" :class="{
            'controlled': isPlayerControlled(hotspot),
            'rival-controlled': isRivalControlled(hotspot),
            'illegal': !hotspot.isLegal
          }" @click="openDetailModal(hotspot)">

            <div v-if="!hotspot.isLegal" class="card-badge illegal">
              <span class="badge-icon">⚠️</span>
              Illegal
            </div>
            <div v-else-if="isRivalControlled(hotspot)" class="card-badge rival">
              <span class="badge-icon">⚔️</span>
              Rival
            </div>
            <div v-else-if="isPlayerControlled(hotspot)" class="card-badge controlled">
              <span class="badge-icon">👑</span>
              Yours
            </div>

            <div class="hotspot-header">
              <div class="hotspot-title-area">
                <h3>{{ hotspot.name }}</h3>
                <div class="hotspot-type">{{ hotspot.type }}</div>
              </div>
              <div class="hotspot-info-icon" @mouseover.stop="showTooltip(hotspot, $event)"
                @mouseleave.stop="hideTooltip">
                <span class="info-icon">ℹ️</span>
              </div>
            </div>

            <div class="hotspot-details">
              <div class="detail-row">
                <div class="detail-item">
                  <span class="detail-label">Location:</span>
                  <span class="detail-value">{{ getHotspotLocation(hotspot) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">Business:</span>
                  <span class="detail-value">{{ hotspot.businessType }}</span>
                </div>
              </div>

              <div class="detail-row">
                <div class="detail-item">
                  <span class="detail-label">Status:</span>
                  <span class="detail-value status" :class="{
                    'controlled': isPlayerControlled(hotspot),
                    'rival': isRivalControlled(hotspot),
                    'neutral': !hotspot.controller
                  }">
                    {{ getHotspotStatus(hotspot) }}
                  </span>
                </div>
                <div class="detail-item" v-if="hotspot.isLegal">
                  <span class="detail-label">Income:</span>
                  <span class="detail-value">${{ formatNumber(hotspot.income) }}/hr</span>
                </div>
              </div>

              <div class="detail-row" v-if="isRivalControlled(hotspot)">
                <div class="detail-item defense-strength">
                  <span class="detail-label">Defense:</span>
                  <span class="detail-value defense" :class="getDefenseClass(hotspot.defenseStrength)">
                    {{ getDefenseLabel(hotspot.defenseStrength) }}
                  </span>
                </div>
              </div>
            </div>

            <div class="hotspot-footer">
              <BaseButton v-if="!hotspot.isLegal" variant="danger" small
                @click.stop="openActionModal(hotspot, TerritoryActionType.EXTORTION)">
                Extort
              </BaseButton>
              <BaseButton v-else-if="!isPlayerControlled(hotspot)"
                :variant="isRivalControlled(hotspot) ? 'danger' : 'primary'" small
                @click.stop="openActionModal(hotspot, TerritoryActionType.TAKEOVER)">
                Takeover
              </BaseButton>
              <BaseButton v-else-if="isPlayerControlled(hotspot) && hotspot.pendingCollection > 0" variant="primary"
                small @click.stop="openActionModal(hotspot, TerritoryActionType.COLLECTION)">
                Collect
              </BaseButton>
              <BaseButton v-else-if="isPlayerControlled(hotspot)" variant="secondary" small
                @click.stop="openActionModal(hotspot, TerritoryActionType.DEFEND)">
                Defend
              </BaseButton>
            </div>
          </div>

          <div v-if="displayedHotspots.length === 0" class="empty-state">
            <div class="empty-icon">🔍</div>
            <h4>No Businesses Found</h4>
            <p>No businesses match your current filters.</p>
            <BaseButton @click="resetFilters">Reset Filters</BaseButton>
          </div>
        </div>

        <!-- Map View -->
        <div v-else-if="viewMode === 'map'" class="territory-map">
          <div class="map-container">
            <div class="map-placeholder">
              <div class="placeholder-icon">🗺️</div>
              <h4>Map View Coming Soon</h4>
              <p>Our cartographers are still charting the city's criminal landscape.</p>
              <BaseButton @click="viewMode = 'grid'">Switch to Grid View</BaseButton>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recent Activity Tab Content -->
    <div v-else-if="activeTab === 'recent'" class="tab-content recent-tab">
      <div class="recent-header">
        <h3>Recent Territory Operations</h3>
        <p class="subtitle">The last 20 actions performed in your criminal empire</p>
      </div>

      <div class="timeline">
        <div v-if="recentActions.length > 0" class="timeline-container">
          <div v-for="(action, index) in recentActions" :key="action.id" class="timeline-item"
            :class="{ 'success': action.result?.success, 'failure': action.result && !action.result.success }">
            <div class="timeline-icon">
              {{ getActionIcon(action.type) }}
            </div>

            <div class="timeline-content">
              <div class="timeline-header">
                <div class="action-type">{{ getActionTypeLabel(action.type) }}</div>
                <div class="action-time">{{ formatTimeAgo(action.timestamp) }}</div>
              </div>

              <div class="action-target">
                {{ getHotspotName(action.hotspotId) }}
              </div>

              <div class="action-details">
                <div class="resources-used">
                  <div class="resource" v-if="action.resources.crew > 0">
                    <span class="resource-icon">👥</span>
                    <span class="resource-value">{{ action.resources.crew }}</span>
                  </div>
                  <div class="resource" v-if="action.resources.weapons > 0">
                    <span class="resource-icon">🔫</span>
                    <span class="resource-value">{{ action.resources.weapons }}</span>
                  </div>
                  <div class="resource" v-if="action.resources.vehicles > 0">
                    <span class="resource-icon">🚗</span>
                    <span class="resource-value">{{ action.resources.vehicles }}</span>
                  </div>
                </div>

                <div class="action-result" v-if="action.result">
                  <div class="result-message">{{ action.result.message }}</div>

                  <div class="result-effects">
                    <div v-if="action.result.moneyGained" class="effect positive">
                      +${{ formatNumber(action.result.moneyGained) }}
                    </div>
                    <div v-if="action.result.moneyLost" class="effect negative">
                      -${{ formatNumber(action.result.moneyLost) }}
                    </div>
                    <div v-if="action.result.respectGained" class="effect positive">
                      +{{ action.result.respectGained }} Respect
                    </div>
                    <div v-if="action.result.heatGenerated" class="effect negative">
                      +{{ action.result.heatGenerated }} Heat
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="empty-state">
          <div class="empty-icon">📜</div>
          <h4>No Recent Activity</h4>
          <p>You haven't performed any territory actions yet.</p>
          <BaseButton @click="activeTab = 'explore'">Explore Territory</BaseButton>
        </div>
      </div>
    </div>

    <!-- Action Modal -->
    <BaseModal v-model="showActionModal" :title="actionModalTitle" class="action-modal">
      <div v-if="selectedHotspot" class="action-modal-content">
        <div class="hotspot-summary">
          <div class="summary-title">
            <h3>{{ selectedHotspot.name }}</h3>
            <div class="hotspot-type">{{ selectedHotspot.type }} - {{ selectedHotspot.businessType }}</div>
          </div>

          <div class="summary-location">
            <div class="location-icon">📍</div>
            <div class="location-text">{{ getHotspotLocation(selectedHotspot, true) }}</div>
          </div>

          <div class="summary-details">
            <div class="summary-column">
              <div class="summary-item" v-if="selectedHotspot.isLegal">
                <div class="item-icon">💵</div>
                <div class="item-details">
                  <div class="item-label">Income:</div>
                  <div class="item-value">${{ formatNumber(selectedHotspot.income) }}/hr</div>
                </div>
              </div>

              <div class="summary-item"
                v-if="isPlayerControlled(selectedHotspot) && selectedHotspot.pendingCollection > 0">
                <div class="item-icon">💰</div>
                <div class="item-details">
                  <div class="item-label">Pending Collection:</div>
                  <div class="item-value">${{ formatNumber(selectedHotspot.pendingCollection) }}</div>
                </div>
              </div>
            </div>

            <div class="summary-column">
              <div class="summary-item" v-if="selectedHotspot.controller">
                <div class="item-icon">👑</div>
                <div class="item-details">
                  <div class="item-label">Controller:</div>
                  <div class="item-value"
                    :class="{ 'controlled': isPlayerControlled(selectedHotspot), 'rival': isRivalControlled(selectedHotspot) }">
                    {{ isPlayerControlled(selectedHotspot) ? 'You' : selectedHotspot.controllerName }}
                  </div>
                </div>
              </div>

              <div class="summary-item" v-if="selectedHotspot.controller">
                <div class="item-icon">🛡️</div>
                <div class="item-details">
                  <div class="item-label">Defense:</div>
                  <div class="item-value defense" :class="getDefenseClass(selectedHotspot.defenseStrength)">
                    {{ getDefenseLabel(selectedHotspot.defenseStrength) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="action-info">
          <div class="action-header">
            <div class="action-icon">{{ getActionIcon(selectedAction) }}</div>
            <div class="action-name">{{ getActionTypeLabel(selectedAction) }}</div>
          </div>

          <div class="action-description">
            {{ getActionDescription(selectedAction, selectedHotspot) }}
          </div>
        </div>

        <div class="resource-allocation">
          <h4>Allocate Resources</h4>

          <div class="resource-allocation-ui">
            <div class="allocation-presets">
              <h5>Quick Allocation:</h5>
              <div class="preset-buttons">
                <button v-for="preset in allocationPresets" :key="preset.name" @click="applyAllocationPreset(preset)"
                  :class="{
                    'active':
                      actionResources.crew === preset.allocation.crew &&
                      actionResources.weapons === preset.allocation.weapons &&
                      actionResources.vehicles === preset.allocation.vehicles
                  }" class="preset-button">
                  {{ preset.name }}
                </button>
              </div>
            </div>

            <div class="resource-controls">
              <!-- Crew Resource Control -->
              <div class="resource-control">
                <div class="control-header">
                  <div class="resource-label">
                    <span class="resource-icon">👥</span>
                    <label>Crew</label>
                  </div>
                  <span class="resource-available">Current Allocation: {{ actionResources.crew }} / {{ availableCrew
                    }}</span>
                </div>

                <div class="control-actions">
                  <div class="adjustment-buttons">
                    <button class="adjust-btn large-adjust" @click="adjustResource('crew', -10)"
                      :disabled="actionResources.crew < 10">-10</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('crew', -5)"
                      :disabled="actionResources.crew < 5">-5</button>
                    <button class="adjust-btn" @click="adjustResource('crew', -1)"
                      :disabled="actionResources.crew <= 0">-</button>
                    <div class="number-input">
                      <input type="number" v-model.number="actionResources.crew" :min="0" :max="availableCrew"
                        @input="validateResourceInput('crew')" />
                    </div>
                    <button class="adjust-btn" @click="adjustResource('crew', 1)"
                      :disabled="actionResources.crew >= availableCrew">+</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('crew', 5)"
                      :disabled="actionResources.crew + 5 > availableCrew">+5</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('crew', 10)"
                      :disabled="actionResources.crew + 10 > availableCrew">+10</button>
                  </div>
                </div>

                <div class="resource-allocation-bar">
                  <div class="allocation-fill"
                    :style="{ width: `${(actionResources.crew / Math.max(1, availableCrew)) * 100}%` }"></div>
                </div>
              </div>

              <!-- Weapons Resource Control -->
              <div class="resource-control">
                <div class="control-header">
                  <div class="resource-label">
                    <span class="resource-icon">🔫</span>
                    <label>Weapons</label>
                  </div>
                  <span class="resource-available">Current Allocation: {{ actionResources.weapons }} / {{
                    availableWeapons }}</span>
                </div>

                <div class="control-actions">
                  <div class="adjustment-buttons">
                    <button class="adjust-btn large-adjust" @click="adjustResource('weapons', -10)"
                      :disabled="actionResources.weapons < 10">-10</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('weapons', -5)"
                      :disabled="actionResources.weapons < 5">-5</button>
                    <button class="adjust-btn" @click="adjustResource('weapons', -1)"
                      :disabled="actionResources.weapons <= 0">-</button>
                    <div class="number-input">
                      <input type="number" v-model.number="actionResources.weapons" :min="0" :max="availableWeapons"
                        @input="validateResourceInput('weapons')" />
                    </div>
                    <button class="adjust-btn" @click="adjustResource('weapons', 1)"
                      :disabled="actionResources.weapons >= availableWeapons">+</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('weapons', 5)"
                      :disabled="actionResources.weapons + 5 > availableWeapons">+5</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('weapons', 10)"
                      :disabled="actionResources.weapons + 10 > availableWeapons">+10</button>
                  </div>
                </div>

                <div class="resource-allocation-bar">
                  <div class="allocation-fill"
                    :style="{ width: `${(actionResources.weapons / Math.max(1, availableWeapons)) * 100}%` }"></div>
                </div>
              </div>

              <!-- Vehicles Resource Control -->
              <div class="resource-control">
                <div class="control-header">
                  <div class="resource-label">
                    <span class="resource-icon">🚗</span>
                    <label>Vehicles</label>
                  </div>
                  <span class="resource-available">Current Allocation: {{ actionResources.vehicles }} / {{
                    availableVehicles }}</span>
                </div>

                <div class="control-actions">
                  <div class="adjustment-buttons">
                    <button class="adjust-btn large-adjust" @click="adjustResource('vehicles', -10)"
                      :disabled="actionResources.vehicles < 10">-10</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('vehicles', -5)"
                      :disabled="actionResources.vehicles < 5">-5</button>
                    <button class="adjust-btn" @click="adjustResource('vehicles', -1)"
                      :disabled="actionResources.vehicles <= 0">-</button>
                    <div class="number-input">
                      <input type="number" v-model.number="actionResources.vehicles" :min="0" :max="availableVehicles"
                        @input="validateResourceInput('vehicles')" />
                    </div>
                    <button class="adjust-btn" @click="adjustResource('vehicles', 1)"
                      :disabled="actionResources.vehicles >= availableVehicles">+</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('vehicles', 5)"
                      :disabled="actionResources.vehicles + 5 > availableVehicles">+5</button>
                    <button class="adjust-btn large-adjust" @click="adjustResource('vehicles', 10)"
                      :disabled="actionResources.vehicles + 10 > availableVehicles">+10</button>
                  </div>
                </div>

                <div class="resource-allocation-bar">
                  <div class="allocation-fill"
                    :style="{ width: `${(actionResources.vehicles / Math.max(1, availableVehicles)) * 100}%` }"></div>
                </div>
              </div>
            </div>
          </div>

          <div class="success-meter">
            <div class="meter-label">Success Chance:</div>
            <div class="meter-bar">
              <div class="meter-fill" :style="{ width: `${successChance}%` }"
                :class="getSuccessChanceClass(successChance)"></div>
            </div>
            <div class="meter-value" :class="getSuccessChanceClass(successChance)">
              {{ successChance }}%
            </div>
          </div>

          <div class="resource-warning" v-if="actionWarning">
            <div class="warning-icon">⚠️</div>
            <div class="warning-message">{{ actionWarning }}</div>
          </div>

          <div class="potential-rewards"
            v-if="selectedAction === TerritoryActionType.EXTORTION || selectedAction === TerritoryActionType.COLLECTION">
            <h4>Potential Rewards</h4>
            <div class="rewards-content">
              <div class="reward-item">
                <div class="reward-icon">💵</div>
                <div class="reward-details">
                  <div class="reward-label">Money:</div>
                  <div class="reward-value">
                    ${{ formatNumber(getPotentialReward(selectedAction, selectedHotspot)) }}
                  </div>
                </div>
              </div>

              <div class="risk-item" v-if="selectedAction === TerritoryActionType.EXTORTION">
                <div class="risk-icon">🚨</div>
                <div class="risk-details">
                  <div class="risk-label">Heat Generated:</div>
                  <div class="risk-value">+{{ getPotentialHeat(selectedAction) }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-footer-actions">
          <BaseButton variant="text" @click="closeActionModal">
            Cancel
          </BaseButton>
          <BaseButton :variant="successChance >= 50 ? 'primary' : 'danger'"
            :disabled="!canPerformAction || isPerformingAction" :loading="isPerformingAction" @click="performAction">
            {{ getActionButtonLabel(selectedAction) }}
          </BaseButton>
        </div>
      </template>
    </BaseModal>

    <!-- Result Modal -->
    <BaseModal v-model="showResultModal" :title="resultModalTitle" class="result-modal">
      <div class="action-result" :class="{ 'success': actionSuccess, 'failure': !actionSuccess }">
        <div class="result-icon">
          {{ actionSuccess ? '✅' : '❌' }}
        </div>

        <div class="result-message">
          {{ actionResult?.message || 'Operation completed.' }}
        </div>

        <div class="result-details" v-if="actionResult">
          <div class="result-columns">
            <div class="result-column">
              <h4>Resources</h4>
              <div v-if="actionResult.moneyGained" class="result-item positive">
                <span class="item-icon">💵</span>
                <span class="item-value">+${{ formatNumber(actionResult.moneyGained) }}</span>
              </div>
              <div v-if="actionResult.moneyLost" class="result-item negative">
                <span class="item-icon">💵</span>
                <span class="item-value">-${{ formatNumber(actionResult.moneyLost) }}</span>
              </div>

              <div v-if="actionResult.crewGained" class="result-item positive">
                <span class="item-icon">👥</span>
                <span class="item-value">+{{ actionResult.crewGained }} crew</span>
              </div>
              <div v-if="actionResult.crewLost" class="result-item negative">
                <span class="item-icon">👥</span>
                <span class="item-value">-{{ actionResult.crewLost }} crew</span>
              </div>

              <div v-if="actionResult.weaponsGained" class="result-item positive">
                <span class="item-icon">🔫</span>
                <span class="item-value">+{{ actionResult.weaponsGained }} weapons</span>
              </div>
              <div v-if="actionResult.weaponsLost" class="result-item negative">
                <span class="item-icon">🔫</span>
                <span class="item-value">-{{ actionResult.weaponsLost }} weapons</span>
              </div>

              <div v-if="actionResult.vehiclesGained" class="result-item positive">
                <span class="item-icon">🚗</span>
                <span class="item-value">+{{ actionResult.vehiclesGained }} vehicles</span>
              </div>
              <div v-if="actionResult.vehiclesLost" class="result-item negative">
                <span class="item-icon">🚗</span>
                <span class="item-value">-{{ actionResult.vehiclesLost }} vehicles</span>
              </div>
            </div>

            <div class="result-column">
              <h4>Reputation</h4>
              <div v-if="actionResult.respectGained" class="result-item positive">
                <span class="item-icon">👊</span>
                <span class="item-value">+{{ actionResult.respectGained }} respect</span>
              </div>
              <div v-if="actionResult.respectLost" class="result-item negative">
                <span class="item-icon">👊</span>
                <span class="item-value">-{{ actionResult.respectLost }} respect</span>
              </div>

              <div v-if="actionResult.influenceGained" class="result-item positive">
                <span class="item-icon">🌟</span>
                <span class="item-value">+{{ actionResult.influenceGained }} influence</span>
              </div>
              <div v-if="actionResult.influenceLost" class="result-item negative">
                <span class="item-icon">🌟</span>
                <span class="item-value">-{{ actionResult.influenceLost }} influence</span>
              </div>

              <div v-if="actionResult.heatGenerated" class="result-item negative">
                <span class="item-icon">🚨</span>
                <span class="item-value">+{{ actionResult.heatGenerated }} heat</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <BaseButton variant="primary" @click="closeResultModal">
          Continue
        </BaseButton>
      </template>
    </BaseModal>

    <!-- Hotspot Tooltip -->
    <HotspotTooltip v-if="tooltipHotspot" :hotspot="tooltipHotspot" :visible="tooltipVisible"
      :position="tooltipPosition" />

    <!-- Hotspot Detail Modal -->
    <HotspotDetailModal :visible="showDetailModal" @update:visible="showDetailModal = $event"
      :hotspot="detailModalHotspot" @open-action-modal="handleOpenActionModal" />
  </div>
</template>

<style lang="scss">
.territory-view {
  // @include page-container;

  .hotspot-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: $spacing-md;

    .hotspot-title-area {
      flex: 1;
    }

    h3 {
      margin: 0 0 $spacing-xs 0;
      font-size: $font-size-lg;
    }

    .hotspot-type {
      color: $text-secondary;
      font-size: $font-size-sm;
    }

    .hotspot-info-icon {
      cursor: help;
      height: 24px;
      width: 24px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      background-color: rgba($background-lighter, 0.2);
      transition: all 0.2s ease;

      &:hover {
        background-color: rgba($gold-color, 0.2);
        transform: scale(1.1);
      }

      .info-icon {
        font-size: 14px;
      }
    }
  }

  .page-header {
    @include flex-column;
    gap: $spacing-lg;
    margin-bottom: $spacing-xl;

    .page-title {
      h2 {
        @include gold-accent;
        margin-bottom: $spacing-xs;
      }

      .subtitle {
        color: $text-secondary;
      }
    }

    .territory-stats {
      display: flex;
      gap: $spacing-md;
      flex-wrap: wrap;

      .stat-card {
        // @include card-dark;
        flex: 1;
        min-width: 150px;
        display: flex;
        align-items: center;
        padding: $spacing-md;
        gap: $spacing-md;

        .stat-icon {
          font-size: 28px;
        }

        .stat-content {
          .stat-value {
            font-size: $font-size-xl;
            font-weight: 600;
            @include gold-accent;
          }

          .stat-label {
            font-size: $font-size-sm;
            color: $text-secondary;
          }
        }
      }
    }
  }

  .territory-tabs {
    display: flex;
    background-color: rgba($background-darker, 0.5);
    border-radius: $border-radius-md;
    margin-bottom: $spacing-xl;
    overflow: hidden;

    .tab-btn {
      flex: 1;
      background: none;
      border: none;
      color: $text-secondary;
      padding: $spacing-md;
      font-size: $font-size-md;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: $spacing-sm;
      transition: $transition-base;

      &:hover {
        background-color: rgba($background-lighter, 0.1);
      }

      &.active {
        background-color: $background-lighter;
        color: $text-color;
        font-weight: 500;
        box-shadow: 0 2px 0 $secondary-color;
      }

      .tab-icon {
        font-size: 20px;
      }
    }
  }

  .tab-content {
    min-height: 400px;
  }

  /* Empire Tab Styles */
  .empire-tab {

    /* Reversed the order of sections */
    .hotspots-section {
      margin-bottom: $spacing-xl;

      .section-header {
        @include flex-between;
        margin-bottom: $spacing-lg;

        // h4 {
        //   @include section-title;
        // }

        .section-filters {
          display: flex;
          gap: $spacing-md;

          .filter-group {
            display: flex;
            align-items: center;
            gap: $spacing-sm;

            select {
              background-color: $background-lighter;
              color: $text-color;
              border: 1px solid $border-color;
              border-radius: $border-radius-sm;
              padding: 6px 12px;
            }
          }
        }
      }
    }

    .empire-header {
      @include flex-between;
      margin-bottom: $spacing-lg;

      // h3 {
      //   @include section-title;
      // }
    }

    .empire-regions-overview {
      // @include card-dark;
      margin-bottom: $spacing-xl;

      .overview-header {
        margin-bottom: $spacing-lg;

        h4 {
          margin: 0 0 $spacing-xs 0;
        }

        .help-text {
          font-size: $font-size-sm;
          color: $text-secondary;
        }
      }

      .regions-chart {
        @include flex-column;
        gap: $spacing-md;

        .region-bar {
          display: flex;
          align-items: center;
          gap: $spacing-md;

          .region-name {
            width: 100px;
            font-weight: 500;
          }

          .bar-wrapper {
            flex: 1;
            height: 24px;
            background-color: $background-darker;
            border-radius: $border-radius-sm;
            position: relative;
            overflow: hidden;

            .bar-fill {
              height: 100%;
              background: linear-gradient(to right, $primary-color, $secondary-color);
              transition: width 0.3s ease;

              &.powerful {
                background: linear-gradient(to right, $secondary-color, $success-color);
              }
            }

            .bar-value {
              position: absolute;
              right: $spacing-sm;
              top: 50%;
              transform: translateY(-50%);
              font-size: $font-size-sm;
              font-weight: 500;
              color: $text-color;
              text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
            }
          }
        }

        .no-regions {
          @include flex-column;
          align-items: center;
          justify-content: center;
          text-align: center;
          color: $text-secondary;
          padding: $spacing-xl;
        }
      }
    }
  }

  /* Empire Grid Styles */
  .empire-grid {
    .hotspot-card {
      position: relative;

      &.has-pending {
        box-shadow: 0 0 0 2px $secondary-color;
      }

      /* Added the golden glow effect on hover */
      &:hover {
        box-shadow: 0 0 15px rgba(255, 215, 0, 0.7), 0 0 30px rgba(255, 215, 0, 0.4);
        border-color: rgba(255, 215, 0, 0.8);
        transform: translateY(-5px);
        transition: all 0.3s ease;
      }

      .card-badge {
        position: absolute;
        top: -10px;
        right: 10px;
        background-color: $secondary-color;
        color: $background-darker;
        font-weight: 600;
        font-size: $font-size-sm;
        padding: 3px 10px;
        border-radius: $border-radius-md;
        display: flex;
        align-items: center;
        gap: 5px;
        box-shadow: $shadow-sm;

        .badge-icon {
          font-size: 14px;
        }
      }

      .detail-row {
        display: flex;
        justify-content: space-between;
        margin-bottom: $spacing-sm;

        .detail-item {
          .detail-label {
            font-size: $font-size-sm;
            color: $text-secondary;
          }

          .detail-value {
            font-weight: 500;

            &.defense {
              &.high {
                color: $success-color;
              }

              &.medium {
                color: $warning-color;
              }

              &.low {
                color: $danger-color;
              }
            }
          }
        }

        &.defense-allocation {
          margin-top: $spacing-md;

          .resource-allocation {
            display: flex;
            gap: $spacing-sm;

            .resource-item {
              background-color: rgba($background-lighter, 0.3);
              border-radius: $border-radius-sm;
              padding: 3px 8px;
              display: flex;
              align-items: center;
              gap: 5px;
              font-size: $font-size-sm;

              .resource-icon {
                font-size: 12px;
              }
            }
          }
        }
      }

      .hotspot-footer {
        display: flex;
        gap: $spacing-sm;
        justify-content: flex-end;
      }
    }
  }

  /* Explore Tab Styles */
  .explore-tab {
    .explore-header {
      @include flex-column;
      gap: $spacing-md;
      margin-bottom: $spacing-lg;

      .filters {
        display: flex;
        flex-wrap: wrap;
        gap: $spacing-md;

        .filter-group {
          display: flex;
          align-items: center;
          gap: $spacing-sm;

          label {
            font-weight: 500;
            white-space: nowrap;
          }

          select {
            background-color: $background-lighter;
            color: $text-color;
            border: 1px solid $border-color;
            border-radius: $border-radius-sm;
            padding: 6px 12px;

            &:disabled {
              opacity: 0.5;
              cursor: not-allowed;
            }
          }

          &.region-filter {
            select {
              border-left: 3px solid $primary-color;
            }
          }
        }
      }

      .filter-tabs {
        display: flex;
        background-color: $background-darker;
        border-radius: $border-radius-md;
        overflow: hidden;

        .filter-tab {
          flex: 1;
          background: none;
          border: none;
          color: $text-secondary;
          padding: $spacing-sm $spacing-md;
          cursor: pointer;
          transition: $transition-base;

          &:hover {
            background-color: rgba($background-lighter, 0.1);
          }

          &.active {
            background-color: $background-lighter;
            color: $text-color;
          }
        }
      }

      .view-toggle {
        display: flex;
        gap: 1px;
        background-color: $border-color;
        border-radius: $border-radius-sm;
        align-self: flex-end;

        .toggle-btn {
          background-color: $background-darker;
          border: none;
          padding: $spacing-sm $spacing-md;
          color: $text-secondary;
          cursor: pointer;
          transition: $transition-base;
          display: flex;
          align-items: center;
          gap: $spacing-xs;

          &:first-child {
            border-radius: $border-radius-sm 0 0 $border-radius-sm;
          }

          &:last-child {
            border-radius: 0 $border-radius-sm $border-radius-sm 0;
          }

          &.active {
            background-color: $primary-color;
            color: $text-color;
          }
        }
      }
    }
  }

  /* Explore Grid Styles */
  .explore-grid {
    .hotspot-card {
      position: relative;

      /* Added the golden glow effect on hover */
      &:hover {
        box-shadow: 0 0 15px rgba(255, 215, 0, 0.7), 0 0 30px rgba(255, 215, 0, 0.4);
        border-color: rgba(255, 215, 0, 0.8);
        transform: translateY(-5px);
        transition: all 0.3s ease;
      }

      .card-badge {
        position: absolute;
        top: -10px;
        right: 10px;
        font-weight: 600;
        font-size: $font-size-sm;
        padding: 3px 10px;
        border-radius: $border-radius-md;
        display: flex;
        align-items: center;
        gap: 5px;
        box-shadow: $shadow-sm;

        &.illegal {
          background-color: $danger-color;
          color: $background-darker;
        }

        &.rival {
          background-color: $warning-color;
          color: $background-darker;
        }

        &.controlled {
          background-color: $success-color;
          color: $background-darker;
        }

        .badge-icon {
          font-size: 14px;
        }
      }

      &.controlled {
        border-left: 4px solid $success-color;
      }

      &.rival-controlled {
        border-left: 4px solid $warning-color;
      }

      &.illegal {
        border-left: 4px solid $danger-color;
      }
    }
  }

  /* Recent Activity Tab Styles */
  .recent-tab {
    .recent-header {
      margin-bottom: $spacing-lg;

      h3 {
        // @include section-title;
        margin-bottom: $spacing-xs;
      }

      .subtitle {
        color: $text-secondary;
      }
    }

    .timeline {
      .timeline-container {
        @include flex-column;
        gap: $spacing-md;

        .timeline-item {
          display: flex;
          gap: $spacing-md;
          padding: $spacing-md;
          background-color: $background-card;
          border-radius: $border-radius-md;
          border-left: 4px solid $border-color;

          &.success {
            border-left-color: $success-color;
          }

          &.failure {
            border-left-color: $danger-color;
          }

          .timeline-icon {
            font-size: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
            width: 40px;
            height: 40px;
            background-color: $background-darker;
            border-radius: 50%;
          }

          .timeline-content {
            flex: 1;

            .timeline-header {
              @include flex-between;
              margin-bottom: $spacing-xs;

              .action-type {
                font-weight: 600;
              }

              .action-time {
                font-size: $font-size-sm;
                color: $text-secondary;
              }
            }

            .action-target {
              font-size: $font-size-md;
              margin-bottom: $spacing-sm;
            }

            .action-details {
              .resources-used {
                display: flex;
                gap: $spacing-sm;
                margin-bottom: $spacing-sm;

                .resource {
                  background-color: rgba($background-lighter, 0.3);
                  border-radius: $border-radius-sm;
                  padding: 3px 8px;
                  display: flex;
                  align-items: center;
                  gap: 5px;
                  font-size: $font-size-sm;
                }
              }

              .action-result {
                .result-message {
                  font-size: $font-size-sm;
                  margin-bottom: $spacing-xs;
                }

                .result-effects {
                  display: flex;
                  flex-wrap: wrap;
                  gap: $spacing-sm;

                  .effect {
                    font-size: $font-size-sm;
                    padding: 2px 8px;
                    border-radius: $border-radius-sm;

                    &.positive {
                      background-color: rgba($success-color, 0.2);
                      color: $success-color;
                    }

                    &.negative {
                      background-color: rgba($danger-color, 0.2);
                      color: $danger-color;
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

  /* Shared Grid Styles */
  .hotspots-grid {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
    gap: $spacing-md;

    @include respond-to(sm) {
      grid-template-columns: repeat(2, 1fr);
    }

    @include respond-to(md) {
      grid-template-columns: repeat(3, 1fr);
    }

    @include respond-to(lg) {
      grid-template-columns: repeat(4, 1fr);
    }

    .hotspot-card {
      @include card;
      cursor: pointer;
      transition: all 0.3s ease;
      /* Updated for smoother transitions */
      display: flex;
      flex-direction: column;
      border: 1px solid rgba($border-color, 0.5);

      /* Removed the original hover effect as it's replaced by the golden glow */
      &:hover {
        transform: translateY(-5px);
      }

      .hotspot-header {
        margin-bottom: $spacing-md;

        h3 {
          margin: 0 0 $spacing-xs 0;
          font-size: $font-size-lg;
        }

        .hotspot-type {
          color: $text-secondary;
          font-size: $font-size-sm;
        }
      }

      .hotspot-details {
        flex: 1;

        .detail-item {
          .detail-label {
            color: $text-secondary;
          }

          .detail-value {
            &.status {
              &.controlled {
                color: $success-color;
              }

              &.rival {
                color: $danger-color;
              }

              &.neutral {
                color: $text-secondary;
              }
            }
          }
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
        font-size: 36px;
        margin-bottom: $spacing-sm;
      }

      h4 {
        margin: 0;
        color: $text-color;
      }
    }
  }

  /* Map View */
  .territory-map {
    .map-container {
      width: 100%;
      height: 600px;
      background-color: $background-card;
      border-radius: $border-radius-md;
      overflow: hidden;

      .map-placeholder {
        @include flex-column;
        align-items: center;
        justify-content: center;
        height: 100%;
        gap: $spacing-md;

        .placeholder-icon {
          font-size: 48px;
          margin-bottom: $spacing-sm;
        }

        h4 {
          margin: 0;
          @include gold-accent;
        }
      }
    }
  }

  /* Action Modal Styles */
  .action-modal {
    .action-modal-content {
      @include flex-column;
      gap: $spacing-lg;

      .hotspot-summary {
        .summary-title {
          margin-bottom: $spacing-md;

          h3 {
            margin: 0 0 $spacing-xs 0;
            @include gold-accent;
          }

          .hotspot-type {
            color: $text-secondary;
          }
        }

        .summary-location {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          margin-bottom: $spacing-md;
          padding: $spacing-sm $spacing-md;
          background-color: $background-darker;
          border-radius: $border-radius-sm;

          .location-text {
            font-weight: 500;
          }
        }

        .summary-details {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: $spacing-md;

          .summary-column {
            @include flex-column;
            gap: $spacing-sm;
          }

          .summary-item {
            display: flex;
            gap: $spacing-sm;

            .item-icon {
              font-size: 20px;
            }

            .item-details {
              .item-label {
                font-size: $font-size-sm;
                color: $text-secondary;
              }

              .item-value {
                font-weight: 500;

                &.controlled {
                  color: $success-color;
                }

                &.rival {
                  color: $danger-color;
                }

                &.defense {
                  &.high {
                    color: $success-color;
                  }

                  &.medium {
                    color: $warning-color;
                  }

                  &.low {
                    color: $danger-color;
                  }
                }
              }
            }
          }
        }
      }

      .action-info {
        @include flex-column;
        gap: $spacing-sm;
        padding: $spacing-md;
        background-color: rgba($background-lighter, 0.1);
        border-radius: $border-radius-md;

        .action-header {
          display: flex;
          align-items: center;
          gap: $spacing-sm;
          margin-bottom: $spacing-sm;

          .action-icon {
            font-size: 24px;
          }

          .action-name {
            font-size: $font-size-lg;
            font-weight: 600;
          }
        }

        .action-description {
          font-size: $font-size-sm;
          color: $text-secondary;
          line-height: 1.5;
        }
      }

      .resource-allocation {
        h4 {
          margin-bottom: $spacing-md;
        }

        /* New Resource Allocation UI Styles */
        .resource-allocation-ui {
          @include flex-column;
          gap: $spacing-md;
          margin-bottom: $spacing-lg;

          .allocation-presets {
            margin-bottom: $spacing-sm;

            h5 {
              margin: 0 0 $spacing-sm 0;
              font-size: $font-size-md;
            }

            .preset-buttons {
              display: flex;
              flex-wrap: wrap;
              gap: $spacing-sm;

              .preset-button {
                background-color: $background-lighter;
                border: 1px solid $border-color;
                border-radius: $border-radius-sm;
                padding: $spacing-sm $spacing-md;
                font-size: $font-size-sm;
                color: $text-color;
                cursor: pointer;
                transition: $transition-base;

                &:hover {
                  background-color: rgba($secondary-color, 0.1);
                  border-color: $secondary-color;
                }

                &.active {
                  background-color: $secondary-color;
                  border-color: $secondary-color;
                  color: $background-darker;
                  font-weight: 500;
                }
              }
            }
          }

          .resource-controls {
            @include flex-column;
            gap: $spacing-md;

            .resource-control {
              background-color: rgba($background-lighter, 0.1);
              border-radius: $border-radius-md;
              padding: $spacing-md;
              @include flex-column;
              gap: $spacing-sm;

              .control-header {
                @include flex-between;
                margin-bottom: $spacing-xs;

                .resource-label {
                  display: flex;
                  align-items: center;
                  gap: $spacing-xs;
                  font-weight: 500;

                  .resource-icon {
                    font-size: 18px;
                  }
                }

                .resource-available {
                  font-weight: 600;
                  color: $secondary-color;
                }
              }

              .control-actions {
                .adjustment-buttons {
                  display: flex;
                  align-items: center;
                  justify-content: space-between;
                  gap: $spacing-xs;

                  .adjust-btn {
                    background-color: $background-lighter;
                    border: 1px solid $border-color;
                    color: $text-color;
                    border-radius: $border-radius-sm;
                    padding: 4px 8px;
                    font-size: $font-size-sm;
                    cursor: pointer;
                    transition: $transition-base;
                    min-width: 32px;

                    &:hover:not(:disabled) {
                      background-color: $secondary-color;
                      color: $background-darker;
                    }

                    &:disabled {
                      opacity: 0.5;
                      cursor: not-allowed;
                    }
                  }

                  .number-input {
                    flex: 1;
                    max-width: 80px;

                    input {
                      width: 100%;
                      background-color: $background-darker;
                      border: 1px solid $border-color;
                      color: $text-color;
                      border-radius: $border-radius-sm;
                      padding: 4px 8px;
                      text-align: center;
                      font-size: $font-size-md;
                      font-weight: 600;

                      &:focus {
                        outline: none;
                        border-color: $secondary-color;
                      }

                      /* Remove spinner arrows in Chrome, Safari, Edge, Opera */
                      &::-webkit-outer-spin-button,
                      &::-webkit-inner-spin-button {
                        -webkit-appearance: none;
                        margin: 0;
                      }

                      /* Remove spinner arrows in Firefox */
                      // &[type=number] {
                      //   -moz-appearance: textfield;
                      // }
                    }
                  }
                }
              }

              .resource-allocation-bar {
                height: 8px;
                background-color: $background-darker;
                border-radius: $border-radius-sm;
                overflow: hidden;
                margin-top: $spacing-xs;

                .allocation-fill {
                  height: 100%;
                  background: linear-gradient(to right, $primary-color, $secondary-color);
                  border-radius: $border-radius-sm;
                  transition: width 0.3s ease;
                }
              }
            }
          }
        }

        .success-meter {
          display: flex;
          align-items: center;
          gap: $spacing-md;
          margin-bottom: $spacing-md;

          .meter-label {
            font-weight: 500;
            min-width: 120px;
          }

          .meter-bar {
            flex: 1;
            height: 10px;
            background-color: $background-darker;
            border-radius: 5px;
            overflow: hidden;

            .meter-fill {
              height: 100%;
              transition: width 0.3s ease;

              &.high {
                background-color: $success-color;
              }

              &.medium {
                background-color: $warning-color;
              }

              &.low {
                background-color: $danger-color;
              }
            }
          }

          .meter-value {
            min-width: 50px;
            font-weight: 600;
            text-align: right;

            &.high {
              color: $success-color;
            }

            &.medium {
              color: $warning-color;
            }

            &.low {
              color: $danger-color;
            }
          }
        }

        .resource-warning {
          display: flex;
          gap: $spacing-sm;
          padding: $spacing-md;
          background-color: rgba($warning-color, 0.1);
          border-left: 3px solid $warning-color;
          border-radius: $border-radius-sm;
          margin-bottom: $spacing-md;

          .warning-message {
            font-size: $font-size-sm;
          }
        }

        .potential-rewards {
          background-color: rgba($success-color, 0.1);
          border-radius: $border-radius-md;
          padding: $spacing-md;

          h4 {
            margin-top: 0;
            margin-bottom: $spacing-md;
            color: $success-color;
          }

          .rewards-content {
            display: flex;
            justify-content: space-between;

            .reward-item,
            .risk-item {
              display: flex;
              align-items: center;
              gap: $spacing-sm;

              .reward-icon,
              .risk-icon {
                font-size: 20px;
              }

              .reward-details,
              .risk-details {

                .reward-label,
                .risk-label {
                  font-size: $font-size-sm;
                  color: $text-secondary;
                }

                .reward-value {
                  font-weight: 600;
                  color: $success-color;
                }

                .risk-value {
                  font-weight: 600;
                  color: $danger-color;
                }
              }
            }
          }
        }
      }
    }
  }

  /* Result Modal Styles */
  .result-modal {
    .action-result {
      @include flex-column;
      align-items: center;
      text-align: center;
      gap: $spacing-md;
      padding: $spacing-lg;
      border-radius: $border-radius-md;

      &.success {
        background-color: rgba($success-color, 0.1);
      }

      &.failure {
        background-color: rgba($danger-color, 0.1);
      }

      .result-icon {
        font-size: 48px;
        margin-bottom: $spacing-sm;
      }

      .result-message {
        font-size: $font-size-lg;
        font-weight: 500;
        margin-bottom: $spacing-md;
      }

      .result-details {
        width: 100%;

        .result-columns {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: $spacing-xl;

          @include respond-to(sm-down) {
            grid-template-columns: 1fr;
            gap: $spacing-lg;
          }

          .result-column {
            @include flex-column;
            gap: $spacing-sm;

            h4 {
              margin: 0 0 $spacing-sm 0;
              padding-bottom: $spacing-xs;
              border-bottom: 1px solid $border-color;
            }

            .result-item {
              display: flex;
              align-items: center;
              gap: $spacing-sm;

              &.positive {
                color: $success-color;
              }

              &.negative {
                color: $danger-color;
              }

              .item-icon {
                font-size: 18px;
              }

              .item-value {
                font-weight: 500;
              }
            }
          }
        }
      }
    }
  }

  /* Income Timer Styles */
  .income-timer {
    margin-top: $spacing-sm;
    padding-top: $spacing-sm;
    border-top: 1px dashed rgba($border-color, 0.5);
  }

  .detail-value.income {
    color: $secondary-color;
    font-weight: 600;
  }

  .detail-value.countdown {
    font-weight: 500;

    &.soon {
      color: $success-color;
      animation: pulse 1s infinite;
    }
  }

  /* New Section Separator */
  .section-separator {
    position: relative;
    text-align: center;
    margin: $spacing-xl;
    height: 20px;

    &:before {
      content: "";
      position: absolute;
      top: 50%;
      left: 0;
      width: 100%;
      height: 2px;
      background: linear-gradient(90deg,
          rgba($gold-color, 0) 0%,
          rgba($gold-color, 0.5) 15%,
          rgba($gold-color, 1) 50%,
          rgba($gold-color, 0.5) 85%,
          rgba($gold-color, 0) 100%);
    }

    .separator-text {
      position: relative;
      background-color: $background-dark;
      padding: 0 $spacing-md;
      font-size: $font-size-md;
      font-weight: 600;
      color: $gold-color;
      text-transform: uppercase;
      letter-spacing: 1px;
    }
  }

  @keyframes pulse {
    0% {
      opacity: 1;
    }

    50% {
      opacity: 0.7;
    }

    100% {
      opacity: 1;
    }
  }

  /* Card-badge styling for pending collections */
  .hotspot-card.has-pending .card-badge {
    background-color: $secondary-color;
    animation: pulse 2s infinite;
  }

  /* Gold glow pulse animation for important elements */
  @keyframes goldPulse {
    0% {
      box-shadow: 0 0 10px rgba(255, 215, 0, 0.5);
    }

    50% {
      box-shadow: 0 0 20px rgba(255, 215, 0, 0.8), 0 0 30px rgba(255, 215, 0, 0.4);
    }

    100% {
      box-shadow: 0 0 10px rgba(255, 215, 0, 0.5);
    }
  }

  /* Added additional polish to the hotspot cards */
  .hotspot-card.controlled {
    border-top: 2px solid $success-color;

    &.has-pending {
      animation: goldPulse 3s infinite;
    }
  }
}
</style>
