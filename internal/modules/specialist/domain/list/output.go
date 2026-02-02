package list

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/value-objects/pagination/cursor"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

// ListSearchOutput é um Value Object que representa o resultado de uma
// busca avançada de especialistas com paginação por cursor.
//
// Este Value Object encapsula tanto os dados retornados quanto os metadados
// de paginação necessários para navegação.
//
// Responsabilidades:
//   - Encapsular lista de especialistas encontrados
//   - Fornecer metadados de paginação (cursores, flags de navegação)
//   - Fornecer métodos de conveniência para verificar estado do resultado
//
// Exemplo de uso:
//
//	output := list.NewListSearchOutput(specialists, cursorOutput)
//
//	if output.IsEmpty() {
//	    return "Nenhum especialista encontrado"
//	}
//
//	for _, specialist := range output.Specialists() {
//	    fmt.Printf("Especialista: %s\n", specialist.Name)
//	}
//
//	if output.HasNextPage() {
//	    nextCursor := output.NextCursor()
//	    // Buscar próxima página usando nextCursor
//	}
type ListSearchOutput struct {
	Specialists  []*domain.Specialist
	CursorOutput *cursor.CursorPaginationOutput
}

// NewListSearchOutput cria um novo output de busca avançada.
//
// Parâmetros:
//   - specialists: lista de especialistas encontrados
//   - cursorOutput: metadados de paginação por cursor
//
// Exemplo:
//
//	output := list.NewListSearchOutput(specialists, cursorOutput)
func NewListSearchOutput(
	specialists []*domain.Specialist,
	cursorOutput *cursor.CursorPaginationOutput,
) *ListSearchOutput {
	return &ListSearchOutput{
		Specialists:  specialists,
		CursorOutput: cursorOutput,
	}
}

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
