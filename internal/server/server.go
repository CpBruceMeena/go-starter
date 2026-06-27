package server

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/CpBruceMeena/go-starter/internal/business"
	"github.com/CpBruceMeena/go-starter/internal/cache"
	"github.com/CpBruceMeena/go-starter/internal/config"
	"github.com/CpBruceMeena/go-starter/internal/database"
	"github.com/CpBruceMeena/go-starter/internal/logger"
	"github.com/CpBruceMeena/go-starter/internal/repository"
	"github.com/CpBruceMeena/go-starter/internal/router"
	"gorm.io/gorm"
)

// Server represents the application server
type Server struct {
	echo        *echo.Echo
	db          *gorm.DB
	config      *config.Config
	log         *logger.Logger
	cache       *cache.TTLCache
	userService business.UserService
}

// New creates a new server instance
func New(cfg *config.Config, log *logger.Logger) (*Server, error) {
	// Initialize database
	db, err := database.InitDB(cfg.DatabaseDSN, &cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Setup database monitoring with connection pool limits
	if err := database.ApplyConnectionPoolSettings(db, database.MonitorConfig{
		SlowQueryThreshold: parseDuration(cfg.Database.SlowQueryThreshold, 100*time.Millisecond),
		QueryTimeout:       parseDuration(cfg.Database.QueryTimeout, 5*time.Second),
		MaxOpenConns:       cfg.Database.MaxOpenConns,
		MaxIdleConns:       cfg.Database.MaxIdleConns,
		ConnMaxLifetime:    parseDuration(cfg.Database.ConnMaxLifetime, time.Hour),
		Enabled:            true,
	}); err != nil {
		return nil, fmt.Errorf("failed to setup database monitoring: %w", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	e := echo.New()

	return &Server{
		echo:   e,
		db:     db,
		cache:  nil,
		config: cfg,
		log:    log,
	}, nil
}

// SetupServices initializes all services without HTTP routes
func (s *Server) SetupServices() error {
	// Initialize cache if not already done
	if s.cache == nil {
		s.cache = cache.NewTTLCache()
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(s.db)

	// Initialize business services
	s.userService = business.NewUserService(userRepo, s.cache, s.log)

	return nil
}

// Setup initializes all routes and services
func (s *Server) Setup() error {
	// Setup services first
	if err := s.SetupServices(); err != nil {
		return err
	}

	// Setup routes
	router.SetupRoutes(s.echo, s.userService, s.log)

	return nil
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.ServerPort)
	s.log.Info("starting server", "port", s.config.ServerPort)
	return s.echo.Start(addr)
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("shutting down server")

	// Close database connection
	if err := database.Close(s.db); err != nil {
		s.log.Error("error closing database", "error", err.Error())
	}

	// Shutdown HTTP server with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.echo.Shutdown(shutdownCtx)
}

// Echo returns the echo instance
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// DB returns the database connection
func (s *Server) DB() *gorm.DB {
	return s.db
}

// GetUserService returns the user service
func (s *Server) GetUserService() business.UserService {
	return s.userService
}

// parseDuration parses a duration string with a fallback default
func parseDuration(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}
