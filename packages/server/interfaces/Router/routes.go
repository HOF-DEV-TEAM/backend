package Router

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "bitbucket.org/hofng/hofApp/docs"
	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bitbucket.org/hofng/hofApp/pkg/user"

	"bitbucket.org/hofng/hofApp/interfaces"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func BuildRoutes(router *chi.Mux, logger *zap.Logger, db *sql.DB, config *config.ServerConfig) {
	router.Handle("/swagger/*", httpSwagger.WrapHandler)

	userRepo := user.NewRepository(db, logger)
	userService := user.NewService(userRepo, logger, &config.Security)

	// TODO - group routing better
	//setup routes

	//Serve static admin bundle
	router.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
        workDir, _ := os.Getwd()
        filesDir := filepath.Join(workDir, "admin")

		staticFilePath := filesDir+r.URL.Path

        if _, err := os.Stat(filesDir + r.URL.Path); errors.Is(err, os.ErrNotExist) {

			staticFilePath = filepath.Join(filesDir, "index.html")
        }
		
		w.WriteHeader(http.StatusOK)
		http.ServeFile(w, r, staticFilePath)
    })

	router.Group(func (r chi.Router){		
		r.Use(jwtauth.Verify(config.Security.TokenAuth))
		r.Use(jwtauth.Authenticator)
	})

	//unprotected routes
	router.Group(func(r chi.Router) {
		buildUserEndpoints(router, userService)
		buildSessionEndpoints(router, userService)
	})

}

func buildUserEndpoints(router *chi.Mux, svc user.Service) {
	userRouter := chi.NewRouter()

	createUserHandler := interfaces.CreateGetUserHandler(svc)
	forgotPasswordHandler := interfaces.ForgotPasswordHandler(svc)
	resetPasswordHandler := interfaces.ResetPasswordHandler(svc)

	userRouter.Post("/", createUserHandler)
	userRouter.Post("/forgotPassword", forgotPasswordHandler)
	userRouter.Post("/resetPassword/{token}", resetPasswordHandler)
	router.Mount("/user", userRouter)
}

func buildSessionEndpoints(router *chi.Mux, svc user.Service) {
	sessionsRouter := chi.NewRouter()

	createLoginHandler := interfaces.CreatePostLoginHandler(svc)
	sessionsRouter.Post("/login", createLoginHandler)
	router.Mount("/session", sessionsRouter)
}
