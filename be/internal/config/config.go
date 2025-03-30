// internal/config/config.go

package config

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
type Config struct {
	Environment string       `yaml:"environment"`
	Server      ServerConfig `yaml:"server"`
	Database    DBConfig     `yaml:"database"`
	JWT         JWTConfig    `yaml:"jwt"`
	Game        GameConfig   `yaml:"game"`
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
}

// ResourceLimitConfig contains limits for game resources
type ResourceLimitConfig struct {
	InitialCrew     int `yaml:"initial_crew"`
	InitialWeapons  int `yaml:"initial_weapons"`
	InitialVehicles int `yaml:"initial_vehicles"`
	InitialMoney    int `yaml:"initial_money"`
	MaxCrew         int `yaml:"max_crew"`
	MaxWeapons      int `yaml:"max_weapons"`
	MaxVehicles     int `yaml:"max_vehicles"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
