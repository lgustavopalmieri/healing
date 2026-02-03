# Elasticsearch Setup Instructions

## 1. Install Elasticsearch Go Client

Run the following command to add the Elasticsearch dependency:

```bash
go get github.com/elastic/go-elasticsearch/v8@latest
```

## 2. Start Elasticsearch

```bash
# Start Elasticsearch container
make es-up

# Check health
make es-health

# View logs
make es-logs
```

## 3. Run the Application

```bash
# The application will automatically:
# - Connect to Elasticsearch
# - Create the "specialists" index with proper mappings
# - Be ready to handle search requests

go run cmd/grpcserver/main.go
```

## 4. Verify Setup

Check the logs for:
```
🔍 Connecting to Elasticsearch ([http://elasticsearch:9200])...
🔄 Creating Elasticsearch indices...
✅ Elasticsearch connected successfully (Index: specialists)
```

## 5. Elasticsearch Commands

```bash
# Start Elasticsearch
make es-up

# Stop Elasticsearch
make es-down

# View logs
make es-logs

# Check health
make es-health
```

## 6. Environment Variables

Already configured in `.env`:
```env
ELASTICSEARCH_ADDRESSES=http://elasticsearch:9200
ELASTICSEARCH_INDEX_SPECIALISTS=specialists
ELASTICSEARCH_MAX_RETRIES=3
ELASTICSEARCH_RETRY_BACKOFF=1s
```

## 7. Index Mapping

The index is automatically created with:
- **Text fields**: name, description, specialty (with standard analyzer)
- **Keyword fields**: id, email, phone, license_number, keywords
- **Date fields**: created_at, updated_at
- **Boolean**: agreed_to_share

## 8. Testing the Search

Once the application is running, you can test the search functionality through the gRPC service.

## Architecture

```
cmd/grpcserver/
├── bootstrap/
│   └── elasticsearch.go          # Elasticsearch initialization
├── config/
│   ├── config.go                 # Config struct with Elasticsearch
│   ├── load.go                   # Load Elasticsearch config
│   └── validate.go               # Validate Elasticsearch config

internal/
├── platform/
│   └── elasticsearch/
│       ├── client.go             # Elasticsearch client creation
│       └── index.go              # Index creation and mapping
└── modules/specialist/features/search/
    ├── application/
    │   ├── command.go            # Search command (simplified)
    │   ├── new_command.go        # Command constructor
    │   ├── interface.go          # Repository interface
    │   └── constants.go          # Constants and errors
    └── infra/
        └── elasticsearch/
            ├── repository.go     # High-performance implementation
            ├── new.go            # Repository constructor
            ├── errors.go         # Elasticsearch-specific errors
            └── README.md         # Detailed documentation
```

## Performance Features

✅ Cursor-based pagination with `search_after`
✅ Multi-field search with relevance scoring
✅ Optimized queries (no total hits tracking)
✅ Proper field mappings and analyzers
✅ Connection pooling and retry logic
✅ Error logging with context

## Next Steps

After running `go get github.com/elastic/go-elasticsearch/v8@latest`:

1. Start Elasticsearch: `make es-up`
2. Run the application: `go run cmd/grpcserver/main.go`
3. The search repository is ready to use!
