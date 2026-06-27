# Running the Application

This guide covers deployment and environment-specific configuration for the Go Starter application.

See [USAGE.md](../USAGE.md) for all commands, modes, and general usage.

## Quick Start

```bash
make run  # Start HTTP server (after running make setup)
```

Visit http://localhost:8080/swagger/index.html for API documentation.

## Application Modes

### HTTP Server Mode (Default)

Runs the application as a RESTful API server.

```bash
# Set mode (or leave empty for default)
export APP_MODE=http

# Start application
./bin/go-starter
```

**Output**:
```json
{"level":"info","msg":"application starting","env":"development","mode":"http","port":8080}
```

### Worker Mode

Runs the application as a background job processor with scheduled tasks.

```bash
# Set mode to worker
export APP_MODE=worker

# Start application
./bin/go-starter
```

**Output**:
```json
{"level":"info","msg":"application starting","env":"development","mode":"worker"}
{"level":"info","msg":"worker started","task_count":5}
```

**Running Tasks**:
- Cleanup old data (hourly)
- Sync external data (every 30 minutes)
- Health checks (every 5 minutes)
- Daily report generation
- Process notifications (every 5 minutes)

## Environment-Specific Configuration

### Development

```bash
# Use development config
cp config/.env.development .env

# Or manually set
export ENV=development
export LOG_LEVEL=debug
export DATABASE_DSN=test.db
export ENABLE_SWAGGER=true

# Run
make run
```

**Characteristics**:
- ✅ Text-based logging (human-readable)
- ✅ Debug level logs
- ✅ Swagger UI enabled
- ✅ SQLite database
- ✅ Hot reload with `make dev`

### UAT (User Acceptance Testing)

```bash
# Use UAT config
cp config/.env.uat .env

# Set environment variables
export ENV=uat
export AWS_REGION=us-east-1
export AWS_SECRETS_NAME=go-starter-uat-secrets

# Or for local database:
export DATABASE_URL=postgresql://postgres:postgres@localhost:5432/go_starter_uat

# Create database if needed
createdb go_starter_uat

# Run
make run
```

**Characteristics**:
- ✅ JSON structured logging
- ✅ Info level logs
- ✅ Swagger UI enabled (for testing)
- ✅ PostgreSQL database
- ✅ AWS Secrets Manager integration
- ✅ Higher timeouts for stability

### Staging

```bash
# Use staging config
cp config/.env.staging .env

# Or set environment variables
export ENV=staging
export LOG_LEVEL=warn
export ENABLE_SWAGGER=false
export AWS_REGION=us-east-1
export AWS_SECRETS_NAME=go-starter-staging-secrets

# Database loaded from AWS Secrets Manager
# No need to set DATABASE_URL (loaded automatically)

# Run
make run
```

**Characteristics**:
- ✅ JSON structured logging
- ✅ Warn level logs (reduced verbosity)
- ✅ Swagger UI disabled
- ✅ RDS database
- ✅ AWS Secrets Manager enabled
- ✅ Production-like configuration

### Production

```bash
# Production should NOT use .env files
# Set all environment variables in deployment:

# In ECS Task Definition:
export ENV=production
export LOG_LEVEL=error
export ENABLE_SWAGGER=false
export AWS_REGION=us-east-1
export AWS_SECRETS_NAME=go-starter-prod-secrets

# Run
./bin/go-starter
```

**Characteristics**:
- ✅ JSON structured logging only
- ✅ Error level logs only
- ✅ Swagger UI disabled
- ✅ RDS Multi-AZ database
- ✅ AWS Secrets Manager enforced
- ✅ Auto-scaling enabled
- ✅ CloudWatch monitoring enabled

## Docker Deployment

### Local Docker (Development)

```bash
# Build image
make docker-build

# Run with development config
docker run -p 8080:8080 \
  -e ENV=development \
  -e LOG_LEVEL=debug \
  -e DATABASE_DSN=test.db \
  go-starter:latest
```

### Docker Compose (with PostgreSQL)

```bash
# Start all services
docker-compose up

# Run in worker mode
docker-compose up -d
docker exec go-starter-app sh -c "export APP_MODE=worker && ./app"

# View logs
docker-compose logs -f app

# Stop all
docker-compose down
```

## Kubernetes Deployment

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-starter
spec:
  replicas: 2
  selector:
    matchLabels:
      app: go-starter
  template:
    metadata:
      labels:
        app: go-starter
    spec:
      containers:
      - name: go-starter
        image: your-registry/go-starter:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: APP_MODE
          value: "http"
        - name: AWS_REGION
          value: "us-east-1"
        - name: AWS_SECRETS_NAME
          valueFrom:
            secretKeyRef:
              name: go-starter-secrets
              key: secrets-name
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
```

### Worker Deployment

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: go-starter-worker
spec:
  schedule: "*/5 * * * *"  # Every 5 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: go-starter
            image: your-registry/go-starter:1.0.0
            env:
            - name: ENV
              value: "production"
            - name: APP_MODE
              value: "worker"
            - name: AWS_REGION
              value: "us-east-1"
            - name: AWS_SECRETS_NAME
              valueFrom:
                secretKeyRef:
                  name: go-starter-secrets
                  key: secrets-name
          restartPolicy: Never
```

## AWS ECS Deployment

### Task Definition Example

```json
{
  "family": "go-starter",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "go-starter",
      "image": "ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-starter:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENV",
          "value": "production"
        },
        {
          "name": "APP_MODE",
          "value": "http"
        },
        {
          "name": "AWS_REGION",
          "value": "us-east-1"
        },
        {
          "name": "AWS_SECRETS_NAME",
          "value": "go-starter-prod-secrets"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/go-starter",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

### Worker Task Definition

```bash
# Run worker as a one-time task
aws ecs run-task \
  --cluster my-cluster \
  --task-definition go-starter:1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx]}" \
  --overrides '{"containerOverrides":[{"name":"go-starter","environment":[{"name":"APP_MODE","value":"worker"}]}]}'
```

## Logging Output Examples

### Development (Text Format)

```
2024-06-27T10:30:45.123456Z    INFO    application starting    env=development mode=http port=8080
2024-06-27T10:30:46.234567Z    DEBUG   cache stats     entries=42
2024-06-27T10:30:47.345678Z    INFO    task started    name=cleanup_old_data
```

### Production (JSON Format)

```json
{"level":"info","msg":"application starting","env":"production","mode":"http","port":8080}
{"level":"info","msg":"cache stats","entries":42}
{"level":"info","msg":"task started","name":"cleanup_old_data","duration_ms":1234}
```

## Monitoring & Health Checks

### Health Check Endpoint

```bash
curl http://localhost:8080/api/v1/health

# Output:
{"status":"ok"}
```

### Logging Health Checks

All requests are logged with:
- Request method and path
- Response status code
- Request duration
- Client IP address
- Request ID (for tracing)

```json
{
  "level": "info",
  "msg": "HTTP request completed",
  "method": "POST",
  "path": "/api/v1/users",
  "status": 201,
  "duration_ms": 45,
  "remote_addr": "127.0.0.1",
  "request_id": "abc123def456"
}
```

## Troubleshooting

### Issue: "Port already in use"

```bash
# Change port via environment variable
export SERVER_PORT=8081
make run
```

### Issue: "Database connection error"

```bash
# Check database configuration
echo $DATABASE_URL
echo $DATABASE_DSN

# For PostgreSQL, test connection:
psql postgresql://user:password@localhost:5432/dbname -c "SELECT 1"
```

### Issue: "Worker tasks not running"

```bash
# Enable debug logging
export LOG_LEVEL=debug
export APP_MODE=worker
make run

# Check task registration
# You should see: "task registered name=... interval=..."
```

### Issue: "AWS Secrets Manager access denied"

```bash
# Verify IAM role/policy
aws iam get-role-policy --role-name your-role --policy-name your-policy

# Test Secrets Manager access
aws secretsmanager get-secret-value --secret-id go-starter-secrets
```

## Performance Tuning

### For HTTP Server

```bash
# Increase connection timeout
export SERVER_TIMEOUT=120

# Adjust cache TTL
export CACHE_TTL=7200

# For high traffic, use Kubernetes HPA
```

### For Worker

```bash
# Run multiple worker instances in Kubernetes
# Or use scheduled Lambda functions

# Adjust task timeouts in internal/worker/examples.go
Task{
  Name: "sync_data",
  Interval: 30 * time.Minute,
  Timeout: 5 * time.Minute,  // Adjust this
}
```

## Next Steps

- See [Architecture Guide](../docs/ARCHITECTURE.md) for design patterns
- See [AWS Integration Guide](../docs/AWS_INTEGRATION.md) for cloud setup
- See [Cache Usage Guide](../docs/CACHE_USAGE.md) for caching patterns
- See [Suggestions](../docs/SUGGESTIONS.md) for enhancement ideas
