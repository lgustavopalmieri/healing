package listener

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

type CreateSpecialistCredentialHandler struct {
	credentialRepository      CredentialRepository
	setPasswordTokenGenerator SetPasswordTokenGenerator
	eventPublisher            event.EventDispatcher
}

func NewCreateSpecialistCredentialHandler(
	credentialRepository CredentialRepository,
	setPasswordTokenGenerator SetPasswordTokenGenerator,
	eventPublisher event.EventDispatcher,
) *CreateSpecialistCredentialHandler {
	return &CreateSpecialistCredentialHandler{
		credentialRepository:      credentialRepository,
		setPasswordTokenGenerator: setPasswordTokenGenerator,
		eventPublisher:            eventPublisher,
	}
}
