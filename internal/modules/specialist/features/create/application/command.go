package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/create"
)

func (c *CreateSpecialistCommand) Execute(contx context.Context, input CreateSpecialistDTO) (*domain.Specialist, error) {
	ctx, span := c.tracer.Start(contx, CreateSpecialistSpanName)
	defer span.End()

	specialist, err := create.CreateSpecialist(create.CreateSpecialistInput{
		Name:          input.Name,
		Email:         input.Email,
		Phone:         input.Phone,
		Specialty:     input.Specialty,
		LicenseNumber: input.LicenseNumber,
		Description:   input.Description,
		Keywords:      input.Keywords,
		AgreedToShare: input.AgreedToShare,
	})
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, err.Error(), observability.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	if err := c.validateUniquenessConstraints(ctx, span, specialist.ID, specialist.Email, specialist.LicenseNumber); err != nil {
		return nil, err
	}

	savedSpecialist, err := c.repository.Save(ctx, specialist)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrSaveSpecialistMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "error", Value: err.Error()})
		return nil, ErrSaveSpecialist
	}

	c.publishSpecialistCreatedEvent(ctx, savedSpecialist)

	c.logger.Info(ctx, SpecialistCreatedSuccessMessage,
		observability.Field{Key: "id", Value: savedSpecialist.ID},
		observability.Field{Key: "email", Value: savedSpecialist.Email})

	return savedSpecialist, nil
}

func (c *CreateSpecialistCommand) validateUniquenessConstraints(ctx context.Context, span observability.Span, id, email, licenseNumber string) error {
	err := c.repository.ValidateUniqueness(ctx, id, email, licenseNumber)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrUniquenessValidationMessage,
			observability.Field{Key: "id", Value: id},
			observability.Field{Key: "email", Value: email},
			observability.Field{Key: "licenseNumber", Value: licenseNumber},
			observability.Field{Key: "error", Value: err.Error()})
		return err
	}
	return nil
}

func (c *CreateSpecialistCommand) publishSpecialistCreatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistCreatedEvent := event.NewEvent(SpecialistCreatedEventName, map[string]any{
		"id":            specialist.ID,
		"email":         specialist.Email,
		"licenseNumber": specialist.LicenseNumber,
		"specialty":     specialist.Specialty,
	})

	if err := c.eventPublisher.Dispatch(ctx, specialistCreatedEvent); err != nil {
		c.logger.Warn(ctx, ErrEventPublishMessage,
			observability.Field{Key: "id", Value: specialist.ID},
			observability.Field{Key: "error", Value: err.Error()})
	}
}
