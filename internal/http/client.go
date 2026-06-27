package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
	"github.com/CpBruceMeena/go-starter/internal/logger"
)

// Client wraps http.Client with circuit breaker support
type Client struct {
	client     *http.Client
	circuitMap map[string]*gobreaker.CircuitBreaker
	logger     *logger.Logger
	timeout    time.Duration
	enabled    bool // Whether circuit breaker is enabled
}

// CircuitBreakerConfig holds circuit breaker settings
type CircuitBreakerConfig struct {
	Name        string
	MaxRequests uint32        // consecutive successful requests before closing
	Interval    time.Duration // time to reset counter
	Timeout     time.Duration // time to transition from open to half-open
	Threshold   float64       // failure rate threshold (0.0 to 1.0)
}

// New creates a new HTTP client with circuit breaker support
func New(timeout time.Duration, log *logger.Logger) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
		},
		circuitMap: make(map[string]*gobreaker.CircuitBreaker),
		logger:     log,
		timeout:    timeout,
		enabled:    true,
	}
}

// DisableCircuitBreaker disables circuit breaker functionality
func (c *Client) DisableCircuitBreaker() {
	c.enabled = false
	c.circuitMap = make(map[string]*gobreaker.CircuitBreaker)
}

// AddCircuitBreaker registers a circuit breaker for a specific endpoint
func (c *Client) AddCircuitBreaker(cfg CircuitBreakerConfig) {
	if !c.enabled {
		return
	}

	if cfg.MaxRequests == 0 {
		cfg.MaxRequests = 1
	}
	if cfg.Interval == 0 {
		cfg.Interval = time.Second * 60
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = time.Second * 30
	}
	if cfg.Threshold == 0 {
		cfg.Threshold = 0.5 // 50% failure rate
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= cfg.Threshold
		},
	})

	c.circuitMap[cfg.Name] = cb
	c.logger.Info("circuit breaker registered",
		"name", cfg.Name,
		"max_requests", cfg.MaxRequests,
		"interval", cfg.Interval.String(),
		"timeout", cfg.Timeout.String(),
		"threshold", fmt.Sprintf("%.0f%%", cfg.Threshold*100),
	)
}

// AddCircuitBreakerFromConfigString adds circuit breaker from config with string durations
func (c *Client) AddCircuitBreakerFromConfigString(name string, maxRequests uint32, intervalStr, timeoutStr string, thresholdPercent uint32) error {
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return fmt.Errorf("invalid interval duration: %w", err)
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return fmt.Errorf("invalid timeout duration: %w", err)
	}

	// Convert percentage (0-100) to ratio (0.0-1.0)
	threshold := float64(thresholdPercent) / 100.0
	if threshold < 0 || threshold > 1 {
		threshold = 0.5
	}

	c.AddCircuitBreaker(CircuitBreakerConfig{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
		Threshold:   threshold,
	})

	return nil
}

// Do executes an HTTP request with circuit breaker protection
func (c *Client) Do(ctx context.Context, req *http.Request, circuitName string) (*http.Response, error) {
	c.logger.InfoContext(ctx, "HTTP request",
		"method", req.Method,
		"url", req.URL.String(),
		"circuit", circuitName,
	)

	// Get circuit breaker if registered
	var cb *gobreaker.CircuitBreaker
	if circuitName != "" {
		cb = c.circuitMap[circuitName]
	}

	// Execute with or without circuit breaker
	var resp *http.Response
	var err error

	if cb != nil {
		resp, err = c.executeWithCircuitBreaker(ctx, req, cb)
	} else {
		resp, err = c.client.Do(req.WithContext(ctx))
	}

	if err != nil {
		c.logger.ErrorContext(ctx, "HTTP request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"circuit", circuitName,
			"error", err.Error(),
		)
		return nil, err
	}

	c.logger.InfoContext(ctx, "HTTP request succeeded",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"circuit", circuitName,
	)

	return resp, nil
}

// executeWithCircuitBreaker executes request through circuit breaker
func (c *Client) executeWithCircuitBreaker(ctx context.Context, req *http.Request, cb *gobreaker.CircuitBreaker) (*http.Response, error) {
	result, err := cb.Execute(func() (interface{}, error) {
		return c.client.Do(req.WithContext(ctx))
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			c.logger.WarnContext(ctx, "circuit breaker open",
				"url", req.URL.String(),
				"circuit", cb.Name(),
			)
		}
		return nil, err
	}

	return result.(*http.Response), nil
}

// GetWithCircuitBreaker performs a GET request with circuit breaker
func (c *Client) GetWithCircuitBreaker(ctx context.Context, url string, circuitName string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req, circuitName)
}

// PostWithCircuitBreaker performs a POST request with circuit breaker
func (c *Client) PostWithCircuitBreaker(ctx context.Context, url string, body io.Reader, circuitName string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return c.Do(ctx, req, circuitName)
}

// CircuitBreakerStatus returns the current state of a circuit breaker
func (c *Client) CircuitBreakerStatus(name string) gobreaker.Counts {
	if cb, exists := c.circuitMap[name]; exists {
		return cb.Counts()
	}
	return gobreaker.Counts{}
}
