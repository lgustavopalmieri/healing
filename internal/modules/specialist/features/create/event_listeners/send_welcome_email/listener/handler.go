package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

func (h *SendWelcomeEmailHandler) Handle(ctx context.Context, evt event.Event) error {
	raw, ok := evt.Payload.([]byte)
	if !ok {
		return ErrInvalidEventPayload
	}

	var payload SpecialistCreatedPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("%s: %w", ErrUnmarshalEventPayloadMessage, err)
	}

	msg := email.Message{
		To: email.Recipient{
			Email: payload.Email,
			Name:  "",
		},
		Template: WelcomeEmailTemplate,
		Data: map[string]any{
			"specialty":      payload.Specialty,
			"license_number": payload.LicenseNumber,
		},
		Locale: DefaultLocale,
	}

	if err := h.emailSender.Send(ctx, msg); err != nil {
		return fmt.Errorf("%s: %w", ErrSendEmailMessage, err)
	}

	return nil
}
