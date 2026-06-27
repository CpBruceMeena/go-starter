package config

import (
	"context"
	"encoding/json"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/joho/godotenv"
)

// CircuitBreakerConfig defines circuit breaker settings for external API clients
type CircuitBreakerConfig struct {
	Name        string `json:"name"` // Client name (e.g., "payment-api")
	MaxRequests uint32 `json:"max_requests"`
	Interval    string `json:"interval"`  // Duration string (e.g., "60s")
	Timeout     string `json:"timeout"`   // Duration string (e.g., "30s")
	Threshold   uint32 `json:"threshold"` // Failure threshold (0-100 percent)
}

// DatabaseConfig defines database query monitoring settings
type DatabaseConfig struct {
	// SlowQueryThreshold is the duration after which queries are logged as slow (e.g., "1s", "500ms")
	SlowQueryThreshold string `json:"slow_query_threshold"`
	// QueryTimeout is the maximum allowed query duration before cancellation (e.g., "5s", "30s")
	QueryTimeout string `json:"query_timeout"`
	// MaxOpenConns limits the number of open database connections
	MaxOpenConns int `json:"max_open_conns"`
	// MaxIdleConns limits the number of idle database connections
	MaxIdleConns int `json:"max_idle_conns"`
	// ConnMaxLifetime is the maximum connection lifetime (e.g., "1h", "30m")
	ConnMaxLifetime string `json:"conn_max_lifetime"`
}

// Config holds all application configuration
type Config struct {
	// Server
	ServerPort    int    `json:"server_port"`
	ServerEnv     string `json:"server_env"`
	ServerTimeout int    `json:"server_timeout"` // seconds

	// Database
	DatabaseURL string `json:"database_url"`
	DatabaseDSN string `json:"database_dsn"`
	Database    DatabaseConfig `json:"database"`

	// AWS
	AWSRegion string `json:"aws_region"`
	AWSRole   string `json:"aws_role"`

	// Features (configurable - skip components not needed)
	EnableSwagger    bool   `json:"enable_swagger"`
	EnableDatabase   bool   `json:"enable_database"`
	EnableCache      bool   `json:"enable_cache"`
	EnableHTTPClient bool   `json:"enable_http_client"`
	EnableConsumer   bool   `json:"enable_consumer"`
	EnableWorker     bool   `json:"enable_worker"`
	LogLevel         string `json:"log_level"`

	// Cache
	CacheTTL int `json:"cache_ttl"` // seconds

	// External APIs & Circuit Breaker
	ExternalAPITimeout    int                             `json:"external_api_timeout"` // seconds
	CircuitBreakerEnabled bool                            `json:"circuit_breaker_enabled"`
	CircuitBreakerConfigs map[string]CircuitBreakerConfig `json:"circuit_breaker_configs"`

	// Consumer (SQS/Kafka)
	ConsumerType       string   `json:"consumer_type"` // "sqs", "kafka"
	SQSQueueURL        string   `json:"sqs_queue_url"`
	SQSMaxMessages     int      `json:"sqs_max_messages"`
	KafkaBrokers       []string `json:"kafka_brokers"`
	KafkaTopic         string   `json:"kafka_topic"`
	KafkaConsumerGroup string   `json:"kafka_consumer_group"`
}

// IsFeatureEnabled checks if a feature is enabled
func (c *Config) IsFeatureEnabled(feature string) bool {
	switch feature {
	case "database":
		return c.EnableDatabase
	case "cache":
		return c.EnableCache
	case "http_client":
		return c.EnableHTTPClient
	case "consumer":
		return c.EnableConsumer
	case "worker":
		return c.EnableWorker
	default:
		return false
	}
}

// Load loads configuration from environment variables and optionally from AWS Secrets Manager
func Load(ctx context.Context) (*Config, error) {
	// First, try to load from .env file (development)
	_ = godotenv.Load()

	cfg := &Config{
		// Server
		ServerPort:    getEnvInt("SERVER_PORT", 8080),
		ServerEnv:     getEnv("ENV", "development"),
		ServerTimeout: getEnvInt("SERVER_TIMEOUT", 30),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DatabaseDSN: getEnv("DATABASE_DSN", ""),
		Database: DatabaseConfig{
			SlowQueryThreshold: getEnv("DB_SLOW_QUERY_THRESHOLD", "1s"),
			QueryTimeout:       getEnv("DB_QUERY_TIMEOUT", "30s"),
			MaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime:    getEnv("DB_CONN_MAX_LIFETIME", "1h"),
		},

		// AWS
		AWSRegion: getEnv("AWS_REGION", "us-east-1"),
		AWSRole:   getEnv("AWS_ROLE", ""),

		// Features (default: all enabled for backward compatibility, except consumer/worker)
		EnableSwagger:    getEnvBool("ENABLE_SWAGGER", true),
		EnableDatabase:   getEnvBool("ENABLE_DATABASE", true),
		EnableCache:      getEnvBool("ENABLE_CACHE", true),
		EnableHTTPClient: getEnvBool("ENABLE_HTTP_CLIENT", true),
		EnableConsumer:   getEnvBool("ENABLE_CONSUMER", false),
		EnableWorker:     getEnvBool("ENABLE_WORKER", false),
		LogLevel:         getEnv("LOG_LEVEL", "info"),

		// Cache
		CacheTTL: getEnvInt("CACHE_TTL", 3600),

		// External APIs & Circuit Breaker
		ExternalAPITimeout:    getEnvInt("EXTERNAL_API_TIMEOUT", 10),
		CircuitBreakerEnabled: getEnvBool("CIRCUIT_BREAKER_ENABLED", true),
		CircuitBreakerConfigs: make(map[string]CircuitBreakerConfig),

		// Consumer
		ConsumerType:       getEnv("CONSUMER_TYPE", ""),
		SQSQueueURL:        getEnv("SQS_QUEUE_URL", ""),
		SQSMaxMessages:     getEnvInt("SQS_MAX_MESSAGES", 10),
		KafkaTopic:         getEnv("KAFKA_TOPIC", ""),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", ""),
	}

	// If AWS Secrets Manager secret name is provided, load from there
	if secretName := getEnv("AWS_SECRETS_NAME", ""); secretName != "" {
		awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.AWSRegion))
		if err != nil {
			return nil, err
		}

		secretCfg, err := loadFromSecretsManager(ctx, awsCfg, secretName)
		if err != nil {
			return nil, err
		}

		// Merge secrets (secrets override env vars)
		mergeConfig(cfg, secretCfg)
	}

	return cfg, nil
}

// loadFromSecretsManager loads configuration from AWS Secrets Manager
func loadFromSecretsManager(ctx context.Context, awsCfg aws.Config, secretName string) (*Config, error) {
	client := secretsmanager.NewFromConfig(awsCfg)

	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal([]byte(*result.SecretString), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// mergeConfig merges secret config into main config (secrets override)
func mergeConfig(dst, src *Config) {
	if src.ServerPort != 0 {
		dst.ServerPort = src.ServerPort
	}
	if src.ServerEnv != "" {
		dst.ServerEnv = src.ServerEnv
	}
	if src.DatabaseURL != "" {
		dst.DatabaseURL = src.DatabaseURL
	}
	if src.DatabaseDSN != "" {
		dst.DatabaseDSN = src.DatabaseDSN
	}
	if src.AWSRegion != "" {
		dst.AWSRegion = src.AWSRegion
	}
}

// Helper functions to read environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
