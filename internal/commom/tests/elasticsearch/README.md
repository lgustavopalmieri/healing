# Elasticsearch Test Utilities

Test utilities for Elasticsearch integration tests using testcontainers.

## Usage

### TestMain Pattern

```go
var testHelper *elasticsearch.TestHelper

func TestMain(m *testing.M) {
	testHelper = elasticsearch.NewTestHelper()
	testHelper.RunTestMain(m)
}

func TestRepository_Search(t *testing.T) {
	client, indexName, cleanup := testHelper.SetupTestIndex(t, indexes.CreateSpecialistsIndex)
	defer cleanup()

	// Your test code here
}
```

### Factory Functions

```go
// Use predefined specialists
specialists := elasticsearch.GetPredefinedSpecialists()
elasticsearch.IndexSpecialists(t, ctx, client, indexName, specialists)

// Use factory with overrides
doc := elasticsearch.SpecialistDocumentFactory(func(d *elasticsearch.SpecialistDocument) {
	d.Name = "Custom Name"
	d.Specialty = "Custom Specialty"
})
```

## Container Configuration

- Image: `docker.elastic.co/elasticsearch/elasticsearch:8.17.0`
- Single node mode
- Security disabled for testing
- Memory: 512MB (Xms and Xmx)
