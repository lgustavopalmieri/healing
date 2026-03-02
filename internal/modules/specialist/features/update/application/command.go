package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/update"
)

func (c *UpdateSpecialistCommand) Execute(contx context.Context, input UpdateSpecialistDTO) (*domain.Specialist, error) {
	ctx, span := c.tracer.Start(contx, UpdateSpecialistSpanName)
	defer span.End()

	existing, err := c.repository.FindByID(ctx, input.ID)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrSpecialistNotFoundMessage,
			observability.Field{Key: "id", Value: input.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return nil, ErrSpecialistNotFound
	}

	updated, err := update.UpdateSpecialist(existing, update.UpdateSpecialistInput{
		ID:            input.ID,
		Name:          input.Name,
		Email:         input.Email,
		Phone:         input.Phone,
		Specialty:     input.Specialty,
		LicenseNumber: input.LicenseNumber,
		Description:   input.Description,
		Keywords:      input.Keywords,
		AgreedToShare: input.AgreedToShare,
		Status:        input.Status,
	})

	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, err.Error(), observability.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	saved, err := c.repository.Update(ctx, updated)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrUpdateSpecialistMessage,
			observability.Field{Key: "id", Value: updated.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return nil, ErrUpdateSpecialist
	}

	c.publishSpecialistUpdatedEvent(ctx, saved)

	return saved, nil
}

func (c *UpdateSpecialistCommand) publishSpecialistUpdatedEvent(ctx context.Context, specialist *domain.Specialist) {
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
