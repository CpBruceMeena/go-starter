# Go Starter - Production-Ready Template

A comprehensive Go application starter template with best practices, modern patterns, and production-ready features built-in. Includes database query monitoring, slow query detection, and connection pool management.

## Features

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