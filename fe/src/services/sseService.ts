//  src/services/sseService.ts

import { reactive } from 'vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { useOperationsStore } from '@/stores/modules/operations';
import { useCampaignStore } from '@/stores/modules/campaign';
import { useTravelStore } from '@/stores/modules/travel';
import { Hotspot } from '@/types/territory';
import { Notification } from '@/types/player';
import { POI, MissionOperation } from '@/types/campaign';
import { Operation, OperationsRefreshInfo } from '@/types/operations';

// SSE event types
export enum SSEEventType {
  CONNECTED = 'connected',
  HEARTBEAT = 'heartbeat',
  INCOME_GENERATED = 'income_generated',
  HOTSPOT_UPDATED = 'hotspot_updated',
  HOTSPOTS_UPDATED = 'hotspots_updated',
  NOTIFICATION = 'notification',
  PLAYER_REGION_CHANGED = 'player_region_changed',

  // Campaign-related types
  CAMPAIGN_ACTION_TRACKED = 'campaign_action_tracked',
  CAMPAIGN_CHOICE_UPDATED = 'campaign_choice_updated',
  CAMPAIGN_POI_UPDATED = 'campaign_poi_updated',
  CAMPAIGN_OPERATION_UPDATED = 'campaign_operation_updated',

  // Operations-related types
  OPERATIONS_REFRESHED = 'operations_refreshed',
}

// Define SSE event payloads
export interface IncomeUpdate {
  hotspotId: string;
  hotspotName: string;
  newIncome: number;
  pendingCollection: number;
  lastIncomeTime: string;
  nextIncomeTime: string;
}

export interface IncomeGeneratedEvent {
  updates: IncomeUpdate[];
  totalPending: number;
  timestamp: string;
}

export interface HotspotUpdatedEvent {
  hotspot: Hotspot;
}

export interface HotspotsUpdatedEvent {
  hotspots: Hotspot[];
}

export interface NotificationEvent {
  notification: Notification;
}

export interface PlayerRegionChangedEvent {
  event: string;
  playerId: string;
  regionId: string;
  regionName: string;
  timestamp: string;
}

// New campaign event payload interfaces
export interface CampaignActionTrackedEvent {
  actionType: string;
  actionValue: string;
  missionId?: string;
  choiceId?: string;
  conditionCompleted?: boolean;
}

export interface CampaignChoiceUpdatedEvent {
  missionId: string;
  choiceId: string;
  isCompleted: boolean;
}

export interface CampaignPOIUpdatedEvent {
  poi: POI;
  isActivated?: boolean;
}

export interface CampaignOperationUpdatedEvent {
  operation: MissionOperation;
  isActivated?: boolean;
}

export interface OperationsRefreshedEvent {
  operations: Operation[];
  timestamp: string;
  refreshInfo?: OperationsRefreshInfo;
}


// SSE service state
const state = reactive({
  connected: false,
  reconnecting: false,
  eventSource: null as EventSource | null,
  lastEventId: '',
  error: null as Error | null
});

/**
 * Invalidate cache and reload all affected data when player region changes
 */
async function invalidateAndReloadAllData() {
  const playerStore = usePlayerStore();
  const territoryStore = useTerritoryStore();
  const operationsStore = useOperationsStore();
  const travelStore = useTravelStore();
  const campaignStore = useCampaignStore();

  console.log('Invalidating cache and reloading all data after region change...');

  // Stop any active timers before reloading data
  territoryStore.stopIncomeTimer();
  operationsStore.stopOperationTimer();

  try {
    // Clear existing data to force fresh load
    territoryStore.$reset();
    operationsStore.$reset();
    travelStore.$reset();

    // Reload all player data
    await playerStore.fetchProfile();

    // Reload travel data
    await travelStore.fetchCurrentRegion();
    await travelStore.fetchAvailableRegions();

    // Reload territory data
    await territoryStore.fetchTerritoryData();
    await territoryStore.fetchRecentActions();

    // Reload operations data
    await operationsStore.fetchAvailableOperations();
    await operationsStore.fetchPlayerOperations();
    await operationsStore.fetchOperationsRefreshInfo();

    // Reload campaign data
    await campaignStore.fetchCampaigns();
    await campaignStore.fetchActivePOIs();
    await campaignStore.fetchActiveMissionOperations();

    // Restart timers
    territoryStore.startIncomeTimer();
    operationsStore.startOperationTimer();

    console.log('Successfully reloaded all data after region change');
  } catch (error) {
    console.error('Error reloading data after region change:', error);
  }
}

/**
 * Establishes an SSE connection with the server
 */
function connect() {
  // Close existing connection if any
  if (state.eventSource) {
    state.eventSource.close();
    state.eventSource = null;
  }

  const token = localStorage.getItem('auth_token');
  if (!token) {
    state.error = new Error('Authentication token not found');
    return;
  }

  // Create a new connection with token in the URL
  const url = `http://localhost:8000/api/sse?token=${encodeURIComponent(token)}`;
  state.eventSource = new EventSource(url);

  // Set up event listeners
  state.eventSource.onopen = () => {
    state.connected = true;
    state.reconnecting = false;
    state.error = null;
    console.log('SSE connection established');
  };

  state.eventSource.onerror = event => {
    console.error('SSE connection error:', event);
    state.connected = false;
    state.error = new Error('Connection error');

    // Attempt to reconnect
    if (!state.reconnecting) {
      state.reconnecting = true;
      setTimeout(() => {
        connect();
      }, 5000); // Try to reconnect after 5 seconds
    }
  };

  // Set up event handlers
  setupEventHandlers(state.eventSource);
}

/**
 * Sets up event handlers for the SSE connection
 */
function setupEventHandlers(eventSource: EventSource) {
  const playerStore = usePlayerStore();
  const territoryStore = useTerritoryStore();
  const campaignStore = useCampaignStore();

  // Connected event
  eventSource.addEventListener(SSEEventType.CONNECTED, event => {
    const data = JSON.parse(event.data);
    console.log('Connected event:', data);
  });

  // Add heartbeat event handler
  eventSource.addEventListener(SSEEventType.HEARTBEAT, event => {
    // Just log the heartbeat or update a last-heartbeat timestamp
    console.log('Heartbeat received:', JSON.parse(event.data));

    // Update connection status to ensure UI shows connected
    state.connected = true;
    state.reconnecting = false;
    state.error = null;
  });

  // Player region changed event handler
  eventSource.addEventListener(SSEEventType.PLAYER_REGION_CHANGED, async event => {
    try {
      const data = JSON.parse(event.data) as PlayerRegionChangedEvent;
      console.log('Player region changed event received:', data);

      // Update player store with new region information
      const playerStore = usePlayerStore();
      if (playerStore.profile) {
        playerStore.profile.currentRegionId = data.regionId;
        playerStore.profile.currentRegionName = data.regionName;
      }

      // Invalidate all cache and reload all affected data
      await invalidateAndReloadAllData();

      // Show a notification about the region change
      const notification: Notification = {
        playerId: data.playerId,
        message: `You have arrived in ${data.regionName}`,
        type: 'travel' as any,
        timestamp: new Date().toISOString(),
        read: false
      };
      playerStore.addNotification(notification);
    } catch (error) {
      console.error('Error processing player region changed event:', error);
    }
  });

  eventSource.addEventListener(SSEEventType.INCOME_GENERATED, event => {
    try {
      const data = JSON.parse(event.data);
      console.log('Income generated event received:', data);

      // Process hotspot update
      if (data.hotspot) {
        const update = data.hotspot;

        // Ensure the timestamp properties are proper ISO strings
        if (update.lastIncomeTime && typeof update.lastIncomeTime !== 'string') {
          update.lastIncomeTime = new Date(update.lastIncomeTime).toISOString();
        }

        if (update.nextIncomeTime && typeof update.nextIncomeTime !== 'string') {
          update.nextIncomeTime = new Date(update.nextIncomeTime).toISOString();
        } else if (update.lastIncomeTime && !update.nextIncomeTime) {
          // Calculate nextIncomeTime if missing
          const lastIncomeTime = new Date(update.lastIncomeTime);
          const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
          update.nextIncomeTime = nextIncomeTime.toISOString();
        }

        console.log('Processing income update for hotspot after processing:', {
          id: update.id,
          name: update.name || 'Unknown',
          lastIncomeTime: update.lastIncomeTime,
          nextIncomeTime: update.nextIncomeTime,
          pendingCollection: update.pendingCollection
        });

        // Update the hotspot in the store
        territoryStore.updateHotspot({
          id: update.id,
          pendingCollection: update.pendingCollection,
          lastIncomeTime: update.lastIncomeTime,
          nextIncomeTime: update.nextIncomeTime
        });

        // Also update player's total pending collections
        if (playerStore.profile) {
          const totalPending = territoryStore.controlledHotspots.reduce(
            (sum, hotspot) => sum + hotspot.pendingCollection,
            0
          );
          playerStore.profile.pendingCollections = totalPending;
        }
      }
    } catch (error) {
      console.error('Error processing income generated event:', error);
    }
  });

  eventSource.addEventListener(SSEEventType.HOTSPOT_UPDATED, event => {
    try {
      const data = JSON.parse(event.data);
      console.log('Hotspot updated event received:', data);

      if (data.hotspot) {
        const hotspot = data.hotspot;

        // Ensure date fields are proper ISO strings
        if (hotspot.lastIncomeTime && typeof hotspot.lastIncomeTime !== 'string') {
          hotspot.lastIncomeTime = new Date(hotspot.lastIncomeTime).toISOString();
          console.log('Converted lastIncomeTime to ISO string:', hotspot.lastIncomeTime);
        }

        // If nextIncomeTime is missing but lastIncomeTime exists, calculate it
        if (!hotspot.nextIncomeTime && hotspot.lastIncomeTime) {
          const lastIncomeTime = new Date(hotspot.lastIncomeTime);
          const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
          hotspot.nextIncomeTime = nextIncomeTime.toISOString();
          console.log('Calculated missing nextIncomeTime:', hotspot.nextIncomeTime);
        }
        // Otherwise, ensure nextIncomeTime is a proper ISO string if it exists
        else if (hotspot.nextIncomeTime && typeof hotspot.nextIncomeTime !== 'string') {
          hotspot.nextIncomeTime = new Date(hotspot.nextIncomeTime).toISOString();
          console.log('Converted nextIncomeTime to ISO string:', hotspot.nextIncomeTime);
        }

        // Log the processed hotspot
        console.log('Processing hotspot update after processing:', {
          id: hotspot.id,
          name: hotspot.name,
          lastIncomeTime: hotspot.lastIncomeTime,
          nextIncomeTime: hotspot.nextIncomeTime
        });

        // Update the hotspot in the territory store
        territoryStore.updateHotspot(hotspot);
      }
    } catch (error) {
      console.error('Error processing hotspot updated event:', error);
    }
  });

  // Hotspots updated event
  eventSource.addEventListener(SSEEventType.HOTSPOTS_UPDATED, event => {
    try {
      const data = JSON.parse(event.data);
      console.log('Hotspots updated event received:', data);

      // Update all hotspots
      if (data.hotspots && Array.isArray(data.hotspots)) {
        data.hotspots.forEach((hotspot: any) => {
          // Ensure date fields are proper ISO strings
          if (hotspot.lastIncomeTime && typeof hotspot.lastIncomeTime !== 'string') {
            hotspot.lastIncomeTime = new Date(hotspot.lastIncomeTime).toISOString();
          }

          if (hotspot.nextIncomeTime && typeof hotspot.nextIncomeTime !== 'string') {
            hotspot.nextIncomeTime = new Date(hotspot.nextIncomeTime).toISOString();
          } else if (hotspot.lastIncomeTime && !hotspot.nextIncomeTime) {
            // Calculate nextIncomeTime if missing
            const lastIncomeTime = new Date(hotspot.lastIncomeTime);
            const nextIncomeTime = new Date(lastIncomeTime.getTime() + 60 * 60 * 1000);
            hotspot.nextIncomeTime = nextIncomeTime.toISOString();
          }

          territoryStore.updateHotspot(hotspot);
        });
      }
    } catch (error) {
      console.error('Error processing hotspots updated event:', error);
    }
  });

  // Notification event
  eventSource.addEventListener(SSEEventType.NOTIFICATION, event => {
    const data = JSON.parse(event.data) as NotificationEvent;
    console.log('Notification event:', data);

    // Add notification
    if (data.notification) {
      playerStore.addNotification(data.notification);
    }
  });

  // Campaign action tracking event handler
  eventSource.addEventListener(SSEEventType.CAMPAIGN_ACTION_TRACKED, event => {
    try {
      const data = JSON.parse(event.data) as CampaignActionTrackedEvent;
      console.log('Campaign action tracked event:', data);

      // Update campaign store if a condition was completed
      if (data.conditionCompleted && data.choiceId) {
        // Refresh POIs and mission operations
        campaignStore.fetchActivePOIs();
        campaignStore.fetchActiveMissionOperations();

        // If the action completed the entire choice, refresh the mission
        if (data.missionId) {
          campaignStore.fetchMission(data.missionId);
        }
      }
    } catch (error) {
      console.error('Error processing campaign action tracked event:', error);
    }
  });

  // Campaign choice updated event handler
  eventSource.addEventListener(SSEEventType.CAMPAIGN_CHOICE_UPDATED, event => {
    try {
      const data = JSON.parse(event.data) as CampaignChoiceUpdatedEvent;
      console.log('Campaign choice updated event:', data);

      // Refresh the mission with the updated choice
      if (data.missionId) {
        campaignStore.fetchMission(data.missionId);
      }
    } catch (error) {
      console.error('Error processing campaign choice updated event:', error);
    }
  });

  // Campaign POI updated event handler
  eventSource.addEventListener(SSEEventType.CAMPAIGN_POI_UPDATED, event => {
    try {
      const data = JSON.parse(event.data) as CampaignPOIUpdatedEvent;
      console.log('Campaign POI updated event:', data);

      // Update the POI in the store
      if (data.poi) {
        // Find the POI in the activePOIs array and update it
        const index = campaignStore.activePOIs.findIndex(p => p.id === data.poi.id);
        if (index >= 0) {
          campaignStore.activePOIs[index] = data.poi;
        } else if (data.isActivated) {
          // If it's a newly activated POI, add it to the array
          campaignStore.activePOIs.push(data.poi);
        }
      }
    } catch (error) {
      console.error('Error processing campaign POI updated event:', error);
    }
  });

  // Campaign mission operation updated event handler
  eventSource.addEventListener(SSEEventType.CAMPAIGN_OPERATION_UPDATED, event => {
    try {
      const data = JSON.parse(event.data) as CampaignOperationUpdatedEvent;
      console.log('Campaign operation updated event:', data);

      // Update the operation in the store
      if (data.operation) {
        // Find the operation in the activeMissionOperations array and update it
        const index = campaignStore.activeMissionOperations.findIndex(op => op.id === data.operation.id);
        if (index >= 0) {
          campaignStore.activeMissionOperations[index] = data.operation;
        } else if (data.isActivated) {
          // If it's a newly activated operation, add it to the array
          campaignStore.activeMissionOperations.push(data.operation);
        }
      }
    } catch (error) {
      console.error('Error processing campaign operation updated event:', error);
    }
  });

  eventSource.addEventListener(SSEEventType.OPERATIONS_REFRESHED, event => {
    try {
      const data = JSON.parse(event.data) as OperationsRefreshedEvent;
      console.log('Operations refreshed event received:', data);

      // Update the operations store
      const operationsStore = useOperationsStore();
      operationsStore.handleOperationsRefreshed(data.operations, data.refreshInfo);
    } catch (error) {
      console.error('Error processing operations refreshed event:', error);
    }
  });

}

/**
 * Disconnects from the SSE server
 */
function disconnect() {
  if (state.eventSource) {
    state.eventSource.close();
    state.eventSource = null;
    state.connected = false;
  }
}

export default {
  connect,
  disconnect,
  state
};
