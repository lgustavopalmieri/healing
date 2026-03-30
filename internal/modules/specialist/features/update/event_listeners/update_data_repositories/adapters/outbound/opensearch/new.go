package opensearch

import (
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type Repository struct {
	Client    *opensearchapi.Client
	IndexName string
}

func NewRepository(client *opensearchapi.Client, indexName string) *Repository {
	return &Repository{
		Client:    client,
		IndexName: indexName,
	}
}
