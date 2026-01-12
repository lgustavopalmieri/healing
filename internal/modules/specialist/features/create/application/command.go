package application

import (
	"context"
	"errors"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

func (c *CreateSpecialistCommand) Execute(contx context.Context, input CreateSpecialistDTO) (*domain.Specialist, error) {
	ctx, cancel := context.WithCancel(contx)
	defer cancel()

	ctx, span := c.tracer.Start(ctx, CreateSpecialistSpanName)
	defer span.End()

	apiCtx, apiCancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer apiCancel()

	type apiResult struct {
		result bool
		err    error
	}

	apiCh := make(chan apiResult, 1)

	go func() {
		apiCtx, apiSpan := c.tracer.Start(apiCtx, "ValidateLicenseExternal")
		defer apiSpan.End()

		res, err := c.validateLicenseWithExternalGateway(apiCtx, apiSpan, input.LicenseNumber)
		apiCh <- apiResult{result: res, err: err}
	}()

	specialist, err := domain.CreateSpecialist(domain.CreateSpecialistInput{
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
		cancel()
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-apiCh:
		if res.err != nil {
			if errors.Is(res.err, context.DeadlineExceeded) {
				cancel()
				return nil, ErrExternalValidationTimeout
			}
			cancel()
			return nil, res.err
		}
		if !res.result {
			return nil, errors.New(InvalidLicenseNumberMessage)
		}
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

func (c *CreateSpecialistCommand) validateLicenseWithExternalGateway(ctx context.Context, span observability.Span, licenseNumber string) (bool, error) {
	isValidLicense, err := c.externalGateway.ValidateLicenseNumber(ctx, licenseNumber)
	if err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, ErrLicenseValidationMessage,
			observability.Field{Key: "licenseNumber", Value: licenseNumber},
			observability.Field{Key: "error", Value: err.Error()})
		return false, ErrLicenseValidation
	}
	if !isValidLicense {
		c.logger.Warn(ctx, InvalidLicenseNumberMessage, observability.Field{Key: "licenseNumber", Value: licenseNumber})
		return false, domain.ErrInvalidLicense
	}
	return true, nil
}

func (c *CreateSpecialistCommand) publishSpecialistCreatedEvent(ctx context.Context, specialist *domain.Specialist) {
	specialistCreatedEvent := event.NewEvent(SpecialistCreatedEventName, map[string]interface{}{
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
