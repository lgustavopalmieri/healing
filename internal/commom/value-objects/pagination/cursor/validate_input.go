package cursor

import "encoding/base64"

func (c *CursorPaginationInput) validate() error {
	if err := c.validatePageSize(); err != nil {
		return err
	}

	if err := c.validateDirection(); err != nil {
		return err
	}

	if err := c.validateCursor(); err != nil {
		return err
	}

	return nil
}

func (c *CursorPaginationInput) validatePageSize() error {
	if c.PageSize <= 0 {
		return ErrInvalidPageSize
	}

	const maxPageSize = 100
	if c.PageSize > maxPageSize {
		return ErrPageSizeTooLarge
	}

	return nil
}

func (c *CursorPaginationInput) validateDirection() error {
	if c.Direction != DirectionNext && c.Direction != DirectionPrevious {
		return ErrInvalidDirection
	}
	return nil
}

func (c *CursorPaginationInput) validateCursor() error {
	if c.EncodedCursor == nil || *c.EncodedCursor == "" {
		return nil
	}

	_, err := base64.StdEncoding.DecodeString(*c.EncodedCursor)
	if err != nil {
		return ErrInvalidCursorFormat
	}

	return nil
}
