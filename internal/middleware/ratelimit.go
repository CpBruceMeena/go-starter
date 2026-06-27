package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/CpBruceMeena/go-starter/internal/response"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerSecond int
	Burst             int
	CleanupInterval   time.Duration
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     time.Duration
	burst    int
}

// Visitor tracks request counts per IP
type Visitor struct {
	lastSeen time.Time
	count    int
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     time.Second / time.Duration(cfg.RequestsPerSecond),
		burst:    cfg.Burst,
	}

	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}

	go rl.cleanupVisitors(cfg.CleanupInterval)
	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &Visitor{count: 1}
		return true
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.lastSeen = time.Now()

	if v.count > rl.burst {
		return false
	}

	v.count++
	return true
}

// cleanupVisitors removes old visitors
func (rl *RateLimiter) cleanupVisitors(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			v.mu.Lock()
			if time.Since(v.lastSeen) > 3*interval {
				delete(rl.visitors, ip)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware returns middleware that rate limits requests
func RateLimitMiddleware(rl *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip, _, _ := net.SplitHostPort(c.RealIP())
			if ip == "" {
				ip = c.RealIP()
			}

			if !rl.Allow(ip) {
				return c.JSON(http.StatusTooManyRequests, response.Error("RATE_LIMITED", "Too many requests"))
			}

			return next(c)
		}
	}
}