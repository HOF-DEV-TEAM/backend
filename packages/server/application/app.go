package application

import (
	// "bitbucket.org/hofng/hofApp/domain/repository"
	"context"
	"fmt"
	log2 "log"
	"net/http"

	"bitbucket.org/hofng/hofApp/domain/repository"
	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bitbucket.org/hofng/hofApp/infrastructure/persistence"
	"bitbucket.org/hofng/hofApp/interfaces"
	"bitbucket.org/hofng/hofApp/interfaces/Router"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"

	"go.uber.org/zap"
)

type application struct {
	logger      *zap.Logger
	config      *config.ServerConfig
	mongoClient *mongo.Client
	router      *chi.Mux
	repo        repository.Repositories
	httpHandler interfaces.HTTPHandler
}

// New instances a new application
// The application contains all the related components that allow the execution of the service
func New(logger *zap.Logger) (*application, error) {
	var app application
	var err error

	app.logger = logger
	app.config, err = app.buildConfig()

	if err != nil {
		return nil, err
	}
	//build application clients
	app.mongoClient = app.buildMongoClient()
	app.repo = persistence.New(app.mongoClient, app.config.Database.DbName, app.logger)
	app.httpHandler = *interfaces.New(app.repo)

	return &app, nil
}

// Run executes the application
func (app *application) Run() error {
	svr := http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.HTTPPort),
		Handler: Router.Router(&app.httpHandler),
	}
	err := svr.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (app *application) buildConfig() (*config.ServerConfig, error) {
	return config.Read(*app.logger)
}

func (app *application) buildMongoClient() *mongo.Client {
	clientOpts := app.config.GetClientOptions()

	mongoClient, err := mongo.NewClient(clientOpts)

	err = mongoClient.Connect(context.Background())
	if err != nil {
		app.logger.Info("msg", zap.String("msg", "failed to connect to database"))
		log2.Fatal(err)

	}
	return mongoClient
}
