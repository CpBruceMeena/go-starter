# Architecture Guide

## Overview

This Go starter template follows clean architecture principles with clear separation of concerns. The architecture is designed to be:

- **Scalable** - Easy to add new features
- **Maintainable** - Clear organization and dependencies
- **Testable** - Dependencies are injected, making unit tests simple
- **Flexible** - Easy to swap implementations (database, caching, logging)

## Architecture Layers

```
┌─────────────────────────────────────────────┐
│         HTTP Router / Handlers              │
│          (internal/router/)                 │
└───────────────┬─────────────────────────────┘
                │
┌───────────────▼─────────────────────────────┐
│         Business Logic Services             │
│          (internal/business/)               │
│   - Validation                              │
│   - Core business rules                     │
│   - Caching orchestration                   │
└───────────────┬─────────────────────────────┘
                │
        ┌───────┴────────┬──────────┬───────────┐
        │                │          │           │
┌───────▼──────┐  ┌──────▼─┐  ┌────▼──┐  ┌────▼──┐
│  Repository  │  │ Cache  │  │Logger │  │ HTTP  │
│ (Data Access)│  │(TTL/IM)│  │(slog) │  │Client │
│(internal/    │  │        │  │       │  │   +   │
│repository/)  │  │        │  │       │  │ CB    │
└───────┬──────┘  └────────┘  └───────┘  └────┬──┘
        │                                      │
┌───────▼────────────────────────────────────┬▼┐
│            Database                         │ │
│          (PostgreSQL/SQLite)         External│ │
│              (GORM)                   APIs   │ │
└────────────────────────────────────────────┴─┘
```

## Layer Descriptions

### 1. **Router/Handler Layer** (`internal/router/`)

**Responsibility**: HTTP request handling and response formatting

```
router/
├── routes.go      # Route definitions and Swagger docs
└── handlers.go    # Handler functions for each endpoint
```

**Characteristics**:
- Minimal business logic
- Converts HTTP requests to service calls
- Formats responses consistently
- HTTP status codes and error handling

**Example**:
```go
func CreateUser(c echo.Context, svc business.UserService) error {
    var req models.CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{...})
    }
    
    ctx := c.Request().Context()
    user, err := svc.CreateUser(ctx, &req)
    if err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{...})
    }
    
    return c.JSON(http.StatusCreated, user)
}
```

### 2. **Business Logic Layer** (`internal/business/`)

**Responsibility**: Core application logic and business rules

```
business/
└── user.go        # User service with all business logic
```

**Characteristics**:
- No HTTP knowledge
- No database queries directly (uses repository)
- Validates business rules
- Orchestrates caching
- Uses context for request tracking

**Example**:
```go
func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
    // Validation
    existing, err := s.repo.GetByEmail(ctx, req.Email)
    if existing != nil {
        return nil, fmt.Errorf("user already exists")
    }
    
    // Create
    user := &models.User{Email: req.Email, Name: req.Name}
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Invalidate cache
    s.cache.Delete("users:list")
    
    return user.ToResponse(), nil
}
```

### 3. **Data Access Layer** (`internal/repository/`)

**Responsibility**: Database operations

```
repository/
└── user.go        # User repository interface and implementation
```

**Characteristics**:
- Interface-based design (for testability)
- Only database queries
- No business logic
- Context-aware operations
- Always using GORM

**Example**:
```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
    List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

type userRepository struct {
    db *gorm.DB
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}
```

### 4. **Model Layer** (`internal/models/`)

**Responsibility**: Data structure definitions

```
models/
└── user.go        # User struct and related models
```

**IMPORTANT**: Models contain ONLY struct definitions. No methods with business logic.

**Allowed in models**:
```go
type User struct {
    ID    string
    Email string
    Name  string
}

// Helper: Convert to response (simple transformation)
func (u *User) ToResponse() *UserResponse {
    return &UserResponse{...}
}

// Request/Response models
type CreateUserRequest struct {
    Email string `json:"email" binding:"required,email"`
    Name  string `json:"name" binding:"required"`
}
```

**NOT allowed in models**:
```go
// ❌ Business logic methods
func (u *User) IsValidEmail() bool { ... }
func (u *User) GenerateToken() string { ... }
func (u *User) ValidatePassword(pwd string) bool { ... }

// ❌ Database operations
func (u *User) Save(db *gorm.DB) error { ... }
func (u *User) Delete(db *gorm.DB) error { ... }
```

## Cross-Cutting Concerns

### Logging (`internal/logger/`)

- Uses `slog` (stdlib log/slog)
- Context-aware (request ID propagation)
- JSON output for structured logging
- Request ID automatically added to all logs

```go
logger.InfoContext(ctx, "user created", 
    "user_id", user.ID,
    "email", user.Email,
)
```

### Caching (`internal/cache/`)

Two cache strategies available:

**TTL Cache**: Simple time-based expiration
```go
cache := cache.NewTTLCache()
cache.Set("user:123", user, 5*time.Minute)
user, exists := cache.Get("user:123")
```

**In-Memory Cache with Locks**: For critical sections
```go
cache := cache.NewInMemoryCache()
entry := cache.GetOrCreate("critical-data")
entry.Lock.Lock()
entry.Value = data
entry.Lock.Unlock()
```

### HTTP Client (`internal/http/`)

- Built-in circuit breaker pattern
- Logging for all requests
- Request/response tracking
- Configurable per endpoint

```go
client.AddCircuitBreaker(http.CircuitBreakerConfig{
    Name:      "payment-api",
    Threshold: 0.5,
    Timeout:   30 * time.Second,
})

resp, err := client.GetWithCircuitBreaker(ctx, url, "payment-api")
```

### Configuration (`internal/config/`)

- Environment variables (local)
- AWS Secrets Manager (production)
- Automatic merge (Secrets override Env)

```go
cfg, err := config.Load(ctx)
// Loads from: .env file + AWS Secrets Manager
```

### Middleware (`internal/middleware/`)

- Request ID generation
- Request/response logging
- CORS (can be added)
- Authentication (can be added)

## Dependency Injection Pattern

Dependencies flow downward:

```
main
  ↓
server (knows about: config, logger)
  ↓
services (know about: repos, cache, logger)
  ↓
repositories (know about: database)
```

**Example**:
```go
// main.go
func main() {
    cfg := config.Load(ctx)
    log := logger.New()
    db := database.InitDB(cfg.DatabaseDSN)
    
    repo := repository.NewUserRepository(db)
    cache := cache.NewTTLCache()
    service := business.NewUserService(repo, cache, log)
    
    server := server.New(cfg, log)
    server.Setup(service)
    server.Start()
}
```

## Design Patterns Used

### 1. **Repository Pattern**
- Abstraction over data access
- Easy to mock for testing
- Can swap database implementations

### 2. **Service Pattern**
- Encapsulate business logic
- Reusable across different handlers
- Easy to test in isolation

### 3. **Dependency Injection**
- All dependencies injected at creation
- No global state
- Easy to test

### 4. **Circuit Breaker Pattern**
- Prevents cascading failures
- Automatic recovery attempts
- Request level visibility

### 5. **Factory Pattern**
- `NewUserRepository()`, `NewUserService()`, etc.
- Consistent object creation
- Easy to modify initialization logic

## Error Handling

**Consistent error responses**:

```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable message"
}
```

**Error flow**:
```
Repository Error
  ↓
Service catches and logs
  ↓
Handler formats as HTTP response
```

## Testing Strategy

Each layer can be tested independently:

```go
// Unit test for service
func TestCreateUser(t *testing.T) {
    mockRepo := &MockUserRepository{}
    mockCache := cache.NewTTLCache()
    service := business.NewUserService(mockRepo, mockCache, logger)
    
    user, err := service.CreateUser(ctx, &req)
    assert.NoError(t, err)
}

// Unit test for handler
func TestCreateUserHandler(t *testing.T) {
    mockService := &MockUserService{}
    e := echo.New()
    router.SetupRoutes(e, mockService, logger)
    
    req := httptest.NewRequest("POST", "/api/v1/users", body)
    rec := httptest.NewRecorder()
    e.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

## Adding New Features

Follow this checklist:

1. **Define Model** (`internal/models/your-feature.go`)
   - Add struct(s)
   - Add request/response models
   - Add ToResponse() helper

2. **Create Repository** (`internal/repository/your-feature.go`)
   - Define interface
   - Implement CRUD operations
   - Use context for all operations

3. **Create Service** (`internal/business/your-feature.go`)
   - Inject repository and cache
   - Implement business logic
   - Handle validation
   - Orchestrate caching

4. **Create Handlers** (`internal/router/handlers.go`)
   - Add handler function
   - Add route in `routes.go`
   - Add Swagger documentation

5. **Test**
   - Write unit tests for service
   - Write unit tests for handlers
   - Run `make test`

6. **Quality Checks**
   - Run `make fmt`
   - Run `make lint`
   - Run `make check-file-sizes`
   - Check file is not over 500KB

## Performance Considerations

### Caching Strategy
- **Frequently accessed data**: TTL cache with 5min TTL
- **Critical sections**: In-memory cache with locks
- **Invalidate on writes**: Delete related cache entries

### Database
- **Query optimization**: Add indexes on frequently queried fields
- **N+1 prevention**: Use eager loading with GORM
- **Connection pooling**: Automatic with GORM

### HTTP Clients
- **Circuit breaker**: Prevents cascading failures
- **Timeouts**: Configured per endpoint
- **Logging**: All requests logged for debugging

### File Size Limits
- Keep files under 500KB
- Split large files into smaller modules
- Use `make check-file-sizes` to monitor

## Security Considerations

1. **Never log sensitive data** - No passwords, tokens in logs
2. **Context timeouts** - Prevent resource exhaustion
3. **Input validation** - Validate all user input
4. **Circuit breaker** - Prevent DoS on external services
5. **Request ID tracking** - For audit logs
6. **Environment secrets** - Use AWS Secrets Manager in production

---

For more details on specific components, see the relevant guides:
- [Cache Usage](CACHE_USAGE.md)
- [AWS Integration](AWS_INTEGRATION.md)
