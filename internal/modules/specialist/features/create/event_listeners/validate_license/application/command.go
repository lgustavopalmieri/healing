package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

func (c *ValidateLicenseCommand) Execute(contx context.Context, payload ValidateLicenseEventPayload) error {
	ctx, span := c.tracer.Start(contx, ValidateLicenseSpanName)
	defer span.End()

	c.logger.Info(ctx, StartingLicenseValidationMessage,
		observability.Field{Key: "id", Value: payload.ID},
		observability.Field{Key: "licenseNumber", Value: payload.LicenseNumber})

	specialist, err := c.repository.FindByID(ctx, payload.ID)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrSpecialistNotFoundMessage,
			observability.Field{Key: "id", Value: payload.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return ErrSpecialistNotFound
	}

	isValid, err := c.gateway.Validate(ctx, specialist.LicenseNumber)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrLicenseValidationMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "licenseNumber", Value: specialist.LicenseNumber},
			observability.Field{Key: "error", Value: err.Error()})
		return ErrLicenseValidation
	}

	if !isValid {
		span.RecordError(ErrInvalidLicense)
		c.logger.Error(ctx, ErrInvalidLicenseMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "licenseNumber", Value: specialist.LicenseNumber})
		return ErrInvalidLicense
	}

	updatedSpecialist, err := c.repository.UpdateStatus(ctx, specialist.ID, domain.StatusActive)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrUpdateStatusMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return ErrUpdateStatus
	}

	c.logger.Info(ctx, SpecialistStatusUpdatedMessage,
		observability.Field{Key: "id", Value: updatedSpecialist.ID})

	c.publishSpecialistUpdatedEvent(ctx, updatedSpecialist)

	c.logger.Info(ctx, LicenseValidatedSuccessMessage,
		observability.Field{Key: "id", Value: updatedSpecialist.ID},
		observability.Field{Key: "email", Value: updatedSpecialist.Email})

	return nil
}

func (c *ValidateLicenseCommand) publishSpecialistUpdatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistUpdatedEvent := event.NewEvent(SpecialistUpdatedEventName, map[string]any{
		"id":            specialist.ID,
		"email":         specialist.Email,
		"licenseNumber": specialist.LicenseNumber,
		"specialty":     specialist.Specialty,
	})

	if err := c.eventPublisher.Dispatch(ctx, specialistUpdatedEvent); err != nil {
		c.logger.Warn(ctx, ErrEventPublishMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "error", Value: err.Error()})
	}
}
