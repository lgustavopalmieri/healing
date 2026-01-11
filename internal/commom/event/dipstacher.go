package event

import "context"

type EventDispatcher interface {
	Dispatch(ctx context.Context, event Event) error
}
