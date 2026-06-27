# Cache Usage Guide

This template provides two types of caching mechanisms for different use cases. Understanding when and how to use each is critical for optimal performance.

## Overview

```
┌─────────────────────────────────────────────┐
│  TTL Cache                                  │
│  - Simple time-based expiration             │
│  - Automatic cleanup                        │
│  - Use for: API responses, DB results       │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│  In-Memory Cache with Locks                 │
│  - Manual lock management                   │
│  - Simultaneous read/write safety           │
│  - Use for: Critical state, computations    │
└─────────────────────────────────────────────┘
```

## 1. TTL Cache

**Best for**: Simple caching with automatic expiration

### Characteristics
- Automatic expiration after TTL
- Background cleanup every minute
- Thread-safe (RWMutex internally)
- Simple API

### When to Use
- API response caching
- Database query results
- Expensive computations with known TTL
- Session data
- User preferences

### API

```go
// Create cache
cache := cache.NewTTLCache()

// Set with 5-minute TTL
cache.Set("user:123", userData, 5*time.Minute)

// Set with no expiration (TTL=0)
cache.Set("permanent:key", data, 0)

// Get value
value, exists := cache.Get("user:123")
if exists {
    user := value.(*User)
    // Use user
}

// Delete manually
cache.Delete("user:123")

// Clear all
cache.Clear()

// Get size
count := cache.Size()
```

### Example: Service Implementation

```go
type userService struct {
    repo  repository.UserRepository
    cache *cache.TTLCache
    log   *logger.Logger
}

func (s *userService) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("user:%s", id)
    if cached, exists := s.cache.Get(cacheKey); exists {
        s.log.DebugContext(ctx, "cache hit", "key", cacheKey)
        return cached.(*User), nil
    }
    
    s.log.DebugContext(ctx, "cache miss", "key", cacheKey)
    
    // Get from database
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    s.cache.Set(cacheKey, user, 5*time.Minute)
    
    return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *User) error {
    // Update in database
    if err := s.repo.Update(ctx, user); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%s", user.ID)
    s.cache.Delete(cacheKey)
    
    // Also invalidate list cache
    s.cache.Delete("users:list")
    
    return nil
}
```

### Caching Strategy for APIs

```go
// List endpoint caching
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
    cacheKey := fmt.Sprintf("users:list:%d:%d", limit, offset)
    
    if cached, exists := s.cache.Get(cacheKey); exists {
        return cached.([]*User), nil
    }
    
    users, err := s.repo.List(ctx, limit, offset)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    s.cache.Set(cacheKey, users, 5*time.Minute)
    return users, nil
}

// Invalidate on write
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    user := &User{...}
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Invalidate all list cache entries
    // Option 1: Delete specific patterns (manual)
    s.cache.Delete("users:list:10:0")
    s.cache.Delete("users:list:20:0")
    
    // Option 2: Use prefix-based invalidation (implement helper)
    s.invalidateCachePattern("users:list:")
    
    return user, nil
}

// Helper function to invalidate by pattern
func (s *userService) invalidateCachePattern(pattern string) {
    // Note: TTLCache doesn't have built-in pattern deletion
    // You can maintain a list of cache keys or implement custom logic
}
```

## 2. In-Memory Cache with Locks

**Best for**: Critical state that needs simultaneous read/write protection

### Characteristics
- Manual lock management (RWMutex)
- No automatic expiration
- Perfect for mutable state
- Fine-grained lock control

### When to Use
- Application state/settings
- Configuration that changes at runtime
- Counters and statistics
- Complex objects being modified
- Shared mutable state

### API

```go
// Create cache
cache := cache.NewInMemoryCache()

// Get or create entry
entry := cache.GetOrCreate("app-config")

// Lock for read
entry.Lock.RLock()
config := entry.Value.(AppConfig)
entry.Lock.RUnlock()

// Lock for write
entry.Lock.Lock()
entry.Value = newConfig
entry.Lock.Unlock()

// Lock for read-modify-write
entry := cache.GetOrCreate("counter")
entry.Lock.Lock()
count := entry.Value.(int)
count++
entry.Value = count
entry.Lock.Unlock()

// Get without creating
entry, exists := cache.Get("counter")

// Manual set
cache.Set("key", value)

// Delete
cache.Delete("key")

// Clear
cache.Clear()

// Size
count := cache.Size()
```

### Example: Mutable State

```go
// Application state cache
type StateCache struct {
    cache *cache.InMemoryCache
    log   *logger.Logger
}

func NewStateCache(log *logger.Logger) *StateCache {
    return &StateCache{
        cache: cache.NewInMemoryCache(),
        log:   log,
    }
}

// GetConfig retrieves application config
func (sc *StateCache) GetConfig() AppConfig {
    entry := sc.cache.GetOrCreate("app-config")
    entry.Lock.RLock()
    defer entry.Lock.RUnlock()
    
    if entry.Value == nil {
        return AppConfig{} // default
    }
    return entry.Value.(AppConfig)
}

// UpdateConfig updates application config
func (sc *StateCache) UpdateConfig(cfg AppConfig) {
    entry := sc.cache.GetOrCreate("app-config")
    entry.Lock.Lock()
    defer entry.Lock.Unlock()
    
    entry.Value = cfg
    sc.log.Info("config updated")
}

// IncrementCounter increments a counter
func (sc *StateCache) IncrementCounter(key string) int {
    entry := sc.cache.GetOrCreate("counter:" + key)
    entry.Lock.Lock()
    defer entry.Lock.Unlock()
    
    count := 0
    if entry.Value != nil {
        count = entry.Value.(int)
    }
    count++
    entry.Value = count
    
    return count
}
```

### Example: Statistics Collection

```go
type Stats struct {
    TotalRequests  int64
    FailedRequests int64
    AverageLatency float64
}

type StatsCollector struct {
    cache *cache.InMemoryCache
}

// RecordRequest records a request
func (sc *StatsCollector) RecordRequest(latencyMs int64, success bool) {
    entry := sc.cache.GetOrCreate("stats")
    entry.Lock.Lock()
    defer entry.Lock.Unlock()
    
    var stats Stats
    if entry.Value != nil {
        stats = entry.Value.(Stats)
    }
    
    stats.TotalRequests++
    if !success {
        stats.FailedRequests++
    }
    
    // Update average
    total := stats.AverageLatency * float64(stats.TotalRequests-1)
    stats.AverageLatency = (total + float64(latencyMs)) / float64(stats.TotalRequests)
    
    entry.Value = stats
}

// GetStats returns current stats
func (sc *StatsCollector) GetStats() Stats {
    entry := sc.cache.GetOrCreate("stats")
    entry.Lock.RLock()
    defer entry.Lock.RUnlock()
    
    if entry.Value == nil {
        return Stats{}
    }
    return entry.Value.(Stats)
}
```

## Comparison Table

| Feature | TTL Cache | In-Memory Cache |
|---------|-----------|-----------------|
| **Expiration** | Automatic | None |
| **Use Case** | Data retrieval | Mutable state |
| **Lock Type** | RWMutex (internal) | Manual (user) |
| **Memory** | Cleanup every 1min | Manual cleanup |
| **Complexity** | Simple | Medium |
| **Best For** | API responses | Application state |

## Caching Patterns

### Pattern 1: Cache-Aside (Lazy Loading)

```go
func (s *Service) GetData(ctx context.Context, id string) (Data, error) {
    // Try cache
    key := fmt.Sprintf("data:%s", id)
    if cached, exists := s.cache.Get(key); exists {
        return cached.(Data), nil
    }
    
    // Load from source
    data, err := s.repo.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache.Set(key, data, 5*time.Minute)
    return data, nil
}
```

### Pattern 2: Write-Through Cache

```go
func (s *Service) UpdateData(ctx context.Context, data Data) error {
    // Update source
    if err := s.repo.Update(ctx, data); err != nil {
        return err
    }
    
    // Update cache
    key := fmt.Sprintf("data:%s", data.ID)
    s.cache.Set(key, data, 5*time.Minute)
    
    return nil
}
```

### Pattern 3: Cache Invalidation

```go
func (s *Service) DeleteData(ctx context.Context, id string) error {
    // Delete from source
    if err := s.repo.Delete(ctx, id); err != nil {
        return err
    }
    
    // Invalidate cache
    key := fmt.Sprintf("data:%s", id)
    s.cache.Delete(key)
    
    // Invalidate related caches
    s.cache.Delete("data:list")
    
    return nil
}
```

## Best Practices

### ✅ DO

1. **Use TTL cache for external data**
   ```go
   // Good: API responses cached with TTL
   s.cache.Set(key, apiResponse, 5*time.Minute)
   ```

2. **Use In-Memory cache for internal state**
   ```go
   // Good: Application config with locks
   entry.Lock.Lock()
   entry.Value = newConfig
   entry.Lock.Unlock()
   ```

3. **Invalidate on write**
   ```go
   // Good: Clear cache when data changes
   s.cache.Delete(cacheKey)
   ```

4. **Use context-aware keys**
   ```go
   // Good: Include tenant ID in key
   key := fmt.Sprintf("tenant:%s:users:list", tenantID)
   ```

5. **Log cache hits/misses in debug**
   ```go
   s.log.DebugContext(ctx, "cache operation", "key", key, "hit", exists)
   ```

### ❌ DON'T

1. **Cache user-sensitive data without TTL**
   ```go
   // Bad: No TTL means user sees stale auth data
   s.cache.Set("user:auth", data, 0)
   
   // Good: Always set appropriate TTL
   s.cache.Set("user:auth", data, 5*time.Minute)
   ```

2. **Forget to invalidate related caches**
   ```go
   // Bad: Cache gets stale
   if err := s.repo.Update(ctx, user); err != nil {
       return err
   }
   // ❌ Forgot to delete cache!
   
   // Good: Invalidate related entries
   s.cache.Delete(fmt.Sprintf("user:%s", user.ID))
   s.cache.Delete("users:list")
   ```

3. **Use In-Memory cache for high-volume data**
   ```go
   // Bad: Memory exhaustion
   cache := cache.NewInMemoryCache()
   for i := 0; i < 1000000; i++ {
       cache.Set(fmt.Sprintf("item:%d", i), largeData)
   }
   
   // Good: Use TTL cache with cleanup
   cache := cache.NewTTLCache()
   cache.Set(key, data, 5*time.Minute)
   ```

4. **Hold locks for too long**
   ```go
   // Bad: Locks held during I/O
   entry.Lock.Lock()
   data.Field = s.expensiveOperation()  // Long operation
   entry.Lock.Unlock()
   
   // Good: Keep critical section small
   result := s.expensiveOperation()
   entry.Lock.Lock()
   entry.Value = result
   entry.Lock.Unlock()
   ```

## Monitoring Cache Health

```go
// Cache stats
func (s *userService) CacheStats(ctx context.Context) {
    size := s.cache.Size()
    s.log.InfoContext(ctx, "cache stats", "entries", size)
    
    // Monitor circuit breaker for HTTP calls
    cbStatus := s.httpClient.CircuitBreakerStatus("api-name")
    s.log.InfoContext(ctx, "circuit breaker status",
        "requests", cbStatus.Requests,
        "failures", cbStatus.TotalFailures,
    )
}
```

## Testing Cache Logic

```go
func TestCacheHitAndMiss(t *testing.T) {
    cache := cache.NewTTLCache()
    
    // Initial miss
    _, exists := cache.Get("key1")
    assert.False(t, exists)
    
    // Set and get
    cache.Set("key1", "value1", 5*time.Minute)
    value, exists := cache.Get("key1")
    assert.True(t, exists)
    assert.Equal(t, "value1", value)
    
    // Delete
    cache.Delete("key1")
    _, exists = cache.Get("key1")
    assert.False(t, exists)
}

func TestInMemoryCacheWithLocks(t *testing.T) {
    cache := cache.NewInMemoryCache()
    
    // Get or create
    entry := cache.GetOrCreate("counter")
    entry.Lock.Lock()
    entry.Value = 0
    entry.Lock.Unlock()
    
    // Concurrent increment
    wg := sync.WaitGroup{}
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            entry := cache.GetOrCreate("counter")
            entry.Lock.Lock()
            count := entry.Value.(int)
            entry.Value = count + 1
            entry.Lock.Unlock()
        }()
    }
    wg.Wait()
    
    // Verify
    entry.Lock.RLock()
    assert.Equal(t, 100, entry.Value.(int))
    entry.Lock.RUnlock()
}
```

---

For more information, see [Architecture Guide](ARCHITECTURE.md) and the main [README](../README.md).
