package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"go.uber.org/zap"
	"gopkg.in/mail.v2"
)

// Mailer sends transactional emails via SMTP.
type Mailer struct {
	cfg    config.MailerConfig
	log    *zap.Logger
	dialer *mail.Dialer
}

// New creates a Mailer ready to send via the configured SMTP server.
func New(cfg config.MailerConfig, log *zap.Logger) *Mailer {
	d := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	return &Mailer{cfg: cfg, log: log, dialer: d}
}

// SendPasswordReset delivers a password-reset OTP to the recipient.
func (m *Mailer) SendPasswordReset(to, name, otp string) error {
	data := map[string]any{
		"Name": name,
		"OTP":  otp,
	}
	return m.send(to, "Password Reset", "reset_password.page.tmpl", data)
}

// SendEmailVerification delivers a verification link to the recipient.
func (m *Mailer) SendEmailVerification(to, name, link string) error {
	data := map[string]any{
		"Name": name,
		"Link": link,
	}
	return m.send(to, "Verify Your Email", "verify_email.page.tmpl", data)
}

func (m *Mailer) send(to, subject, templateFile string, data map[string]any) error {
	tmplPath := filepath.Join(m.cfg.TemplatePath, templateFile)
	basePath := filepath.Join(m.cfg.TemplatePath, "base.layout.tmpl")

	t, err := template.ParseFiles(basePath, tmplPath)
	if err != nil {
		return fmt.Errorf("parsing email template %s: %w", templateFile, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("rendering email template: %w", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", m.cfg.Header, m.cfg.Email))
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", buf.String())

	if err := m.dialer.DialAndSend(msg); err != nil {
		m.log.Error("failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("sending email: %w", err)
	}

	m.log.Info("email sent", zap.String("to", to), zap.String("subject", subject))
	return nil
}
