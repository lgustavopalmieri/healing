package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/create"
)

func (c *CreateSpecialistUseCase) Execute(ctx context.Context, input CreateSpecialistDTO) (*domain.Specialist, error) {
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
		return nil, err
	}

	savedSpecialist, err := c.repository.SaveWithValidation(ctx, specialist)
	if err != nil {
		return nil, err
	}

	go c.publishSpecialistCreatedEvent(context.WithoutCancel(ctx), savedSpecialist)

	return savedSpecialist, nil
}

func (c *CreateSpecialistUseCase) publishSpecialistCreatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistCreatedEvent := event.NewEvent(SpecialistCreatedEventName, map[string]any{
		"id":            specialist.ID,
		"email":         specialist.Email,
		"licenseNumber": specialist.LicenseNumber,
		"specialty":     specialist.Specialty,
	})

	if err := c.eventPublisher.Dispatch(ctx, specialistCreatedEvent); err != nil {
		c.logger.Error(ctx, ErrEventPublishMessage,
			observability.Field{Key: "specialist_id", Value: specialist.ID},
			observability.Field{Key: "event", Value: SpecialistCreatedEventName},
			observability.Field{Key: "error", Value: err.Error()},
		)
	}
}
