# Go Starter - Usage Guide

Complete documentation for running, developing, and deploying the Go Starter application.

## Quick Start

See [README.md](../README.md) for setup. After setup:

```bash
make run        # Start HTTP server on port 8080
```

## Feature Flags

By default, all features are **disabled**. Enable only what you need in `.env`:

| Flag | Default | Description |
|------|---------|-------------|
| `ENABLE_DATABASE` | `false` | Database and user API endpoints |
| `ENABLE_CACHE` | `false` | Caching layer (optional performance) |
| `ENABLE_HTTP_CLIENT` | `false` | External HTTP client with circuit breaker |
| `ENABLE_CONSUMER` | `false` | SQS/Kafka consumers |
| `ENABLE_WORKER` | `false` | Background worker mode |
| `ENABLE_SWAGGER` | `true` | Swagger UI documentation |

Example minimal `.env` for an API without database:
```bash
ENABLE_DATABASE=false
ENABLE_CACHE=false
SERVER_PORT=8080
```

## Make Commands

### Setup & Development

```bash
make help              # Show all available commands
make setup             # Setup dependencies (run after cloning)
make init-repo         # Initialize Git and create GitHub repo
make install-tools     # Install development tools
make run-worker        # Run as background worker (cron jobs)
make run-uat           # Run with UAT configuration
make run-staging       # Run with staging configuration
make dev               # Run with live reload (requires air)
```

### Code Quality

```bash
make fmt               # Format code
make lint              # Run linter
make test              # Run tests with coverage
make check-file-sizes  # Check for oversized files
```

### Building & Running

```bash
make build             # Build the application
make run               # Build and run
make clean             # Clean build artifacts
```

### Documentation

```bash
make swagger           # Generate Swagger documentation
```

### Code Generation

```bash
make generate-scaffold RESOURCE=product  # Generate new resource template
```

### Deployment

```bash
make push              # Push to GitHub
make docker-build      # Build Docker image
make docker-run        # Run Docker container
```

## Application Modes

### HTTP Server Mode (Default)

Runs the application as a RESTful API server on port 8080.

### Worker Mode

Runs the application as a background job processor with scheduled tasks.

```bash
export APP_MODE=worker
make run
```

**Running Tasks:**
- Cleanup old data (hourly)
- Sync external data (every 30 minutes)
- Health checks (every 5 minutes)
- Daily report generation
- Process notifications (every 5 minutes)

## Environment-Specific Configuration

### Development

- Text-based logging (human-readable)
- Debug level logs
- Swagger UI enabled
- SQLite database
- Hot reload with `make dev`

### UAT (User Acceptance Testing)

- JSON structured logging
- Info level logs
- Swagger UI enabled
- PostgreSQL database
- AWS Secrets Manager integration

### Staging

- JSON structured logging
- Warn level logs
- Swagger UI disabled
- RDS database
- AWS Secrets Manager enabled
- Production-like configuration

### Production

- JSON structured logging only
- Error level logs only
- Swagger UI disabled
- RDS Multi-AZ database
- AWS Secrets Manager enforced

## Features Explained

### 1. Two Run Modes

**HTTP Server Mode** (default):
```bash
make run
# Runs RESTful API on port 8080
```

**Worker Mode** (background jobs):
```bash
make run-worker
# Processes scheduled tasks and cron jobs
```

See [Running Guide](docs/RUNNING_GUIDE.md) for details.

### 2. JSON Logging (Production-Ready)

All logs are JSON-formatted in production/staging:

```json
{
  "level": "info",
  "msg": "user created",
  "user_id": "usr-123",
  "email": "user@example.com",
  "request_id": "req-abc123",
  "duration_ms": 45
}
```

Development uses human-readable text. Control via `LOG_LEVEL` env var.

### 3. HTTP Client with Circuit Breaker

```go
client := http.New(30*time.Second, logger)
client.AddCircuitBreaker(http.CircuitBreakerConfig{
    Name:      "payment-api",
    Threshold: 0.5,  // 50% failure rate
})

resp, err := client.GetWithCircuitBreaker(ctx, url, "payment-api")
```

### 4. Dual Caching System

TTL Cache (simple):
```go
cache.Set("key", data, 5*time.Minute)
value, exists := cache.Get("key")
```

In-Memory with Locks (critical sections):
```go
entry := cache.GetOrCreate("critical")
entry.Lock.Lock()
entry.Value = data
entry.Lock.Unlock()
```

### 5. AWS Integration

Secrets Manager auto-load:
```go
cfg, err := config.Load(ctx)
// Loads from: env vars → AWS Secrets Manager
```

### 6. Repository Pattern

Clean data access layer:
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
}

type userRepository struct {
    db *gorm.DB
}
```

### 7. Background Worker/Cron Jobs

```go
w.RegisterTask(Task{
    Name:     "cleanup",
    Interval: 1 * time.Hour,
    Fn:       cleanupTask,
})

w.Start(ctx)
```

### 8. Database Query Monitoring

Monitor slow queries and enforce connection limits:

```bash
# Configure in .env
DB_SLOW_QUERY_THRESHOLD=1s    # Log warnings for queries slower than 1 second
DB_QUERY_TIMEOUT=30s          # Cancel queries after 30 seconds
DB_MAX_OPEN_CONNS=25          # Maximum open connections
DB_MAX_IDLE_CONNS=5           # Maximum idle connections
DB_CONN_MAX_LIFETIME=1h       # Connection max lifetime
```

Slow query log output:
```json
{
  "level": "warn",
  "msg": "slow database query detected",
  "duration_ms": 1500,
  "sql": "SELECT * FROM users WHERE...",
  "rows_affected": 100
}
```

## Logger Recommendation

This template uses **slog** (Go 1.21+ stdlib) as the primary logger:

### Why slog?
- ✅ Zero external dependencies
- ✅ Built-in context support
- ✅ Structured JSON logging
- ✅ Full Go team backing
- ✅ Perfect for cloud-native apps

### Alternative: zerolog

If you need maximum performance (high-throughput):

```bash
go get github.com/rs/zerolog
```

See [JSON Logging Guide](docs/JSON_LOGGING_GUIDE.md) for details.

## Error Handling

Standard error response format:

```json
{
  "error": "USER_NOT_FOUND",
  "message": "User with ID 123 not found"
}
```

## Contributing

When adding new features:

1. Create models in `internal/models/`
2. Create repository in `internal/repository/`
3. Create service in `internal/business/`
4. Create handlers in `internal/router/`
5. Add routes in `internal/router/routes.go`
6. Add Swagger documentation tags
7. Run `make check-file-sizes` before commit
8. Run `make test` and `make lint` before PR

## Security Best Practices

1. **Never commit .env files** - Use `.env.example`
2. **Secrets in AWS Secrets Manager** - Not in code
3. **Context timeout on all operations** - Prevents hangs
4. **Circuit breaker on external APIs** - Prevents cascading failures
5. **Request ID logging** - For debugging and tracing
6. **Structured logging** - For security audits

## Performance Considerations

- **Caching**: Use TTL cache for frequently accessed data
- **Database**: Add indexes on frequently queried fields
- **Circuit breaker**: Configure thresholds based on SLA
- **File size limits**: Keep files under 500KB for readability

## Troubleshooting

### Port already in use
```bash
export SERVER_PORT=8081
```

### Database connection error
```bash
echo $DATABASE_URL
echo $DATABASE_DSN
psql postgresql://user:password@localhost:5432/dbname -c "SELECT 1"
```

### Swagger not showing
```bash
make swagger
# Visit http://localhost:8080/swagger/index.html
```

### Circuit breaker always open
```go
status := httpClient.CircuitBreakerStatus("endpoint-name")
fmt.Printf("Requests: %d, Failures: %d", status.Requests, status.TotalFailures)
```

## Documentation

- [Running Guide](docs/RUNNING_GUIDE.md) - How to run in different modes and environments
- [Architecture Guide](docs/ARCHITECTURE.md) - Detailed architecture and patterns
- [Cache Usage Guide](docs/CACHE_USAGE.md) - How to use caching correctly
- [AWS Integration Guide](docs/AWS_INTEGRATION.md) - AWS setup and configuration
- [JSON Logging Guide](docs/JSON_LOGGING_GUIDE.md) - Detailed logging configuration
- [Enhancement Suggestions](docs/SUGGESTIONS.md) - Future improvements and roadmap