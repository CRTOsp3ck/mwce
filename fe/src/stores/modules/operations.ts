// src/stores/modules/operations.ts

import { defineStore } from 'pinia';
import operationsService from '@/services/operationsService';
import {
  Operation,
  OperationType,
  OperationResources,
  OperationResult,
  OperationStatus,
  OperationAttempt
} from '@/types/operations';
import { usePlayerStore } from './player';

interface OperationsState {
  availableOperations: Operation[];
  currentOperations: OperationAttempt[];
  completedOperations: OperationAttempt[];
  selectedOperationId: string | null;
  isLoading: boolean;
  error: string | null;
  timerRefreshCounter: number; // Used to force timer updates
  incomeTimerInterval: number | null;
}

export const useOperationsStore = defineStore('operations', {
  state: (): OperationsState => ({
    availableOperations: [],
    currentOperations: [],
    completedOperations: [],
    selectedOperationId: null,
    isLoading: false,
    error: null,
    timerRefreshCounter: 0,
    incomeTimerInterval: null
  }),

  getters: {
    selectedOperation: (state) => {
      return state.selectedOperationId
        ? state.availableOperations.find(o => o.id === state.selectedOperationId)
        : null;
    },

    basicOperations: (state) => {
      return state.availableOperations.filter(o => !o.isSpecial);
    },

    specialOperations: (state) => {
      return state.availableOperations.filter(o => o.isSpecial);
    },

    inProgressOperations: (state) => {
      return state.currentOperations.filter(o =>
        o.status === OperationStatus.IN_PROGRESS
      );
    },

    completedSuccessfulOperations: (state) => {
      return state.completedOperations.filter(o =>
        o.status === OperationStatus.COMPLETED && o.result && o.result.success
      );
    },

    completedFailedOperations: (state) => {
      return state.completedOperations.filter(o =>
        o.status === OperationStatus.FAILED || (o.result && !o.result.success)
      );
    },

    // Time-related getters with reactivity through timerRefreshCounter
    getTimeRemaining: (state) => (operationId: string): string => {
      // Use the refresh counter to make this getter reactive to timer changes
      const _ = state.timerRefreshCounter;

      const inProgressOp = state.currentOperations.find(op => op.id === operationId);
      if (!inProgressOp || inProgressOp.status !== OperationStatus.IN_PROGRESS) {
        return 'Completed';
      }

      const operation = state.availableOperations.find(op => op.id === inProgressOp.operationId);
      if (!operation) return 'Unknown';

      const startTime = new Date(inProgressOp.timestamp);
      const endTime = new Date(startTime.getTime() + (operation.duration * 1000));
      const now = new Date();

      if (now >= endTime) {
        return 'Ready to collect';
      }

      const remainingSeconds = Math.floor((endTime.getTime() - now.getTime()) / 1000);
      return formatDuration(remainingSeconds);
    },

    isOpCompletionSoon: (state) => (operationId: string): boolean => {
      // Use the refresh counter to make this getter reactive to timer changes
      const _ = state.timerRefreshCounter;

      const inProgressOp = state.currentOperations.find(op => op.id === operationId);
      if (!inProgressOp || inProgressOp.status !== OperationStatus.IN_PROGRESS) return false;

      const operation = state.availableOperations.find(op => op.id === inProgressOp.operationId);
      if (!operation) return false;

      const startTime = new Date(inProgressOp.timestamp);
      const endTime = new Date(startTime.getTime() + (operation.duration * 1000));
      const now = new Date();

      // If already passed or less than 5 minutes remaining
      const diffMs = endTime.getTime() - now.getTime();
      return diffMs <= 5 * 60 * 1000 && diffMs >= 0;
    }
  },

  actions: {
    async fetchAvailableOperations() {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await operationsService.getAvailableOperations();
        if (response.success && response.data) {
          this.availableOperations = response.data;
        } else {
          throw new Error('Failed to get operations data');
        }
      } catch (error) {
        this.error = 'Failed to load available operations';
        console.error('Error fetching available operations:', error);
      } finally {
        this.isLoading = false;
      }
    },

    async fetchPlayerOperations() {
      this.isLoading = true;

      try {
        // Get current operations
        const currentResponse = await operationsService.getCurrentOperations();
        if (currentResponse.success && currentResponse.data) {
          this.currentOperations = currentResponse.data;
        }

        // Get completed operations
        const completedResponse = await operationsService.getCompletedOperations();
        if (completedResponse.success && completedResponse.data) {
          this.completedOperations = completedResponse.data;
        }
      } catch (error) {
        console.error('Error fetching player operations:', error);
      } finally {
        this.isLoading = false;
      }
    },

    selectOperation(operationId: string | null) {
      this.selectedOperationId = operationId;
    },

    async startOperation(operationId: string, resources: OperationResources) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await operationsService.startOperation(operationId, resources);
        if (!response.success || !response.data) {
          throw new Error('Failed to start operation');
        }

        const newOperation = response.data;

        // Add to current operations
        this.currentOperations.push(newOperation);

        // Update player resources
        const playerStore = usePlayerStore();
        if (playerStore.profile) {
          // Deduct resources
          playerStore.profile.crew -= resources.crew;
          playerStore.profile.weapons -= resources.weapons;
          playerStore.profile.vehicles -= resources.vehicles;

          if (resources.money) {
            playerStore.profile.money -= resources.money;
          }
        }

        // Start the operation timer
        this.startOperationTimer();

        return {
          operation: newOperation,
          gameMessage: response.gameMessage
        };
      } catch (error) {
        this.error = 'Failed to start operation';
        console.error('Error starting operation:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async cancelOperation(operationId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await operationsService.cancelOperation(operationId);
        if (!response.success) {
          throw new Error('Failed to cancel operation');
        }

        // Find the operation and update its status
        const operationIndex = this.currentOperations.findIndex(o => o.id === operationId);
        if (operationIndex !== -1) {
          this.currentOperations[operationIndex].status = OperationStatus.CANCELLED;

          // Move to completed list
          this.completedOperations.unshift(this.currentOperations[operationIndex]);
          this.currentOperations.splice(operationIndex, 1);
        }

        return {
          status: 'cancelled',
          gameMessage: response.gameMessage
        };
      } catch (error) {
        this.error = 'Failed to cancel operation';
        console.error('Error cancelling operation:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async collectOperation(operationId: string) {
      this.isLoading = true;
      this.error = null;

      try {
        const response = await operationsService.collectOperation(operationId);
        if (!response.success || !response.data) {
          throw new Error('Failed to collect operation');
        }

        const result = response.data;

        // Find the operation and update it
        const operationIndex = this.currentOperations.findIndex(o => o.id === operationId);

        if (operationIndex !== -1) {
          this.currentOperations[operationIndex].result = result;
          this.currentOperations[operationIndex].status = result.success
            ? OperationStatus.COMPLETED
            : OperationStatus.FAILED;
          this.currentOperations[operationIndex].completionTime = new Date().toISOString();

          // Move to completed list
          this.completedOperations.unshift(this.currentOperations[operationIndex]);
          this.currentOperations.splice(operationIndex, 1);

          // Apply resource changes to the player
          this.applyOperationResultToPlayer(result);
        }

        return {
          result,
          gameMessage: response.gameMessage
        };
      } catch (error) {
        this.error = 'Failed to collect operation rewards';
        console.error('Error collecting operation rewards:', error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    // Helper function to apply operation result to player resources
    applyOperationResultToPlayer(result: OperationResult) {
      const playerStore = usePlayerStore();
      if (playerStore.profile) {
        // Update money
        if (result.moneyGained) {
          playerStore.profile.money += result.moneyGained;
        }
        if (result.moneyLost) {
          playerStore.profile.money = Math.max(0, playerStore.profile.money - result.moneyLost);
        }

        // Update crew
        if (result.crewGained) {
          playerStore.profile.crew += result.crewGained;
        }
        if (result.crewLost) {
          playerStore.profile.crew = Math.max(0, playerStore.profile.crew - result.crewLost);
        }

        // Update weapons
        if (result.weaponsGained) {
          playerStore.profile.weapons += result.weaponsGained;
        }
        if (result.weaponsLost) {
          playerStore.profile.weapons = Math.max(0, playerStore.profile.weapons - result.weaponsLost);
        }

        // Update vehicles
        if (result.vehiclesGained) {
          playerStore.profile.vehicles += result.vehiclesGained;
        }
        if (result.vehiclesLost) {
          playerStore.profile.vehicles = Math.max(0, playerStore.profile.vehicles - result.vehiclesLost);
        }

        // Update respect/influence
        if (result.respectGained) {
          playerStore.profile.respect += result.respectGained;
        }
        if (result.influenceGained) {
          playerStore.profile.influence += result.influenceGained;
        }

        // Update heat
        if (result.heatGenerated) {
          playerStore.profile.heat += result.heatGenerated;
        }
        if (result.heatReduced) {
          playerStore.profile.heat = Math.max(0, playerStore.profile.heat - result.heatReduced);
        }
      }
    },

    // Start timer to update operation progress
    startOperationTimer() {
      // Clean up existing timer if any
      if (this.incomeTimerInterval) {
        clearInterval(this.incomeTimerInterval);
        this.incomeTimerInterval = null;
      }

      // Set up new timer that updates every second
      this.incomeTimerInterval = window.setInterval(() => {
        // Increment the refresh counter to trigger reactivity
        this.timerRefreshCounter++;

        // Check for operations that have reached their end time
        this.checkOperationStatus();
      }, 1000);

      console.log('Operation timer started');
    },

    // Stop timer when component unmounts
    stopOperationTimer() {
      if (this.incomeTimerInterval) {
        clearInterval(this.incomeTimerInterval);
        this.incomeTimerInterval = null;
        console.log('Operation timer stopped');
      }
    },

    // Check and update status of in-progress operations
    checkOperationStatus() {
      const now = new Date();

      this.currentOperations.forEach(operation => {
        if (operation.status === OperationStatus.IN_PROGRESS) {
          // Get the operation details to determine duration
          const operationDetails = this.availableOperations.find(o => o.id === operation.operationId);

          if (operationDetails) {
            // Calculate completion time
            const startTime = new Date(operation.timestamp);
            const completionTime = new Date(startTime.getTime() + (operationDetails.duration * 1000));

            // If operation is complete but not yet collected, update the UI to reflect this
            // We're NOT auto-collecting here, just updating the UI to show it's ready
            if (now >= completionTime) {
              // This is handled by the isOperationReady function now
            }
          }
        }
      });
    },

    // Handle operation completion from SSE notification
    handleOperationCompletion(completedOperation: OperationAttempt) {
      console.log('Handling operation completion from SSE:', completedOperation);

      // For our new flow, operations received from SSE are NOT moved to completed
      // They stay in currentOperations, just with updated status

      // Find if this operation is in our current operations list
      const operationIndex = this.currentOperations.findIndex(op => op.id === completedOperation.id);

      if (operationIndex !== -1) {
        // Update the operation with readiness indicator
        // but keep it in the current operations list
        this.currentOperations[operationIndex] = {
          ...this.currentOperations[operationIndex],
          // We DON'T change the status here - it stays as IN_PROGRESS until collected
          completionTime: completedOperation.completionTime
        };

        console.log('Operation updated and kept in current operations:', completedOperation.id);
      } else {
        // If not found in current operations, check if it's in completed
        const existingCompleted = this.completedOperations.findIndex(op => op.id === completedOperation.id);

        // If not in either list, add it to current operations
        if (existingCompleted === -1) {
          // Make sure the status is still IN_PROGRESS
          this.currentOperations.push({
            ...completedOperation,
            status: OperationStatus.IN_PROGRESS
          });
          console.log('Added new operation to current operations list:', completedOperation.id);
        }
      }
    }
  }
});

// Helper function to format remaining time
function formatDuration(seconds: number): string {
  if (seconds <= 0) {
    return 'Now';
  }

  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m ${secs}s`;
  } else if (minutes > 0) {
    return `${minutes}m ${secs}s`;
  } else {
    return `${secs}s`;
  }
}
