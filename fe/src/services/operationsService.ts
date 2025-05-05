// src/services/operationsService.ts

import api, { ApiResponse } from './api';
import {
  Operation,
  OperationAttempt,
  OperationResources,
  OperationResult,
  OperationsRefreshInfo
} from '@/types/operations';

// Define types that match backend models
export interface StartOperationRequest {
  resources: OperationResources;
}

// Endpoints
const ENDPOINTS = {
  OPERATIONS: '/operations',
  CURRENT: '/operations/current',
  COMPLETED: '/operations/completed',
  REFRESH_INFO: '/operations/refresh-info'
};

export default {
  getOperationsRefreshInfo() {
    return api.get<OperationsRefreshInfo>(ENDPOINTS.REFRESH_INFO);
  },

  /**
   * Get all available operations
   */
  getAvailableOperations() {
    return api.get<Operation[]>(ENDPOINTS.OPERATIONS);
  },

  /**
   * Get a specific operation
   */
  getOperation(operationId: string) {
    return api.get<Operation>(`${ENDPOINTS.OPERATIONS}/${operationId}`);
  },

  /**
   * Get operations in progress
   */
  getCurrentOperations() {
    return api.get<OperationAttempt[]>(ENDPOINTS.CURRENT);
  },

  /**
   * Get completed operations
   */
  getCompletedOperations() {
    return api.get<OperationAttempt[]>(ENDPOINTS.COMPLETED);
  },

  /**
   * Start a new operation
   */
  startOperation(operationId: string, resources: OperationResources) {
    const request: StartOperationRequest = {
      resources
    };
    return api.post<OperationAttempt>(`${ENDPOINTS.OPERATIONS}/${operationId}/start`, request);
  },

  /**
   * Cancel an in-progress operation
   */
  cancelOperation(operationId: string) {
    return api.post<{status: string}>(`${ENDPOINTS.OPERATIONS}/${operationId}/cancel`);
  },

  /**
   * Collect rewards from a completed operation
   */
  collectOperation(operationId: string) {
    return api.post<OperationResult>(`${ENDPOINTS.OPERATIONS}/${operationId}/collect`);
  }
};
