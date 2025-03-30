// src/services/operationsService.ts

import api from './api';
import { 
  Operation, 
  OperationAttempt, 
  OperationResources, 
  OperationResult 
} from '@/types/operations';

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
    return api.post<OperationAttempt>(`${ENDPOINTS.OPERATIONS}/${operationId}/start`, {
      resources
    });
  },
  
  /**
   * Cancel an in-progress operation
   */
  cancelOperation(operationId: string) {
    return api.post(`${ENDPOINTS.OPERATIONS}/${operationId}/cancel`);
  },
  
  /**
   * Collect rewards from a completed operation
   */
  collectOperation(operationId: string) {
    return api.post<OperationResult>(`${ENDPOINTS.OPERATIONS}/${operationId}/collect`);
  }
};