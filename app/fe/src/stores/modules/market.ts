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
        // Get listings
        const listingsResponse = await marketService.getListings();
        if (listingsResponse.success && listingsResponse.data) {
          this.listings = listingsResponse.data;
        }

        // Get transactions
        const transactionsResponse = await marketService.getTransactions();
        if (transactionsResponse.success && transactionsResponse.data) {
          this.transactions = transactionsResponse.data;
        }

        // Get price history
        const historyResponse = await marketService.getPriceHistory();
        if (historyResponse.success && historyResponse.data) {
          this.priceHistory = historyResponse.data;
        }
      } catch (error) {
        this.error = "Failed to load market data";
        console.error("Error fetching market data:", error);
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
        const response = await marketService.buyResource(resourceType, quantity);

        // Check for success and process the response
        if (!response.success || !response.data) {
          throw new Error("Transaction failed");
        }

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
        const updatedListingResponse = await marketService.getListing(resourceType);
        if (updatedListingResponse.success && updatedListingResponse.data) {
          const listingIndex = this.listings.findIndex(
            (l) => l.type === resourceType
          );

          if (listingIndex !== -1) {
            this.listings[listingIndex] = updatedListingResponse.data;
          }
        }

        return {
          transaction,
          gameMessage: response.gameMessage
        };
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
        const response = await marketService.sellResource(resourceType, quantity);

        // Check for success
        if (!response.success || !response.data) {
          throw new Error("Transaction failed");
        }

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
        const updatedListingResponse = await marketService.getListing(resourceType);
        if (updatedListingResponse.success && updatedListingResponse.data) {
          const listingIndex = this.listings.findIndex(
            (l) => l.type === resourceType
          );

          if (listingIndex !== -1) {
            this.listings[listingIndex] = updatedListingResponse.data;
          }
        }

        return {
          transaction,
          gameMessage: response.gameMessage
        };
      } catch (error) {
        this.error = "Failed to sell resource";
        console.error("Error selling resource:", error);
        return null;
      } finally {
        this.isLoading = false;
      }
    }
  }
});
