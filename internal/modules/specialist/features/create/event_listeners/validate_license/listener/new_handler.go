package listener

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/observability"
)

type ValidateLicenseHandler struct {
	repository     ValidateLicenseRepositoryInterface
	gateway        LicenseGatewayInterface
	eventPublisher event.EventDispatcher
	tracer         observability.Tracer
	logger         observability.Logger
}

func NewValidateLicenseHandler(
	repository ValidateLicenseRepositoryInterface,
	gateway LicenseGatewayInterface,
	eventPublisher event.EventDispatcher,
	tracer observability.Tracer,
	logger observability.Logger,
) *ValidateLicenseHandler {
	return &ValidateLicenseHandler{
		repository:     repository,
		gateway:        gateway,
		eventPublisher: eventPublisher,
		tracer:         tracer,
		logger:         logger,
	}
}
