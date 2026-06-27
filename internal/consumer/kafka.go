package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// KafkaConfig holds Kafka specific configuration
type KafkaConfig struct {
	Brokers        []string // Kafka broker addresses
	Topic          string   // Topic to consume
	ConsumerGroup  string   // Consumer group ID
	StartOffset    int64    // Start offset (0 = newest, -1 = oldest, or specific offset)
	SessionTimeout int      // Session timeout in seconds
}

// KafkaConsumer implements Consumer interface for Apache Kafka
type KafkaConsumer struct {
	cfg            *Config
	handler        Handler
	running        bool
	mu             sync.RWMutex
	stopCh         chan struct{}
	wg             sync.WaitGroup
	processingCh   chan *Message
	maxConcurrency int
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(cfg *Config) (Consumer, error) {
	if len(cfg.Kafka.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker must be specified")
	}

	if cfg.Kafka.Topic == "" {
		return nil, fmt.Errorf("topic must be specified")
	}

	if cfg.Kafka.ConsumerGroup == "" {
		cfg.Kafka.ConsumerGroup = "default-group"
	}

	if cfg.Kafka.SessionTimeout <= 0 {
		cfg.Kafka.SessionTimeout = 30
	}

	return &KafkaConsumer{
		cfg:            cfg,
		stopCh:         make(chan struct{}),
		processingCh:   make(chan *Message, cfg.MaxConcurrency),
		maxConcurrency: cfg.MaxConcurrency,
		running:        false,
	}, nil
}

// Register registers a message handler
func (c *KafkaConsumer) Register(handler Handler) {
	c.handler = handler
}

// Start begins consuming messages from Kafka
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("consumer is already running")
	}
	if c.handler == nil {
		c.mu.Unlock()
		return fmt.Errorf("handler not registered")
	}
	c.running = true
	c.mu.Unlock()

	c.cfg.Logger.Info("Kafka consumer starting",
		"brokers", fmt.Sprintf("%v", c.cfg.Kafka.Brokers),
		"topic", c.cfg.Kafka.Topic,
		"group", c.cfg.Kafka.ConsumerGroup,
	)

	// Start message processors
	for i := 0; i < c.maxConcurrency; i++ {
		c.wg.Add(1)
		go c.processMessages(ctx)
	}

	// Start message poller
	c.wg.Add(1)
	go c.pollMessages(ctx)

	return nil
}

// Stop gracefully stops the consumer
func (c *KafkaConsumer) Stop(ctx context.Context) error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	c.mu.Unlock()

	c.cfg.Logger.Info("Kafka consumer stopping")
	close(c.stopCh)

	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.cfg.Logger.Info("Kafka consumer stopped gracefully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

// IsRunning checks if consumer is actively consuming
func (c *KafkaConsumer) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// pollMessages polls messages from Kafka
func (c *KafkaConsumer) pollMessages(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// TODO: Implement actual Kafka polling using segmentio/kafka-go or confluent-kafka-go
			// For now, this is a placeholder
			// messages, err := c.readMessages(ctx)
			// if err != nil {
			//     if c.cfg.ErrorHandler != nil {
			//         c.cfg.ErrorHandler(err)
			//     }
			//     continue
			// }
			// for _, msg := range messages {
			//     select {
			//     case c.processingCh <- msg:
			//     case <-c.stopCh:
			//         return
			//     }
			// }
		}
	}
}

// processMessages processes messages from the channel
func (c *KafkaConsumer) processMessages(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		case msg := <-c.processingCh:
			if msg == nil {
				return
			}

			// Create context with timeout for handler
			handlerCtx, cancel := context.WithTimeout(ctx, time.Duration(c.cfg.HandlerTimeout)*time.Second)

			if err := c.handler(handlerCtx, msg); err != nil {
				c.cfg.Logger.Error("message handler error",
					"message_id", msg.ID,
					"topic", c.cfg.Kafka.Topic,
					"error", err.Error(),
				)
				if c.cfg.ErrorHandler != nil {
					c.cfg.ErrorHandler(err)
				}
			} else {
				c.cfg.Logger.Debug("message processed",
					"message_id", msg.ID,
					"topic", c.cfg.Kafka.Topic,
				)
			}

			cancel()
		}
	}
}
