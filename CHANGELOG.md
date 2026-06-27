# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Database query monitoring library with configurable limits and slow query logging
- Connection pool settings: `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`
- Slow query threshold configuration: `DB_SLOW_QUERY_THRESHOLD`
- Query timeout configuration: `DB_QUERY_TIMEOUT`
- Automatic warning logs for queries exceeding slow query threshold
- Integration with existing logger for structured slow query logs

### Fixed
- Duplicate package declaration in `internal/consumer/consumer.go`
- Circuit breaker status return type in `internal/http/client.go`