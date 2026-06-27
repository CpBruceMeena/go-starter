package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger for consistent logging across the application
type Logger struct {
	*slog.Logger
}

// Global logger instance
var defaultLogger *Logger

func init() {
	defaultLogger = New()
}

// New creates a new logger instance
// Uses JSON handler for production, text handler for development
func New() *Logger {
	var handler slog.Handler
	var level slog.Level

	env := os.Getenv("ENV")
	logLevel := os.Getenv("LOG_LEVEL")

	// Determine log level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		if env == "development" {
			level = slog.LevelDebug
		} else {
			level = slog.LevelInfo
		}
	}

	// Use text handler only for development, JSON for all other environments
	if env == "development" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		// JSON handler for production, staging, and test environments
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// Default returns the global logger instance
func Default() *Logger {
	return defaultLogger
}

// SetDefault sets the global logger instance
func SetDefault(l *Logger) {
	defaultLogger = l
}

// WithContext adds context values to logger (request ID, trace ID, etc.)
func (l *Logger) WithContext(ctx context.Context) *Logger {
	var args []any

	// Extract request ID if present
	if requestID := ctx.Value("request_id"); requestID != nil {
		args = append(args, slog.String("request_id", requestID.(string)))
	}

	// Extract trace ID if present
	if traceID := ctx.Value("trace_id"); traceID != nil {
		args = append(args, slog.String("trace_id", traceID.(string)))
	}

	// Extract user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		args = append(args, slog.String("user_id", userID.(string)))
	}

	if len(args) == 0 {
		return l
	}

	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// InfoContext logs info level message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Logger.InfoContext(ctx, msg, args...)
}

// WarnContext logs warn level message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Logger.WarnContext(ctx, msg, args...)
}

// ErrorContext logs error level message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Logger.ErrorContext(ctx, msg, args...)
}

// DebugContext logs debug level message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Logger.DebugContext(ctx, msg, args...)
}
