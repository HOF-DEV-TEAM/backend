package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)


type ServerConfig struct {
	HTTPPort 	int `env:"HTTP_SERVE_PORT" envDefault:"80" envWhitelisted:"true"`
	Database 	DatabaseConfig
}

type DatabaseConfig struct {
	Host 		string	`env:"HOST,required"`
	Port 		string 	`env:"PORT,required"`
	Timeout 	int    	`env:"CONNECTION_TIMEOUT_SECONDS" envDefault:"10"`
	DbName 		string 	`env:"DB_NAME" envDefault:"hof_backend"`
	UserName 	string 	`env:"USERNAME,required"`
	Password 	string  `env:"PASSWORD,required"`
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

func (svrConfig *ServerConfig) GetClientOptions() *options.ClientOptions {
	return options.Client().
	SetConnectTimeout(time.Duration(svrConfig.Database.Timeout) * time.Second).
	SetHosts([]string{svrConfig.Database.Host +  svrConfig.Database.Port}).
	SetAuth(options.Credential{		
		AuthMechanism: "SCRAM-SHA-256",
		Username: svrConfig.Database.UserName,
		Password: svrConfig.Database.Password,
	})
}