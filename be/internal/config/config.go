// internal/config/config.go

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	Environment string       `yaml:"environment"`
	Server      ServerConfig `yaml:"server"`
	Database    DBConfig     `yaml:"database"`
	JWT         JWTConfig    `yaml:"jwt"`
	Game        *GameConfig  `yaml:"-"` // Loaded separately
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port            int `yaml:"port"`
	TimeoutRead     int `yaml:"timeout_read"`
	TimeoutWrite    int `yaml:"timeout_write"`
	TimeoutIdle     int `yaml:"timeout_idle"`
	TimeoutShutdown int `yaml:"timeout_shutdown"`
}

// DBConfig holds the database configuration
type DBConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	SSLMode      string `yaml:"sslmode"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

// JWTConfig holds the JWT configuration
type JWTConfig struct {
	Secret        string        `yaml:"secret"`
	TokenLifetime time.Duration `yaml:"token_lifetime"`
}

// GameConfig holds game-specific configuration
type GameConfig struct {
	MechanicsFile             string              `yaml:"mechanics_file"`
	TerritoryStructureFile    string              `yaml:"territory_structure_file"`
	OperationsFile            string              `yaml:"operations_file"`
	DailyOperationsCount      int                 `yaml:"daily_operations_count"`
	SpecialOperationsCount    int                 `yaml:"special_operations_count"`
	OperationsRefreshInterval int                 `yaml:"operations_refresh_interval"`  // in minutes
	MarketPriceUpdateInterval int                 `yaml:"market_price_update_interval"` // in minutes
	ResourceLimit             ResourceLimitConfig `yaml:"resource_limit"`
	Mechanics                 *MechanicsConfig    `yaml:"-"` // Loaded separately
}

// ResourceLimitConfig contains limits for game resources
type ResourceLimitConfig struct {
	InitialRespect   int `yaml:"initial_respect"`
	InitialInfluence int `yaml:"initial_influence"`
	InitialHeat      int `yaml:"initial_heat"`
	InitialCrew      int `yaml:"initial_crew"`
	InitialWeapons   int `yaml:"initial_weapons"`
	InitialVehicles  int `yaml:"initial_vehicles"`
	InitialMoney     int `yaml:"initial_money"`
	MaxCrew          int `yaml:"max_crew"`
	MaxWeapons       int `yaml:"max_weapons"`
	MaxVehicles      int `yaml:"max_vehicles"`
}

// LoadConfig loads a configuration from a YAML file
func LoadConfig(path string, config interface{}) error {
	// Read the file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Print first 100 characters for debugging
	preview := string(data)
	if len(preview) > 100 {
		preview = preview[:100]
	}
	fmt.Printf("First 100 chars of %s: %s...\n", filepath.Base(path), preview)

	// Unmarshal the data
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to unmarshal config from %s: %w", path, err)
	}

	return nil
}

// CheckFileExists verifies that a file exists and is readable
func CheckFileExists(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Check if file is readable
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("file exists but cannot be opened (permissions?): %w", err)
	}
	defer file.Close()

	return nil
}

// LoadAllConfigs loads the app config, game config, and mechanics config
func LoadAllConfigs(appConfigPath string) (*Config, error) {
	fmt.Printf("Loading configs from base path: %s\n", appConfigPath)

	// Check if app config exists
	if err := CheckFileExists(appConfigPath); err != nil {
		return nil, fmt.Errorf("app config error: %w", err)
	}

	// Load app config
	config := &Config{}
	if err := LoadConfig(appConfigPath, config); err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}
	fmt.Printf("App config loaded successfully\n")

	// Determine gameConfigPath based on app config location
	baseDir := filepath.Dir(appConfigPath)
	gameConfigPath := filepath.Join(baseDir, "game.yaml")
	fmt.Printf("Game config path: %s\n", gameConfigPath)

	// Check if game config exists
	if err := CheckFileExists(gameConfigPath); err != nil {
		return nil, fmt.Errorf("game config error: %w", err)
	}

	// Load game config
	gameConfig := &GameConfig{}
	if err := LoadConfig(gameConfigPath, gameConfig); err != nil {
		return nil, fmt.Errorf("failed to load game config: %w", err)
	}
	fmt.Printf("Game config loaded successfully\n")

	// Print game config values for debugging
	fmt.Printf("Game Config Values:\n")
	fmt.Printf("  DailyOpCount: %d\n", gameConfig.DailyOperationsCount)
	fmt.Printf("  SpecialOpCount: %d\n", gameConfig.SpecialOperationsCount)
	fmt.Printf("  MarketUpdateInterval: %d\n", gameConfig.MarketPriceUpdateInterval)
	fmt.Printf("  Mechanics File: %s\n", gameConfig.MechanicsFile)

	// Set game config in main config
	config.Game = gameConfig

	mechanicsPath := filepath.Join(baseDir, "mechanics.yaml")
	fmt.Printf("Mechanics config path: %s\n", mechanicsPath)

	// Check if mechanics config exists
	if err := CheckFileExists(mechanicsPath); err != nil {
		return nil, fmt.Errorf("mechanics config error: %w", err)
	}

	mechanicsConfig := &MechanicsConfig{}
	if err := LoadConfig(mechanicsPath, mechanicsConfig); err != nil {
		return nil, fmt.Errorf("failed to load mechanics config: %w", err)
	}
	fmt.Printf("Mechanics config loaded successfully\n")

	// Verify market section is loaded
	if mechanicsConfig.Market.PriceFluctuationRange == 0 {
		fmt.Printf("WARNING: Market price fluctuation range not loaded (zero value)\n")
	} else {
		fmt.Printf("Market price fluctuation range: %d\n", mechanicsConfig.Market.PriceFluctuationRange)
	}

	if len(mechanicsConfig.Market.BasePrices) == 0 {
		fmt.Printf("WARNING: Market base prices not loaded (empty map)\n")
	} else {
		fmt.Printf("Market base prices loaded with %d entries\n", len(mechanicsConfig.Market.BasePrices))
	}

	gameConfig.Mechanics = mechanicsConfig

	return config, nil
}

// MechanicsConfig holds the game mechanics configuration
type MechanicsConfig struct {
	SuccessChances map[string]SuccessChance `yaml:"success_chances"`
	DefenseValues  DefenseValues            `yaml:"defense_values"`
	Income         IncomeConfig             `yaml:"income"`
	Market         MarketConfig             `yaml:"market"`
	Progression    ProgressionConfig        `yaml:"progression"`
	Heat           HeatConfig               `yaml:"heat"`
	Notifications  NotificationConfig       `yaml:"notifications"`
	Operations     OperationsConfig         `yaml:"operations"`
}

// SuccessChance represents success chance configuration for an action
type SuccessChance struct {
	BaseChance         int                `yaml:"base_chance"`
	ResourceMultiplier map[string]float64 `yaml:"resource_multiplier"`
}

// DefenseValues represents resource values used for defense calculation
type DefenseValues struct {
	Crew     int `yaml:"crew"`
	Weapons  int `yaml:"weapons"`
	Vehicles int `yaml:"vehicles"`
}

// IncomeConfig represents territory income generation configuration
type IncomeConfig struct {
	BaseRates   map[string]int     `yaml:"base_rates"`
	Multipliers map[string]float64 `yaml:"multipliers"`
}

// MarketConfig represents market configuration
type MarketConfig struct {
	PriceFluctuationRange int            `yaml:"price_fluctuation_range"`
	PriceUpdateInterval   int            `yaml:"price_update_interval"`
	BasePrices            map[string]int `yaml:"base_prices"`
	MinPrices             map[string]int `yaml:"min_prices"`
	MaxPrices             map[string]int `yaml:"max_prices"`
}

// ProgressionConfig represents player progression configuration
type ProgressionConfig struct {
	TitleRequirements    map[string]map[string]int     `yaml:"title_requirements"`
	ResourceCapIncreases map[string]map[string]float64 `yaml:"resource_cap_increases"`
}

// HeatConfig represents heat mechanics configuration
type HeatConfig struct {
	DecayRate int                               `yaml:"decay_rate"`
	MaxHeat   int                               `yaml:"max_heat"`
	Effects   map[string]map[string]map[int]int `yaml:"effects"`
}

// NotificationConfig represents notification settings
type NotificationConfig struct {
	MaxUnread int `yaml:"max_unread"`
	MaxTotal  int `yaml:"max_total"`
}

// OperationsConfig represents operation generation settings
type OperationsConfig struct {
	DurationRange                map[string]int     `yaml:"duration_range"`
	RewardScaling                map[string]int     `yaml:"reward_scaling"`
	DifficultyMultipliers        map[string]float64 `yaml:"difficulty_multipliers"`
	SpecialOperationRequirements map[string]int     `yaml:"special_operation_requirements"`
}
