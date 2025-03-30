//  src/services/sseService.ts

import { reactive } from 'vue';
import { usePlayerStore } from '@/stores/modules/player';
import { useTerritoryStore } from '@/stores/modules/territory';

// SSE event types
export enum SSEEventType {
  CONNECTED = 'connected',
  HEARTBEAT = 'heartbeat',
  INCOME_GENERATED = 'income_generated',
  HOTSPOT_UPDATED = 'hotspot_updated',
  NOTIFICATION = 'notification'
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
  // const url = `${import.meta.env.VITE_API_URL}/sse?token=${encodeURIComponent(token)}`;
  const url = `http://localhost:8000/api/sse?token=${encodeURIComponent(token)}`;
  state.eventSource = new EventSource(url);
  
  // Set up event listeners
  state.eventSource.onopen = () => {
    state.connected = true;
    state.reconnecting = false;
    state.error = null;
    console.log('SSE connection established');
  };
  
  state.eventSource.onerror = (event) => {
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
  eventSource.addEventListener(SSEEventType.CONNECTED, (event) => {
    const data = JSON.parse(event.data);
    console.log('Connected event:', data);
  });

  // Add heartbeat event handler
  eventSource.addEventListener(SSEEventType.HEARTBEAT, (event) => {
    // Just log the heartbeat or update a last-heartbeat timestamp
    console.log('Heartbeat received:', JSON.parse(event.data));
    
    // Update connection status to ensure UI shows connected
    state.connected = true;
    state.reconnecting = false;
    state.error = null;
  });
  
  // Income generated event
  eventSource.addEventListener(SSEEventType.INCOME_GENERATED, (event) => {
    const data = JSON.parse(event.data);
    console.log('Income generated event:', data);
    
    // Process income updates
    if (data.updates && Array.isArray(data.updates)) {
      // Update hotspots
      data.updates.forEach((update: any) => {
        territoryStore.updateHotspotIncome(
          update.hotspotId,
          update.newIncome,
          update.pendingCollection,
          update.lastIncomeTime,
          update.nextIncomeTime
        );
      });
      
      // Update total pending collections
      if (playerStore.profile && typeof data.totalPending === 'number') {
        playerStore.profile.pendingCollections = data.totalPending;
      }
    }
  });
  
  // Hotspot updated event
  eventSource.addEventListener(SSEEventType.HOTSPOT_UPDATED, (event) => {
    const data = JSON.parse(event.data);
    console.log('Hotspot updated event:', data);
    
    // Update the hotspot
    if (data.hotspot) {
      territoryStore.updateHotspot(data.hotspot);
    }
  });
  
  // Notification event
  eventSource.addEventListener(SSEEventType.NOTIFICATION, (event) => {
    const data = JSON.parse(event.data);
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