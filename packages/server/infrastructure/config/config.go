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
	AppEnv           string `env:"APP_ENV" envDefault:"dev" envWhitelisted:"true"`
	HTTPPort         int    `env:"PORT" envDefault:"8080" envWhitelisted:"true"`
	Database         DatabaseConfig
	AwsConfiguration AwsConfiguration
	Security         security.SecurityConfig
}

type AwsConfiguration struct {
	Region   string `env:"AWS_REGION"`
	Endpoint string `env:"AWS_ENDPOINT"`
	Secret   string `env:"AWS_SECRET"`
	Bucket   string `env:"AWS_BUCKET" envDefault:"hof-media" envWhitelisted:"true"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	Timeout  int    `env:"CONNECTION_TIMEOUT_SECONDS" envDefault:"10"`
	DbName   string `env:"DB_NAME" envDefault:"hof_backend"`
	UserName string `env:"DB_USERNAME"`
	Password string `env:"DB_PASSWORD"`
	DbUrl    string `env:"DATABASE_URL" envDefault:"" envWhitelisted:"true"`
}

func Read(logger zap.Logger) (*ServerConfig, error) {
	var serverConfig ServerConfig

	for _, target := range []interface{}{
		&serverConfig,
		&serverConfig.Database,
		&serverConfig.Security,
		&serverConfig.AwsConfiguration,
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
