package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type Repository struct {
	Client          *elasticsearch.Client
	IndexName       string
	Logger          observability.Logger
	EventDispatcher event.EventDispatcher
}

func NewRepository(client *elasticsearch.Client, indexName string, logger observability.Logger, eventDispatcher event.EventDispatcher) *Repository {
	return &Repository{
		Client:          client,
		IndexName:       indexName,
		Logger:          logger,
		EventDispatcher: eventDispatcher,
	}
}
