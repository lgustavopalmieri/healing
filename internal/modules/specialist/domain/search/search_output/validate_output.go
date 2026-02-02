package searchoutput

import "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"

func (l *ListSearchOutput) IsEmpty() bool {
	return len(l.Specialists) == 0
}

func (l *ListSearchOutput) Count() int {
	return len(l.Specialists)
}

func (l *ListSearchOutput) HasNextPage() bool {
	if l.CursorOutput == nil {
		return false
	}
	return l.CursorOutput.HasNextPage
}

func (l *ListSearchOutput) HasPreviousPage() bool {
	if l.CursorOutput == nil {
		return false
	}
	return l.CursorOutput.HasPreviousPage
}

func (l *ListSearchOutput) NextCursor() *string {
	if l.CursorOutput == nil {
		return nil
	}
	return l.CursorOutput.NextCursor
}

func (l *ListSearchOutput) PreviousCursor() *string {
	if l.CursorOutput == nil {
		return nil
	}
	return l.CursorOutput.PreviousCursor
}

func (l *ListSearchOutput) FirstSpecialist() *domain.Specialist {
	if len(l.Specialists) == 0 {
		return nil
	}
	return l.Specialists[0]
}

func (l *ListSearchOutput) LastSpecialist() *domain.Specialist {
	if len(l.Specialists) == 0 {
		return nil
	}
	return l.Specialists[len(l.Specialists)-1]
}
