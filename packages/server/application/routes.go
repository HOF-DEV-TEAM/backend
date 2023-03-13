package application

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "bitbucket.org/hofng/hofApp/docs"
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/pkg/audio_message"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/subscription/paystack"
	"bitbucket.org/hofng/hofApp/pkg/user"

	"bitbucket.org/hofng/hofApp/pkg/uploader"
	"github.com/go-chi/chi/v5"
)

func (app *application) buildRoutes() {
	app.router.Handle("/swagger/*", httpSwagger.WrapHandler)

	userRepo := user.NewRepository(app.db, app.logger)
	userService := user.NewService(userRepo, app.logger, &app.config.Security)

	subscritpionRepo := subscription.NewRepository(app.db, app.logger)

	subProvider := paystack.NewPaystackService(
		paystack.NewPayStackHttpClient(
			&app.config.PaystackConfig,
			http_helper.NewHTTPCaller(),
			app.logger,
		),
		userRepo,
		subscritpionRepo,
		&app.config.Security,
	)

	subscriptionSvc := subscription.NewService(subProvider, subscritpionRepo, &app.config.Security, userRepo)
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
		buildSubscriptionEndpoints(r, subscriptionSvc)
	})

	//unprotected routes
	app.router.Group(func(r chi.Router) {
		buildSessionEndpoints(r, userService)
		//webhook

		subEvent := subscription.NewSubEvent(userRepo, subscritpionRepo, app.logger)
		createSubscriptionHookHandler := subscription.CreateSubscriptionHookHandler(subEvent)
		r.Post("/subscription/webhook", createSubscriptionHookHandler)
	})

}

func buildUserEndpoints(router chi.Router, svc user.Service) {
	userRouter := chi.NewRouter()
	router.Mount("/user", userRouter)
}

func buildSessionEndpoints(router chi.Router, svc user.Service) {
	sessionsRouter := chi.NewRouter()

	signInHandler := user.SignInHandler(svc)
	signUpUserHandler := user.GetUserHandler(svc)
	forgotResetPasswordHandler := user.ForgotPasswordHandler(svc)
	verifyResetPasswordOTPHandler := user.VerifyPasswordResetOTPHandler(svc)
	resetPasswordHandler := user.ResetPasswordHandler(svc)

	// authenticateHandler := user.AuthenticateHandler(svc)

	sessionsRouter.Post("/sign_in", signInHandler)
	sessionsRouter.Post("/sign_up", signUpUserHandler)
	sessionsRouter.Post("/forgot_password", forgotResetPasswordHandler)
	sessionsRouter.Put("/verify_token", verifyResetPasswordOTPHandler)
	sessionsRouter.Post("/reset_password/{token}", resetPasswordHandler)
	// sessionsRouter.Post("/authenticate/{token}", resetPasswordHandler)

	router.Mount("/session", sessionsRouter)
}

func buildAudioMessageEndpoints(router chi.Router, svc audio_message.Service) {
	audioMessageRouter := chi.NewRouter()

	createAudioMessageHandler := audio_message.CreateAudioMessageHandler(svc)
	getAudioMessagesHandler := audio_message.GetAudioMessagesHandler(svc)
	getAudioMessageByIDHandler := audio_message.GetAudioMessageByIDHandler(svc)
	updateAudioMesageByIDHandler := audio_message.UpdateAudioMessagesByIDHandler(svc)
	deleteAudioMesageByIDHandler := audio_message.DeleteAudioMessagesByIDHandler(svc)

	audioMessageRouter.Get("/", getAudioMessagesHandler)
	audioMessageRouter.Post("/", createAudioMessageHandler)
	audioMessageRouter.Get("/id/message/{message_id}", getAudioMessageByIDHandler)
	audioMessageRouter.Put("/update/{message_id}", updateAudioMesageByIDHandler)
	audioMessageRouter.Delete("/delete/{message_id}", deleteAudioMesageByIDHandler)

	router.Mount("/audio_message", audioMessageRouter)
}

func buildAudioSeriesEndpoints(router chi.Router, svc audio_message.Service) {
	audioSeriesRouter := chi.NewRouter()

	createAudioSeriesHandler := audio_message.CreateAudioSeriesHandler(svc)
	getAudioSeriesHandler := audio_message.GetAudioSeriesHandler(svc)
	getAudioSeriesByIDHandler := audio_message.GetAudioSeriesByIDHandler(svc)
	updateAudioSeriesByIDHandler := audio_message.UpdateAudioSeriesByIDHandler(svc)
	deleteAudioSeriesByIDHandler := audio_message.DeleteAudioSeriesByIDHandler(svc)

	audioSeriesRouter.Post("/", createAudioSeriesHandler)
	audioSeriesRouter.Get("/", getAudioSeriesHandler)
	audioSeriesRouter.Get("/id/series/{series_id}", getAudioSeriesByIDHandler)
	audioSeriesRouter.Put("/update/{series_id}", updateAudioSeriesByIDHandler)
	audioSeriesRouter.Delete("/delete/{series_id}", deleteAudioSeriesByIDHandler)

	router.Mount("/audio_series", audioSeriesRouter)
}

func buildUploadEndpoints(router chi.Router, svc uploader.Service) {
	uploadFileHandler := uploader.UploadFile(svc)
	router.Post("/upload", uploadFileHandler)
}

func buildSubscriptionEndpoints(router chi.Router, svc subscription.Service) {
	subRouter := chi.NewRouter()

	createSubscriptionHandler := subscription.CreateSubscriptionHandler(svc)
	createSubscriptionPlanHandler := subscription.CreateSubscriptionPlanHandler(svc)
	createSubscriptionOfferingHandler := subscription.CreateSubscriptionOfferingHandler(svc)
	getSubscriptionPlanOfferings := subscription.GetSubscriptionPlanOfferingsHandler(svc)
	createSubscritionPlanOfferings := subscription.CreateSubscriptionPlanOfferingHandler(svc)
	verifySubscriptionHandler := subscription.VerifySubscriptionHandler(svc)

	subRouter.Post("/", createSubscriptionHandler)
	// subRouter.Get("/", createSubscriptionHandler)
	subRouter.Post("/plan", createSubscriptionPlanHandler)
	subRouter.Get("/verify/{ref_id}", verifySubscriptionHandler)

	subRouter.Post("/offering", createSubscriptionOfferingHandler)
	subRouter.Get("/plan/offering", getSubscriptionPlanOfferings)
	subRouter.Post("/plan/offering", createSubscritionPlanOfferings)

	router.Mount("/subscription", subRouter)
}
