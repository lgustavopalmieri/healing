package grpcservice

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/adapters/outbound/database"
)

type Dependencies struct {
	DB             *sql.DB
	EventPublisher event.EventDispatcher
	Tracer         observability.Tracer
	Logger         observability.Logger
}

func NewSpecialistCreateService(deps Dependencies) *SpecialistCreateGRPCService {
	repository := database.NewSpecialistCreateRepository(deps.DB)

	command := application.NewCreateSpecialistCommand(
		repository,
		deps.EventPublisher,
		deps.Tracer,
		deps.Logger,
	)

	service := NewSpecialistCreateGRPCService(command)

	return service
}
