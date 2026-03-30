package opensearch

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	opensearchgo "github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/awsv2"
)

type Config struct {
	Addresses   []string
	Region      string
	IndexPrefix string
}

func NewClient(cfg Config) (*opensearchapi.Client, error) {
	osCfg := opensearchgo.Config{
		Addresses: cfg.Addresses,
	}

	if cfg.Region != "" {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(cfg.Region))
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		signer, err := requestsigner.NewSignerWithService(awsCfg, "es")
		if err != nil {
			return nil, fmt.Errorf("failed to create SigV4 signer: %w", err)
		}
		osCfg.Signer = signer
	}

	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: osCfg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create opensearch client: %w", err)
	}

	res, err := client.Ping(context.Background(), nil)
	if res != nil && res.Body != nil {
		res.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to ping opensearch: %w", err)
	}

	return client, nil
}
