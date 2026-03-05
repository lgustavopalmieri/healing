# Elasticsearch Repository - Specialist Search

High-performance Elasticsearch repository implementation for specialist search with cursor-based pagination.

## Features

- **Full-text search** across multiple fields (name, description, specialty, keywords)
- **Advanced filtering** with exact match and full-text support
- **Cursor-based pagination** using Elasticsearch `search_after`
- **Multi-field sorting** with consistent ordering
- **Optimized queries** with proper analyzers and field mappings

## Index Mapping

### Text Fields (Full-text search with analyzer)
- `name`: Standard analyzer + keyword sub-field for sorting
- `description`: Standard analyzer for full-text search
- `specialty`: Standard analyzer + keyword sub-field for exact match filters

### Keyword Fields (Exact match)
- `keywords`: Array of keywords for precise filtering
- `id`, `email`, `phone`, `license_number`: Exact match only

### Date Fields
- `created_at`, `updated_at`: ISO 8601 format for sorting and filtering

## Query Strategy

### Search Term
Uses `multi_match` query with:
- **Fields**: `name^3`, `description^2`, `specialty^2`, `keywords`
- **Type**: `best_fields` for relevance scoring
- **Fuzziness**: `AUTO` for typo tolerance
- **Operator**: `or` for flexible matching

### Filters
- **Keywords**: `term` query for exact match
- **Specialty**: `match` query with `and` operator for full-text
- **Name/Description**: `match` query with `and` operator

### Sorting
- Default: `created_at DESC, id ASC`
- Custom: User-defined fields with ID as tiebreaker
- Text fields use `.keyword` sub-field for sorting

## Cursor Pagination

### How it works
1. Fetch `pageSize + 1` documents to detect if there's a next page
2. Use `search_after` with sort values from last document
3. Encode cursor with: `base64(sortField:sortValue:id)`
4. Decode cursor to extract sort values for next query

### Performance Benefits
- No offset/limit (O(1) vs O(n) complexity)
- Consistent results even with concurrent writes
- Efficient deep pagination

## Usage Example

```go
// Create repository
repo := elasticsearch.NewRepository(esClient, "specialists", logger)

// Build search input
input, _ := searchinput.NewListSearchInput(
    &searchTerm,
    filters,
    sort,
    pagination,
)

// Execute search
output, err := repo.Search(ctx, input)
if err != nil {
    // Handle error
}

// Access results
for _, specialist := range output.Specialists {
    fmt.Printf("Found: %s\n", specialist.Name)
}

// Check pagination
if output.CursorOutput.HasNextPage {
    nextCursor := output.CursorOutput.NextCursor
    // Use nextCursor for next page
}
```

## Performance Considerations

1. **Index Settings**
   - Single shard for small datasets (< 50GB)
   - No replicas in development
   - Standard analyzer with English stopwords

2. **Query Optimization**
   - `track_total_hits: false` to skip count aggregation
   - Fetch only `pageSize + 1` documents
   - Use keyword sub-fields for sorting text fields

3. **Connection Pooling**
   - Reuse Elasticsearch client across requests
   - Configure max retries and backoff strategy
   - Enable metrics for monitoring

## Error Handling

All errors are logged with context and wrapped with descriptive messages:
- Query encoding failures
- Network/connection errors
- Elasticsearch response errors
- Cursor decoding failures

Only errors are logged (no info/debug logs) to reduce noise in production.
