package infra

import (
	"log"

	"github.com/lgustavopalmieri/healing-specialist/cmd/server/config"
	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
	"github.com/lgustavopalmieri/healing-specialist/internal/platform/email/smtp"
)

func InitEmailSender(cfg *config.Config) email.EmailSender {
	log.Printf("Initializing email sender (smtp=%s:%d)...", cfg.Email.SMTPHost, cfg.Email.SMTPPort)

	sender := smtp.NewSMTPEmailSender(smtp.Config{
		Host:        cfg.Email.SMTPHost,
		Port:        cfg.Email.SMTPPort,
		FromAddress: cfg.Email.FromAddress,
		FromName:    cfg.Email.FromName,
	})

	log.Println("Email sender initialized")
	return sender
}
