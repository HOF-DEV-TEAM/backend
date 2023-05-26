package mailer

import (
	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bytes"
	"errors"
	"go.uber.org/zap"
	"path/filepath"
	"text/template"

	"gopkg.in/gomail.v2"
)

type Request struct {
	to      string
	subject string
	body    string
	log     *zap.Logger
	config  *config.MailerConfig
}

// NewRequest holds new request data
func NewRequest(to string, subject string, logger *zap.Logger, config *config.MailerConfig) *Request {
	return &Request{
		to:      to,
		subject: subject,
		log:     logger,
		config:  config,
	}
}

func (r *Request) AppSendMail(tempName string, item interface{}) error {
	err := r.CreateTemplate(tempName, item)
	if err != nil {
		r.log.Error("CreateTemplate", zap.Error(err))
		return err
	}
	if ok := r.sendEmail(); ok {
		r.log.Info("Message sent", zap.String("subject", r.subject))
		return nil
	} else {
		r.log.Error("sendEmail", zap.Error(err))
		return errors.New("error sending email")
	}

}

var functions = template.FuncMap{}

func (r *Request) CreateTemplate(tmpl string, data interface{}) error {

	tc, err := r.CreateTemplateCache()
	if err != nil {
		r.log.Error("CreateTemplateCache", zap.Error(err))
	}
	t, ok := tc[tmpl]
	if !ok {
		r.log.Error("tmpl", zap.Error(errors.New("passed template does not match any available template")))
		return err
	}
	buf := new(bytes.Buffer)

	if err = t.Execute(buf, data); err != nil {
		r.log.Error("Execute Data", zap.Error(err))
		return err
	}
	r.body = buf.String()
	return nil
}

func (r *Request) CreateTemplateCache() (map[string]*template.Template, error) {

	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		r.log.Error("filepath.Glob pages", zap.Error(err))
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			r.log.Error("ParseFiles", zap.Error(err))
			return nil, err
		}
		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			r.log.Error("filepath.Glob base layout", zap.Error(err))
			return nil, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return nil, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}

func (r *Request) sendEmail() bool {

	m := gomail.NewMessage()

	m.SetAddressHeader("From", r.config.Email, r.config.Header)
	m.SetHeader("To", r.to)
	m.SetHeader("Subject", r.subject)
	m.SetBody("text/html", r.body)

	d := gomail.NewDialer(r.config.Smtp, r.config.Port, r.config.UserName, r.config.Password)

	if err := d.DialAndSend(m); err != nil {
		r.log.Error("DialAndSend", zap.Error(err))
		return false
	}
	r.log.Info("Email Sent to: " + r.to)
	return true
}
