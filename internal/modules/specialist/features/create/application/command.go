package application

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

func (c *CreateSpecialistCommand) Execute(ctx context.Context, input CreateSpecialistDTO) (*domain.Specialist, error) {

}
