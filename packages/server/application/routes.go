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
	"bitbucket.org/hofng/hofApp/pkg/auth"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/subscription/paystack"
	"bitbucket.org/hofng/hofApp/pkg/uploader"
	"bitbucket.org/hofng/hofApp/pkg/user"
	"github.com/go-chi/chi/v5"
)

func (app *application) buildRoutes() {
	app.router.Handle("/swagger/*", httpSwagger.WrapHandler)

	userRepo := user.NewRepository(app.db, app.logger)
	userService := user.NewService(userRepo, app.logger, &app.config.Security)

	//Subscription
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

	authService := auth.NewService(userRepo, subscriptionSvc, app.logger, &app.config.Security)

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
		buildSessionEndpoints(r, authService, userService)
		//webhook

		subEvent := subscription.NewSubEvent(userRepo, subscritpionRepo, app.logger)
		createSubscriptionHookHandler := subscription.CreateSubscriptionHookHandler(subEvent)
		r.Post("/subscription/webhook", createSubscriptionHookHandler)
	})

}

func buildUserEndpoints(router chi.Router, svc user.Service) {
	userRouter := chi.NewRouter()
	updateUserProfileHandler := user.UpdateUserProfileHandler(svc)

	favRouter := buildFavEndpoints(svc)
	deviceRouter := buildDeviceEndpoints(svc)
	appVersionRouter := buildAppVersionEndpoints(svc)

	resetPasswordHandler := user.ResetPasswordHandler(svc)
	changePasswordHandler := user.ChangePasswordHandler(svc)

	router.Route("/user", func(r chi.Router) {
		r.Mount("/favourite", favRouter)
		r.Mount("/devices", deviceRouter)
		r.Mount("/app_version", appVersionRouter)
		r.Mount("/", userRouter)
		r.Post("/reset_password", resetPasswordHandler)
		r.Post("/change_password", changePasswordHandler)
		r.Post("/update", updateUserProfileHandler)
	})
}

func buildSessionEndpoints(router chi.Router, authSvc auth.Service, userSvc user.Service) {
	sessionsRouter := chi.NewRouter()

	signInHandler := auth.SignInHandler(authSvc)
	signUpUserHandler := user.GetUserHandler(userSvc)
	forgotResetPasswordHandler := user.ForgotPasswordHandler(userSvc)
	verifyResetPasswordOTPHandler := user.VerifyPasswordResetOTPHandler(userSvc)
	authenticateHandler := auth.AuthenticateHandler(authSvc)
	buildDevicesHandler := user.BuildDeviceHandler(userSvc)

	sessionsRouter.Post("/authenticate", authenticateHandler)
	sessionsRouter.Post("/sign_in", signInHandler)
	sessionsRouter.Post("/sign_up", signUpUserHandler)
	sessionsRouter.Post("/forgot_password", forgotResetPasswordHandler)
	sessionsRouter.Put("/verify_token", verifyResetPasswordOTPHandler)
	// sessionsRouter.Post("/authenticate/{token}", resetPasswordHandler)
	sessionsRouter.Post("/device/{email}", buildDevicesHandler)

	router.Mount("/session", sessionsRouter)
}

func buildAudioMessageEndpoints(router chi.Router, svc audio_message.Service) {
	audioMessageRouter := chi.NewRouter()

	createAudioMessageHandler := audio_message.CreateAudioMessageHandler(svc)
	getAudioMessagesHandler := audio_message.GetAudioMessagesHandler(svc)
	getAudioMessageByIDHandler := audio_message.GetAudioMessageByIDHandler(svc)
	updateAudioMesageByIDHandler := audio_message.UpdateAudioMessagesByIDHandler(svc)
	deleteAudioMesageByIDHandler := audio_message.DeleteAudioMessagesByIDHandler(svc)
	createMeditationHandler := audio_message.CreateMeditationHandler(svc)
	updateMeditationByIDHandler := audio_message.UpdateMeditationByIDHandler(svc)

	audioMessageRouter.Get("/", getAudioMessagesHandler)
	audioMessageRouter.Post("/", createAudioMessageHandler)
	audioMessageRouter.Get("/id/message/{message_id}", getAudioMessageByIDHandler)
	audioMessageRouter.Put("/update/{message_id}", updateAudioMesageByIDHandler)
	audioMessageRouter.Delete("/delete/{message_id}", deleteAudioMesageByIDHandler)
	audioMessageRouter.Post("/meditation", createMeditationHandler)
	audioMessageRouter.Put("/meditation/{meditation_id}", updateMeditationByIDHandler)

	router.Mount("/audio_message", audioMessageRouter)
}

func buildAudioSeriesEndpoints(router chi.Router, svc audio_message.Service) {
	audioSeriesRouter := chi.NewRouter()

	createAudioSeriesHandler := audio_message.CreateAudioSeriesHandler(svc)
	getAudioSeriesHandler := audio_message.GetAudioSeriesHandler(svc)
	getAudioSeriesByIDHandler := audio_message.GetAudioSeriesByIDHandler(svc)
	updateAudioSeriesByIDHandler := audio_message.UpdateAudioSeriesByIDHandler(svc)
	deleteAudioSeriesByIDHandler := audio_message.DeleteAudioSeriesByIDHandler(svc)
	homepageHandler := audio_message.HomePageDirectoryHandler(svc)

	audioSeriesRouter.Post("/", createAudioSeriesHandler)
	audioSeriesRouter.Get("/", getAudioSeriesHandler)
	audioSeriesRouter.Get("/id/series/{series_id}", getAudioSeriesByIDHandler)
	audioSeriesRouter.Put("/update/{series_id}", updateAudioSeriesByIDHandler)
	audioSeriesRouter.Delete("/delete/{series_id}", deleteAudioSeriesByIDHandler)
	audioSeriesRouter.Get("/home", homepageHandler)

	router.Mount("/audio_series", audioSeriesRouter)
}

func buildUploadEndpoints(router chi.Router, svc uploader.Service) {
	uploadFileHandler := uploader.UploadFile(svc)
	router.Post("/upload", uploadFileHandler)
}

func buildSubscriptionEndpoints(router chi.Router, svc subscription.Service) {
	subRouter := chi.NewRouter()

	createSubscriptionPlanHandler := subscription.CreateSubscriptionPlanHandler(svc)
	createSubscriptionOfferingHandler := subscription.CreateSubscriptionOfferingHandler(svc)
	getSubscriptionPlanOfferings := subscription.GetSubscriptionPlanOfferingsHandler(svc)
	createSubscritionPlanOfferings := subscription.CreateSubscriptionPlanOfferingHandler(svc)
	verifySubscriptionHandler := subscription.VerifySubscriptionHandler(svc)

	subRouter.Post("/plan", createSubscriptionPlanHandler)
	subRouter.Post("/verify", verifySubscriptionHandler)

	subRouter.Post("/offering", createSubscriptionOfferingHandler)
	subRouter.Get("/plan/offering", getSubscriptionPlanOfferings)
	subRouter.Post("/plan/offering", createSubscritionPlanOfferings)

	router.Mount("/subscription", subRouter)
}

func buildFavEndpoints(svc user.Service) http.Handler {
	favRouter := chi.NewRouter()
	createFavouriteHandler := user.CreateFavouriteHandler(svc)
	getAllFavouritesHandler := user.GetFavouritesHandler(svc)
	deleteFavouriteHandler := user.DeleteFavouritesHandler(svc)

	favRouter.Post("/", createFavouriteHandler)
	favRouter.Get("/favs", getAllFavouritesHandler)
	favRouter.Delete("/delete/{message_id}", deleteFavouriteHandler)

	return favRouter
}

func buildDeviceEndpoints(svc user.Service) http.Handler {
	deviceRouter := chi.NewRouter()
	getAllDevicesHandler := user.GetDevicesHandler(svc)
	deleteDeviceHandler := user.DeleteDeviceHandler(svc)
	updateDeviceHandler := user.UpdateDeviceHandler(svc)

	deviceRouter.Get("/all", getAllDevicesHandler)
	deviceRouter.Delete("/delete/{identifier}", deleteDeviceHandler)
	deviceRouter.Put("/update/{identifier}/{status}", updateDeviceHandler)
	return deviceRouter
}

func buildAppVersionEndpoints(svc user.Service) http.Handler {
	appVersionRouter := chi.NewRouter()
	updateAppVersionHandler := user.UpdateAppVersion(svc)
	getAppVersionHandler := user.GetAppVersion(svc)

	appVersionRouter.Put("/admin/update", updateAppVersionHandler)
	appVersionRouter.Get("/version/{version_id}", getAppVersionHandler)

	return appVersionRouter
}
