package elasticsearch

import "errors"

var (
	ErrSearchFailed     = errors.New("elasticsearch search failed")
	ErrInvalidQuery     = errors.New("invalid search query")
	ErrConnectionFailed = errors.New("elasticsearch connection failed")
	ErrIndexNotFound    = errors.New("index not found")
	ErrInvalidCursor    = errors.New("invalid cursor format")
	ErrEncodingFailed   = errors.New("failed to encode query")
	ErrDecodingFailed   = errors.New("failed to decode response")
)
