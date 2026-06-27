package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSConsumer implements Consumer interface for AWS SQS
type SQSConsumer struct {
	cfg            *Config
	handler        Handler
	running        bool
	mu             sync.RWMutex
	stopCh         chan struct{}
	wg             sync.WaitGroup
	processingCh   chan *Message
	maxConcurrency int
	sqsClient      *sqs.Client
	queueURL       string
}

// NewSQSConsumer creates a new SQS consumer
func NewSQSConsumer(cfg *Config) (Consumer, error) {
	if cfg.SQS.MaxMessages < 1 || cfg.SQS.MaxMessages > 10 {
		cfg.SQS.MaxMessages = 10
	}

	if cfg.SQS.WaitTimeSecond < 0 || cfg.SQS.WaitTimeSecond > 20 {
		cfg.SQS.WaitTimeSecond = 20
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(cfg.SQS.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	consumer := &SQSConsumer{
		cfg:            cfg,
		stopCh:         make(chan struct{}),
		processingCh:   make(chan *Message, cfg.MaxConcurrency),
		maxConcurrency: cfg.MaxConcurrency,
		running:        false,
		sqsClient:      sqs.NewFromConfig(awsCfg),
		queueURL:       cfg.SQS.QueueURL,
	}

	cfg.Logger.Info("SQS consumer configured", "queue_url", cfg.SQS.QueueURL, "max_messages", cfg.SQS.MaxMessages)
	return consumer, nil
}

// Register registers a message handler
func (c *SQSConsumer) Register(handler Handler) {
	c.handler = handler
}

// Start begins consuming messages from SQS
func (c *SQSConsumer) Start(ctx context.Context) error {
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

	c.cfg.Logger.Info("SQS consumer starting", "queue_url", c.queueURL)

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
func (c *SQSConsumer) Stop(ctx context.Context) error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	c.mu.Unlock()

	c.cfg.Logger.Info("SQS consumer stopping")
	close(c.stopCh)

	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.cfg.Logger.Info("SQS consumer stopped gracefully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

// IsRunning checks if consumer is actively consuming
func (c *SQSConsumer) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// pollMessages polls messages from SQS
func (c *SQSConsumer) pollMessages(ctx context.Context) {
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
			messages, err := c.receiveMessages(ctx)
			if err != nil {
				if c.cfg.ErrorHandler != nil {
					c.cfg.ErrorHandler(err)
				}
				continue
			}
			for _, msg := range messages {
				select {
				case c.processingCh <- msg:
				case <-c.stopCh:
					return
				}
			}
		}
	}
}

// receiveMessages fetches messages from SQS
func (c *SQSConsumer) receiveMessages(ctx context.Context) ([]*Message, error) {
	result, err := c.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: int32(c.cfg.SQS.MaxMessages),
		WaitTimeSeconds:     int32(c.cfg.SQS.WaitTimeSecond),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	messages := make([]*Message, len(result.Messages))
	for i, msg := range result.Messages {
		messages[i] = &Message{
			ID:            *msg.MessageId,
			Body:          *msg.Body,
			ReceiptHandle: *msg.ReceiptHandle,
			Raw:           msg,
		}
	}
	return messages, nil
}

// processMessages processes messages from the channel
func (c *SQSConsumer) processMessages(ctx context.Context) {
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
				c.cfg.Logger.Error("SQS message handler error",
					"message_id", msg.ID,
					"error", err.Error(),
				)
				if c.cfg.ErrorHandler != nil {
					c.cfg.ErrorHandler(err)
				}
			} else {
				// Delete message from queue on success
				if err := c.deleteMessage(ctx, msg.ReceiptHandle); err != nil {
					c.cfg.Logger.Error("failed to delete message", "message_id", msg.ID, "error", err.Error())
				}
			}

			cancel()
		}
	}
}

// deleteMessage removes a message from the SQS queue
func (c *SQSConsumer) deleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := c.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}
