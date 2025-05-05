// internal/controller/operations.go

package controller

import (
	"encoding/json"
	"net/http"

	"mwce-be/internal/middleware"
	"mwce-be/internal/model"
	"mwce-be/internal/service"
	"mwce-be/internal/util"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// OperationsController handles operations-related HTTP requests
type OperationsController struct {
	operationsService service.OperationsService
	logger            zerolog.Logger
}

// NewOperationsController creates a new operations controller
func NewOperationsController(operationsService service.OperationsService, logger zerolog.Logger) *OperationsController {
	return &OperationsController{
		operationsService: operationsService,
		logger:            logger,
	}
}

// GetAvailableOperations handles getting available operations
func (c *OperationsController) GetAvailableOperations(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get available operations
	operations, err := c.operationsService.GetAvailableOperations(playerID, false)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get available operations")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get available operations")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, operations)
}

// GetOperation handles getting a specific operation
func (c *OperationsController) GetOperation(w http.ResponseWriter, r *http.Request) {
	// Get operation ID from URL
	operationID := chi.URLParam(r, "id")
	if operationID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Operation ID is required")
		return
	}

	// Get the operation
	operation, err := c.operationsService.GetOperationByID(operationID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get operation")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get operation")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, operation)
}

// GetCurrentOperations handles getting in-progress operations
func (c *OperationsController) GetCurrentOperations(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get current operations
	operations, err := c.operationsService.GetCurrentOperations(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get current operations")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get current operations")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, operations)
}

// GetCompletedOperations handles getting completed operations
func (c *OperationsController) GetCompletedOperations(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get completed operations
	operations, err := c.operationsService.GetCompletedOperations(playerID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get completed operations")
		util.RespondWithError(w, http.StatusInternalServerError, "Failed to get completed operations")
		return
	}

	// Return success response
	util.RespondWithJSON(w, http.StatusOK, operations)
}

// StartOperation handles starting a new operation
func (c *OperationsController) StartOperation(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get operation ID from URL
	operationID := chi.URLParam(r, "id")
	if operationID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Operation ID is required")
		return
	}

	// Parse request body
	var request model.StartOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Start the operation
	attempt, err := c.operationsService.StartOperation(playerID, operationID, request.Resources)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to start operation")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusCreated,
		attempt,
		util.GameMessageTypeSuccess,
		"Operation started successfully. Check back later for results.",
	)
}

// CancelOperation handles canceling an in-progress operation
func (c *OperationsController) CancelOperation(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get operation ID from URL
	operationID := chi.URLParam(r, "id")
	if operationID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Operation ID is required")
		return
	}

	// Cancel the operation
	if err := c.operationsService.CancelOperation(playerID, operationID); err != nil {
		c.logger.Error().Err(err).Msg("Failed to cancel operation")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		map[string]string{"status": "cancelled"},
		util.GameMessageTypeInfo,
		"Operation cancelled. 50% of committed resources have been returned.",
	)
}

// CollectOperation handles collecting a completed operation
func (c *OperationsController) CollectOperation(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get operation ID from URL
	operationID := chi.URLParam(r, "id")
	if operationID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Operation ID is required")
		return
	}

	// Collect the operation
	result, err := c.operationsService.CollectOperation(playerID, operationID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to collect operation")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response with game message
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		result,
		func() string {
			if result.Success {
				return util.GameMessageTypeSuccess
			}
			return util.GameMessageTypeError
		}(),
		func() string {
			if result.Success {
				return "Operation completed! Collect your rewards."
			} else {
				return "Operation failed. Check the details and collect to process losses."
			}
		}(),
	)
}

// CollectOperationReward handles collecting rewards from a completed operation
func (c *OperationsController) CollectOperationReward(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		util.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get operation ID from URL
	operationID := chi.URLParam(r, "id")
	if operationID == "" {
		util.RespondWithError(w, http.StatusBadRequest, "Operation ID is required")
		return
	}

	// Collect the operation reward
	result, err := c.operationsService.CollectOperationReward(playerID, operationID)
	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to collect operation reward")
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response with game message
	util.RespondWithGameMessage(
		w,
		http.StatusOK,
		result,
		func() string {
			if result.Success {
				return util.GameMessageTypeSuccess
			}
			return util.GameMessageTypeError
		}(),
		func() string {
			if result.Success {
				return "Rewards collected successfully!"
			} else {
				return "Operation losses processed."
			}
		}(),
	)
}
