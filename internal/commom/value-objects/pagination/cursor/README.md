# Paginação por Cursor - Guia Completo

## 📚 O que é Paginação por Cursor?

Paginação por cursor é uma técnica de paginação onde, ao invés de usar números de página (1, 2, 3...) ou offset/limit, usamos um **"ponteiro"** (cursor) que aponta para uma posição específica no conjunto de dados.

### Analogia do Mundo Real

Imagine que você está lendo um livro:

- **Paginação Offset**: "Vá para a página 50" → Você precisa contar 50 páginas desde o início
- **Paginação Cursor**: "Continue de onde parou (marcador na página 50)" → Você abre direto no marcador

O cursor é como um **marcador de livro** que guarda exatamente onde você parou.

## 🎯 Por que usar Cursor ao invés de Offset?

### Problema com Offset/Limit

```
Página 1: SELECT * FROM items ORDER BY created_at OFFSET 0 LIMIT 10
Página 2: SELECT * FROM items ORDER BY created_at OFFSET 10 LIMIT 10
```

**Problemas:**

1. **Duplicatas**: Se um novo item for inserido entre as requisições, você pode ver o mesmo item duas vezes
2. **Performance**: `OFFSET 10000` força o banco a ler e descartar 10.000 registros
3. **Itens pulados**: Se itens forem deletados, você pode pular registros

### Solução com Cursor

```
Página 1: SELECT * FROM items WHERE id > 0 ORDER BY created_at LIMIT 10
Página 2: SELECT * FROM items WHERE created_at < '2024-01-01' ORDER BY created_at LIMIT 10
```

**Vantagens:**

1. ✅ **Sem duplicatas**: Sempre busca a partir de um ponto específico
2. ✅ **Performance**: Usa índices eficientemente (WHERE ao invés de OFFSET)
3. ✅ **Consistência**: Não importa se dados foram inseridos/deletados

## 🏗️ Arquitetura do Value Object

### Componentes Principais

```
┌─────────────────────────────────────────────────────────────┐
│                  CursorPaginationInput                      │
│  (O que o cliente envia)                                    │
├─────────────────────────────────────────────────────────────┤
│  - encodedCursor: *string    // "eyJpZCI6MTIzfQ=="         │
│  - pageSize: int             // 20                          │
│  - direction: Direction      // "next" ou "previous"        │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                    [Repository/Service]
                    Busca dados no banco
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                 CursorPaginationOutput                      │
│  (O que o sistema retorna)                                  │
├─────────────────────────────────────────────────────────────┤
│  - nextCursor: *string       // Cursor para próxima página  │
│  - previousCursor: *string   // Cursor para página anterior │
│  - hasNextPage: bool         // Existe próxima página?      │
│  - hasPreviousPage: bool     // Existe página anterior?     │
│  - totalItemsInPage: int     // Quantos itens retornados    │
└─────────────────────────────────────────────────────────────┘
```

## 📖 Como Usar - Passo a Passo

### 1️⃣ Primeira Requisição (Sem Cursor)

```go
// Cliente quer buscar os primeiros 20 itens
input, err := cursor.NewCursorPaginationInput(
    nil,                      // sem cursor (primeira página)
    20,                       // 20 itens por página
    cursor.DirectionNext,     // navegando para frente
)

if err != nil {
    // Tratar erro de validação
    return err
}

// Usar no repository
items, output := repository.FindItems(ctx, input)
```

### 2️⃣ Próxima Página (Com Cursor)

```go
// Cliente recebeu o output da primeira requisição
// e quer buscar a próxima página

if output.HasNextPage() {
    nextCursor := output.NextCursor()
    
    input, err := cursor.NewCursorPaginationInput(
        nextCursor,               // cursor da página anterior
        20,                       // mesmo page size
        cursor.DirectionNext,     // continua para frente
    )
    
    items, output := repository.FindItems(ctx, input)
}
```

### 3️⃣ Página Anterior (Navegação Reversa)

```go
// Cliente quer voltar para a página anterior

if output.HasPreviousPage() {
    prevCursor := output.PreviousCursor()
    
    input, err := cursor.NewCursorPaginationInput(
        prevCursor,                  // cursor da página atual
        20,                          // mesmo page size
        cursor.DirectionPrevious,    // navegando para trás
    )
    
    items, output := repository.FindItems(ctx, input)
}
```

## 🔧 Implementação no Repository

### Estrutura Básica

```go
func (r *Repository) FindItems(
    ctx context.Context,
    input *cursor.CursorPaginationInput,
) ([]Item, *cursor.CursorPaginationOutput, error) {
    
    // 1. Decodificar cursor (se existir)
    var lastID string
    var lastCreatedAt time.Time
    
    if !input.IsFirstPage() {
        decoded, err := input.DecodeCursor()
        if err != nil {
            return nil, nil, err
        }
        
        lastID = decoded.ID
        timestamp, _ := cursor.ParseIntFromCursor(decoded)
        lastCreatedAt = time.Unix(timestamp, 0)
    }
    
    // 2. Construir query baseada na direção
    query := r.buildQuery(input, lastID, lastCreatedAt)
    
    // 3. Buscar N+1 itens (para detectar se há próxima página)
    items, err := r.executeQuery(ctx, query, input.PageSize()+1)
    if err != nil {
        return nil, nil, err
    }
    
    // 4. Detectar se há próxima página
    hasNextPage := len(items) > input.PageSize()
    if hasNextPage {
        items = items[:input.PageSize()] // remover item extra
    }
    
    // 5. Gerar cursores
    var nextCursor *string
    if hasNextPage {
        lastItem := items[len(items)-1]
        encoded := cursor.EncodeCursor(
            lastItem.ID,
            lastItem.CreatedAt.Unix(),
            "created_at",
        )
        nextCursor = &encoded
    }
    
    var prevCursor *string
    if !input.IsFirstPage() {
        firstItem := items[0]
        encoded := cursor.EncodeCursor(
            firstItem.ID,
            firstItem.CreatedAt.Unix(),
            "created_at",
        )
        prevCursor = &encoded
    }
    
    // 6. Construir output
    output := cursor.NewCursorPaginationOutput(
        nextCursor,
        prevCursor,
        hasNextPage,
        !input.IsFirstPage(),
        len(items),
    )
    
    return items, output, nil
}
```

### Query SQL para Direção "Next"

```sql
-- Primeira página (sem cursor)
SELECT * FROM items
ORDER BY created_at DESC, id DESC
LIMIT 21;  -- pageSize + 1

-- Páginas seguintes (com cursor)
SELECT * FROM items
WHERE 
    created_at < ? OR 
    (created_at = ? AND id < ?)
ORDER BY created_at DESC, id DESC
LIMIT 21;
```

### Query SQL para Direção "Previous"

```sql
-- Navegação reversa
SELECT * FROM items
WHERE 
    created_at > ? OR 
    (created_at = ? AND id > ?)
ORDER BY created_at ASC, id ASC  -- ordem invertida!
LIMIT 21;

-- Depois inverter os resultados antes de retornar ao cliente
```

## 🎨 Estrutura do Cursor

### Formato Interno

O cursor é uma string base64 que contém:

```
Formato: base64("sortField:sortValue:id")
Exemplo: base64("created_at:1632489600:123")
Resultado: "Y3JlYXRlZF9hdDoxNjMyNDg5NjAwOjEyMw=="
```

### Por que este formato?

1. **sortField**: Indica qual campo está sendo usado para ordenação
2. **sortValue**: O valor daquele campo no último item visualizado
3. **id**: ID único para desempate (garante ordenação consistente)

### Exemplo Prático

```go
// Item: {ID: "123", CreatedAt: 2021-09-24 10:00:00, Name: "John"}

// Gerar cursor
cursor := cursor.EncodeCursor(
    "123",                    // ID
    1632489600,               // CreatedAt.Unix()
    "created_at",             // campo de ordenação
)
// Resultado: "Y3JlYXRlZF9hdDoxNjMyNDg5NjAwOjEyMw=="

// Decodificar cursor
input, _ := cursor.NewCursorPaginationInput(&cursor, 20, cursor.DirectionNext)
decoded, _ := input.DecodeCursor()

// decoded.SortField = "created_at"
// decoded.SortValue = "1632489600"
// decoded.ID = "123"
```

## 🚨 Tratamento de Erros

### Erros de Domínio

```go
// Page size inválido
_, err := cursor.NewCursorPaginationInput(nil, 0, cursor.DirectionNext)
// err = ErrInvalidPageSize

// Page size muito grande
_, err := cursor.NewCursorPaginationInput(nil, 1000, cursor.DirectionNext)
// err = ErrPageSizeTooLarge

// Direção inválida
_, err := cursor.NewCursorPaginationInput(nil, 20, "invalid")
// err = ErrInvalidDirection

// Cursor corrompido
invalidCursor := "not-base64!!!"
_, err := cursor.NewCursorPaginationInput(&invalidCursor, 20, cursor.DirectionNext)
// err = ErrInvalidCursorFormat
```

### Como Tratar

```go
input, err := cursor.NewCursorPaginationInput(cursor, pageSize, direction)
if err != nil {
    if cursor.IsDomainError(err) {
        // Erro de validação do cliente
        // Retornar 400 Bad Request
        return BadRequestResponse(err.Error())
    }
    
    // Erro técnico
    // Retornar 500 Internal Server Error
    return InternalErrorResponse()
}
```

## 🎯 Casos de Uso Comuns

### 1. Feed Infinito (Scroll Infinito)

```go
// Cliente carrega mais itens ao rolar para baixo
func LoadMoreItems(currentCursor *string) {
    input, _ := cursor.NewCursorPaginationInput(
        currentCursor,
        20,
        cursor.DirectionNext,
    )
    
    items, output := api.GetItems(input)
    
    // Adicionar itens à lista
    appendToList(items)
    
    // Guardar cursor para próxima requisição
    if output.HasNextPage() {
        saveNextCursor(output.NextCursor())
    }
}
```

### 2. Navegação com Botões (Anterior/Próximo)

```go
// UI com botões "Anterior" e "Próximo"
type PaginationState struct {
    CurrentCursor *string
    PrevCursor    *string
    NextCursor    *string
}

func (s *PaginationState) GoNext() {
    if s.NextCursor != nil {
        input, _ := cursor.NewCursorPaginationInput(
            s.NextCursor,
            20,
            cursor.DirectionNext,
        )
        
        items, output := api.GetItems(input)
        
        // Atualizar estado
        s.CurrentCursor = s.NextCursor
        s.PrevCursor = output.PreviousCursor()
        s.NextCursor = output.NextCursor()
        
        renderItems(items)
    }
}

func (s *PaginationState) GoPrevious() {
    if s.PrevCursor != nil {
        input, _ := cursor.NewCursorPaginationInput(
            s.PrevCursor,
            20,
            cursor.DirectionPrevious,
        )
        
        items, output := api.GetItems(input)
        
        // Atualizar estado
        s.CurrentCursor = s.PrevCursor
        s.PrevCursor = output.PreviousCursor()
        s.NextCursor = output.NextCursor()
        
        renderItems(items)
    }
}
```

### 3. Diferentes Ordenações

```go
// Ordenar por data de criação (mais recente primeiro)
cursor1 := cursor.EncodeCursor(item.ID, item.CreatedAt.Unix(), "created_at")

// Ordenar por score (maior primeiro)
cursor2 := cursor.EncodeCursor(item.ID, item.Score, "score")

// Ordenar por nome (alfabética)
cursor3 := cursor.EncodeCursor(item.ID, item.Name, "name")
```

## ✅ Boas Práticas

### 1. Sempre busque N+1 itens

```go
// Buscar pageSize + 1 para detectar se há próxima página
items := repository.Find(cursor, pageSize + 1)

hasNextPage := len(items) > pageSize
if hasNextPage {
    items = items[:pageSize] // remover item extra
}
```

### 2. Use índices compostos no banco

```sql
-- Para ordenação por created_at + id
CREATE INDEX idx_items_created_at_id ON items(created_at DESC, id DESC);
```

### 3. Cursor é opaco para o cliente

```go
// ❌ NUNCA faça isso no cliente
cursor := base64.decode(cursorString)
id := cursor.split(":")[2]

// ✅ Sempre trate cursor como string opaca
nextInput := cursor.NewCursorPaginationInput(
    output.NextCursor(), // apenas passe adiante
    20,
    cursor.DirectionNext,
)
```

### 4. Valide sempre o input

```go
// O construtor já valida automaticamente
input, err := cursor.NewCursorPaginationInput(cursor, pageSize, direction)
if err != nil {
    // Tratar erro de validação
    return err
}
```

## 🔍 Comparação: Offset vs Cursor

| Aspecto | Offset/Limit | Cursor |
|---------|--------------|--------|
| **Performance** | ❌ Lenta em grandes offsets | ✅ Sempre rápida (usa índices) |
| **Consistência** | ❌ Duplicatas/pulos possíveis | ✅ Sempre consistente |
| **Complexidade** | ✅ Simples de implementar | ⚠️ Mais complexa |
| **Pular páginas** | ✅ Pode ir direto para página N | ❌ Só pode navegar sequencialmente |
| **Total de páginas** | ✅ Pode calcular total | ❌ Não sabe total de páginas |
| **Uso ideal** | Tabelas pequenas, paginação tradicional | Feeds, APIs, grandes datasets |

## 📚 Referências e Leitura Adicional

- [Relay Cursor Connections Specification](https://relay.dev/graphql/connections.htm)
- [Pagination Best Practices](https://www.citusdata.com/blog/2016/03/30/five-ways-to-paginate/)
- [Why Cursor Pagination?](https://slack.engineering/evolving-api-pagination-at-slack/)

## 🎓 Resumo

**Paginação por Cursor** é ideal quando você precisa de:
- ✅ Performance consistente em grandes datasets
- ✅ Navegação sem duplicatas ou pulos
- ✅ Feeds em tempo real (scroll infinito)
- ✅ APIs públicas com alto volume

**Use Offset/Limit** quando você precisa de:
- ✅ Pular para páginas específicas (página 5, 10, etc.)
- ✅ Mostrar "Página X de Y"
- ✅ Datasets pequenos e estáveis
- ✅ Implementação mais simples

---

**Dúvidas?** Leia os exemplos em `example_usage.go` para ver casos práticos de uso!
