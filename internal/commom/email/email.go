package email

import "context"

type Recipient struct {
	Email string
	Name  string
}

type Message struct {
	To       Recipient
	Template string
	Data     map[string]any
	Locale   string
}

type EmailSender interface {
	Send(ctx context.Context, msg Message) error
}
