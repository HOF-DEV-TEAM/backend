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
	"bitbucket.org/hofng/hofApp/pkg/audio_message"
	"bitbucket.org/hofng/hofApp/pkg/user"

	"bitbucket.org/hofng/hofApp/interfaces"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
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

		staticFilePath := filesDir + r.URL.Path

		if _, err := os.Stat(filesDir + r.URL.Path); errors.Is(err, os.ErrNotExist) {

			staticFilePath = filepath.Join(filesDir, "index.html")
		}

		w.WriteHeader(http.StatusOK)
		http.ServeFile(w, r, staticFilePath)
	})

	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(config.Security.TokenAuth))
		r.Use(jwtauth.Authenticator)

		audioMessageRepo := audio_message.NewRepository(db, logger)
		audioMessageService := audio_message.NewService(audioMessageRepo, logger, &config.Security)

		buildUserEndpoints(r, userService)
		buildAudioMessageEndpoints(r, audioMessageService)
		buildAudioSeriesEndpoints(r, audioMessageService)
	})

	//unprotected routes
	router.Group(func(r chi.Router) {
		buildSessionEndpoints(r, userService)
	})

}

func buildUserEndpoints(router chi.Router, svc user.Service) {
	userRouter := chi.NewRouter()
	router.Mount("/user", userRouter)
}

func buildSessionEndpoints(router chi.Router, svc user.Service) {
	sessionsRouter := chi.NewRouter()

	signInHandler := interfaces.CreateSignInHandler(svc)
	signUpUserHandler := interfaces.NewHTTPHandler(interfaces.CreateGetUserHandler, svc)
	forgotPasswordHandler := interfaces.ForgotPasswordHandler(svc)
	resetPasswordHandler := interfaces.ResetPasswordHandler(svc)

	sessionsRouter.Post("/sign_in", signInHandler)
	sessionsRouter.Post("/sign_up", signUpUserHandler)
	sessionsRouter.Post("/forgot_password", forgotPasswordHandler)
	sessionsRouter.Post("/reset_password/{token}", resetPasswordHandler)

	router.Mount("/session", sessionsRouter)
}

func buildAudioMessageEndpoints(router chi.Router, svc audio_message.Service) {
	audioMessageRouter := chi.NewRouter()

	createAudioMessageHandler := interfaces.NewHTTPHandler(interfaces.CreateAudioMessageHandler, svc)
	getAudioMessagesHandler := interfaces.NewHTTPHandler(interfaces.GetAudioMessagesHandler, svc)
	getAudioMessageByIDHandler := interfaces.NewHTTPHandler(interfaces.GetAudioMessageByIDHandler, svc)

	audioMessageRouter.Get("/", getAudioMessagesHandler)
	audioMessageRouter.Post("/", createAudioMessageHandler)
	audioMessageRouter.Get("/id/{id}", getAudioMessageByIDHandler)

	router.Mount("/audio_message", audioMessageRouter)
}

func buildAudioSeriesEndpoints(router chi.Router, svc audio_message.Service) {
	audioSeriesRouter := chi.NewRouter()

	createAudioSeriesHandler := interfaces.NewHTTPHandler(interfaces.CreateAudioSeriesHandler, svc)
	getAudioSeriesHandler := interfaces.NewHTTPHandler(interfaces.GetAudioSeriesHandler, svc)
	getAudioSeriesByIDHandler := interfaces.NewHTTPHandler(interfaces.GetAudioSeriesByIDHandler, svc)

	audioSeriesRouter.Post("/", createAudioSeriesHandler)
	audioSeriesRouter.Get("/", getAudioSeriesHandler)
	audioSeriesRouter.Get("/series_id/{id}", getAudioSeriesByIDHandler)

	router.Mount("/audio_series", audioSeriesRouter)
}
