package grpcservice

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/outbound/database"
)

type Dependencies struct {
	DB             *sql.DB
	EventPublisher event.EventDispatcher
	Tracer         observability.Tracer
	Logger         observability.Logger
}

func NewSpecialistUpdateService(deps Dependencies) *SpecialistUpdateGRPCService {
	repository := database.NewSpecialistUpdateRepository(deps.DB)

	command := application.NewUpdateSpecialistCommand(
		repository,
		deps.EventPublisher,
		deps.Tracer,
		deps.Logger,
	)

	return NewSpecialistUpdateGRPCService(command)
}
