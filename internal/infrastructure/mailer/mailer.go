// Package mailer provides SMTP email delivery helpers.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"go.uber.org/zap"
)

// Mailer sends transactional emails via MailerSend API.
type Mailer struct {
	cfg    *config.MailerConfig
	log    *zap.Logger
	client *http.Client
	apiURL string
}

// New creates a Mailer ready to send via the MailerSend API.
func New(cfg *config.MailerConfig, log *zap.Logger) *Mailer {
	if cfg == nil {
		cfg = &config.MailerConfig{}
	}
	return &Mailer{
		cfg:    cfg,
		log:    log,
		client: &http.Client{Timeout: 10 * time.Second},
		apiURL: "https://api.mailersend.com/v1/email",
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
	if execErr := t.Execute(&buf, map[string]any{"DataMap": data}); execErr != nil {
		return fmt.Errorf("rendering email template: %w", execErr)
	}

	// Prepare MailerSend API request
	emailData := map[string]interface{}{
		"from": map[string]interface{}{
			"email": m.cfg.Email,
			"name":  m.cfg.Header,
		},
		"to": []map[string]interface{}{
			{
				"email": to,
			},
		},
		"subject": subject,
		"html":    buf.String(),
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return fmt.Errorf("marshaling email data: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", m.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+m.cfg.MailerSendAPIKey)

	// Send request
	resp, err := m.client.Do(req)
	if err != nil {
		m.log.Error("failed to send email via API",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("sending email via API: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			m.log.Error("failed to close response body", zap.Error(err))
		}
	}()

	if resp.StatusCode >= 400 {
		// Read error response body for debugging
		body, _ := io.ReadAll(resp.Body)
		m.log.Error("MailerSend API error",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return fmt.Errorf("MailerSend API returned status %d: %s", resp.StatusCode, string(body))
	}

	m.log.Info("email sent successfully via API",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}

// keep for smtp use.
// func (m *Mailer) send(to, subject, templateFile string, data map[string]any) error {
//	tmplPath := filepath.Join(m.cfg.TemplatePath, templateFile)
//	basePath := filepath.Join(m.cfg.TemplatePath, "base.layout.tmpl")
//
//	// Parse the page template first so that t.Execute() runs the page file
//	// (which calls {{template "base" .}}), NOT the empty layout wrapper.
//	t, err := template.ParseFiles(tmplPath, basePath)
//	if err != nil {
//		return fmt.Errorf("parsing email template %s: %w", templateFile, err)
//	}
//
//	// Wrap data in DataMap to match {{.DataMap.X}} references in all templates.
//	// HofRoundLogo is injected from config so the URL can be changed without redeploying.
//	data["HofRoundLogo"] = m.cfg.LogoURL
//
//	var buf bytes.Buffer
//	if err := t.Execute(&buf, map[string]any{"DataMap": data}); err != nil {
//		return fmt.Errorf("rendering email template: %w", err)
//	}
//
//	msg := mail.NewMessage()
//	msg.SetHeader("From", fmt.Sprintf("%s <%s>", m.cfg.Header, m.cfg.Email))
//	msg.SetHeader("To", to)
//	msg.SetHeader("Subject", subject)
//	msg.SetBody("text/html", buf.String())
//
//	if err := m.dialer.DialAndSend(msg); err != nil {
//		m.log.Error("failed to send email",
//			zap.String("to", to),
//			zap.String("subject", subject),
//			zap.Error(err),
//		)
//		return fmt.Errorf("sending email: %w", err)
//	}
//
//	m.log.Info("email sent", zap.String("to", to), zap.String("subject", subject))
//	return nil
//}
