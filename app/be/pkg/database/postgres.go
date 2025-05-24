// pkg/database/postgres.go
package database

import (
	"fmt"

	"mwce-be/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database is the interface for database operations
type Database interface {
	GetDB() *gorm.DB
	Close() error
}

// PostgresDB is a postgres implementation of the Database interface
type PostgresDB struct {
	db *gorm.DB
}

// NewPostgresDB creates a new postgres database connection
func NewPostgresDB(cfg config.DBConfig) (Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	gormConfig := &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	return &PostgresDB{db: db}, nil
}

// GetDB returns the gorm.DB instance
func (p *PostgresDB) GetDB() *gorm.DB {
	return p.db
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
