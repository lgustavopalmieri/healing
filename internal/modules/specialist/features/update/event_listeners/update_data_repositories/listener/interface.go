package listener

import (
	"context"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
)

type SourceRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Specialist, error)
}

type DataRepository interface {
	Update(ctx context.Context, specialist *domain.Specialist) error
	PublishDLQ(ctx context.Context, specialist *domain.Specialist, reason error) error
}
