// Package mailer provides SMTP email delivery helpers.
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

// Mailer sends transactional emails via MailerSend API.
type Mailer struct {
	cfg    *config.MailerConfig
	log    *zap.Logger
	dialer *mail.Dialer
}

// New creates a Mailer ready to send via the MailerSend API.
func New(cfg *config.MailerConfig, log *zap.Logger) *Mailer {
	if cfg == nil {
		cfg = &config.MailerConfig{}
	}

	d := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	return &Mailer{
		cfg:    cfg,
		log:    log,
		dialer: d,
	}
}

// SendPasswordReset delivers a password-reset OTP to the recipient.
// The OTP is valid for 5 minutes (enforced in the application layer).
func (m *Mailer) SendPasswordReset(to, name, otp string) error {
	data := map[string]any{
		"User":      name,
		"OTP":       otp,
		"ExpiresIn": "5",
	}
	return m.send(to, "Password Reset", "reset_password.page.tmpl", data)
}

// SendEmailVerification delivers a 24-hour verification link to the recipient.
func (m *Mailer) SendEmailVerification(to, name, link string) error {
	data := map[string]any{
		"User":             name,
		"VerificationLink": link,
		"ExpiresIn":        "24 hours",
	}
	return m.send(to, "Verify Your Email", "verify_email.page.tmpl", data)
}

func (m *Mailer) send(to, subject, templateFile string, data map[string]any) error {
	tmplPath := filepath.Join(m.cfg.TemplatePath, templateFile)
	basePath := filepath.Join(m.cfg.TemplatePath, "base.layout.tmpl")

	// Parse the page template first so that t.Execute() runs the page file
	// (which calls {{template "base" .}}), NOT the empty layout wrapper.
	t, err := template.ParseFiles(tmplPath, basePath)
	if err != nil {
		return fmt.Errorf("parsing email template %s: %w", templateFile, err)
	}

	// Wrap data in DataMap to match {{.DataMap.X}} references in all templates.
	// HofRoundLogo is injected from config so the URL can be changed without redeploying.
	data["HofRoundLogo"] = m.cfg.LogoURL

	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]any{"DataMap": data}); err != nil {
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
