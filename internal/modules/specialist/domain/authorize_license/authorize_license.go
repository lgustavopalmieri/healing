package authorizelicense

import (
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

func AuthorizeLicense(specialist *domain.Specialist) (*domain.Specialist, error) {
	if specialist.Status != domain.StatusPending {
		return nil, ErrInvalidStatusTransition
	}

	specialist.Status = domain.StatusAuthorizedLicense
	specialist.UpdatedAt = time.Now().UTC()

	return specialist, nil
}
