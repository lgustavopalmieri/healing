package listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	authorizelicense "github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/authorize_license"
)

func (h *ValidateLicenseHandler) Handle(ctx context.Context, evt event.Event) error {
	payload := ValidateLicenseEventPayload{}
	err := json.Unmarshal(evt.Payload.([]byte), &payload)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrUnmarshalEventPayloadMessage, err)
	}

	return h.execute(ctx, payload)
}

func (h *ValidateLicenseHandler) execute(contx context.Context, payload ValidateLicenseEventPayload) error {
	ctx, span := h.tracer.Start(contx, ValidateLicenseSpanName)
	defer span.End()

	h.logger.Info(ctx, StartingLicenseValidationMessage,
		observability.Field{Key: "id", Value: payload.ID},
		observability.Field{Key: "licenseNumber", Value: payload.LicenseNumber})

	specialist, err := h.repository.FindByID(ctx, payload.ID)
	if err != nil {
		span.RecordError(err)
		h.logger.Error(ctx, ErrSpecialistNotFoundMessage,
			observability.Field{Key: "id", Value: payload.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return ErrSpecialistNotFound
	}

	isValid, err := h.gateway.Validate(ctx, specialist.LicenseNumber)
	if err != nil {
		span.RecordError(err)
		h.logger.Error(ctx, ErrLicenseValidationMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "licenseNumber", Value: specialist.LicenseNumber},
			observability.Field{Key: "error", Value: err.Error()})
		return ErrLicenseValidation
	}

	if !isValid {
		span.RecordError(ErrInvalidLicense)
		h.logger.Error(ctx, ErrInvalidLicenseMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "licenseNumber", Value: specialist.LicenseNumber})
		return ErrInvalidLicense
	}

	authorized, err := authorizelicense.AuthorizeLicense(specialist)
	if err != nil {
		span.RecordError(err)
		h.logger.Error(ctx, ErrInvalidStatusTransitionMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "status", Value: string(specialist.Status)},
			observability.Field{Key: "error", Value: err.Error()})
		return err
	}

	err = h.repository.UpdateStatus(ctx, authorized.ID, authorized.Status)
	if err != nil {
		span.RecordError(err)
		h.logger.Error(ctx, ErrUpdateStatusMessage,
			observability.Field{Key: "id", Value: authorized.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return ErrUpdateStatus
	}

	h.logger.Info(ctx, SpecialistStatusUpdatedMessage,
		observability.Field{Key: "id", Value: authorized.ID})

	h.publishSpecialistUpdatedEvent(ctx, authorized)

	h.logger.Info(ctx, LicenseValidatedSuccessMessage,
		observability.Field{Key: "id", Value: authorized.ID},
		observability.Field{Key: "email", Value: authorized.Email})

	return nil
}

func (h *ValidateLicenseHandler) publishSpecialistUpdatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistUpdatedEvent := event.NewEvent(SpecialistUpdatedEventName, map[string]any{
		"id":            specialist.ID,
		"email":         specialist.Email,
		"licenseNumber": specialist.LicenseNumber,
		"specialty":     specialist.Specialty,
	})

	if err := h.eventPublisher.Dispatch(ctx, specialistUpdatedEvent); err != nil {
		h.logger.Warn(ctx, ErrEventPublishMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "error", Value: err.Error()})
	}
}
