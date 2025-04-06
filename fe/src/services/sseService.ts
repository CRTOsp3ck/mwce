//  src/services/sseService.ts

import { reactive } from 'vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';
import { Hotspot } from '@/types/territory';
import { Notification } from '@/types/player';

// SSE event types
export enum SSEEventType {
  CONNECTED = 'connected',
  HEARTBEAT = 'heartbeat',
  INCOME_GENERATED = 'income_generated',
  HOTSPOT_UPDATED = 'hotspot_updated',
  HOTSPOTS_UPDATED = 'hotspots_updated',
  NOTIFICATION = 'notification'
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

// SSE service state
const state = reactive({
  connected: false,
  reconnecting: false,
  eventSource: null as EventSource | null,
  lastEventId: '',
  error: null as Error | null
});

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
        data.hotspots.forEach(hotspot => {
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
