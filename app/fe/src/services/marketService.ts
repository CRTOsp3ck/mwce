// src/services/marketService.ts

import api from './api';
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
    return api.get<MarketListing[]>(ENDPOINTS.LISTINGS);
  },

  /**
   * Get a specific listing
   */
  getListing(resourceType: ResourceType) {
    return api.get<MarketListing>(`${ENDPOINTS.LISTING_BY_TYPE}/${resourceType}`);
  },

  /**
   * Get player's transaction history
   */
  getTransactions() {
    return api.get<MarketTransaction[]>(ENDPOINTS.TRANSACTIONS);
  },

  /**
   * Get market price history
   */
  getPriceHistory() {
    return api.get<MarketHistory[]>(ENDPOINTS.HISTORY);
  },

  /**
   * Get price history for a specific resource
   */
  getResourcePriceHistory(resourceType: ResourceType) {
    return api.get<MarketHistory>(`${ENDPOINTS.RESOURCE_HISTORY}/${resourceType}`);
  },

  /**
   * Buy a resource from the market
   */
  buyResource(resourceType: ResourceType, quantity: number) {
    const request: ResourceTransaction = {
      resourceType,
      quantity
    };
    return api.post<MarketTransaction>(ENDPOINTS.BUY, request);
  },

  /**
   * Sell a resource to the market
   */
  sellResource(resourceType: ResourceType, quantity: number) {
    const request: ResourceTransaction = {
      resourceType,
      quantity
    };
    return api.post<MarketTransaction>(ENDPOINTS.SELL, request);
  }
};
