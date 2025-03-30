// src/types/market.ts

export interface MarketListing {
  id: string;
  type: ResourceType;
  price: number;
  quantity: number;
  trend: PriceTrend;
  trendPercentage: number;
}

export enum ResourceType {
  CREW = "crew",
  WEAPONS = "weapons",
  VEHICLES = "vehicles",
}

export enum PriceTrend {
  UP = "up",
  DOWN = "down",
  STABLE = "stable",
}

export interface MarketTransaction {
  id: string;
  playerId: string;
  resourceType: ResourceType;
  quantity: number;
  price: number;
  totalCost: number;
  timestamp: string;
  transactionType: TransactionType;
}

export enum TransactionType {
  BUY = "buy",
  SELL = "sell",
}

export interface MarketHistory {
  resourceType: ResourceType;
  timePoints: string[]; // Timestamps
  prices: number[];
}
