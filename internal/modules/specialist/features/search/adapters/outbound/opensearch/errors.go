package opensearch

import "errors"

var (
	ErrSearchFailed     = errors.New("opensearch search failed")
	ErrInvalidQuery     = errors.New("invalid search query")
	ErrConnectionFailed = errors.New("opensearch connection failed")
	ErrIndexNotFound    = errors.New("index not found")
	ErrInvalidCursor    = errors.New("invalid cursor format")
	ErrEncodingFailed   = errors.New("failed to encode query")
	ErrDecodingFailed   = errors.New("failed to decode response")
)
