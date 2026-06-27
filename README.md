# Go Starter - Production-Ready Template

A production-ready Go application template with clean architecture, comprehensive observability, and enterprise deployment patterns. Built for scalability, maintainability, and operational excellence.

## Features

- ✅ **Latest Go Version** (1.23+) - Using modern Go features with slog and generics
- ✅ **Clean Architecture** - Repository pattern with clear separation of concerns
- ✅ **Swagger/OpenAPI** - Auto-generated API documentation
- ✅ **Structured Logging** - JSON logging with context propagation, request tracing
- ✅ **HTTP Client with Circuit Breaker** - Resilience patterns for external API calls
- ✅ **Dual Caching System** - TTL cache + in-memory cache with lock control
- ✅ **Database Monitoring** - Slow query detection, connection pool management, timeouts
- ✅ **AWS Integration** - Secrets Manager, SQS consumers, SNS publishing
- ✅ **Prometheus Metrics** - Built-in observability with HTTP and DB metrics
- ✅ **OpenTelemetry Support** - Distributed tracing integration ready
- ✅ **Rate Limiting** - Configurable per-IP rate limiting middleware
- ✅ **Input Validation** - Struct-tag based validation with error handling
- ✅ **Docker Ready** - Multi-stage Dockerfile with health checks
- ✅ **Worker Mode** - Background job processing with scheduled tasks
- ✅ **Comprehensive Testing** - Mock-ready architecture with test examples

## Quick Start

```bash
make setup    # Install dependencies and initialize the project
make run      # Start HTTP server on port 8080
```

Visit http://localhost:8080/swagger/index.html for API documentation.

## Project Structure

```
go-starter/
├── cmd/                    # Application entry points
│   └── app/
│       └── main.go        # Main application file
├── internal/               # Private application code
│   ├── models/           # Data models
│   ├── business/         # Business logic (services)
│   ├── repository/       # Data access layer
│   ├── router/           # HTTP routes and handlers
│   ├── server/           # Server initialization
│   ├── database/         # Database configuration
│   ├── config/           # Configuration management
│   ├── logger/           # Logging setup
│   ├── http/             # HTTP client with circuit breaker
│   ├── cache/            # Caching implementation
│   ├── middleware/       # HTTP middleware
│   ├── aws/              # AWS integrations
│   └── external/         # External service integrations
├── pkg/                   # Public packages
├── migrations/            # Database migration files
├── docs/                  # Documentation
├── scripts/               # Utility scripts
├── config/                # Configuration files
├── go.mod                 # Module definition
├── Makefile               # Build automation
├── .env.example           # Environment variables template
└── docker-compose.yml       # Docker compose for dependencies
```

## Next Steps

- See [USAGE.md](USAGE.md) for all commands, modes, and detailed documentation
- See [docs/RUNNING_GUIDE.md](docs/RUNNING_GUIDE.md) for running in different environments
- See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design patterns

## License

MIT