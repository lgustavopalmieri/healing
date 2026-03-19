package cursor

type CursorPaginationOutput struct {
	NextCursor       *string
	PreviousCursor   *string
	HasNextPage      bool
	HasPreviousPage  bool
	TotalItemsInPage int
}

func NewCursorPaginationOutput(
	nextCursor *string,
	previousCursor *string,
	hasNextPage bool,
	hasPreviousPage bool,
	totalItemsInPage int,
) *CursorPaginationOutput {
	return &CursorPaginationOutput{
		NextCursor:       nextCursor,
		PreviousCursor:   previousCursor,
		HasNextPage:      hasNextPage,
		HasPreviousPage:  hasPreviousPage,
		TotalItemsInPage: totalItemsInPage,
	}
}

func (c *CursorPaginationOutput) IsEmpty() bool {
	return c.TotalItemsInPage == 0
}

func (c *CursorPaginationOutput) IsPartialPage(requestedPageSize int) bool {
	return c.TotalItemsInPage < requestedPageSize && c.TotalItemsInPage > 0
}
