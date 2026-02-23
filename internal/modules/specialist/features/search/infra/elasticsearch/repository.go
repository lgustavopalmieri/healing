package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

func (r *Repository) Search(ctx context.Context, input *searchinput.ListSearchInput) (*searchoutput.ListSearchOutput, error) {
	query, err := r.buildQuery(input)
	if err != nil {
		r.logger.Error(ctx, "failed to build elasticsearch query",
			observability.Field{Key: "error", Value: err.Error()})
		return nil, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		r.logger.Error(ctx, "failed to encode elasticsearch query",
			observability.Field{Key: "error", Value: err.Error()})
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.indexName),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(false),
	)
	if err != nil {
		r.logger.Error(ctx, "elasticsearch search request failed",
			observability.Field{Key: "error", Value: err.Error()})
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		r.logger.Error(ctx, "elasticsearch returned error",
			observability.Field{Key: "status", Value: res.Status()},
			observability.Field{Key: "body", Value: string(body)})
		return nil, fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	var esResponse elasticsearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		r.logger.Error(ctx, "failed to decode elasticsearch response",
			observability.Field{Key: "error", Value: err.Error()})
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	specialists := make([]*domain.Specialist, 0, len(esResponse.Hits.Hits))
	for _, hit := range esResponse.Hits.Hits {
		specialist := r.mapToSpecialist(hit.Source)
		specialists = append(specialists, specialist)
	}

	pageSize := input.Pagination.PageSize
	hasNext := len(specialists) > pageSize

	if hasNext {
		specialists = specialists[:pageSize]
	}

	cursorOutput := r.buildCursorOutput(input, specialists, esResponse.Hits.Hits, hasNext)

	return searchoutput.NewListSearchOutput(specialists, cursorOutput), nil
}
