package sqs

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Config struct {
	Region      string
	QueuePrefix string
	Endpoint    string
}

func NewClient(ctx context.Context, cfg Config) (*sqs.Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{}

	if cfg.Region != "" {
		opts = append(opts, awsconfig.WithRegion(cfg.Region))
	}

	if cfg.Endpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", "test"),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	sqsOpts := []func(*sqs.Options){}
	if cfg.Endpoint != "" {
		sqsOpts = append(sqsOpts, func(o *sqs.Options) {
			o.BaseEndpoint = &cfg.Endpoint
		})
	}

	client := sqs.NewFromConfig(awsCfg, sqsOpts...)
	return client, nil
}
