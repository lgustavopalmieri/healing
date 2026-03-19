package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

func (r *Repository) Search(ctx context.Context, input *searchinput.ListSearchInput) (*searchoutput.SearchResult, error) {
	query, err := r.buildQuery(input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.indexName),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(false),
	)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.Status())
	}

	var esResponse elasticsearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	hits := esResponse.Hits.Hits
	pageSize := input.Pagination.PageSize
	hasNext := len(hits) > pageSize

	if hasNext {
		hits = hits[:pageSize]
	}

	specialists := make([]*domain.Specialist, 0, len(hits))
	for _, hit := range hits {
		specialists = append(specialists, r.mapToSpecialist(hit.Source))
	}

	var firstSortValues []any
	var lastSortValues []any

	if len(hits) > 0 {
		firstSortValues = hits[0].Sort
		lastSortValues = hits[len(hits)-1].Sort
	}

	return &searchoutput.SearchResult{
		Specialists:     specialists,
		HasNextPage:     hasNext,
		FirstSortValues: firstSortValues,
		LastSortValues:  lastSortValues,
	}, nil
}
