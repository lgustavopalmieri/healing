package list

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

// ListSearchOutput represents the search result with cursor-based pagination
type ListSearchOutput struct {
	Specialists []*domain.Specialist
	Cursor      *cursor.Output
}
