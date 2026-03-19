package listener

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type ValidateLicenseHandler struct {
	repository     ValidateLicenseRepositoryInterface
	gateway        LicenseGatewayInterface
	eventPublisher event.EventDispatcher
}

func NewValidateLicenseHandler(
	repository ValidateLicenseRepositoryInterface,
	gateway LicenseGatewayInterface,
	eventPublisher event.EventDispatcher,
) *ValidateLicenseHandler {
	return &ValidateLicenseHandler{
		repository:     repository,
		gateway:        gateway,
		eventPublisher: eventPublisher,
	}
}
