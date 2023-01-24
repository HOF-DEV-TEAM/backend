package application

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	log2 "log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bitbucket.org/hofng/hofApp/interfaces/Router"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jackc/tern/migrate"

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

	if err := app.buildRouter(); err != nil {
		return nil, err
	}

	return &app, nil
}

// Run executes the application
func (app *application) Run() error {
	defer app.db.Close()
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

type EmbeddedF5 struct {
	dirname string
	filename string
	glob string
}


// TODO - Move migration logic to separate module
func (e EmbeddedF5) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (e EmbeddedF5) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func (e EmbeddedF5) Glob(pattern string) (matches []string, err error) {
	return filepath.Glob(pattern)
}

func NewEmbeddedFS() migrate.MigratorFS {
	return EmbeddedF5{}
}


func (app *application) buildSqlClient() *sql.DB {
	dbUrl := app.getUri(app.config.Database.Host, app.config.Database.Port, app.config.Database.UserName, app.config.Database.Password, app.config.Database.DbName)	
	db, err := sql.Open("pgx", dbUrl)

	if err != nil {
		app.logger.Info("msg", zap.String("msg", "failed to connect to database"))
		log2.Fatal(err)
	}

	conn, err := db.Conn(context.Background())

	err = conn.Raw(func (driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn)		//conn is a *pgx.Conn
		opts := migrate.MigratorOptions{
				MigratorFS: NewEmbeddedFS(),
		}

		schema := "public"
		table := fmt.Sprintf("%s.schema_version", schema)
		
		_, b, _, _ := runtime.Caller(0)

		migrationPath, _ := filepath.Abs(fmt.Sprintf("%s/../migrations/", filepath.Dir(b)))
		migrator, err := migrate.NewMigratorEx(context.Background(), conn.Conn(), table, &opts)

		migrator.LoadMigrations(migrationPath)

		if err != nil {
			app.logger.Info("msg", zap.String("msg", "failed to connect to migrator"))
			log2.Fatal(err)
		}
		
		if err := migrator.Migrate(context.Background()); err != nil {
			app.logger.Info("msg", zap.String("msg", "failed to run migrations"))
			log2.Fatal(err)
		}
		return nil
	})

	
	return db
}

func (app *application) getUri(host, port, username, password, dbName string) string {
	format := "postgres://%s:%s@%s:%s/%s"
	return fmt.Sprintf(format, username, password, host, port, dbName)
}