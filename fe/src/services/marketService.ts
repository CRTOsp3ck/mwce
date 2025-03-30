// src/services/marketService.ts

import api from './api';
import { 
  MarketListing, 
  ResourceType, 
  MarketTransaction,
  MarketHistory
} from '@/types/market';

// Endpoints
const ENDPOINTS = {
  LISTINGS: '/market/listings',
  TRANSACTIONS: '/market/transactions',
  HISTORY: '/market/history',
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
    return api.get<MarketListing>(`${ENDPOINTS.LISTINGS}/${resourceType}`);
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
    return api.get<MarketHistory>(`${ENDPOINTS.HISTORY}/${resourceType}`);
  },
  
  /**
   * Buy a resource from the market
   */
  buyResource(resourceType: ResourceType, quantity: number) {
    return api.post<MarketTransaction>(ENDPOINTS.BUY, {
      resourceType,
      quantity
    });
  },
  
  /**
   * Sell a resource to the market
   */
  sellResource(resourceType: ResourceType, quantity: number) {
    return api.post<MarketTransaction>(ENDPOINTS.SELL, {
      resourceType,
      quantity
    });
  }
};