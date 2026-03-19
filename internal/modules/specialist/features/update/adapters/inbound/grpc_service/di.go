package grpcservice

import (
	"database/sql"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/adapters/outbound/database"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/update/application"
)

type Dependencies struct {
	DB             *sql.DB
	EventPublisher event.EventDispatcher
}

func NewSpecialistUpdateService(deps Dependencies) *SpecialistUpdateGRPCService {
	repository := database.NewSpecialistUpdateRepository(deps.DB)

	useCase := application.NewUpdateSpecialistUseCase(
		repository,
		deps.EventPublisher,
	)

	return NewSpecialistUpdateGRPCService(useCase)
}
