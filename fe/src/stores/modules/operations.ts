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
}

export const useOperationsStore = defineStore('operations', {
  state: (): OperationsState => ({
    availableOperations: [],
    currentOperations: [],
    completedOperations: [],
    selectedOperationId: null,
    isLoading: false,
    error: null
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
          this.completedOperations.push(this.currentOperations[operationIndex]);
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
          this.completedOperations.push(this.currentOperations[operationIndex]);
          this.currentOperations.splice(operationIndex, 1);

          // Update player resources based on result
          const playerStore = usePlayerStore();
          if (playerStore.profile) {
            // Update money
            if (result.moneyGained) {
              playerStore.profile.money += result.moneyGained;
            }
            if (result.moneyLost) {
              playerStore.profile.money -= result.moneyLost;
            }

            // Update crew
            if (result.crewGained) {
              playerStore.profile.crew += result.crewGained;
            }
            if (result.crewLost) {
              playerStore.profile.crew -= result.crewLost;
            }

            // Update weapons
            if (result.weaponsGained) {
              playerStore.profile.weapons += result.weaponsGained;
            }
            if (result.weaponsLost) {
              playerStore.profile.weapons -= result.weaponsLost;
            }

            // Update vehicles
            if (result.vehiclesGained) {
              playerStore.profile.vehicles += result.vehiclesGained;
            }
            if (result.vehiclesLost) {
              playerStore.profile.vehicles -= result.vehiclesLost;
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
              playerStore.profile.heat -= result.heatReduced;
              if (playerStore.profile.heat < 0) {
                playerStore.profile.heat = 0;
              }
            }
          }
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

            // If operation is complete
            if (now >= completionTime) {
              this.collectOperation(operation.id);
            }
          }
        }
      });
    }
  }
});
