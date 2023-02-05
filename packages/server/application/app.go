package application

import (
	"context"
	"database/sql"
	"fmt"
	log2 "log"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bitbucket.org/hofng/hofApp/infrastructure/db"
	"bitbucket.org/hofng/hofApp/interfaces/Router"
	"bitbucket.org/hofng/hofApp/pkg/uploader"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

type application struct {
	logger      *zap.Logger
	config      *config.ServerConfig
	db  		*sql.DB
	awsClient 	*uploader.AWSClient
	router      *chi.Mux
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
	app.db = app.buildSqlClient()
	app.awsClient = app.getAwsS3Uploader()

	if err := app.db.PingContext(context.Background()); err != nil {
		app.logger.Info("msg", zap.String("msg", "failed to ping to database"))
		log2.Fatal(err)
	}

	if err := app.buildRouter(); err != nil {
		return nil, err
	}

	return &app, nil
}

// Run executes the application
func (app *application) Run() error {
	defer app.db.Close()

	app.router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("Welcome to HOF Server.."))
	})

	svr := http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.HTTPPort),
		Handler: app.router,
	}
	err := svr.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (app *application) buildRouter() error {
	app.router = chi.NewRouter()

	app.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))



	Router.BuildRoutes(
		app.router, 
		app.logger, 
		app.db, 
		app.config,
		app.awsClient,
	)
	
	return nil
}


func (app *application) buildConfig() (*config.ServerConfig, error) {
	return config.Read(*app.logger)
}

func (app *application) buildSqlClient() *sql.DB {
	db := database.Database{Config: app.config, Log: app.logger}

	dbConn, err := db.ConnectDB()

	if err != nil {
		log2.Fatal(err)
	}

	if err := db.RunMigration(dbConn); err != nil {
		log2.Fatal(err)
	}


	return dbConn
}

// Allow aws fail silently?
func (app *application) getAwsS3Uploader() *uploader.AWSClient {	
	awsClient := uploader.AWSClient{Config: app.config, Log: app.logger}
	awsClient.ConnectAWS()

	return &awsClient
}