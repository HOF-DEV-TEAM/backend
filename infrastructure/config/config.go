package config

import (
	"fmt"
	"net/url"
	"strings"

	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type ServerConfig struct {
	ServerUrl        string `env:"SERVER_URL" envDefault:"https://app.hoftech.org"`
	AppEnv           string `env:"APP_ENV" envDefault:"dev" envWhitelisted:"true"`
	HTTPPort         int    `env:"PORT" envDefault:"8080" envWhitelisted:"true"`
	Database         DatabaseConfig
	AwsConfiguration AwsConfiguration
	Security         security.SecurityConfig
	PaystackConfig   PaystackConfig
	Mailer           MailerConfig
}

type PaystackConfig struct {
	Addr           string `env:"PAYSTACK_ADDR" envDefault:"https://api.paystack.co"`
	PaystackSecret string `env:"PAYSTACK_SECRET"`
}

type AwsConfiguration struct {
	Region     string `env:"AWS_REGION" envDefault:"us-east-1"`
	Endpoint   string `env:"AWS_ENDPOINT" envDefault:"AKIATT2RU3YXQR3XSCNG"`
	Secret     string `env:"AWS_SECRET"  envDefault:"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"`
	Bucket     string `env:"AWS_BUCKET" envDefault:"hof-s3" envWhitelisted:"true"`
	BucketPath string `env:"AWS_BUCKET_PATH" envDefault:"goninja/hof/"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	Timeout  int    `env:"CONNECTION_TIMEOUT_SECONDS" envDefault:"10"`
	DbName   string `env:"DB_NAME" envDefault:"postgres"`
	UserName string `env:"DB_USERNAME"`
	Password string `env:"DB_PASSWORD"`
	DbUrl    string `env:"DATABASE_URL" envDefault:"postgres://postgres:hof_db@2023@db.fzeqoeqecuajgxnllbls.supabase.co:5432/postgres?sslmode=disable" envWhitelisted:"true"`
}
type MailerConfig struct {
	Email                 string `env:"MAILER_EMAIL" envDefault:"no-reply@hofng.org"`
	Smtp                  string `env:"MAILER_HOST" envDefault:"smtp-relay.sendinblue.com"`
	UserName              string `env:"MAILER_USERNAME" envDefault:"hofchurchnig@gmail.com"`
	Password              string `env:"MAILER_PASSWORD" envDefault:"DUBcPE8KYabH1gnr"`
	Port                  int    `env:"MAILER_PORT" envDefault:"2525"`
	Header                string `env:"MAIL_HEADER" envDefault:"Heritage of Faith Church"`
	TemplatePath          string `env:"TEMPLATE_PATH" envDefault:"./files/templates/"`
	PasswordResetMailPath string `env:"MAIL_PATH" envDefault:"reset_password.page.tmpl"`
}

func Read(logger zap.Logger) (*ServerConfig, error) {
	var serverConfig ServerConfig

	for _, target := range []interface{}{
		&serverConfig,
		&serverConfig.Database,
		&serverConfig.Security,
		&serverConfig.AwsConfiguration,
		&serverConfig.PaystackConfig,
		&serverConfig.Mailer,
	} {
		if err := env.Parse(target); err != nil {
			return &serverConfig, err
		}
	}

	serverConfig.Security.JWTContextKey = security.JWTContextKey
	serverConfig.Security.JWTClaimsContextKey = security.JWTClaimsContextKey
	serverConfig.Security.JWTExpiration = security.JWTLifeTime

	out := serverConfig.formartUri()
	logger.Info(out)
	logger.Info(serverConfig.Mailer.PasswordResetMailPath + " " + serverConfig.Mailer.Email)
	return &serverConfig, nil
}

func (config *ServerConfig) formartUri() string {
	format := "database: {host: %s port:%s timeout:%d, username-hidden password-hidden}"
	host := config.Database.Host
	port := config.Database.Port
	timeout := config.Database.Timeout

	if config.Database.DbUrl != "" {
		if connString, err := url.Parse(config.Database.DbUrl); err == nil {

			result := strings.Split(connString.Host, ":")
			host = result[0]
			port = result[1]
		}
	}

	return fmt.Sprintf(format, host, port, timeout)
}

func (config *ServerConfig) GetUri() string {
	if len(config.Database.DbUrl) > 0 {
		return config.Database.DbUrl
	}

	format := "postgres://%s:%s@%s:%s/%s"
	return fmt.Sprintf(
		format,
		config.Database.UserName,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DbName,
	)
}
