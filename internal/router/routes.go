package router

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/your-org/go-starter/internal/business"
	"github.com/your-org/go-starter/internal/logger"
	"github.com/your-org/go-starter/internal/middleware"
)

// SetupRoutes configures all application routes
func SetupRoutes(e *echo.Echo, userService business.UserService, log *logger.Logger) {
	// Add middleware
	e.Use(middleware.RequestIDMiddleware(log))
	e.Use(middleware.LoggingMiddleware(log))

	// Swagger endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API v1
	api := e.Group("/api/v1")

	// User routes
	setupUserRoutes(api, userService)

	// Health check
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})
}

// setupUserRoutes sets up user API routes
func setupUserRoutes(g *echo.Group, svc business.UserService) {
	users := g.Group("/users")

	users.POST("", func(c echo.Context) error {
		return CreateUser(c, svc)
	})

	users.GET("", func(c echo.Context) error {
		return ListUsers(c, svc)
	})

	users.GET("/:id", func(c echo.Context) error {
		return GetUser(c, svc)
	})

	users.PUT("/:id", func(c echo.Context) error {
		return UpdateUser(c, svc)
	})

	users.DELETE("/:id", func(c echo.Context) error {
		return DeleteUser(c, svc)
	})
}
