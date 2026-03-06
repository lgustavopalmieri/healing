package elasticsearch

import (
	"fmt"

	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
)

func (r *Repository) buildQuery(input *searchinput.ListSearchInput) (map[string]any, error) {
	query := map[string]any{
		"query": r.buildBoolQuery(input),
		"sort":  r.buildSort(input),
		"size":  input.Pagination.PageSize + 1,
	}

	searchAfter, err := r.buildSearchAfter(input)
	if err != nil {
		return nil, fmt.Errorf("failed to build search_after: %w", err)
	}
	if searchAfter != nil {
		query["search_after"] = searchAfter
	}

	return query, nil
}

func (r *Repository) buildBoolQuery(input *searchinput.ListSearchInput) map[string]any {
	must := make([]any, 0)

	if input.HasSearchTerm() {
		searchTerm := *input.SearchTerm

		if len(searchTerm) <= 3 {
			must = append(must, map[string]any{
				"bool": map[string]any{
					"should": []any{
						map[string]any{
							"wildcard": map[string]any{
								"name": map[string]any{
									"value":            searchTerm + "*",
									"case_insensitive": true,
								},
							},
						},
						map[string]any{
							"wildcard": map[string]any{
								"description": map[string]any{
									"value":            "*" + searchTerm + "*",
									"case_insensitive": true,
								},
							},
						},
						map[string]any{
							"wildcard": map[string]any{
								"specialty": map[string]any{
									"value":            "*" + searchTerm + "*",
									"case_insensitive": true,
								},
							},
						},
						map[string]any{
							"term": map[string]any{
								"keywords": searchTerm,
							},
						},
					},
					"minimum_should_match": 1,
				},
			})
		} else {
			must = append(must, map[string]any{
				"multi_match": map[string]any{
					"query":     searchTerm,
					"fields":    []string{"name^3", "description^2", "specialty^2", "keywords"},
					"type":      "best_fields",
					"operator":  "or",
					"fuzziness": "AUTO",
				},
			})
		}
	}

	if input.HasFilters() {
		for _, filter := range input.Filters {
			must = append(must, r.buildFilterQuery(filter))
		}
	}

	if len(must) == 0 {
		return map[string]any{
			"match_all": map[string]any{},
		}
	}

	return map[string]any{
		"bool": map[string]any{
			"must": must,
		},
	}
}

func (r *Repository) buildFilterQuery(filter searchinput.Filter) map[string]any {
	switch filter.Field {
	case searchinput.FieldKeywords:
		return map[string]any{
			"term": map[string]any{
				"keywords": filter.Value,
			},
		}
	case searchinput.FieldStatus:
		if len(filter.Values) > 0 {
			return map[string]any{
				"terms": map[string]any{
					"status": filter.Values,
				},
			}
		}
		return map[string]any{
			"term": map[string]any{
				"status": filter.Value,
			},
		}
	case searchinput.FieldSpecialty:
		return map[string]any{
			"match": map[string]any{
				"specialty": map[string]any{
					"query":    filter.Value,
					"operator": "and",
				},
			},
		}
	case searchinput.FieldName:
		return map[string]any{
			"match": map[string]any{
				"name": map[string]any{
					"query":    filter.Value,
					"operator": "and",
				},
			},
		}
	case searchinput.FieldDescription:
		return map[string]any{
			"match": map[string]any{
				"description": map[string]any{
					"query":    filter.Value,
					"operator": "and",
				},
			},
		}
	default:
		return map[string]any{
			"match": map[string]any{
				string(filter.Field): filter.Value,
			},
		}
	}
}

func (r *Repository) buildSort(input *searchinput.ListSearchInput) []any {
	sort := make([]any, 0)

	if input.HasSort() {
		for _, s := range input.Sort {
			fieldName := r.mapSortFieldToES(s.Field)
			sort = append(sort, map[string]any{
				fieldName: map[string]any{
					"order": string(s.Order),
				},
			})
		}
	}

	sort = append(sort, map[string]any{
		"id": map[string]any{
			"order": "asc",
		},
	})

	return sort
}

func (r *Repository) buildSearchAfter(input *searchinput.ListSearchInput) ([]any, error) {
	if input.Pagination.IsFirstPage() {
		return nil, nil
	}

	decoded, err := input.Pagination.DecodeMultiSortCursor()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}

	if decoded == nil {
		return nil, nil
	}

	return decoded.SortValues, nil
}
