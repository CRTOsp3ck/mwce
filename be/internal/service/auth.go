// internal/service/auth.go

package service

import (
	"errors"
	"time"

	"mwce-be/internal/config"
	"mwce-be/internal/model"
	"mwce-be/internal/repository"
	"mwce-be/internal/util"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication-related business logic
type AuthService interface {
	Register(request model.RegisterRequest) (*model.AuthResponse, error)
	Login(request model.LoginRequest) (*model.AuthResponse, error)
	ValidateToken(token string) (string, error)
}

type authService struct {
	playerRepo repository.PlayerRepository
	jwtConfig  config.JWTConfig
	logger     zerolog.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(playerRepo repository.PlayerRepository, jwtConfig config.JWTConfig, logger zerolog.Logger) AuthService {
	return &authService{
		playerRepo: playerRepo,
		jwtConfig:  jwtConfig,
		logger:     logger,
	}
}

// Register registers a new user
func (s *authService) Register(request model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if email already exists
	_, err := s.playerRepo.GetPlayerByEmail(request.Email)
	if err == nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create player
	now := time.Now()
	player := &model.Player{
		Name:        request.Name,
		Email:       request.Email,
		Password:    string(hashedPassword),
		Title:       util.PlayerTitleAssociate, // Starting title
		Money:       10000,                     // Starting money
		Crew:        5,                         // Starting crew
		MaxCrew:     25,                        // Max crew
		Weapons:     3,                         // Starting weapons
		MaxWeapons:  30,                        // Max weapons
		Vehicles:    1,                         // Starting vehicles
		MaxVehicles: 12,                        // Max vehicles
		Respect:     10,                        // Starting respect
		Influence:   5,                         // Starting influence
		Heat:        0,                         // Starting heat
		CreatedAt:   now,
		LastActive:  now,
	}

	// Create player in database
	if err := s.playerRepo.CreatePlayer(player); err != nil {
		return nil, errors.New("failed to create player")
	}

	// Create player stats
	stats := &model.PlayerStats{
		PlayerID:  player.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.playerRepo.UpdatePlayerStats(stats); err != nil {
		s.logger.Error().Err(err).Msg("Failed to create player stats")
	}

	// Generate JWT token
	token, err := util.GenerateToken(player.ID, s.jwtConfig.Secret, s.jwtConfig.TokenLifetime)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.AuthResponse{
		Token:  token,
		Player: *player,
	}, nil
}

// Login authenticates a user
func (s *authService) Login(request model.LoginRequest) (*model.AuthResponse, error) {
	// Get player by email
	player, err := s.playerRepo.GetPlayerByEmail(request.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(player.Password), []byte(request.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Update last active timestamp
	player.LastActive = time.Now()
	if err := s.playerRepo.UpdatePlayer(player); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update last active timestamp")
	}

	// Generate JWT token
	token, err := util.GenerateToken(player.ID, s.jwtConfig.Secret, s.jwtConfig.TokenLifetime)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &model.AuthResponse{
		Token:  token,
		Player: *player,
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *authService) ValidateToken(token string) (string, error) {
	// Parse and validate token
	claims, err := util.ParseToken(token, s.jwtConfig.Secret)
	if err != nil {
		return "", err
	}

	// Get the user ID from the claims
	userID := claims.UserID

	// Verify that the user exists
	_, err = s.playerRepo.GetPlayerByID(userID)
	if err != nil {
		return "", errors.New("invalid token: user not found")
	}

	return userID, nil
}
