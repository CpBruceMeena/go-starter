package aws

import (
	"context"
	"io"
	"time"
)

// S3Client defines S3 operations interface
type S3Client interface {
	Upload(ctx context.Context, bucket, key string, body io.Reader, contentType string) error
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
}

// SNSClient defines SNS operations interface
type SNSClient interface {
	Publish(ctx context.Context, topicARN, message, subject string) error
	PublishToTopic(ctx context.Context, topicName, message, subject string) error
}

// SQSClient defines SQS operations interface
type SQSClient interface {
	SendMessage(ctx context.Context, queueURL, message string) error
	ReceiveMessages(ctx context.Context, queueURL string, maxMessages int) ([]SQSMessage, error)
	DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error
}

// SQSMessage represents an SQS message
type SQSMessage struct {
	ID            string
	Body          string
	ReceiptHandle string
}