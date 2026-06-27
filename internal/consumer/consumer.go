package consumer

import (
	"context"
	"fmt"

	"github.com/your-org/go-starter/internal/logger"
)

// Message represents a message received from queue/broker
type Message struct {
	ID    string
	Body  string
	Raw   interface{} // Raw message for custom processing
}

// Handler processes a message from the consumer
type Handler func(ctx context.Context, msg *Message) error

// Consumer defines the interface for message consumers
type Consumer interface {
	// Start begins consuming messages
	Start(ctx context.Context) error

	// Stop gracefully stops consuming messages
	Stop(ctx context.Context) error

	// Register registers a message handler
	Register(handler Handler)

	// IsRunning checks if consumer is actively consuming
	IsRunning() bool
}

// Config holds consumer configuration
type Config struct {
	Type            string // "sqs" or "kafka"
	SQS             *SQSConfig
	Kafka           *KafkaConfig
	MaxConcurrency  int           // Number of concurrent message processors
	HandlerTimeout  int           // Timeout for handler in seconds
	ErrorHandler    func(error)   // Custom error handler
	Logger          *logger.Logger
}

// New creates a new consumer based on config
func New(cfg *Config) (Consumer, error) {
	if cfg.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if cfg.Type == "" {
		return nil, fmt.Errorf("consumer type must be specified (sqs or kafka)")
	}

	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 1
	}

	if cfg.HandlerTimeout <= 0 {
		cfg.HandlerTimeout = 30
	}

	switch cfg.Type {
	case "sqs":
		if cfg.SQS == nil {
			return nil, fmt.Errorf("SQS config is required")
		}
		return NewSQSConsumer(cfg)

	case "kafka":
		if cfg.Kafka == nil {
			return nil, fmt.Errorf("Kafka config is required")
		}
		return NewKafkaConsumer(cfg)

	default:
		return nil, fmt.Errorf("unsupported consumer type: %s", cfg.Type)
	}
}
