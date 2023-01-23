package application

import (
	"context"
	"database/sql"
	"fmt"
	log2 "log"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bitbucket.org/hofng/hofApp/interfaces/Router"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"
)

type application struct {
	logger      *zap.Logger
	config      *config.ServerConfig
	db  		*sql.DB
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

	if err := app.db.PingContext(context.Background()); err != nil {
		app.logger.Info("msg", zap.String("msg", "failed to ping to database"))
		log2.Fatal(err)
	}

	defer app.db.Close()

	if err := app.buildRouter(); err != nil {
		return nil, err
	}

	return &app, nil
}

// Run executes the application
func (app *application) Run() error {

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

	Router.BuildRoutes(app.router, app.logger, app.db)
	
	return nil
}


func (app *application) buildConfig() (*config.ServerConfig, error) {
	return config.Read(*app.logger)
}

func (app *application) buildSqlClient() *sql.DB {
	dbUrl := app.getUri(app.config.Database.Host, app.config.Database.Port, app.config.Database.UserName, app.config.Database.Password, app.config.Database.DbName)	
	db, err := sql.Open("pgx", dbUrl)

	if err != nil {
		app.logger.Info("msg", zap.String("msg", "failed to connect to database"))
		log2.Fatal(err)
	}

	return db
}

func (app *application) getUri(host, port, username, password, dbName string) string {
	format := "postgres://%s:%s@%s:%s/%s"
	return fmt.Sprintf(format, username, password, host, port, dbName)
}