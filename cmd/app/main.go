package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/CpBruceMeena/go-starter/internal/config"
	"github.com/CpBruceMeena/go-starter/internal/logger"
	"github.com/CpBruceMeena/go-starter/internal/server"
	"github.com/CpBruceMeena/go-starter/internal/worker"
)

// @title Go Starter API
// @version 1.0
// @description A production-ready Go starter template with best practices
// @host localhost:8080
// @basePath /
// @schemes http https
func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log := logger.New()

	// Determine run mode
	mode := os.Getenv("APP_MODE")
	if mode == "" {
		mode = "http" // Default to HTTP server
	}

	log.Info("application starting",
		"env", cfg.ServerEnv,
		"mode", mode,
		"port", cfg.ServerPort,
	)

	switch mode {
	case "worker":
		runWorker(ctx, cfg, log)
	case "http":
		fallthrough
	default:
		runHTTPServer(ctx, cfg, log)
	}
}

// runHTTPServer runs the application as an HTTP server
func runHTTPServer(ctx context.Context, cfg *config.Config, log *logger.Logger) {
	// Create server
	srv, err := server.New(cfg, log)
	if err != nil {
		log.Error("failed to create server", "error", err.Error())
		os.Exit(1)
	}

	// Setup routes and services
	if err := srv.Setup(); err != nil {
		log.Error("failed to setup server", "error", err.Error())
		os.Exit(1)
	}

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Error("server error", "error", err.Error())
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received")

	// Graceful shutdown
	if err := srv.Stop(ctx); err != nil {
		log.Error("shutdown error", "error", err.Error())
		os.Exit(1)
	}

	log.Info("application stopped")
}

// runWorker runs the application as a background worker
func runWorker(ctx context.Context, cfg *config.Config, log *logger.Logger) {
	// Import worker package
	// This is imported at top: "github.com/CpBruceMeena/go-starter/internal/worker"

	// Create server (for database and services)
	srv, err := server.New(cfg, log)
	if err != nil {
		log.Error("failed to create server", "error", err.Error())
		os.Exit(1)
	}

	// Setup services (but not HTTP routes)
	if err := srv.SetupServices(); err != nil {
		log.Error("failed to setup services", "error", err.Error())
		os.Exit(1)
	}

	// Create and setup worker
	w := worker.New(log)

	// Register example tasks (can be customized based on needs)
	userService := srv.GetUserService()
	worker.RegisterExampleTasks(w, log, userService)

	// Start worker
	if err := w.Start(ctx); err != nil {
		log.Error("failed to start worker", "error", err.Error())
		os.Exit(1)
	}

	log.Info("worker started", "task_count", w.TaskCount())

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received")

	// Graceful shutdown
	if err := w.Stop(ctx); err != nil {
		log.Error("worker stop error", "error", err.Error())
	}

	if err := srv.Stop(ctx); err != nil {
		log.Error("shutdown error", "error", err.Error())
		os.Exit(1)
	}

	log.Info("application stopped")
}
