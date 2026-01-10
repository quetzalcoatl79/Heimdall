package database

import (
	"fmt"

	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps gorm.DB
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig) (*DB, error) {
	logLevel := logger.Silent
	if cfg.SSLMode == "disable" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return &DB{db}, nil
}

// AutoMigrate runs GORM auto migrations (for development)
func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Plugin{},
		&models.Job{},
		&models.AuditLog{},
		&models.RefreshToken{},
	)
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
