package observability

import "context"

type Metrics interface {
	Counter(name string) Counter
	Histogram(name string) Histogram
	Gauge(name string) Gauge
}

type Counter interface {
	Add(ctx context.Context, value float64, labels ...Label)
}

type Histogram interface {
	Record(ctx context.Context, value float64, labels ...Label)
}

type Gauge interface {
	Set(ctx context.Context, value float64, labels ...Label)
}

type Label struct {
	Key   string
	Value string
}