package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

func (h *SendCredentialsEmailHandler) Handle(ctx context.Context, evt event.Event) error {
	raw, ok := evt.Payload.([]byte)
	if !ok {
		return ErrInvalidEventPayload
	}

	var payload CredentialPendingPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("%s: %w", ErrUnmarshalEventPayloadMessage, err)
	}

	link := h.setPasswordURL + "?token=" + payload.SetPasswordToken

	msg := email.Message{
		To: email.Recipient{
			Email: payload.Email,
		},
		Template: SetPasswordEmailTemplate,
		Data: map[string]any{
			"link": link,
			"role": payload.Role,
		},
		Locale: DefaultLocale,
	}

	if err := h.emailSender.Send(ctx, msg); err != nil {
		return fmt.Errorf("%s: %w", ErrSendEmailMessage, err)
	}

	return nil
}
