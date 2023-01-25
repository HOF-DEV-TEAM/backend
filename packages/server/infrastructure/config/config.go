package config

import (
	"fmt"

	"github.com/caarlos0/env"	
	"go.uber.org/zap"
)


type ServerConfig struct {
	AppEnv		string 	`env:"APP_ENV" envDefault:"dev" envWhitelisted:"true"`
	HTTPPort 	int 	`env:"PORT" envDefault:"8080" envWhitelisted:"true"`
	Database 	DatabaseConfig
}

type DatabaseConfig struct {
	Host 		string	`env:"DB_HOST"`
	Port 		string 	`env:"DB_PORT"`
	Timeout 	int    	`env:"CONNECTION_TIMEOUT_SECONDS" envDefault:"10"`
	DbName 		string 	`env:"DB_NAME" envDefault:"hof_backend"`
	UserName 	string 	`env:"DB_USERNAME"`
	Password 	string  `env:"DB_PASSWORD"`
	DbUrl		string  `env:"DATABASE_URL" envDefault:"" envWhitelisted:"true"`
}


func Read(logger zap.Logger) (*ServerConfig, error) {
	var serverConfig ServerConfig
	
	for _, target := range []interface{} {
		&serverConfig,
		&serverConfig.Database,
	} {
		if err := env.Parse(target); err != nil {
			return &serverConfig, err
		}
	}

	format := "database: {host: %s port:%s timeout:%d, username-hidden password-hidden}"
	out := fmt.Sprintf(format, serverConfig.Database.Host, serverConfig.Database.Port, serverConfig.Database.Timeout)
	logger.Info(out)
	return &serverConfig, nil
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