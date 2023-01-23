package application

import (
	"context"
	"database/sql"
	"fmt"
	log2 "log"
	"net/http"

	"bitbucket.org/hofng/hofApp/domain/repository"
	"bitbucket.org/hofng/hofApp/infrastructure/config"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/zap"
)

type application struct {
	logger      *zap.Logger
	config      *config.ServerConfig
	sqlClient   *sql.DB
	router      *chi.Mux
	repo        repository.Repositories
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
	app.sqlClient = app.buildSqlClient()

	if err := app.sqlClient.PingContext(context.Background()); err != nil {
		app.logger.Info("msg", zap.String("msg", "failed to ping to database"))
		log2.Fatal(err)
	}

	defer app.sqlClient.Close()
	return &app, nil
}

// Run executes the application
func (app *application) Run() error {
	rtr := chi.NewRouter()

	rtr.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to HOF"))
	})
	svr := http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.HTTPPort),
		Handler: rtr,
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

func (app *application) buildSqlClient() *sql.DB {
	dbUrl := app.getUri(app.config.Database.Host, app.config.Database.Port, app.config.Database.UserName, app.config.Database.Password, app.config.Database.DbName)
	fmt.Println(dbUrl)
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