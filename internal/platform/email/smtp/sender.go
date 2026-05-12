package smtp

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/email"
)

type Config struct {
	Host        string
	Port        int
	FromAddress string
	FromName    string
}

type SMTPEmailSender struct {
	host        string
	port        int
	fromAddress string
	fromName    string
}

func NewSMTPEmailSender(cfg Config) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:        cfg.Host,
		port:        cfg.Port,
		fromAddress: cfg.FromAddress,
		fromName:    cfg.FromName,
	}
}

func (s *SMTPEmailSender) Send(ctx context.Context, msg email.Message) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress)
	to := msg.To.Email
	subject := buildSubject(msg.Template, msg.Data)
	body := buildBody(msg.Template, msg.Data)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	raw := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s",
		from, to, subject, mime, body)

	err := smtp.SendMail(addr, nil, s.fromAddress, []string{to}, []byte(raw))
	if err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}

func buildSubject(template string, data map[string]any) string {
	switch template {
	case "set-password":
		return "Healing - Configure sua senha"
	case "reset-password":
		return "Healing - Redefinir senha"
	case "welcome":
		return "Healing - Bem-vindo"
	default:
		return "Healing - Notificação"
	}
}

func buildBody(template string, data map[string]any) string {
	var sb strings.Builder

	switch template {
	case "set-password":
		link, _ := data["link"].(string)
		name, _ := data["name"].(string)
		sb.WriteString("<h2>Olá")
		if name != "" {
			sb.WriteString(", ")
			sb.WriteString(name)
		}
		sb.WriteString("!</h2>")
		sb.WriteString("<p>Sua conta foi criada na plataforma Healing.</p>")
		sb.WriteString("<p>Clique no link abaixo para configurar sua senha:</p>")
		sb.WriteString(fmt.Sprintf("<p><a href=\"%s\">Configurar Senha</a></p>", link))
		sb.WriteString("<p>Este link expira em 24 horas.</p>")

	case "reset-password":
		link, _ := data["link"].(string)
		sb.WriteString("<h2>Redefinição de Senha</h2>")
		sb.WriteString("<p>Você solicitou a redefinição da sua senha.</p>")
		sb.WriteString(fmt.Sprintf("<p><a href=\"%s\">Redefinir Senha</a></p>", link))
		sb.WriteString("<p>Este link expira em 1 hora.</p>")
		sb.WriteString("<p>Se você não solicitou, ignore este email.</p>")

	default:
		message, _ := data["message"].(string)
		if message == "" {
			message = "Você tem uma nova notificação na plataforma Healing."
		}
		sb.WriteString(fmt.Sprintf("<p>%s</p>", message))
	}

	return sb.String()
}
