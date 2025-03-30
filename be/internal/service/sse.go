// internal/service/sse.go

package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SSEService handles server-sent events for real-time updates
type SSEService interface {
	HandleConnection(w http.ResponseWriter, r *http.Request)
	SendEventToPlayer(playerID string, eventType string, data interface{})
	SendEventToAll(eventType string, data interface{})
	Close()
}

type sseService struct {
	clients      map[string]map[string]http.ResponseWriter
	clientsMutex sync.RWMutex
	logger       zerolog.Logger
}

// NewSSEService creates a new SSE service
func NewSSEService(logger zerolog.Logger) SSEService {
	return &sseService{
		clients:      make(map[string]map[string]http.ResponseWriter),
		clientsMutex: sync.RWMutex{},
		logger:       logger,
	}
}

// HandleConnection establishes an SSE connection for a player
func (s *sseService) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Get player ID from context
	playerID, ok := r.Context().Value("userID").(string)
	if !ok {
		s.logger.Error().Msg("Failed to get player ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Generate a unique client ID
	clientID := uuid.New().String()

	// Register the client connection
	s.clientsMutex.Lock()
	if _, exists := s.clients[playerID]; !exists {
		s.clients[playerID] = make(map[string]http.ResponseWriter)
	}
	s.clients[playerID][clientID] = w
	s.clientsMutex.Unlock()

	s.logger.Info().Str("playerID", playerID).Str("clientID", clientID).Msg("New SSE connection established")

	// Send an initial connection established event
	s.sendEvent(w, "connected", map[string]string{
		"message":  "Connection established",
		"playerID": playerID,
	})

	// Keep the connection alive with periodic heartbeats
	go func() {
		heartbeatTicker := time.NewTicker(30 * time.Second)
		defer heartbeatTicker.Stop()

		for {
			select {
			case <-heartbeatTicker.C:
				// Check if client is still registered
				s.clientsMutex.RLock()
				clients, exists := s.clients[playerID]
				_, clientExists := clients[clientID]
				s.clientsMutex.RUnlock()

				if !exists || !clientExists {
					return // Client is no longer registered
				}

				// Send heartbeat
				if err := s.sendEvent(w, "heartbeat", map[string]string{"timestamp": time.Now().Format(time.RFC3339)}); err != nil {
					s.logger.Error().Err(err).Str("playerID", playerID).Str("clientID", clientID).Msg("Failed to send heartbeat, closing connection")

					// Remove the client
					s.clientsMutex.Lock()
					delete(s.clients[playerID], clientID)
					if len(s.clients[playerID]) == 0 {
						delete(s.clients, playerID)
					}
					s.clientsMutex.Unlock()

					return
				}
			}
		}
	}()

	// Keep the connection alive until the client disconnects
	notify := r.Context().Done()
	<-notify

	// Remove the client when connection is closed
	s.clientsMutex.Lock()
	delete(s.clients[playerID], clientID)
	if len(s.clients[playerID]) == 0 {
		delete(s.clients, playerID)
	}
	s.clientsMutex.Unlock()

	s.logger.Info().Str("playerID", playerID).Str("clientID", clientID).Msg("SSE connection closed")
}

// Update the sendEvent method to handle errors
func (s *sseService) sendEvent(w http.ResponseWriter, eventType string, data interface{}) error {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal event data")
		return err
	}

	// Format the event data
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, jsonData)

	// Send the event
	_, err = fmt.Fprint(w, event)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to send SSE event")
		return err
	}

	// Flush the data to the client
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// SendEventToPlayer sends an event to a specific player
func (s *sseService) SendEventToPlayer(playerID string, eventType string, data interface{}) {
	s.clientsMutex.RLock()
	clients, exists := s.clients[playerID]
	s.clientsMutex.RUnlock()

	if !exists {
		return
	}

	for _, client := range clients {
		s.sendEvent(client, eventType, data)
	}
}

// SendEventToAll sends an event to all connected clients
func (s *sseService) SendEventToAll(eventType string, data interface{}) {
	s.clientsMutex.RLock()
	for _, clients := range s.clients {
		for _, client := range clients {
			s.sendEvent(client, eventType, data)
		}
	}
	s.clientsMutex.RUnlock()
}

// Close closes all client connections
func (s *sseService) Close() {
	s.clientsMutex.Lock()
	s.clients = make(map[string]map[string]http.ResponseWriter)
	s.clientsMutex.Unlock()
}
