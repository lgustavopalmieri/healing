package cursor

// Este arquivo contém exemplos práticos de como usar o Value Object
// de paginação por cursor em diferentes cenários.
//
// IMPORTANTE: Este arquivo é apenas para documentação e exemplos.
// Não deve ser usado em produção.

import (
	"fmt"
	"time"
)

// ExampleItem representa um item genérico que será paginado.
// Em um caso real, seria sua entidade de domínio (Specialist, Product, etc.)
type ExampleItem struct {
	ID        string
	Name      string
	CreatedAt time.Time
	Score     int
}

// ExampleScenario1_FirstPageRequest demonstra como fazer a primeira requisição
// (sem cursor) para buscar a primeira página de resultados.
func ExampleScenario1_FirstPageRequest() {
	fmt.Println("=== Cenário 1: Primeira Requisição (Sem Cursor) ===")

	// Cliente quer buscar os primeiros 20 itens
	input, err := NewCursorPaginationInput(
		nil,           // sem cursor (primeira página)
		20,            // 20 itens por página
		DirectionNext, // navegando para frente
	)

	if err != nil {
		fmt.Printf("Erro ao criar input: %v\n", err)
		return
	}

	fmt.Printf("Input criado com sucesso:\n")
	fmt.Printf("  - É primeira página? %v\n", input.IsFirstPage())
	fmt.Printf("  - Page size: %d\n", input.PageSize())
	fmt.Printf("  - Direção: %s\n", input.Direction())

	// Agora você usaria este input no seu repository/service:
	// items, output := repository.FindItems(ctx, input)
}

// ExampleScenario2_NextPageRequest demonstra como buscar a próxima página
// usando o cursor retornado da requisição anterior.
func ExampleScenario2_NextPageRequest() {
	fmt.Println("\n=== Cenário 2: Próxima Página (Com Cursor) ===")

	// Simula o cursor recebido da resposta anterior
	// Em produção, este cursor viria do CursorPaginationOutput.NextCursor()
	previousCursor := EncodeCursor("123", 1632489600, "created_at")

	input, err := NewCursorPaginationInput(
		&previousCursor, // cursor da página anterior
		20,              // mesmo page size
		DirectionNext,   // continua navegando para frente
	)

	if err != nil {
		fmt.Printf("Erro ao criar input: %v\n", err)
		return
	}

	fmt.Printf("Input criado com sucesso:\n")
	fmt.Printf("  - É primeira página? %v\n", input.IsFirstPage())
	fmt.Printf("  - Tem cursor? %v\n", input.EncodedCursor() != nil)
	fmt.Printf("  - Navegando para frente? %v\n", input.IsNavigatingForward())

	// Decodifica o cursor para usar na query
	decoded, err := input.DecodeCursor()
	if err != nil {
		fmt.Printf("Erro ao decodificar cursor: %v\n", err)
		return
	}

	fmt.Printf("  - Cursor decodificado:\n")
	fmt.Printf("    - Campo de ordenação: %s\n", decoded.SortField)
	fmt.Printf("    - Valor de ordenação: %s\n", decoded.SortValue)
	fmt.Printf("    - ID: %s\n", decoded.ID)

	// No repository, você usaria estes valores para construir a query:
	// SELECT * FROM items
	// WHERE created_at < ? OR (created_at = ? AND id < ?)
	// ORDER BY created_at DESC, id DESC
	// LIMIT 21  -- pageSize + 1 para detectar se há próxima página
}

// ExampleScenario3_PreviousPageRequest demonstra como voltar para a página anterior.
func ExampleScenario3_PreviousPageRequest() {
	fmt.Println("\n=== Cenário 3: Página Anterior (Navegação Reversa) ===")

	// Cursor da página atual (para voltar)
	currentCursor := EncodeCursor("100", 1632489000, "created_at")

	input, err := NewCursorPaginationInput(
		&currentCursor,    // cursor da página atual
		20,                // mesmo page size
		DirectionPrevious, // navegando para trás
	)

	if err != nil {
		fmt.Printf("Erro ao criar input: %v\n", err)
		return
	}

	fmt.Printf("Input criado com sucesso:\n")
	fmt.Printf("  - Navegando para trás? %v\n", input.IsNavigatingBackward())

	// No repository, a query seria invertida:
	// SELECT * FROM items
	// WHERE created_at > ? OR (created_at = ? AND id > ?)
	// ORDER BY created_at ASC, id ASC  -- ordem invertida!
	// LIMIT 21
	// Depois inverter os resultados antes de retornar
}

// ExampleScenario4_BuildingOutput demonstra como construir o output
// após buscar os dados no repository.
func ExampleScenario4_BuildingOutput() {
	fmt.Println("\n=== Cenário 4: Construindo Output no Repository ===")

	// Simula que buscamos 21 itens (pageSize + 1) para detectar próxima página
	pageSize := 20
	items := []ExampleItem{
		{ID: "1", Name: "Item 1", CreatedAt: time.Now(), Score: 100},
		{ID: "2", Name: "Item 2", CreatedAt: time.Now(), Score: 95},
		// ... mais 18 itens ...
		{ID: "21", Name: "Item 21", CreatedAt: time.Now(), Score: 50}, // item extra
	}

	// Detecta se há próxima página
	hasNextPage := len(items) > pageSize
	if hasNextPage {
		items = items[:pageSize] // remove o item extra
	}

	// Gera cursor para próxima página (baseado no último item)
	var nextCursor *string
	if hasNextPage {
		lastItem := items[len(items)-1]
		encoded := EncodeCursor(
			lastItem.ID,
			lastItem.CreatedAt.Unix(),
			"created_at",
		)
		nextCursor = &encoded
	}

	// Gera cursor para página anterior (baseado no primeiro item)
	// Assumindo que não é a primeira página
	firstItem := items[0]
	prevEncoded := EncodeCursor(
		firstItem.ID,
		firstItem.CreatedAt.Unix(),
		"created_at",
	)
	previousCursor := &prevEncoded

	// Constrói o output
	output := NewCursorPaginationOutput(
		nextCursor,
		previousCursor,
		hasNextPage,
		true, // hasPreviousPage (não é primeira página)
		len(items),
	)

	fmt.Printf("Output construído:\n")
	fmt.Printf("  - Total de itens: %d\n", output.TotalItemsInPage())
	fmt.Printf("  - Tem próxima página? %v\n", output.HasNextPage())
	fmt.Printf("  - Tem página anterior? %v\n", output.HasPreviousPage())
	fmt.Printf("  - É página vazia? %v\n", output.IsEmpty())
	fmt.Printf("  - É página parcial? %v\n", output.IsPartialPage(pageSize))

	if output.HasNextPage() {
		fmt.Printf("  - Próximo cursor disponível: %v\n", output.NextCursor() != nil)
	}
}

// ExampleScenario5_UsingBuilder demonstra o uso do Builder para criar output.
func ExampleScenario5_UsingBuilder() {
	fmt.Println("\n=== Cenário 5: Usando Builder (Mais Legível) ===")

	nextCursor := EncodeCursor("50", 1632489500, "created_at")

	output := NewCursorBuilder().
		WithNextCursor(&nextCursor).
		WithPreviousCursor(nil).
		WithHasNextPage(true).
		WithHasPreviousPage(false).
		WithTotalItems(20).
		Build()

	fmt.Printf("Output construído com builder:\n")
	fmt.Printf("  - Total de itens: %d\n", output.TotalItemsInPage())
	fmt.Printf("  - Tem próxima página? %v\n", output.HasNextPage())
}

// ExampleScenario6_ErrorHandling demonstra tratamento de erros de validação.
func ExampleScenario6_ErrorHandling() {
	fmt.Println("\n=== Cenário 6: Tratamento de Erros ===")

	// Erro: page size inválido
	_, err := NewCursorPaginationInput(nil, 0, DirectionNext)
	if err != nil {
		fmt.Printf("Erro esperado (page size = 0): %v\n", err)
		fmt.Printf("É erro de domínio? %v\n", IsDomainError(err))
	}

	// Erro: page size muito grande
	_, err = NewCursorPaginationInput(nil, 1000, DirectionNext)
	if err != nil {
		fmt.Printf("Erro esperado (page size = 1000): %v\n", err)
	}

	// Erro: direção inválida
	_, err = NewCursorPaginationInput(nil, 20, "invalid")
	if err != nil {
		fmt.Printf("Erro esperado (direção inválida): %v\n", err)
	}

	// Erro: cursor inválido
	invalidCursor := "not-base64!!!"
	_, err = NewCursorPaginationInput(&invalidCursor, 20, DirectionNext)
	if err != nil {
		fmt.Printf("Erro esperado (cursor inválido): %v\n", err)
	}
}

// ExampleScenario7_CompleteFlow demonstra um fluxo completo de paginação.
func ExampleScenario7_CompleteFlow() {
	fmt.Println("\n=== Cenário 7: Fluxo Completo de Paginação ===")

	// 1. Cliente faz primeira requisição
	fmt.Println("\n1. Primeira requisição:")
	// input1, _ := NewCursorPaginationInput(nil, 10, DirectionNext)
	fmt.Printf("   - Buscando primeira página (10 itens)\n")

	// 2. Sistema retorna primeira página com cursor para próxima
	nextCursor1 := EncodeCursor("10", 1632489000, "created_at")
	output1 := NewCursorPaginationOutput(&nextCursor1, nil, true, false, 10)
	fmt.Printf("   - Retornou 10 itens\n")
	fmt.Printf("   - Tem próxima página? %v\n", output1.HasNextPage())

	// 3. Cliente usa o cursor para buscar próxima página
	fmt.Println("\n2. Segunda requisição (usando cursor):")
	// input2, _ := NewCursorPaginationInput(output1.NextCursor(), 10, DirectionNext)
	fmt.Printf("   - Buscando próxima página com cursor\n")

	// 4. Sistema retorna segunda página
	nextCursor2 := EncodeCursor("20", 1632488000, "created_at")
	prevCursor2 := EncodeCursor("11", 1632488900, "created_at")
	output2 := NewCursorPaginationOutput(&nextCursor2, &prevCursor2, true, true, 10)
	fmt.Printf("   - Retornou 10 itens\n")
	fmt.Printf("   - Tem próxima página? %v\n", output2.HasNextPage())
	fmt.Printf("   - Tem página anterior? %v\n", output2.HasPreviousPage())

	// 5. Cliente decide voltar para página anterior
	fmt.Println("\n3. Terceira requisição (voltando):")
	input3, _ := NewCursorPaginationInput(output2.PreviousCursor(), 10, DirectionPrevious)
	fmt.Printf("   - Voltando para página anterior\n")
	fmt.Printf("   - Direção: %s\n", input3.Direction())

	// 6. Sistema retorna página anterior (mesma que output1)
	output3 := NewCursorPaginationOutput(&nextCursor1, nil, true, false, 10)
	fmt.Printf("   - Retornou à primeira página\n")
	fmt.Printf("   - Tem página anterior? %v\n", output3.HasPreviousPage())
}

// ExampleScenario8_DifferentSortFields demonstra paginação com diferentes campos de ordenação.
func ExampleScenario8_DifferentSortFields() {
	fmt.Println("\n=== Cenário 8: Diferentes Campos de Ordenação ===")

	// Ordenação por timestamp (mais comum)
	fmt.Println("\n1. Ordenação por created_at:")
	cursor1 := EncodeCursor("123", 1632489600, "created_at")
	fmt.Printf("   Cursor: %s\n", cursor1)

	// Ordenação por score (ranking)
	fmt.Println("\n2. Ordenação por score:")
	cursor2 := EncodeCursor("456", 95, "score")
	fmt.Printf("   Cursor: %s\n", cursor2)

	// Ordenação por nome (alfabética)
	fmt.Println("\n3. Ordenação por name:")
	cursor3 := EncodeCursor("789", "John Doe", "name")
	fmt.Printf("   Cursor: %s\n", cursor3)

	// Decodificando para ver o conteúdo
	input, _ := NewCursorPaginationInput(&cursor3, 20, DirectionNext)
	decoded, _ := input.DecodeCursor()
	fmt.Printf("\n   Cursor decodificado:\n")
	fmt.Printf("   - Campo: %s\n", decoded.SortField)
	fmt.Printf("   - Valor: %s\n", decoded.SortValue)
	fmt.Printf("   - ID: %s\n", decoded.ID)
}
