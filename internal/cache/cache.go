package cache

import (
	"sync"
	"time"
)

// Item represents a cached item with expiration
type Item struct {
	Value      interface{}
	Expiration time.Time
}

// TTLCache is a thread-safe in-memory cache with TTL support
type TTLCache struct {
	mu    sync.RWMutex
	items map[string]*Item
}

// NewTTLCache creates a new TTL cache
func NewTTLCache() *TTLCache {
	cache := &TTLCache{
		items: make(map[string]*Item),
	}

	// Cleanup goroutine to remove expired items
	go cache.cleanupExpired()

	return cache
}

// Set stores a value with TTL
func (c *TTLCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Time{}
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.items[key] = &Item{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves a value from cache
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if !item.Expiration.IsZero() && time.Now().After(item.Expiration) {
		// Don't delete here (to avoid upgrade from RLock to Lock)
		return nil, false
	}

	return item.Value, true
}

// Delete removes a value from cache
func (c *TTLCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from cache
func (c *TTLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*Item)
}

// Size returns the number of items in cache
func (c *TTLCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// cleanupExpired periodically removes expired items
func (c *TTLCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if !item.Expiration.IsZero() && now.After(item.Expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// CacheWithLock provides a cache entry protected by a lock
type CacheWithLock struct {
	Value interface{}
	Lock  sync.RWMutex
}

// InMemoryCache is a thread-safe in-memory cache with manual lock management
// Use this for critical sections where you need fine-grained lock control
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]*CacheWithLock
}

// NewInMemoryCache creates a new in-memory cache with lock support
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		items: make(map[string]*CacheWithLock),
	}
}

// GetOrCreate gets an existing cache entry or creates a new one
// Returns a CacheWithLock that you can lock for safe access
func (c *InMemoryCache) GetOrCreate(key string) *CacheWithLock {
	c.mu.RLock()
	if item, exists := c.items[key]; exists {
		c.mu.RUnlock()
		return item
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check again after acquiring write lock
	if item, exists := c.items[key]; exists {
		return item
	}

	item := &CacheWithLock{Value: nil}
	c.items[key] = item
	return item
}

// Get retrieves a cache entry
func (c *InMemoryCache) Get(key string) (*CacheWithLock, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	return item, exists
}

// Set stores a cache entry
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		item.Lock.Lock()
		item.Value = value
		item.Lock.Unlock()
	} else {
		c.items[key] = &CacheWithLock{Value: value}
	}
}

// Delete removes a cache entry
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all entries
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheWithLock)
}

// Size returns the number of entries
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
