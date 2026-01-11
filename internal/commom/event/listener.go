package event

import "context"

type Listener interface {
	Handle(ctx context.Context, event Event) error
}

type ListenerManager struct {
	listeners map[string]Listener
}

func NewListenerManager() *ListenerManager {
	return &ListenerManager{
		listeners: make(map[string]Listener),
	}
}

func (lm *ListenerManager) Register(topic string, listener Listener) {
	lm.listeners[topic] = listener
}

func (lm *ListenerManager) Topics() []string {
	topics := make([]string, 0, len(lm.listeners))
	for topic := range lm.listeners {
		topics = append(topics, topic)
	}
	return topics
}

func (lm *ListenerManager) GetListener(topic string) (Listener, bool) {
	l, ok := lm.listeners[topic]
	return l, ok
}
