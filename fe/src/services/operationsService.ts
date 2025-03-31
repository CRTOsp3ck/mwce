// src/services/operationsService.ts

import api, { ApiResponse } from './api';
import { 
  Operation, 
  OperationAttempt, 
  OperationResources, 
  OperationResult 
} from '@/types/operations';

// Define types that match backend models
export interface StartOperationRequest {
  resources: OperationResources;
}

export interface OperationResultWithMessage {
  result: OperationResult;
  gameMessage: {
    type: string;
    message: string;
  };
}

// Endpoints
const ENDPOINTS = {
  OPERATIONS: '/operations',
  CURRENT: '/operations/current',
  COMPLETED: '/operations/completed'
};

export default {
  /**
   * Get all available operations
   */
  getAvailableOperations() {
    return api.get<ApiResponse<Operation[]>>(ENDPOINTS.OPERATIONS);
  },
  
  /**
   * Get a specific operation
   */
  getOperation(operationId: string) {
    return api.get<ApiResponse<Operation>>(`${ENDPOINTS.OPERATIONS}/${operationId}`);
  },
  
  /**
   * Get operations in progress
   */
  getCurrentOperations() {
    return api.get<ApiResponse<OperationAttempt[]>>(ENDPOINTS.CURRENT);
  },
  
  /**
   * Get completed operations
   */
  getCompletedOperations() {
    return api.get<ApiResponse<OperationAttempt[]>>(ENDPOINTS.COMPLETED);
  },
  
  /**
   * Start a new operation
   */
  startOperation(operationId: string, resources: OperationResources) {
    const request: StartOperationRequest = {
      resources
    };
    return api.post<ApiResponse<OperationAttempt>>(`${ENDPOINTS.OPERATIONS}/${operationId}/start`, request);
  },
  
  /**
   * Cancel an in-progress operation
   */
  cancelOperation(operationId: string) {
    return api.post<ApiResponse<{status: string}>>(`${ENDPOINTS.OPERATIONS}/${operationId}/cancel`);
  },
  
  /**
   * Collect rewards from a completed operation
   */
  collectOperation(operationId: string) {
    return api.post<ApiResponse<OperationResultWithMessage>>(`${ENDPOINTS.OPERATIONS}/${operationId}/collect`);
  }
};