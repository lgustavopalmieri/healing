package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain/update"
)

func (c *UpdateSpecialistUseCase) Execute(ctx context.Context, input UpdateSpecialistDTO) (*domain.Specialist, error) {
	existing, err := c.repository.FindByID(ctx, input.ID)
	if err != nil {
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
		return nil, err
	}

	saved, err := c.repository.Update(ctx, updated)
	if err != nil {
		return nil, ErrUpdateSpecialist
	}

	c.publishSpecialistUpdatedEvent(ctx, saved)

	return saved, nil
}

func (c *UpdateSpecialistUseCase) publishSpecialistUpdatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistUpdatedEvent := event.NewEvent(SpecialistUpdatedEventName, map[string]any{
		"id":            specialist.ID,
		"email":         specialist.Email,
		"licenseNumber": specialist.LicenseNumber,
		"specialty":     specialist.Specialty,
	})

	c.eventPublisher.Dispatch(ctx, specialistUpdatedEvent)
}
