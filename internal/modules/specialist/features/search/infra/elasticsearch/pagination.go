package elasticsearch

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	searchinput "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/search/search_input"
)

func (r *Repository) buildCursorOutput(
	input *searchinput.ListSearchInput,
	specialists []*domain.Specialist,
	hits []elasticsearchHit,
) *cursor.CursorPaginationOutput {
	pageSize := input.Pagination.PageSize
	hasNext := len(specialists) > pageSize

	if hasNext {
		specialists = specialists[:pageSize]
		hits = hits[:pageSize]
	}

	var nextCursor *string
	if hasNext && len(hits) > 0 {
		lastHit := hits[len(hits)-1]
		encoded := r.encodeCursorFromHit(lastHit)
		nextCursor = &encoded
	}

	var prevCursor *string
	if !input.Pagination.IsFirstPage() && len(hits) > 0 {
		firstHit := hits[0]
		encoded := r.encodeCursorFromHit(firstHit)
		prevCursor = &encoded
	}

	return cursor.NewCursorPaginationOutput(
		nextCursor,
		prevCursor,
		hasNext,
		!input.Pagination.IsFirstPage(),
		len(specialists),
	)
}

func (r *Repository) encodeCursorFromHit(hit elasticsearchHit) string {
	if len(hit.Sort) == 0 {
		return ""
	}

	sortValues := make([]interface{}, len(hit.Sort))
	copy(sortValues, hit.Sort)

	return cursor.EncodeCursorMultiSort(sortValues)
}
