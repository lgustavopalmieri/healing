package indexes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

func CreateSpecialistsIndex(ctx context.Context, client *opensearchapi.Client, indexName string) error {
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"standard_analyzer": map[string]interface{}{
						"type":      "standard",
						"stopwords": "_english_",
					},
					"name_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "standard",
						"filter": []string{
							"lowercase",
							"asciifolding",
						},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type":     "text",
					"analyzer": "name_analyzer",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"email": map[string]interface{}{
					"type": "keyword",
				},
				"phone": map[string]interface{}{
					"type": "keyword",
				},
				"specialty": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard_analyzer",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"license_number": map[string]interface{}{
					"type": "keyword",
				},
				"description": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard_analyzer",
				},
				"keywords": map[string]interface{}{
					"type": "keyword",
				},
				"agreed_to_share": map[string]interface{}{
					"type": "boolean",
				},
				"rating": map[string]interface{}{
					"type": "float",
				},
				"status": map[string]interface{}{
					"type": "keyword",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	return createIndex(ctx, client, indexName, mapping)
}

func createIndex(ctx context.Context, client *opensearchapi.Client, indexName string, mapping map[string]interface{}) error {
	_, err := client.Indices.Exists(ctx, opensearchapi.IndicesExistsReq{
		Indices: []string{indexName},
	})
	if err == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("failed to encode mapping: %w", err)
	}

	_, err = client.Indices.Create(ctx, opensearchapi.IndicesCreateReq{
		Index: indexName,
		Body:  &buf,
	})
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}
