package httphandler

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

func NewSpecialistCreateHandler(deps Dependencies) *SpecialistCreateHTTPHandler {
	repository := database.NewSpecialistCreateRepository(deps.DB)

	useCase := application.NewCreateSpecialistUseCase(
		repository,
		deps.EventPublisher,
		deps.Tracer,
		deps.Logger,
	)

	return NewSpecialistCreateHTTPHandler(useCase)
}
