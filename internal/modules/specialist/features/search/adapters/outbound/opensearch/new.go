package opensearch

import (
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type Repository struct {
	client    *opensearchapi.Client
	indexName string
}

func NewRepository(client *opensearchapi.Client, indexName string) *Repository {
	return &Repository{
		client:    client,
		indexName: indexName,
	}
}
