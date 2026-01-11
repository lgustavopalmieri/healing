package event

import "time"

type Event struct {
	Name      string
	Payload   any
	Timestamp time.Time
}

func NewEvent(name string, payload any) Event {
	return Event{
		Name:      name,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// ver implementação em pocs/kafka-poc