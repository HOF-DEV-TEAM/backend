// Package config loads environment-driven application settings.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// ServerConfig holds every environment-driven setting for the application.
type ServerConfig struct {
	ServerURL  string `env:"SERVER_URL" envDefault:"https://my-heritage-app-1e457dfa2e9c.herokuapp.com"`
	AppEnv     string `env:"APP_ENV" envDefault:"dev"`
	HTTPPort   int    `env:"PORT" envDefault:"8080"`
	Database   DatabaseConfig
	AWS        AWSConfig
	Cloudinary CloudinaryConfig
	Security   SecurityConfig
	Paystack   PaystackConfig
	Mailer     MailerConfig
}

// DatabaseConfig holds the PostgreSQL connection parameters.
type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	Name     string `env:"DB_NAME" envDefault:"postgres"`
	Username string `env:"DB_USERNAME"`
	Password string `env:"DB_PASSWORD"`
	URL      string `env:"DATABASE_URL" envDefault:"postgres://postgres:hof_db@2023@db.fzeqoeqecuajgxnllbls.supabase.co:5432/postgres?sslmode=disable"`
	SSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
}

// DSN returns the PostgreSQL data-source name.
// DATABASE_URL takes precedence when set.
func (d *DatabaseConfig) DSN() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.Username, d.Password, d.Name, d.SSLMode,
	)
}

// AWSConfig holds S3 credentials and bucket settings.
type AWSConfig struct {
	Region     string `env:"AWS_REGION" envDefault:"us-east-1"`
	AccessKey  string `env:"AWS_ENDPOINT" envDefault:""`
	Secret     string `env:"AWS_SECRET" envDefault:""`
	Bucket     string `env:"AWS_BUCKET" envDefault:"hof-s3"`
	BucketPath string `env:"AWS_BUCKET_PATH" envDefault:"goninja/hof/"`
	BaseURL    string `env:"AWS_BASE_URL" envDefault:"https://s3.amazonaws.com/"`
}

// CloudinaryConfig holds Cloudinary credentials and settings.
type CloudinaryConfig struct {
	CloudName    string `env:"CLOUDINARY_CLOUD_NAME" envDefault:""`
	APIKey       string `env:"CLOUDINARY_API_KEY" envDefault:""`
	APISecret    string `env:"CLOUDINARY_API_SECRET" envDefault:""`
	UploadPreset string `env:"CLOUDINARY_UPLOAD_PRESET" envDefault:""`
}

// SecurityConfig holds JWT signing keys and related options.
type SecurityConfig struct {
	JWTSecret     string `env:"JWT_SECRET"`
	JWTSigningKey string `env:"JWT_SIGNING_KEY"`
}

// PaystackConfig holds the Paystack API address and secret.
type PaystackConfig struct {
	Addr   string `env:"PAYSTACK_ADDR" envDefault:"https://api.paystack.co"`
	Secret string `env:"PAYSTACK_SECRET"`
}

// MailerConfig holds SMTP settings for outbound email.
type MailerConfig struct {
	Email        string `env:"MAILER_EMAIL" envDefault:"no-reply@hofng.org"`
	Host         string `env:"MAILER_HOST" envDefault:"smtp-relay.brevo.com"`
	Username     string `env:"MAILER_USERNAME"`
	Password     string `env:"MAILER_PASSWORD"`
	Port         int    `env:"MAILER_PORT" envDefault:"587"`
	Header       string `env:"MAIL_HEADER" envDefault:"Heritage of Faith Church"`
	TemplatePath string `env:"TEMPLATE_PATH" envDefault:"./templates/"`
	LogoURL      string `env:"MAIL_LOGO_URL" envDefault:"https://s3.eu-west-2.amazonaws.com/hof--s3/hof/HoF_Logo_White.png"`
}

// Load reads all environment variables into a ServerConfig.
func Load(log *zap.Logger) (*ServerConfig, error) {
	cfg := &ServerConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	log.Info("config loaded",
		zap.String("env", cfg.AppEnv),
		zap.Int("port", cfg.HTTPPort),
	)
	return cfg, nil
}
