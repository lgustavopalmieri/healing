package grpchandler

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/infra/external"
)

type Dependencies struct {
	DB             *sql.DB
	EventPublisher event.EventDispatcher
	Tracer         observability.Tracer
	Logger         observability.Logger
}

func NewSpecialistCreateService(deps Dependencies) *SpecialistCreateGRPCHandler {
	repository := database.NewSpecialistCreateRepository(deps.DB)

	externalGateway := external.NewLicenseValidationGateway()

	command := application.NewCreateSpecialistCommand(
		repository,
		externalGateway,
		deps.EventPublisher,
		deps.Tracer,
		deps.Logger,
	)

	handler := NewSpecialistCreateGRPCHandler(command)

	return handler
}
