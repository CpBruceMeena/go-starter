package router

import (
	"net/http"

	"github.com/CpBruceMeena/go-starter/internal/business"
	"github.com/CpBruceMeena/go-starter/internal/logger"
	"github.com/CpBruceMeena/go-starter/internal/middleware"
	"github.com/CpBruceMeena/go-starter/internal/response"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupRoutes configures all application routes
func SetupRoutes(e *echo.Echo, userService business.UserService, log *logger.Logger) {
	// Add middleware
	e.Use(echomiddleware.Recover())
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
		return c.JSON(http.StatusOK, response.Success(map[string]string{
			"status": "ok",
		}))
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
