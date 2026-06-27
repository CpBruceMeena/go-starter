# JSON Logging Guide

This template uses Go's built-in `log/slog` package for structured, JSON-formatted logging in production environments.

## Quick Overview

### Development (Human-Readable Text)

```bash
export ENV=development
make run
```

Output:
```
2024-06-27T10:30:45.123Z    INFO    starting server    env=development port=8080
2024-06-27T10:30:46.234Z    INFO    user created      user_id=usr-123 email=user@example.com
```

### Production (JSON Structured)

```bash
export ENV=production
make run
```

Output:
```json
{"level":"info","msg":"starting server","env":"production","port":8080}
{"level":"info","msg":"user created","user_id":"usr-123","email":"user@example.com","request_id":"req-abc"}
```

## Log Levels

Control verbosity with `LOG_LEVEL` environment variable:

```bash
export LOG_LEVEL=debug      # Most verbose
export LOG_LEVEL=info       # Default
export LOG_LEVEL=warn       # Warnings and errors only
export LOG_LEVEL=error      # Errors only
```

## Logging in Your Code

### Basic Logging

```go
import "github.com/your-org/go-starter/internal/logger"

log := logger.Default()

// Info level
log.Info("user signup", "user_id", user.ID, "email", user.Email)

// Warning level
log.Warn("high memory usage", "usage_percent", 85)

// Error level
log.Error("database connection failed", "error", err.Error())
```

### Context-Aware Logging

Include request tracking information automatically:

```go
// WithContext extracts request_id, trace_id, user_id from context
log.WithContext(ctx).Info("processing request", "operation", "create_user")

// Or use the convenience methods
log.InfoContext(ctx, "user created", "user_id", user.ID)
log.WarnContext(ctx, "operation slow", "duration_ms", 5000)
log.ErrorContext(ctx, "operation failed", "error", err.Error())
```

### Adding Context Values

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "request_id", "req-123")
ctx = context.WithValue(ctx, "user_id", "usr-456")

log.WithContext(ctx).Info("action performed")
// Output: {"level":"info","msg":"action performed","request_id":"req-123","user_id":"usr-456"}
```

## JSON Output Format

Each log entry contains:

```json
{
  "level": "info",           // Log level: debug, info, warn, error
  "msg": "operation completed",
  "timestamp": "2024-06-27T10:30:45.123Z",
  "request_id": "req-abc123", // If in context
  "user_id": "usr-456",        // If in context  
  "duration_ms": 145,
  "status": 201,
  "error": null
}
```

## Production Logging Best Practices

### ✅ DO

1. **Log important business events**
   ```go
   log.InfoContext(ctx, "user registered",
       "user_id", user.ID,
       "email", user.Email,
       "signup_source", "web",
   )
   ```

2. **Include request IDs for tracing**
   ```go
   // Automatically added by middleware
   log.InfoContext(ctx, "payment processed", "amount", 99.99)
   ```

3. **Log errors with context**
   ```go
   if err != nil {
       log.ErrorContext(ctx, "payment failed",
           "error", err.Error(),
           "payment_id", paymentID,
           "retry_count", retries,
       )
   }
   ```

4. **Use structured fields**
   ```go
   log.Info("API response",
       "method", "POST",
       "path", "/api/users",
       "status", 201,
       "response_time_ms", 45,
   )
   ```

### ❌ DON'T

1. **Don't log sensitive data**
   ```go
   // ❌ Bad: Password in log
   log.Info("user login", "password", user.Password)
   
   // ✅ Good: Only log safe data
   log.Info("user login", "user_id", user.ID, "success", true)
   ```

2. **Don't use string concatenation**
   ```go
   // ❌ Bad: No structure
   log.Info("User " + user.Email + " created with ID " + user.ID)
   
   // ✅ Good: Structured fields
   log.Info("user created", "email", user.Email, "user_id", user.ID)
   ```

3. **Don't forget context propagation**
   ```go
   // ❌ Bad: Missing request_id
   log.Info("processing started")
   
   // ✅ Good: Include context
   log.InfoContext(ctx, "processing started")
   ```

4. **Don't log at wrong level**
   ```go
   // ❌ Bad: Info for error
   log.Info("Database error:", err.Error())
   
   // ✅ Good: Correct level
   log.Error("database connection failed", "error", err.Error())
   ```

## Parsing JSON Logs

### Using jq (Command Line)

```bash
# Pretty print
cat app.log | jq

# Filter by level
cat app.log | jq 'select(.level == "error")'

# Extract specific fields
cat app.log | jq '.msg, .error, .duration_ms'

# Filter by request_id
cat app.log | jq 'select(.request_id == "req-123")'

# Count errors per endpoint
cat app.log | jq 'select(.level == "error") | .path' | sort | uniq -c
```

### Using ELK Stack (Production)

```bash
# Elasticsearch Logstash Kibana
# Automatically parses JSON logs and makes them searchable

# Query: Find all errors in the last hour
GET logs-app-*/_search
{
  "query": {
    "bool": {
      "filter": [
        {"range": {"@timestamp": {"gte": "now-1h"}}},
        {"term": {"level": "error"}}
      ]
    }
  }
}
```

### Using CloudWatch Insights (AWS)

```
fields @timestamp, @message, level, request_id, user_id, duration_ms
| filter level = "error"
| stats count() by level
| sort count() desc
```

## Custom Logger Setup

If you need custom logger configuration:

```go
package logger

import (
    "log/slog"
    "os"
)

func NewWithOptions(level slog.Level, addSource bool) *Logger {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:     level,
        AddSource: addSource,  // Include file:line info
    })
    
    return &Logger{
        Logger: slog.New(handler),
    }
}
```

Usage:
```go
log := logger.NewWithOptions(slog.LevelDebug, true)
```

## Performance Considerations

### Log Volume

- **Development**: Debug logs acceptable (small impact)
- **Production**: Warn+ only (minimal performance impact)

```go
// Control with LOG_LEVEL
export LOG_LEVEL=warn  # Only warn and error in production
```

### Async Logging (Optional)

For high-throughput applications, buffer logs:

```go
// Create buffered handler
var buf strings.Builder
handler := slog.NewJSONHandler(&buf, nil)

// Or use third-party like lumberjack for rotation
// import "gopkg.in/natefinch/lumberjack.v2"
```

## Example Log Sequences

### User Registration Flow

```json
{"level":"info","msg":"registration started","email":"user@example.com","request_id":"req-1"}
{"level":"info","msg":"email validation passed","email":"user@example.com","request_id":"req-1"}
{"level":"info","msg":"user created","user_id":"usr-123","email":"user@example.com","request_id":"req-1"}
{"level":"info","msg":"confirmation email sent","user_id":"usr-123","email":"user@example.com","request_id":"req-1"}
{"level":"info","msg":"registration completed","user_id":"usr-123","duration_ms":245,"request_id":"req-1"}
```

### Error Scenario

```json
{"level":"info","msg":"payment processing started","payment_id":"pay-456","amount":99.99,"request_id":"req-2"}
{"level":"warn","msg":"payment validation warning","warning":"cvv_check_slow","request_id":"req-2"}
{"level":"error","msg":"payment processing failed","error":"insufficient_funds","payment_id":"pay-456","retry_count":1,"request_id":"req-2"}
{"level":"info","msg":"payment retry scheduled","payment_id":"pay-456","retry_at":"2024-06-27T10:35:00Z","request_id":"req-2"}
```

## Monitoring & Alerts

### Log-Based Alerts

```bash
# Alert on error rate
if cat app.log | jq 'select(.level == "error")' | wc -l > 100; then
    send_alert "High error rate detected"
fi

# Alert on slow requests
cat app.log | jq 'select(.duration_ms > 5000)' | wc -l
```

### CloudWatch Alarms

```bash
# Create alarm for error logs
aws logs put-metric-filter \
  --log-group-name /ecs/go-starter \
  --filter-name "ErrorCount" \
  --filter-pattern "[... level=ERROR ...]" \
  --metric-transformations metricName=ErrorCount,metricValue=1
```

## Troubleshooting

### Logs not showing up

```bash
# Check LOG_LEVEL
echo $LOG_LEVEL

# Check ENV
echo $ENV

# Run with debug
export LOG_LEVEL=debug
make run
```

### Can't parse JSON logs

```bash
# Check if valid JSON
cat app.log | jq . 2>&1 | head -20

# Try pretty printing
cat app.log | jq . | head -50
```

### Missing request IDs

```go
// Ensure middleware is adding to context
ctx = context.WithValue(ctx, "request_id", requestID)
log.InfoContext(ctx, "message")  // Use InfoContext
```

---

For more information, see:
- [Running Guide](RUNNING_GUIDE.md) - Environment setup
- [Architecture Guide](ARCHITECTURE.md) - Where to add logging
- [Go slog Documentation](https://pkg.go.dev/log/slog)
