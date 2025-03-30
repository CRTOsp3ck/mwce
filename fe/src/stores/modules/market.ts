// src/stores/modules/market.ts

import { defineStore } from "pinia";
import marketService from "@/services/marketService";
import {
  MarketListing,
  ResourceType,
  PriceTrend,
  MarketTransaction,
  TransactionType,
  MarketHistory,
} from "@/types/market";
import { usePlayerStore } from "./player";

interface MarketState {
  listings: MarketListing[];
  transactions: MarketTransaction[];
  priceHistory: MarketHistory[];
  isLoading: boolean;
  error: string | null;
}

export const useMarketStore = defineStore("market", {
  state: (): MarketState => ({
    listings: [],
    transactions: [],
    priceHistory: [],
    isLoading: false,
    error: null,
  }),

  getters: {
    crewListing: (state) => {
      return state.listings.find((l) => l.type === ResourceType.CREW);
    },

    weaponsListing: (state) => {
      return state.listings.find((l) => l.type === ResourceType.WEAPONS);
    },

    vehiclesListing: (state) => {
      return state.listings.find((l) => l.type === ResourceType.VEHICLES);
    },

    recentTransactions: (state) => {
      return state.transactions.slice(0, 10);
    },

    buyTransactions: (state) => {
      return state.transactions.filter(
        (t) => t.transactionType === TransactionType.BUY
      );
    },

    sellTransactions: (state) => {
      return state.transactions.filter(
        (t) => t.transactionType === TransactionType.SELL
      );
    },

    crewPriceHistory: (state) => {
      return state.priceHistory.find(
        (h) => h.resourceType === ResourceType.CREW
      );
    },

    weaponsPriceHistory: (state) => {
      return state.priceHistory.find(
        (h) => h.resourceType === ResourceType.WEAPONS
      );
    },

    vehiclesPriceHistory: (state) => {
      return state.priceHistory.find(
        (h) => h.resourceType === ResourceType.VEHICLES
      );
    },
  },

  actions: {
    async fetchMarketData() {
      this.isLoading = true;
      this.error = null;

      try {
        const listingsResponse = await marketService.getListings();
        this.listings = listingsResponse.data;
        console.log('Listings response: ', listingsResponse.data)

        const transactionsResponse = await marketService.getTransactions();
        this.transactions = transactionsResponse.data;

        const historyResponse = await marketService.getPriceHistory();
        this.priceHistory = historyResponse.data;
      } catch (error) {
        this.error = "Failed to load market data";
        console.error("Error fetching market data:", error);

        // For development, set mock market data
        // this.setMockMarketData();
      } finally {
        this.isLoading = false;
      }
    },

    async buyResource(resourceType: ResourceType, quantity: number) {
      this.isLoading = true;
      this.error = null;

      try {
        // Find the listing for this resource
        const listing = this.listings.find((l) => l.type === resourceType);

        if (!listing) {
          throw new Error(`No listing found for ${resourceType}`);
        }

        // Calculate total cost
        const totalCost = listing.price * quantity;

        // Check if player has enough money
        const playerStore = usePlayerStore();
        if (!playerStore.profile || playerStore.profile.money < totalCost) {
          throw new Error("Not enough money to complete this purchase");
        }

        // Execute transaction
        const response = await marketService.buyResource(
          resourceType,
          quantity
        );
        const transaction = response.data;

        // Update player resources
        playerStore.profile.money -= totalCost;

        switch (resourceType) {
          case ResourceType.CREW:
            playerStore.profile.crew += quantity;
            break;
          case ResourceType.WEAPONS:
            playerStore.profile.weapons += quantity;
            break;
          case ResourceType.VEHICLES:
            playerStore.profile.vehicles += quantity;
            break;
        }

        // Add transaction to list
        this.transactions.unshift(transaction);

        // Update listing with new price (in case price changed)
        const updatedListing = await marketService.getListing(resourceType);
        const listingIndex = this.listings.findIndex(
          (l) => l.type === resourceType
        );

        if (listingIndex !== -1) {
          this.listings[listingIndex] = updatedListing.data;
        }

        return transaction;
      } catch (error) {
        this.error = "Failed to buy resource";
        console.error("Error buying resource:", error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    async sellResource(resourceType: ResourceType, quantity: number) {
      this.isLoading = true;
      this.error = null;

      try {
        // Find the listing for this resource
        const listing = this.listings.find((l) => l.type === resourceType);

        if (!listing) {
          throw new Error(`No listing found for ${resourceType}`);
        }

        // Check if player has enough of the resource
        const playerStore = usePlayerStore();
        if (!playerStore.profile) {
          throw new Error("Player profile not loaded");
        }

        let playerQuantity = 0;
        switch (resourceType) {
          case ResourceType.CREW:
            playerQuantity = playerStore.profile.crew;
            break;
          case ResourceType.WEAPONS:
            playerQuantity = playerStore.profile.weapons;
            break;
          case ResourceType.VEHICLES:
            playerQuantity = playerStore.profile.vehicles;
            break;
        }

        if (playerQuantity < quantity) {
          throw new Error(`Not enough ${resourceType} to complete this sale`);
        }

        // Calculate total value
        const totalValue = listing.price * quantity;

        // Execute transaction
        const response = await marketService.sellResource(
          resourceType,
          quantity
        );
        const transaction = response.data;

        // Update player resources
        playerStore.profile.money += totalValue;

        switch (resourceType) {
          case ResourceType.CREW:
            playerStore.profile.crew -= quantity;
            break;
          case ResourceType.WEAPONS:
            playerStore.profile.weapons -= quantity;
            break;
          case ResourceType.VEHICLES:
            playerStore.profile.vehicles -= quantity;
            break;
        }

        // Add transaction to list
        this.transactions.unshift(transaction);

        // Update listing with new price (in case price changed)
        const updatedListing = await marketService.getListing(resourceType);
        const listingIndex = this.listings.findIndex(
          (l) => l.type === resourceType
        );

        if (listingIndex !== -1) {
          this.listings[listingIndex] = updatedListing.data;
        }

        return transaction;
      } catch (error) {
        this.error = "Failed to sell resource";
        console.error("Error selling resource:", error);
        return null;
      } finally {
        this.isLoading = false;
      }
    },

    // Mock data for development
    setMockMarketData() {
      // Create mock listings
      this.listings = [
        {
          id: "listing1",
          type: ResourceType.CREW,
          price: 1000,
          quantity: 999,
          trend: PriceTrend.UP,
          trendPercentage: 5,
        },
        {
          id: "listing2",
          type: ResourceType.WEAPONS,
          price: 2000,
          quantity: 999,
          trend: PriceTrend.DOWN,
          trendPercentage: 3,
        },
        {
          id: "listing3",
          type: ResourceType.VEHICLES,
          price: 5000,
          quantity: 999,
          trend: PriceTrend.STABLE,
          trendPercentage: 0,
        },
      ];

      // Create mock transactions
      this.transactions = [
        {
          id: "transaction1",
          playerId: "1",
          resourceType: ResourceType.CREW,
          quantity: 3,
          price: 950,
          totalCost: 2850,
          timestamp: new Date(Date.now() - 2 * 3600 * 1000).toISOString(),
          transactionType: TransactionType.BUY,
        },
        {
          id: "transaction2",
          playerId: "1",
          resourceType: ResourceType.WEAPONS,
          quantity: 2,
          price: 2100,
          totalCost: 4200,
          timestamp: new Date(Date.now() - 6 * 3600 * 1000).toISOString(),
          transactionType: TransactionType.BUY,
        },
        {
          id: "transaction3",
          playerId: "1",
          resourceType: ResourceType.VEHICLES,
          quantity: 1,
          price: 4800,
          totalCost: 4800,
          timestamp: new Date(Date.now() - 24 * 3600 * 1000).toISOString(),
          transactionType: TransactionType.SELL,
        },
      ];

      // Create mock price history (last 7 days)
      const now = new Date();
      const timePoints = [];
      const crewPrices = [];
      const weaponsPrices = [];
      const vehiclesPrices = [];

      for (let i = 6; i >= 0; i--) {
        const date = new Date(now.getTime() - i * 24 * 3600 * 1000);
        timePoints.push(date.toISOString());

        // Generate some price fluctuations
        const crewBasePrice = 1000;
        const weaponsBasePrice = 2000;
        const vehiclesBasePrice = 5000;

        crewPrices.push(crewBasePrice + Math.floor(Math.random() * 200) - 100);
        weaponsPrices.push(
          weaponsBasePrice + Math.floor(Math.random() * 300) - 150
        );
        vehiclesPrices.push(
          vehiclesBasePrice + Math.floor(Math.random() * 500) - 250
        );
      }

      this.priceHistory = [
        {
          resourceType: ResourceType.CREW,
          timePoints,
          prices: crewPrices,
        },
        {
          resourceType: ResourceType.WEAPONS,
          timePoints,
          prices: weaponsPrices,
        },
        {
          resourceType: ResourceType.VEHICLES,
          timePoints,
          prices: vehiclesPrices,
        },
      ];
    },
  },
});
