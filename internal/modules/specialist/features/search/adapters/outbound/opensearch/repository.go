package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
	searchoutput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_output"
)

func (r *Repository) Search(ctx context.Context, input *searchinput.ListSearchInput) (*searchoutput.SearchResult, error) {
	query, err := r.buildQuery(input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}

	query["track_total_hits"] = false

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	searchResp, err := r.client.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{r.indexName},
		Body:    &buf,
	})
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	hits := searchResp.Hits.Hits
	pageSize := input.Pagination.PageSize
	hasNext := len(hits) > pageSize

	if hasNext {
		hits = hits[:pageSize]
	}

	specialists := make([]*domain.Specialist, 0, len(hits))
	for _, hit := range hits {
		var source opensearchSource
		if err := json.Unmarshal(hit.Source, &source); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrDecodingFailed, err)
		}
		specialists = append(specialists, r.mapToSpecialist(source))
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
