package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/CpBruceMeena/go-starter/internal/logger"
)

// RequestIDMiddleware adds a request ID to the context
func RequestIDMiddleware(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add to context
			ctx := context.WithValue(c.Request().Context(), "request_id", requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Add to response header
			c.Response().Header().Set("X-Request-ID", requestID)

			return next(c)
		}
	}
}

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(log *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			ctx := c.Request().Context()

			err := next(c)

			duration := time.Since(start).Milliseconds()

			log.InfoContext(ctx, "HTTP request completed",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"duration_ms", duration,
				"remote_addr", c.RealIP(),
			)

			return err
		}
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
