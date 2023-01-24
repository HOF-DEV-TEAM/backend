package config

import (
	"fmt"

	"github.com/caarlos0/env"	
	"go.uber.org/zap"
)


type ServerConfig struct {
	HTTPPort 	int `env:"HTTP_SERVE_PORT" envDefault:"80" envWhitelisted:"true"`
	Database 	DatabaseConfig
}

type DatabaseConfig struct {
	Host 		string	`env:"DB_HOST,required"`
	Port 		string 	`env:"DB_PORT,required"`
	Timeout 	int    	`env:"CONNECTION_TIMEOUT_SECONDS" envDefault:"10"`
	DbName 		string 	`env:"DB_NAME" envDefault:"hof_backend"`
	UserName 	string 	`env:"DB_USERNAME,required"`
	Password 	string  `env:"DB_PASSWORD,required"`
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

	out := fmt.Sprintf("database: {host: %s port:%s timeout:%d, username-hidden password-hidden}", serverConfig.Database.Host, serverConfig.Database.Port, serverConfig.Database.Timeout)
	logger.Info(out)
	return &serverConfig, nil
}