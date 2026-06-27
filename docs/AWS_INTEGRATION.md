# AWS Integration Guide

This template includes built-in support for AWS services commonly used in production Go applications.

## Supported AWS Services

1. **Secrets Manager** - For configuration and secrets
2. **Task Definition** - For ECS deployments
3. **Environment variable overrides** - Local to production seamless transition

## Configuration Hierarchy

Configuration is loaded in this order (later overrides earlier):

```
1. Default values in code
   ↓
2. Environment variables (.env file)
   ↓
3. AWS Secrets Manager (if configured)
```

## 1. AWS Secrets Manager Setup

### Local Development (No AWS Required)

```bash
cp .env.example .env
# Edit .env with your local values
source .env
make run
```

### Production Setup with AWS

#### Step 1: Create AWS Secrets Manager Secret

```bash
# Create secret in AWS console or CLI
aws secretsmanager create-secret \
  --name my-app-secrets \
  --secret-string '
  {
    "database_url": "postgresql://user:pass@prod-db:5432/myapp",
    "server_port": 8080,
    "log_level": "info",
    "cache_ttl": 3600,
    "aws_region": "us-east-1"
  }
  '
```

#### Step 2: Configure Environment Variable

```bash
export AWS_SECRETS_NAME=my-app-secrets
export AWS_REGION=us-east-1
```

Or add to `.env`:
```
AWS_SECRETS_NAME=my-app-secrets
AWS_REGION=us-east-1
```

#### Step 3: Run Application

```bash
make run
# Application will automatically load from AWS Secrets Manager
```

### Local Configuration

```go
// internal/config/config.go handles the loading automatically

// In your application:
cfg, err := config.Load(ctx)

// cfg now contains:
// - Values from .env (if present)
// - Overridden by AWS Secrets Manager (if configured)
```

## 2. Secrets Manager Configuration Format

Secrets should be stored as JSON:

```json
{
  "server_port": 8080,
  "server_env": "production",
  "server_timeout": 30,
  "database_url": "postgresql://user:password@hostname:5432/dbname",
  "database_dsn": "user=postgres password=secret host=localhost port=5432 dbname=myapp",
  "aws_region": "us-east-1",
  "aws_role": "arn:aws:iam::123456789:role/my-app-role",
  "enable_swagger": false,
  "log_level": "info",
  "cache_ttl": 3600,
  "external_api_timeout": 10
}
```

## 3. ECS Task Definition Setup

### Sample Task Definition for ECS

```json
{
  "family": "go-starter",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "containerDefinitions": [
    {
      "name": "go-starter",
      "image": "YOUR_AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-starter:latest",
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
          "name": "AWS_REGION",
          "value": "us-east-1"
        },
        {
          "name": "AWS_SECRETS_NAME",
          "value": "my-app-secrets"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789:secret:my-app-secrets:database_url::"
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
  ],
  "executionRoleArn": "arn:aws:iam::123456789:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::123456789:role/go-starter-task-role"
}
```

## 4. IAM Roles & Policies

### Task Execution Role (allows ECS to launch tasks)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchGetImage",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchCheckLayerAvailability"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:CreateLogGroup"
      ],
      "Resource": "arn:aws:logs:us-east-1:123456789:log-group:/ecs/*"
    },
    {
      "Effect": "Allow",
      "Action": "secretsmanager:GetSecretValue",
      "Resource": "arn:aws:secretsmanager:us-east-1:123456789:secret:my-app-secrets-*"
    }
  ]
}
```

### Task Role (allows application to access AWS services)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:us-east-1:123456789:secret:my-app-secrets-*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": "arn:aws:s3:::my-app-bucket/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:GetItem",
        "dynamodb:PutItem"
      ],
      "Resource": "arn:aws:dynamodb:us-east-1:123456789:table/my-app-table"
    }
  ]
}
```

## 5. AWS SDK Integration

### Example: Using AWS Secrets Manager in Custom Code

```go
package aws

import (
    "context"
    "encoding/json"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// GetSecret retrieves a secret from AWS Secrets Manager
func GetSecret(ctx context.Context, region, secretName string) (map[string]string, error) {
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, err
    }
    
    client := secretsmanager.NewFromConfig(cfg)
    
    result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: &secretName,
    })
    if err != nil {
        return nil, err
    }
    
    var secrets map[string]string
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return nil, err
    }
    
    return secrets, nil
}
```

### Example: Using AWS S3

```go
package aws

import (
    "context"
    "io"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
    client *s3.Client
}

func NewS3Client(ctx context.Context, region string) (*S3Client, error) {
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, err
    }
    
    return &S3Client{
        client: s3.NewFromConfig(cfg),
    }, nil
}

func (sc *S3Client) UploadFile(ctx context.Context, bucket, key string, data io.Reader) error {
    _, err := sc.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: &bucket,
        Key:    &key,
        Body:   data,
    })
    return err
}

func (sc *S3Client) DownloadFile(ctx context.Context, bucket, key string) ([]byte, error) {
    result, err := sc.client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: &bucket,
        Key:    &key,
    })
    if err != nil {
        return nil, err
    }
    defer result.Body.Close()
    
    return io.ReadAll(result.Body)
}
```

## 6. Local Testing with AWS SDK

### Using AWS CLI Profiles

```bash
# Configure profile
aws configure --profile my-profile

# Use in code
export AWS_PROFILE=my-profile
make run
```

### Using LocalStack for Local AWS Emulation

```bash
# Install LocalStack
brew install localstack

# Start LocalStack
localstack start -d

# Configure AWS CLI for LocalStack
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1

# Create local secret
aws secretsmanager create-secret \
  --name my-app-secrets \
  --secret-string '{"database_url":"sqlite:///test.db"}' \
  --endpoint-url http://localhost:4566
```

## 7. Deployment Checklist

### Pre-Deployment

- [ ] Create Secrets Manager secret with all required values
- [ ] Create IAM roles and policies
- [ ] Create ECR repository
- [ ] Create CloudWatch log group
- [ ] Test locally with `.env` file
- [ ] Test with AWS credentials in development

### Deployment Steps

```bash
# Build Docker image
make docker-build

# Tag for ECR
docker tag go-starter:latest YOUR_AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-starter:latest

# Push to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin YOUR_AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com
docker push YOUR_AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-starter:latest

# Create/Update ECS service
aws ecs create-service \
  --cluster my-cluster \
  --service-name go-starter \
  --task-definition go-starter \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx]}"
```

### Post-Deployment

```bash
# Check service status
aws ecs describe-services \
  --cluster my-cluster \
  --services go-starter

# View logs
aws logs tail /ecs/go-starter --follow

# Health check
curl https://your-app-url/api/v1/health
```

## 8. Environment-Specific Configuration

### Development

```bash
# .env file
ENV=development
LOG_LEVEL=debug
ENABLE_SWAGGER=true
DATABASE_DSN=test.db
```

### Staging

```bash
# .env file (or environment variables)
ENV=staging
LOG_LEVEL=info
ENABLE_SWAGGER=true
DATABASE_URL=postgresql://user:pass@staging-db:5432/app
AWS_SECRETS_NAME=app-staging-secrets
```

### Production

```bash
# Environment variables (no .env file)
ENV=production
LOG_LEVEL=warn
ENABLE_SWAGGER=false
AWS_SECRETS_NAME=app-prod-secrets
# DATABASE_URL loaded from Secrets Manager
```

## 9. Monitoring & Logging

### CloudWatch Logs Integration

```go
// Configure in ECS Task Definition to send logs to CloudWatch
// Application logs automatically go to stdout
// Docker picks them up and sends to CloudWatch

// View logs
aws logs tail /ecs/go-starter --follow
```

### Request Tracing with X-Ray

```go
// Optional: Add AWS X-Ray instrumentation
// In internal/http/client.go, add X-Ray middleware

import "github.com/aws/xray-sdk-go/xray"

// Wrap HTTP client
client = xray.Client(client)
```

## 10. Troubleshooting

### "Unable to locate credentials"

```bash
# Solution 1: Set AWS credentials
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret

# Solution 2: Use AWS profile
export AWS_PROFILE=your-profile

# Solution 3: Check ~/.aws/credentials
cat ~/.aws/credentials
```

### "Access Denied to Secrets Manager"

```bash
# Check IAM policy
aws iam get-role-policy --role-name your-role --policy-name your-policy

# Verify secret exists
aws secretsmanager describe-secret --secret-id my-app-secrets

# Test secret access
aws secretsmanager get-secret-value --secret-id my-app-secrets
```

### "Connection timeout to database"

```bash
# Check security group
aws ec2 describe-security-groups --group-ids sg-xxx

# Check RDS instance status
aws rds describe-db-instances --db-instance-identifier my-db

# Verify connection string format
DATABASE_URL=postgresql://user:password@host:5432/dbname
```

## Resources

- [AWS Secrets Manager Documentation](https://docs.aws.amazon.com/secretsmanager/)
- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)

---

For more information, see [Architecture Guide](ARCHITECTURE.md) and the main [README](../README.md).
