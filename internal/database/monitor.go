package database

import (
	"context"
	"fmt"
	"time"

	"github.com/CpBruceMeena/go-starter/internal/logger"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// MonitorConfig holds configuration for database query monitoring
type MonitorConfig struct {
	SlowQueryThreshold time.Duration
	QueryTimeout       time.Duration
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	Enabled            bool
}

// Monitor wraps database operations with timeout and logging
type Monitor struct {
	db     *gorm.DB
	config MonitorConfig
	log    *logger.Logger
}

// NewMonitor creates a database monitor with configurable limits
func NewMonitor(db *gorm.DB, log *logger.Logger, cfg MonitorConfig) *Monitor {
	return &Monitor{
		db:     db,
		config: cfg,
		log:    log,
	}
}

// WithTimeout wraps a context with query timeout if configured
func (m *Monitor) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if m.config.QueryTimeout > 0 {
		return context.WithTimeout(ctx, m.config.QueryTimeout)
	}
	return context.WithCancel(ctx)
}

// LogSlowQuery logs a slow query warning with duration and SQL
func (m *Monitor) LogSlowQuery(duration time.Duration, sql string, rowsAffected int64) {
	if m.config.SlowQueryThreshold > 0 && duration > m.config.SlowQueryThreshold {
		m.log.Warn("slow database query detected",
			"duration_ms", duration.Milliseconds(),
			"sql", sql,
			"rows_affected", rowsAffected,
			"threshold_ms", m.config.SlowQueryThreshold.Milliseconds(),
		)
	}
}

// GormLogger returns a GORM logger that tracks slow queries
func (m *Monitor) GormLogger() gormLogger.Interface {
	return gormLogger.Default.LogMode(gormLogger.Warn)
}

// ApplyConnectionPoolSettings applies connection pool limits to the database
func ApplyConnectionPoolSettings(db *gorm.DB, cfg MonitorConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	defaultLog := logger.Default()
	defaultLog.Info("database connection pool configured",
		"max_open_conns", cfg.MaxOpenConns,
		"max_idle_conns", cfg.MaxIdleConns,
		"conn_max_lifetime", cfg.ConnMaxLifetime.String(),
	)

	return nil
}

// SetupMonitoring initializes database monitoring with the given configuration
func SetupMonitoring(db *gorm.DB, cfg MonitorConfig) (*Monitor, error) {
	if err := ApplyConnectionPoolSettings(db, cfg); err != nil {
		return nil, err
	}

	return NewMonitor(db, logger.Default(), cfg), nil
}
