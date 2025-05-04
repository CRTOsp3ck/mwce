package service

import (
	"mwce-be/internal/model"
	"time"
)

// SendCampaignUpdate sends a campaign-related update to a player through SSE
func (s *campaignService) SendCampaignUpdate(playerID, eventType string, data interface{}) {
	// Use the sseService to send the event
	if s.sseService != nil {
		s.sseService.SendEventToPlayer(playerID, eventType, data)
	}
}

// SendMissionUpdate sends a mission-related update to a player through SSE
func (s *campaignService) SendMissionUpdate(playerID string, mission *model.Mission, progress *model.PlayerMissionProgress) {
	if s.sseService == nil {
		return
	}

	// Create the event data
	data := map[string]interface{}{
		"mission": map[string]interface{}{
			"id":          mission.ID,
			"title":       mission.Title,
			"description": mission.Description,
			"status":      progress.Status,
			"updatedAt":   time.Now().Format(time.RFC3339),
		},
	}

	// Send the mission update event
	s.sseService.SendEventToPlayer(playerID, "mission_updated", data)
}

// SendChoiceActivatedUpdate sends a choice activation update to a player
func (s *campaignService) SendChoiceActivatedUpdate(playerID string, missionID, choiceID string) {
	if s.sseService == nil {
		return
	}

	// Get the mission and choice details
	mission, err := s.campaignRepo.GetMissionByID(missionID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get mission for SSE update")
		return
	}

	// Find the choice
	var choice *model.MissionChoice
	for _, c := range mission.Choices {
		if c.ID == choiceID {
			choice = &c
			break
		}
	}

	if choice == nil {
		s.logger.Error().Str("choiceID", choiceID).Msg("Choice not found for SSE update")
		return
	}

	// Get the player conditions for this choice
	conditions, err := s.campaignRepo.GetPlayerCompletionConditions(playerID, choiceID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get conditions for SSE update")
		return
	}

	// Create condition data for the event
	conditionData := make([]map[string]interface{}, len(conditions))
	for i, condition := range conditions {
		conditionData[i] = map[string]interface{}{
			"id":              condition.ID,
			"type":            condition.Type,
			"requiredValue":   condition.RequiredValue,
			"additionalValue": condition.AdditionalValue,
			"orderIndex":      condition.OrderIndex,
			"isCompleted":     condition.IsCompleted,
		}
	}

	// Create the event data
	data := map[string]interface{}{
		"missionID":  missionID,
		"choiceID":   choiceID,
		"choiceText": choice.Text,
		"conditions": conditionData,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	// Send the choice activation event
	s.sseService.SendEventToPlayer(playerID, "choice_activated", data)
}

// SendConditionCompletedUpdate sends a condition completion update to a player
func (s *campaignService) SendConditionCompletedUpdate(playerID string, conditionID string) {
	if s.sseService == nil {
		return
	}

	// Get the condition details
	condition, err := s.campaignRepo.GetPlayerCompletionCondition(conditionID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get condition for SSE update")
		return
	}

	// Create the event data
	data := map[string]interface{}{
		"conditionID": conditionID,
		"choiceID":    condition.ChoiceID,
		"isCompleted": true,
		"completedAt": condition.CompletedAt,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Send the condition completed event
	s.sseService.SendEventToPlayer(playerID, "condition_completed", data)
}

// SendPOIActivatedUpdate sends a POI activation update to a player
func (s *campaignService) SendPOIActivatedUpdate(playerID string, playerPOI *model.PlayerPOI) {
	if s.sseService == nil {
		return
	}

	// Create the event data
	data := map[string]interface{}{
		"poi": map[string]interface{}{
			"id":           playerPOI.ID,
			"name":         playerPOI.Name,
			"description":  playerPOI.Description,
			"locationType": playerPOI.LocationType,
			"locationID":   playerPOI.LocationID,
			"missionID":    playerPOI.MissionID,
			"choiceID":     playerPOI.ChoiceID,
			"isActive":     playerPOI.IsActive,
			"isCompleted":  playerPOI.IsCompleted,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Send the POI activated event
	s.sseService.SendEventToPlayer(playerID, "poi_activated", data)
}

// SendPOICompletedUpdate sends a POI completion update to a player
func (s *campaignService) SendPOICompletedUpdate(playerID string, playerPOI *model.PlayerPOI) {
	if s.sseService == nil {
		return
	}

	// Create the event data
	data := map[string]interface{}{
		"poiID":       playerPOI.ID,
		"missionID":   playerPOI.MissionID,
		"choiceID":    playerPOI.ChoiceID,
		"isCompleted": true,
		"completedAt": playerPOI.CompletedAt,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Send the POI completed event
	s.sseService.SendEventToPlayer(playerID, "poi_completed", data)
}

// SendOperationActivatedUpdate sends an operation activation update to a player
func (s *campaignService) SendOperationActivatedUpdate(playerID string, playerOp *model.PlayerMissionOperation) {
	if s.sseService == nil {
		return
	}

	// Create the event data
	data := map[string]interface{}{
		"operation": map[string]interface{}{
			"id":            playerOp.ID,
			"name":          playerOp.Name,
			"description":   playerOp.Description,
			"operationType": playerOp.OperationType,
			"missionID":     playerOp.MissionID,
			"choiceID":      playerOp.ChoiceID,
			"duration":      playerOp.Duration,
			"successRate":   playerOp.SuccessRate,
			"isActive":      playerOp.IsActive,
			"isCompleted":   playerOp.IsCompleted,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Send the operation activated event
	s.sseService.SendEventToPlayer(playerID, "operation_activated", data)
}

// SendOperationCompletedUpdate sends an operation completion update to a player
func (s *campaignService) SendOperationCompletedUpdate(playerID string, playerOp *model.PlayerMissionOperation) {
	if s.sseService == nil {
		return
	}

	// Create the event data
	data := map[string]interface{}{
		"operationID": playerOp.ID,
		"missionID":   playerOp.MissionID,
		"choiceID":    playerOp.ChoiceID,
		"isCompleted": true,
		"completedAt": playerOp.CompletedAt,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Send the operation completed event
	s.sseService.SendEventToPlayer(playerID, "operation_completed", data)
}
