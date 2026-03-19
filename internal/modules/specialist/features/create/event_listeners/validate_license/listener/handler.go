package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	authorizelicense "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/authorize_license"
)

func (h *ValidateLicenseHandler) Handle(ctx context.Context, evt event.Event) error {
	payload := ValidateLicenseEventPayload{}
	err := json.Unmarshal(evt.Payload.([]byte), &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrUnmarshalEventPayloadMessage, err)
	}

	specialist, err := h.repository.FindByID(ctx, payload.ID)
	if err != nil {
		return ErrSpecialistNotFound
	}

	isValid, err := h.gateway.Validate(ctx, specialist.LicenseNumber)
	if err != nil {
		return ErrLicenseValidation
	}

	if !isValid {
		return ErrInvalidLicense
	}

	authorized, err := authorizelicense.AuthorizeLicense(specialist)
	if err != nil {
		return err
	}

	err = h.repository.UpdateStatus(ctx, authorized.ID, authorized.Status)
	if err != nil {
		return ErrUpdateStatus
	}

	h.publishSpecialistUpdatedEvent(ctx, authorized)

	return nil
}

func (h *ValidateLicenseHandler) publishSpecialistUpdatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistUpdatedEvent := event.NewEvent(SpecialistUpdatedEventName, map[string]any{
		"id":            specialist.ID,
		"email":         specialist.Email,
		"licenseNumber": specialist.LicenseNumber,
		"specialty":     specialist.Specialty,
	})

	h.eventPublisher.Dispatch(ctx, specialistUpdatedEvent)
}
