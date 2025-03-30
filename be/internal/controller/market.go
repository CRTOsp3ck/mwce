// internal/controller/market.go

package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"mwce-be/internal/middleware"
	"mwce-be/internal/model"
	"mwce-be/internal/service"
	"mwce-be/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// MarketController handles market-related HTTP requests
type MarketController struct {
	marketService service.MarketService
	logger        zerolog.Logger
}

// NewMarketController creates a new market controller
func NewMarketController(marketService service.MarketService, logger zerolog.Logger) *MarketController {
	return &MarketController{
		marketService: marketService,
		logger:        logger,
	}
}

// GetListings handles getting all market listings
func (c *MarketController) GetListings(w http.ResponseWriter, r *http.Request) {
	// Get all listings
	listings, err := c.marketService.GetListings()
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get market listings")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get market listings")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, listings)
}

// GetListing handles getting a specific market listing
func (c *MarketController) GetListing(w http.ResponseWriter, r *http.Request) {
	// Get resource type from URL
	resourceType := chi.URLParam(r, "type")
	if resourceType == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Resource type is required")
		return
	}

	// Get the listing
	listing, err := c.marketService.GetListingByType(resourceType)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get market listing")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get market listing")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, listing)
}

// GetTransactions handles getting player's market transactions
func (c *MarketController) GetTransactions(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get player's transactions
	transactions, err := c.marketService.GetTransactions(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get market transactions")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get market transactions")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, transactions)
}

// GetPriceHistory handles getting market price history
func (c *MarketController) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	// Default to 7 days of history
	days := 7

	// Get price history
	history, err := c.marketService.GetPriceHistory(days)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get price history")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get price history")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, history)
}

// GetResourcePriceHistory handles getting price history for a specific resource
func (c *MarketController) GetResourcePriceHistory(w http.ResponseWriter, r *http.Request) {
	// Get resource type from URL
	resourceType := chi.URLParam(r, "type")
	if resourceType == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Resource type is required")
		return
	}

	// Default to 7 days of history
	days := 7

	// Get resource price history
	history, err := c.marketService.GetResourcePriceHistory(resourceType, days)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get resource price history")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get resource price history")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, history)
}

// BuyResource handles buying a resource from the market
func (c *MarketController) BuyResource(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request model.ResourceTransaction
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if request.ResourceType == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Resource type is required")
		return
	}

	if request.Quantity <= 0 {
		util.RespondWithError(w, http.StatusBadRequest, "Quantity must be greater than zero")
		return
	}

	// Buy the resource
	transaction, err := c.marketService.BuyResource(playerID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to buy resource")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate success message
	message := generateBuyMessage(request.Quantity, request.ResourceType, transaction.TotalCost)

	// Return success response with game message
	util.RespondWithGameMessage(
		w,
		http.StatusCreated,
		transaction,
		util.GameMessageTypeSuccess,
		message,
	)
}

// SellResource handles selling a resource to the market
func (c *MarketController) SellResource(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request model.ResourceTransaction
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate request
	if request.ResourceType == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Resource type is required")
		return
	}

	if request.Quantity <= 0 {
		util.RespondWithError(w, http.StatusBadRequest, "Quantity must be greater than zero")
		return
	}

	// Sell the resource
	transaction, err := c.marketService.SellResource(playerID, request)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to sell resource")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate success message
	message := generateSellMessage(request.Quantity, request.ResourceType, transaction.TotalCost)

	// Return success response with game message
	util.RespondWithGameMessage(
		w,
		http.StatusCreated,
		transaction,
		util.GameMessageTypeSuccess,
		message,
	)
}

// Helper functions

// Helper functions for the market controller

// generateBuyMessage generates a message for a purchase
func generateBuyMessage(quantity int, resourceType string, totalCost int) string {
	var resourceName string

	switch resourceType {
	case util.ResourceTypeCrew:
		if quantity == 1 {
			resourceName = "crew member"
		} else {
			resourceName = "crew members"
		}
	case util.ResourceTypeWeapons:
		if quantity == 1 {
			resourceName = "weapon"
		} else {
			resourceName = "weapons"
		}
	case util.ResourceTypeVehicles:
		if quantity == 1 {
			resourceName = "vehicle"
		} else {
			resourceName = "vehicles"
		}
	default:
		resourceName = "resources"
	}

	return fmt.Sprintf("You bought %s %s for $%s.", formatNumber(quantity), resourceName, formatMoney(totalCost))
}

// generateSellMessage generates a message for a sale
func generateSellMessage(quantity int, resourceType string, totalValue int) string {
	var resourceName string

	switch resourceType {
	case util.ResourceTypeCrew:
		if quantity == 1 {
			resourceName = "crew member"
		} else {
			resourceName = "crew members"
		}
	case util.ResourceTypeWeapons:
		if quantity == 1 {
			resourceName = "weapon"
		} else {
			resourceName = "weapons"
		}
	case util.ResourceTypeVehicles:
		if quantity == 1 {
			resourceName = "vehicle"
		} else {
			resourceName = "vehicles"
		}
	default:
		resourceName = "resources"
	}

	return fmt.Sprintf("You sold %s %s for $%s.", formatNumber(quantity), resourceName, formatMoney(totalValue))
}

// formatNumber formats a number with commas
func formatNumber(n int) string {
	in := strconv.FormatInt(int64(n), 10)
	out := make([]byte, len(in)+(len(in)-1)/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

// formatMoney formats money values
func formatMoney(amount int) string {
	if amount >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(amount)/1000000)
	} else if amount >= 1000 {
		return fmt.Sprintf("%.1fK", float64(amount)/1000)
	}
	return formatNumber(amount)
}
