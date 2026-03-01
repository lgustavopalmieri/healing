package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/event_listeners/validate_license/application"
)

type ValidateLicenseListener struct {
	command *application.ValidateLicenseCommand
}

func NewValidateLicenseListener(command *application.ValidateLicenseCommand) *ValidateLicenseListener {
	return &ValidateLicenseListener{
		command: command,
	}
}

func (l *ValidateLicenseListener) Handle(ctx context.Context, evt event.Event) error {
	payload := application.ValidateLicenseEventPayload{}
	err := json.Unmarshal(evt.Payload.([]byte), &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", application.ErrUnmarshalEventPayloadMessage, err)
	}

	return l.command.Execute(ctx, payload)
}
