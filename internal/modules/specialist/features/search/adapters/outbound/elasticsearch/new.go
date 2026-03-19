package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
)

type Repository struct {
	client    *elasticsearch.Client
	indexName string
}

func NewRepository(client *elasticsearch.Client, indexName string) *Repository {
	return &Repository{
		client:    client,
		indexName: indexName,
	}
}
