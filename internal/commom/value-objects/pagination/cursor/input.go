package cursor

type CursorPaginationInput struct {
	EncodedCursor *string
	PageSize      int
	Direction     PaginationDirection
}

type PaginationDirection string

const (
	DirectionNext     PaginationDirection = "next"
	DirectionPrevious PaginationDirection = "previous"
)

func NewCursorPaginationInput(
	encodedCursor *string,
	pageSize int,
	direction PaginationDirection,
) (*CursorPaginationInput, error) {
	input := &CursorPaginationInput{
		EncodedCursor: encodedCursor,
		PageSize:      pageSize,
		Direction:     direction,
	}

	if err := input.validate(); err != nil {
		return nil, err
	}

	return input, nil
}

func (c *CursorPaginationInput) IsFirstPage() bool {
	return c.EncodedCursor == nil || *c.EncodedCursor == ""
}

func (c *CursorPaginationInput) IsNavigatingForward() bool {
	return c.Direction == DirectionNext
}

func (c *CursorPaginationInput) IsNavigatingBackward() bool {
	return c.Direction == DirectionPrevious
}
