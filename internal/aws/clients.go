package aws

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Config holds AWS configuration
type Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// Clients holds AWS service clients
type Clients struct {
	S3  *s3.Client
	SNS *sns.Client
	SQS *sqs.Client
}

// NewClients creates AWS service clients
func NewClients(ctx context.Context, cfg Config) (*Clients, error) {
	var awsCfg aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		)
	} else {
		awsCfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Clients{
		S3:  s3.NewFromConfig(awsCfg),
		SNS: sns.NewFromConfig(awsCfg),
		SQS: sqs.NewFromConfig(awsCfg),
	}, nil
}

// S3Client implementation
type s3Client struct {
	client *s3.Client
}

func NewS3Client(clients *Clients) S3Client {
	return &s3Client{client: clients.S3}
}

func (c *s3Client) Upload(ctx context.Context, bucket, key string, body io.Reader, contentType string) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

func (c *s3Client) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (c *s3Client) Delete(ctx context.Context, bucket, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (c *s3Client) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)
	url, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return url.URL, nil
}

// SNSClient implementation
type snsClient struct {
	client     *sns.Client
	topicCache map[string]string
}

func NewSNSClient(clients *Clients) SNSClient {
	return &snsClient{client: clients.SNS, topicCache: make(map[string]string)}
}

func (c *snsClient) Publish(ctx context.Context, topicARN, message, subject string) error {
	_, err := c.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(message),
		Subject:  aws.String(subject),
	})
	return err
}

func (c *snsClient) PublishToTopic(ctx context.Context, topicName, message, subject string) error {
	topicARN, err := c.getOrCreateTopic(ctx, topicName)
	if err != nil {
		return err
	}
	return c.Publish(ctx, topicARN, message, subject)
}

func (c *snsClient) getOrCreateTopic(ctx context.Context, topicName string) (string, error) {
	if arn, ok := c.topicCache[topicName]; ok {
		return arn, nil
	}
	arn, err := c.client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		return "", err
	}
	c.topicCache[topicName] = *arn.TopicArn
	return *arn.TopicArn, nil
}

// SQSClient implementation
type sqsClient struct {
	client *sqs.Client
}

func NewSQSClient(clients *Clients) SQSClient {
	return &sqsClient{client: clients.SQS}
}

func (c *sqsClient) SendMessage(ctx context.Context, queueURL, message string) error {
	_, err := c.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(message),
	})
	return err
}

func (c *sqsClient) ReceiveMessages(ctx context.Context, queueURL string, maxMessages int) ([]SQSMessage, error) {
	result, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: int32(maxMessages),
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]SQSMessage, len(result.Messages))
	for i, msg := range result.Messages {
		messages[i] = SQSMessage{
			ID:            safeString(msg.MessageId),
			Body:          safeString(msg.Body),
			ReceiptHandle: safeString(msg.ReceiptHandle),
		}
	}
	return messages, nil
}

func (c *sqsClient) DeleteMessage(ctx context.Context, queueURL, receiptHandle string) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}