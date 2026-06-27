# Go Starter - Production-Ready Template

A comprehensive Go application starter template with best practices, modern patterns, and production-ready features built-in. Includes database query monitoring, slow query detection, and connection pool management.

## 🎯 Features

- ✅ **Latest Go Version** (1.23+) - Using modern Go features
- ✅ **Swagger/OpenAPI** - API documentation out of the box
- ✅ **Structured Logging** - Using slog (stdlib) with context support
- ✅ **HTTP Client with Circuit Breaker** - Built-in resilience patterns
- ✅ **Dual Caching System** - TTL cache and in-memory cache with locks
- ✅ **Database Query Monitoring** - Slow query detection, connection pool limits, configurable timeouts
- ✅ **AWS Integration Ready** - Secrets Manager, task definitions, local env override
- ✅ **Repository Pattern** - Clean data access layer
- ✅ **Service Layer** - Business logic separation
- ✅ **File Size Guardrails** - Automatic warnings for large files
- ✅ **Makefile Automation** - Setup, build, test, deploy commands
- ✅ **Docker Ready** - Dockerfile included
- ✅ **Database Migrations** - GORM auto-migration support
- ✅ **Middleware Support** - Request ID, logging, CORS ready
- ✅ **Error Handling** - Standardized error responses
- ✅ **Configuration Management** - Environment variables + AWS Secrets Manager

## 🚀 Quick Start (5 minutes)

### Local Development - HTTP Server

```bash
make setup              # Install dependencies
make run                # Start HTTP server on port 8080
```

Visit http://localhost:8080/swagger/index.html

### Local Development - Worker/Cron Jobs

```bash
make setup
make run-worker         # Start background job processor
```

### Different Environments

```bash
make run                # HTTP server (development)
make run-uat            # HTTP server (UAT config)
make run-staging        # HTTP server (staging config)
make run-worker         # Worker mode (background jobs)
```

See [Running Guide](docs/RUNNING_GUIDE.md) for detailed instructions.

## 📋 Project Structure

```
go-starter/
├── cmd/                          # Application entry points
│   └── app/
│       └── main.go              # Main application file
├── internal/                     # Private application code
│   ├── models/                  # Data models (structs only)
│   │   └── user.go
│   ├── business/                # Business logic (services)
│   │   └── user.go
│   ├── repository/              # Data access layer
│   │   └── user.go
│   ├── router/                  # HTTP routes and handlers
│   │   ├── routes.go
│   │   └── handlers.go
│   ├── server/                  # Server initialization
│   │   └── server.go
│   ├── database/                # Database configuration
│   │   └── db.go
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── logger/                  # Logging setup
│   │   └── logger.go
│   ├── http/                    # HTTP client with circuit breaker
│   │   └── client.go
│   ├── cache/                   # Caching implementation
│   │   └── cache.go
│   ├── middleware/              # HTTP middleware
│   │   └── logging.go
│   ├── aws/                     # AWS integrations
│   │   ├── secrets.go
│   │   └── task_definition.go
│   └── external/                # External service integrations
├── pkg/                         # Public packages
│   ├── errors/                  # Error types
│   ├── utils/                   # Utility functions
│   └── health/                  # Health check utilities
├── migrations/                  # Database migration files
├── docs/                        # Documentation
│   └── ARCHITECTURE.md
├── scripts/                     # Utility scripts
├── config/                      # Configuration files (example configs)
├── go.mod                       # Module definition
├── Makefile                     # Build automation
├── README.md                    # This file
├── .gitignore                   # Git ignore rules
├── .env.example                 # Environment variables template
└── docker-compose.yml           # Docker compose for dependencies

```

## 📖 Documentation

- [Running Guide](docs/RUNNING_GUIDE.md) - How to run in different modes and environments
- [Architecture Guide](docs/ARCHITECTURE.md) - Detailed architecture and patterns
- [Cache Usage Guide](docs/CACHE_USAGE.md) - How to use caching correctly
- [AWS Integration Guide](docs/AWS_INTEGRATION.md) - AWS setup and configuration
- [Enhancement Suggestions](docs/SUGGESTIONS.md) - Future improvements and roadmap

## 🛠️ Make Commands

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
make docs              # Generate API docs
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

## 🗂️ Key Patterns & Features
🎯 Features Explained

### 1. **Two Run Modes**

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

### 2. **JSON Logging (Production-Ready)**

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

### 3. **HTTP Client with Circuit Breaker**

```go
client := http.New(30*time.Second, logger)
client.AddCircuitBreaker(http.CircuitBreakerConfig{
    Name:      "payment-api",
    Threshold: 0.5,  // 50% failure rate
})

resp, err := client.GetWithCircuitBreaker(ctx, url, "payment-api")
```

### 4. **Dual Caching System**

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

### 5. **AWS Integration**

Secrets Manager auto-load:
```go
cfg, err := config.Load(ctx)
// Loads from: env vars → AWS Secrets Manager
```

### 6. **Repository Pattern**

Clean data access layer:
```go
// Interface
type UserRepository interface {
    Create(ctx, user) error
}

// Implementation uses GORM
type userRepository struct {
    db *gorm.DB
}
```

### 7. **Background Worker/Cron Jobs**

```go
// Define tasks
w.RegisterTask(Task{
    Name:     "cleanup",
    Interval: 1 * time.Hour,
    Fn:       cleanupTask,
})

// Run
w.Start(ctx)
```

Examples included: cleanup, sync, health checks, reports, notifications.

### 8. **Database Query Monitoring**

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
## 📊 Logger Recommendation

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
# In go.mod: Add zerolog
go get github.com/rs/zerolog

# See internal/logger/logger.go for optional zerolog integration
```

## 🐛 Error Handling

Standard error response format:

```json
{
  "error": "USER_NOT_FOUND",
  "message": "User with ID 123 not found"
}
```

## 🧪 Testing

```bash (production)
- ✅ Text logging (development)
- ✅ Full Go team backing

### Logging by Environment

**Development** (text, debug level):
```
2024-06-27T10:30:45.123Z    INFO    user created    user_id=123 email=user@example.com
```

**Production/UAT/Staging** (JSON, configurable level):
```json
{"level":"info","msg":"user created","user_id":"123","email":"user@example.com","request_id":"abc123"}
```

Control with:
```bash
export LOG_LEVEL=debug      # debug, info, warn, error
export ENV=production       # Triggers JSON format
make build
./bin/go-starter
```

### Docker Deployment
```bash
make docker-build
docker push your-registry/go-starter:latest
```

### AWS Deployment
1. Configure `.env` with AWS Secrets Manager details
2. Set `AWS_SECRETS_NAME` environment variable
3. Deploy with task definition

## 🔒 Security Best Practices

1. **Never commit .env files** - Use `.env.example`
2. **Secrets in AWS Secrets Manager** - Not in code
3. **Context timeout on all operations** - Prevents hangs
4. **Circuit breaker on external APIs** - Prevents cascading failures
5. **Request ID logging** - For debugging and tracing
6. **Structured logging** - For security audits

## 📈 Performance Considerations

- **Caching**: Use TTL cache for frequently accessed data
- **Database**: Add indexes on frequently queried fields
- **Circuit breaker**: Configure thresholds based on SLA
- **File size limits**: Keep files under 500KB for readability

## 🤝 Contributing

When adding new features:

1. Create models in `internal/models/`
2. Create repository in `internal/repository/`
3. Create service in `internal/business/`
4. Create handlers in `internal/router/`
5. Add routes in `internal/router/routes.go`
6. Add Swagger documentation tags
7. Run `make check-file-sizes` before commit
8. Run `make test` and `make lint` before PR

## 📝 License

MIT

## 🆘 Troubleshooting

### Port already in use
```bash
# Change port in .env
SERVER_PORT=8081
```

### Database connection error
```bash
# Check DATABASE_DSN in .env
# For PostgreSQL: postgresql://user:password@host:5432/db
# For SQLite: test.db
```

### Swagger not showing
```bash
make swagger
# Visit http://localhost:8080/swagger/index.html
```

### Circuit breaker always open
Check the endpoint status:
```go
status := httpClient.CircuitBreakerStatus("endpoint-name")
fmt.Printf("Requests: %d, Failures: %d", status.Requests, status.TotalFailures)
```

## 📚 Additional Resources

- [Go Best Practices](https://golang.org/doc/effective_go)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Design Patterns in Go](https://refactoring.guru/design-patterns/go)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)

---

**Ready to build something great!** 🚀
