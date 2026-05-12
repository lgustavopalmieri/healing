package sns

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Config struct {
	Region      string
	Endpoint    string
	TopicPrefix string
}

func NewClient(ctx context.Context, cfg Config) (*sns.Client, error) {
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

	snsOpts := []func(*sns.Options){}
	if cfg.Endpoint != "" {
		snsOpts = append(snsOpts, func(o *sns.Options) {
			o.BaseEndpoint = &cfg.Endpoint
		})
	}

	client := sns.NewFromConfig(awsCfg, snsOpts...)
	return client, nil
}
