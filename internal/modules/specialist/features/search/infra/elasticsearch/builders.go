package elasticsearch

import searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"

func (r *Repository) buildQuery(input *searchinput.ListSearchInput) map[string]interface{} {
	query := map[string]interface{}{
		"query": r.buildBoolQuery(input),
		"sort":  r.buildSort(input),
		"size":  input.Pagination.PageSize + 1,
	}

	if searchAfter := r.buildSearchAfter(input); searchAfter != nil {
		query["search_after"] = searchAfter
	}

	return query
}

func (r *Repository) buildBoolQuery(input *searchinput.ListSearchInput) map[string]interface{} {
	must := make([]interface{}, 0)

	if input.HasSearchTerm() {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     *input.SearchTerm,
				"fields":    []string{"name^3", "description^2", "specialty^2", "keywords"},
				"type":      "best_fields",
				"operator":  "or",
				"fuzziness": "AUTO",
			},
		})
	}

	if input.HasFilters() {
		for _, filter := range input.Filters {
			must = append(must, r.buildFilterQuery(filter))
		}
	}

	if len(must) == 0 {
		return map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must": must,
		},
	}
}

func (r *Repository) buildFilterQuery(filter searchinput.Filter) map[string]interface{} {
	switch filter.Field {
	case searchinput.FieldKeywords:
		return map[string]interface{}{
			"term": map[string]interface{}{
				"keywords": filter.Value,
			},
		}
	case searchinput.FieldSpecialty:
		return map[string]interface{}{
			"match": map[string]interface{}{
				"specialty": map[string]interface{}{
					"query":    filter.Value,
					"operator": "and",
				},
			},
		}
	case searchinput.FieldName:
		return map[string]interface{}{
			"match": map[string]interface{}{
				"name": map[string]interface{}{
					"query":    filter.Value,
					"operator": "and",
				},
			},
		}
	case searchinput.FieldDescription:
		return map[string]interface{}{
			"match": map[string]interface{}{
				"description": map[string]interface{}{
					"query":    filter.Value,
					"operator": "and",
				},
			},
		}
	default:
		return map[string]interface{}{
			"match": map[string]interface{}{
				string(filter.Field): filter.Value,
			},
		}
	}
}

func (r *Repository) buildSort(input *searchinput.ListSearchInput) []interface{} {
	sort := make([]interface{}, 0)

	// change to a domain rule
	if input.HasSort() {
		for _, s := range input.Sort {
			fieldName := r.mapSortFieldToES(s.Field)
			sort = append(sort, map[string]interface{}{
				fieldName: map[string]interface{}{
					"order": string(s.Order),
				},
			})
		}
	} else {
		sort = append(sort, map[string]interface{}{
			"created_at": map[string]interface{}{
				"order": "desc",
			},
		})
	}

	sort = append(sort, map[string]interface{}{
		"id": map[string]interface{}{
			"order": "asc",
		},
	})

	return sort
}

func (r *Repository) buildSearchAfter(input *searchinput.ListSearchInput) []interface{} {
	if input.Pagination.IsFirstPage() {
		return nil
	}

	decoded, err := input.Pagination.DecodeCursor()
	if err != nil {
		return nil
	}

	if decoded == nil {
		return nil
	}

	if decoded.SortField == "created_at" || decoded.SortField == "updated_at" {
		return []interface{}{decoded.SortValue, decoded.ID}
	}

	return []interface{}{decoded.SortValue, decoded.ID}
}
