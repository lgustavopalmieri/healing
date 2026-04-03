package grpcservice

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
)

type Dependencies struct {
	DB             *sql.DB
	EventPublisher event.EventDispatcher
	Logger         observability.Logger
}

func NewSpecialistCreateService(deps Dependencies) *SpecialistCreateGRPCService {
	repository := database.NewSpecialistCreateRepository(deps.DB)

	useCase := application.NewCreateSpecialistUseCase(
		repository,
		deps.EventPublisher,
		deps.Logger,
	)

	return NewSpecialistCreateGRPCService(useCase)
}
