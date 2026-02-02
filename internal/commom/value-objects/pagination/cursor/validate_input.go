package cursor

import "encoding/base64"

// validate executa todas as validações de domínio para o input de paginação.
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

// validatePageSize garante que o tamanho da página é válido.
func (c *CursorPaginationInput) validatePageSize() error {
	if c.PageSize <= 0 {
		return ErrInvalidPageSize
	}

	// Opcional: você pode adicionar um limite máximo para evitar sobrecarga
	const maxPageSize = 100
	if c.PageSize > maxPageSize {
		return ErrPageSizeTooLarge
	}

	return nil
}

// validateDirection garante que a direção é válida.
func (c *CursorPaginationInput) validateDirection() error {
	if c.Direction != DirectionNext && c.Direction != DirectionPrevious {
		return ErrInvalidDirection
	}
	return nil
}

// validateCursor valida o formato do cursor, se fornecido.
func (c *CursorPaginationInput) validateCursor() error {
	// Cursor nil ou vazio é válido (indica primeira página)
	if c.EncodedCursor == nil || *c.EncodedCursor == "" {
		return nil
	}

	// Valida se o cursor está em formato base64 válido
	// (isso é uma validação básica - a validação completa acontece ao decodificar)
	_, err := base64.StdEncoding.DecodeString(*c.EncodedCursor)
	if err != nil {
		return ErrInvalidCursorFormat
	}

	return nil
}

