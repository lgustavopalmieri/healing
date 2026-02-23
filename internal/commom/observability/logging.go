package observability

import "context"

//go:generate mockgen -source=logging.go -destination=mocks/logger_mock.go -package=mocks
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
}

type Field struct {
	Key   string
	Value string
}
