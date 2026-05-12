package listener

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
)

type SendWelcomeEmailHandler struct {
	emailSender email.EmailSender
}

func NewSendWelcomeEmailHandler(emailSender email.EmailSender) *SendWelcomeEmailHandler {
	return &SendWelcomeEmailHandler{
		emailSender: emailSender,
	}
}
