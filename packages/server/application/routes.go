package application

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "bitbucket.org/hofng/hofApp/docs"
	"bitbucket.org/hofng/hofApp/pkg/audio_message"
	"bitbucket.org/hofng/hofApp/pkg/user"

	"bitbucket.org/hofng/hofApp/pkg/uploader"
	"github.com/go-chi/chi/v5"
)

func (app *application) buildRoutes() {
	app.router.Handle("/swagger/*", httpSwagger.WrapHandler)

	userRepo := user.NewRepository(app.db, app.logger)
	userService := user.NewService(userRepo, app.logger, &app.config.Security)

	// TODO - group routing better
	//setup routes

	//Serve static admin bundle
	app.router.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
		workDir, _ := os.Getwd()
		filesDir := filepath.Join(workDir, "admin")

		staticFilePath := filesDir + r.URL.Path

		if _, err := os.Stat(filesDir + r.URL.Path); errors.Is(err, os.ErrNotExist) {

			staticFilePath = filepath.Join(filesDir, "index.html")
		}

		w.WriteHeader(http.StatusOK)
		http.ServeFile(w, r, staticFilePath)
	})

	app.router.Group(func(r chi.Router) {
		r.Use(app.config.Security.Verifier())
		r.Use(app.config.Security.Authenticator)

		audioMessageRepo := audio_message.NewRepository(app.db, app.logger)
		audioMessageService := audio_message.NewService(audioMessageRepo, app.logger, &app.config.Security)
		uploaderService := uploader.NewService(app.awsClient)

		buildUserEndpoints(r, userService)
		buildAudioMessageEndpoints(r, audioMessageService)
		buildAudioSeriesEndpoints(r, audioMessageService)
		buildUploadEndpoints(r, uploaderService)
	})

	//unprotected routes
	app.router.Group(func(r chi.Router) {
		buildSessionEndpoints(r, userService)
	})

}

func buildUserEndpoints(router chi.Router, svc user.Service) {
	userRouter := chi.NewRouter()
	router.Mount("/user", userRouter)
}

func buildSessionEndpoints(router chi.Router, svc user.Service) {
	sessionsRouter := chi.NewRouter()

	signInHandler := user.CreateSignInHandler(svc)
	signUpUserHandler := NewHTTPHandler(user.CreateGetUserHandler, svc)
	forgotPasswordHandler := user.ForgotPasswordHandler(svc)
	resetPasswordHandler := user.ResetPasswordHandler(svc)

	sessionsRouter.Post("/sign_in", signInHandler)
	sessionsRouter.Post("/sign_up", signUpUserHandler)
	sessionsRouter.Post("/forgot_password", forgotPasswordHandler)
	sessionsRouter.Post("/reset_password/{token}", resetPasswordHandler)

	router.Mount("/session", sessionsRouter)
}

func buildAudioMessageEndpoints(router chi.Router, svc audio_message.Service) {
	audioMessageRouter := chi.NewRouter()

	createAudioMessageHandler := NewHTTPHandler(audio_message.CreateAudioMessageHandler, svc)
	getAudioMessagesHandler := NewHTTPHandler(audio_message.GetAudioMessagesHandler, svc)
	getAudioMessageByIDHandler := NewHTTPHandler(audio_message.GetAudioMessageByIDHandler, svc)
	updateAudioMesageByIDHandler := NewHTTPHandler(audio_message.UpdateAudioMessagesByIDHandler, svc)
	deleteAudioMesageByIDHandler := NewHTTPHandler(audio_message.DeleteAudioMessagesByIDHandler, svc)

	audioMessageRouter.Get("/", getAudioMessagesHandler)
	audioMessageRouter.Post("/", createAudioMessageHandler)
	audioMessageRouter.Get("/id/message/{message_id}", getAudioMessageByIDHandler)
	audioMessageRouter.Put("/update/{message_id}", updateAudioMesageByIDHandler)
	audioMessageRouter.Delete("/delete/{message_id}", deleteAudioMesageByIDHandler)

	router.Mount("/audio_message", audioMessageRouter)
}

func buildAudioSeriesEndpoints(router chi.Router, svc audio_message.Service) {
	audioSeriesRouter := chi.NewRouter()

	createAudioSeriesHandler := NewHTTPHandler(audio_message.CreateAudioSeriesHandler, svc)
	getAudioSeriesHandler := NewHTTPHandler(audio_message.GetAudioSeriesHandler, svc)
	getAudioSeriesByIDHandler := NewHTTPHandler(audio_message.GetAudioSeriesByIDHandler, svc)
	updateAudioSeriesByIDHandler := NewHTTPHandler(audio_message.UpdateAudioSeriesByIDHandler, svc)
	deleteAudioSeriesByIDHandler := NewHTTPHandler(audio_message.DeleteAudioSeriesByIDHandler, svc)

	audioSeriesRouter.Post("/", createAudioSeriesHandler)
	audioSeriesRouter.Get("/", getAudioSeriesHandler)
	audioSeriesRouter.Get("/id/series/{series_id}", getAudioSeriesByIDHandler)
	audioSeriesRouter.Put("/update/{series_id}", updateAudioSeriesByIDHandler)
	audioSeriesRouter.Delete("/delete/{series_id}", deleteAudioSeriesByIDHandler)

	router.Mount("/audio_series", audioSeriesRouter)
}

func buildUploadEndpoints(router chi.Router, svc uploader.Service) {
	uploadFileHandler := NewHTTPHandler(uploader.UploadFile, svc)
	router.Post("/upload", uploadFileHandler)
}
