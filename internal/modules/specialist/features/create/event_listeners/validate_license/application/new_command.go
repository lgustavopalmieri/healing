package application

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type ValidateLicenseCommand struct {
	repository     ValidateLicenseRepositoryInterface
	gateway        LicenseGatewayInterface
	eventPublisher event.EventDispatcher
	tracer         observability.Tracer
	logger         observability.Logger
}

func NewValidateLicenseCommand(
	repository ValidateLicenseRepositoryInterface,
	gateway LicenseGatewayInterface,
	eventPublisher event.EventDispatcher,
	tracer observability.Tracer,
	logger observability.Logger,
) *ValidateLicenseCommand {
	return &ValidateLicenseCommand{
		repository:     repository,
		gateway:        gateway,
		eventPublisher: eventPublisher,
		tracer:         tracer,
		logger:         logger,
	}
}
