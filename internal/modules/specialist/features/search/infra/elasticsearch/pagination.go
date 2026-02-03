package elasticsearch

import (
	"fmt"
	"time"

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
		encoded := r.encodeCursorFromHit(lastHit, input)
		nextCursor = &encoded
	}

	var prevCursor *string
	if !input.Pagination.IsFirstPage() && len(hits) > 0 {
		firstHit := hits[0]
		encoded := r.encodeCursorFromHit(firstHit, input)
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

func (r *Repository) encodeCursorFromHit(hit elasticsearchHit, input *searchinput.ListSearchInput) string {
	if len(hit.Sort) < 2 {
		return cursor.EncodeCursor(hit.Source.ID, hit.Source.CreatedAt.Format(time.RFC3339Nano), "created_at")
	}

	sortField := "created_at"
	if input.HasSort() && len(input.Sort) > 0 {
		sortField = string(input.Sort[0].Field)
	}

	sortValue := fmt.Sprintf("%v", hit.Sort[0])
	id := hit.Source.ID

	return cursor.EncodeCursor(id, sortValue, sortField)
}
