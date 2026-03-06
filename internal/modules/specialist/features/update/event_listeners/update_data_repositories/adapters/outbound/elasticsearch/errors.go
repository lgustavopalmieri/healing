package elasticsearch

var (
	FailedToSerializeErr  = "failed to serialize specialist for indexing: %w"
	FailedToIndexErr      = "failed to index specialist: %w"
	IndexErrorResponseErr = "elasticsearch index error: %s"
	FailedToPublishDLQErr = "failed to publish elasticsearch DLQ event: %w"
)

const (
	ElasticsearchUpdateDLQEventName = "specialist.updated.elasticsearch.dlq"
)
