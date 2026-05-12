package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	authevents "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/events"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

func (h *CreateSpecialistCredentialHandler) Handle(ctx context.Context, evt event.Event) error {
	raw, ok := evt.Payload.([]byte)
	if !ok {
		return ErrInvalidEventPayload
	}

	var payload SpecialistCreatedPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("%s: %w", ErrUnmarshalEventPayloadMessage, err)
	}

	existing, err := h.credentialRepository.FindByEmailProviderRole(ctx, payload.Email, provider.Password, role.Specialist)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrFindCredentialMessage, err)
	}
	if existing != nil {
		return nil
	}

	newCredential := credential.NewCredential(credential.NewCredentialInput{
		SubjectID: payload.ID,
		Role:      role.Specialist,
		Provider:  provider.Password,
		Email:     payload.Email,
	})
	if err := h.credentialRepository.Save(ctx, newCredential); err != nil {
		return fmt.Errorf("%s: %w", ErrSaveCredentialMessage, err)
	}

	tokenString, _, err := h.setPasswordTokenGenerator.Generate(ctx, payload.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrGenerateSetPasswordMessage, err)
	}

	h.publishCredentialPendingEvent(ctx, payload, tokenString)
	return nil
}

func (h *CreateSpecialistCredentialHandler) publishCredentialPendingEvent(ctx context.Context, payload SpecialistCreatedPayload, setPasswordToken string) {
	credentialPendingEvent := event.NewEvent(authevents.AuthCredentialPending, map[string]any{
		"subject_id":         payload.ID,
		"role":               role.Specialist.String(),
		"email":              payload.Email,
		"set_password_token": setPasswordToken,
	})

	h.eventPublisher.Dispatch(ctx, credentialPendingEvent)
}
