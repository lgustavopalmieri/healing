package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
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

	if err := c.validateUniquenessConstraints(ctx, specialist.ID, specialist.Email, specialist.LicenseNumber); err != nil {
		return nil, err
	}

	savedSpecialist, err := c.repository.Save(ctx, specialist)
	if err != nil {
		return nil, ErrSaveSpecialist
	}

	c.publishSpecialistCreatedEvent(ctx, savedSpecialist)

	return savedSpecialist, nil
}

func (c *CreateSpecialistUseCase) validateUniquenessConstraints(ctx context.Context, id, email, licenseNumber string) error {
	err := c.repository.ValidateUniqueness(ctx, id, email, licenseNumber)
	if err != nil {
		return err
	}
	return nil
}

func (c *CreateSpecialistUseCase) publishSpecialistCreatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistCreatedEvent := event.NewEvent(SpecialistCreatedEventName, map[string]any{
		"id":            specialist.ID,
		"email":         specialist.Email,
		"licenseNumber": specialist.LicenseNumber,
		"specialty":     specialist.Specialty,
	})

	c.eventPublisher.Dispatch(ctx, specialistCreatedEvent)
}
