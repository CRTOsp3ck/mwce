// src/services/marketService.ts

import api, { ApiResponse } from './api';
import { 
  MarketListing, 
  ResourceType, 
  MarketTransaction,
  MarketHistory
} from '@/types/market';

// Define types that match backend models
export interface ResourceTransaction {
  resourceType: ResourceType;
  quantity: number;
}

export interface MarketTransactionResponse {
  id: string;
  playerID: string;
  resourceType: string;
  quantity: number;
  price: number;
  totalCost: number;
  timestamp: string;
  transactionType: string;
}

export interface GameMessage {
  type: string;
  message: string;
}

export interface MarketTransactionWithMessage {
  result: MarketTransaction;
  gameMessage: GameMessage;
}

// Endpoints
const ENDPOINTS = {
  LISTINGS: '/market/listings',
  LISTING_BY_TYPE: '/market/listings', // + /:type
  TRANSACTIONS: '/market/transactions',
  HISTORY: '/market/history',
  RESOURCE_HISTORY: '/market/history', // + /:type
  BUY: '/market/buy',
  SELL: '/market/sell'
};

export default {
  /**
   * Get all market listings
   */
  getListings() {
    return api.get<ApiResponse<MarketListing[]>>(ENDPOINTS.LISTINGS);
  },
  
  /**
   * Get a specific listing
   */
  getListing(resourceType: ResourceType) {
    return api.get<ApiResponse<MarketListing>>(`${ENDPOINTS.LISTING_BY_TYPE}/${resourceType}`);
  },
  
  /**
   * Get player's transaction history
   */
  getTransactions() {
    return api.get<ApiResponse<MarketTransaction[]>>(ENDPOINTS.TRANSACTIONS);
  },
  
  /**
   * Get market price history
   */
  getPriceHistory() {
    return api.get<ApiResponse<MarketHistory[]>>(ENDPOINTS.HISTORY);
  },
  
  /**
   * Get price history for a specific resource
   */
  getResourcePriceHistory(resourceType: ResourceType) {
    return api.get<ApiResponse<MarketHistory>>(`${ENDPOINTS.RESOURCE_HISTORY}/${resourceType}`);
  },
  
  /**
   * Buy a resource from the market
   */
  buyResource(resourceType: ResourceType, quantity: number) {
    const request: ResourceTransaction = {
      resourceType,
      quantity
    };
    return api.post<ApiResponse<MarketTransactionWithMessage>>(ENDPOINTS.BUY, request);
  },
  
  /**
   * Sell a resource to the market
   */
  sellResource(resourceType: ResourceType, quantity: number) {
    const request: ResourceTransaction = {
      resourceType,
      quantity
    };
    return api.post<ApiResponse<MarketTransactionWithMessage>>(ENDPOINTS.SELL, request);
  }
};