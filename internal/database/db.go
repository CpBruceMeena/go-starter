package database

import (
	"fmt"
	"time"

	"github.com/your-org/go-starter/internal/config"
	"github.com/your-org/go-starter/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// InitDB initializes the database connection
// Supports PostgreSQL and SQLite
func InitDB(databaseDSN string, dbCfg *config.DatabaseConfig) (*gorm.DB, error) {
	if databaseDSN == "" {
		databaseDSN = "test.db" // Default SQLite for local development
	}

	// Parse slow query threshold from config
	var slowThreshold time.Duration
	if dbCfg.SlowQueryThreshold != "" {
		var err error
		slowThreshold, err = time.ParseDuration(dbCfg.SlowQueryThreshold)
		if err != nil {
			return nil, fmt.Errorf("invalid slow_query_threshold: %w", err)
		}
	}

	// Try to detect database type
	var dialector gorm.Dialector
	if len(databaseDSN) > 10 && databaseDSN[:10] == "postgres://" {
		dialector = postgres.Open(databaseDSN)
	} else if len(databaseDSN) > 5 && databaseDSN[:5] == "user=" {
		dialector = postgres.Open(databaseDSN)
	} else {
		dialector = sqlite.Open(databaseDSN)
	}

	gormCfg := &gorm.Config{}
	if slowThreshold > 0 {
		gormCfg.Logger = gormLogger.Default.LogMode(gormLogger.Warn)
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
	)
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
