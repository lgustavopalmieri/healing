package listener

import (
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
)

type SendCredentialsEmailHandler struct {
	emailSender    email.EmailSender
	setPasswordURL string
}

func NewSendCredentialsEmailHandler(emailSender email.EmailSender, setPasswordURL string) *SendCredentialsEmailHandler {
	return &SendCredentialsEmailHandler{
		emailSender:    emailSender,
		setPasswordURL: setPasswordURL,
	}
}
