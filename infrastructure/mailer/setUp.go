package mailer

import (
	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bytes"
	"errors"
	"go.uber.org/zap"
	"gopkg.in/mail.v2"
	"text/template"
)

type MailRequest struct {
	to             string
	subject        string
	body           string
	log            *zap.Logger
	config         *config.MailerConfig
	templatesCache *map[string]*template.Template
	templateStore  *map[string]string
}

// NewRequest holds new request data
func NewRequest(to string, subject string, logger *zap.Logger, config *config.MailerConfig) *MailRequest {
	return &MailRequest{
		to:      to,
		subject: subject,
		log:     logger,
		config:  config,
	}
}

func (r *MailRequest) AppSendMail(tempName string, item interface{}) error {
	err := r.createSingleTemplate(tempName, item)
	if err != nil {
		r.log.Error("createSingleTemplate", zap.Error(err))
		return err
	}
	if ok := r.sendEmail(tempName); ok {
		r.log.Info("Message sent", zap.String("subject", r.subject))
		return nil
	} else {
		r.log.Error("sendEmail", zap.Error(err))
		return errors.New("error sending email")
	}

}

func (r *MailRequest) createSingleTemplate(tmpl string, data interface{}) error {
	if r.templateStore == nil {
		r.templateStore = &map[string]string{}
	}

	t := Template{}
	buf := new(bytes.Buffer)

	err := t.Create(buf, tmpl, data)

	if err != nil {
		r.log.Error("Create SingleTemplate", zap.Error(err))
		return err
	}

	(*r.templateStore)[tmpl] = buf.String()

	return nil
}

func (r *MailRequest) sendEmail(templateName string) bool {
	m := mail.NewMessage()

	tCache := *(r.templateStore)
	mBody := tCache[templateName]

	m.SetAddressHeader("From", r.config.Email, r.config.Header)
	m.SetHeader("To", r.to)
	m.SetHeader("Subject", r.subject)
	m.SetBody("text/html", mBody)

	d := mail.NewDialer(r.config.Smtp, r.config.Port, r.config.UserName, r.config.Password)

	if err := d.DialAndSend(m); err != nil {
		r.log.Error("DialAndSend", zap.Error(err))
		return false
	}
	r.log.Info("Email Sent to: " + r.to)
	return true
}
