package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type Repository struct {
	Client          *elasticsearch.Client
	IndexName       string
	EventDispatcher event.EventDispatcher
}

func NewRepository(client *elasticsearch.Client, indexName string, eventDispatcher event.EventDispatcher) *Repository {
	return &Repository{
		Client:          client,
		IndexName:       indexName,
		EventDispatcher: eventDispatcher,
	}
}
