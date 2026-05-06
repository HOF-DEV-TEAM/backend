// Package http wires the application HTTP routes and middleware.
package http

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	docs "bitbucket.org/hofng/hofApp/docs"
	appAuth "bitbucket.org/hofng/hofApp/internal/application/auth"
	appContent "bitbucket.org/hofng/hofApp/internal/application/content"
	appSub "bitbucket.org/hofng/hofApp/internal/application/subscription"
	appUser "bitbucket.org/hofng/hofApp/internal/application/user"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/storage"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/handler"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/middleware"
)

// NewRouter wires up all routes and returns a ready-to-use Chi router.
func NewRouter(
	jwtSvc *security.JWTService,
	serverURL string,
	templatePath string,
	paystackSecret string,
	authSvc appAuth.Service,
	userSvc appUser.Service,
	contentSvc appContent.Service,
	subSvc appSub.Service,
	fileStorage storage.Storage,
	log *zap.Logger,
) http.Handler {
	// ── Swagger host/scheme — set dynamically from SERVER_URL ────────────────
	if parsed, err := url.Parse(serverURL); err == nil && parsed.Host != "" {
		docs.SwaggerInfo.Host = parsed.Host
		if parsed.Scheme != "" {
			docs.SwaggerInfo.Schemes = []string{parsed.Scheme}
		}
	}

	r := chi.NewRouter()

	// ── Global middleware ─────────────────────────────────────────────────────
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	// Attach JWT claims to context on every request (non-blocking — no 401 yet).
	r.Use(jwtSvc.Middleware)

	// ── API Docs ──────────────────────────────────────────────────────────────
	// Scalar UI (modern): /docs
	r.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1"/>
  <title>HOF Backend API</title>
  <style>body{margin:0}</style>
</head>
<body>
  <script
    id="api-reference"
    data-url="/swagger/doc.json"
    data-configuration='{"theme":"purple","layout":"modern","defaultHttpClient":{"targetKey":"shell","clientKey":"curl"}}'
  ></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`))
	})
	// Raw Swagger UI (legacy): /swagger/*
	r.Handle("/swagger/*", httpSwagger.WrapHandler)

	// ── Health check ──────────────────────────────────────────────────────────
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("HOF Backend — running"))
	})
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// ── Handlers ──────────────────────────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc, log, serverURL, templatePath, func(token string) (uuid.UUID, error) {
		claims, err := jwtSvc.Parse(token)
		if err != nil {
			return uuid.UUID{}, err
		}
		return uuid.Parse(claims.UserID)
	})
	contentH := handler.NewContentHandler(contentSvc)
	subH := handler.NewSubscriptionHandler(subSvc, paystackSecret, log)
	uploadH := handler.NewUploadHandler(fileStorage)
	adminH := handler.NewAdminHandler(subSvc)

	// ── Email verification ────────────────────────────────────────────────────
	// GET  /verify_email/{token} — browser link from the verification email.
	//   The handler owns all JWT validation and renders a branded HTML page.
	// POST /session/send_verify_email — public; call right after sign-up.
	r.Get("/verify_email/{token}", userH.VerifyEmail)

	// ── Public session routes ─────────────────────────────────────────────────
	r.Route("/session", func(r chi.Router) {
		// admin
		r.Post("/sign_up/admin", userH.AdminSignup)
		r.Post("/sign_in/admin", authH.AdminSignIn)

		r.Post("/sign_up", userH.SignUp)
		r.Post("/sign_in", authH.SignIn)
		r.Post("/authenticate", authH.Authenticate)

		r.Post("/forgot_password", userH.ForgotPassword)
		r.Put("/verify_token", userH.VerifyOTP)
		r.Post("/send_verify_email", userH.SendEmailVerification)
	})

	// ── Paystack webhook (public, verified by Paystack HMAC signature) ──────────
	r.Post("/subscription/webhook", subH.PaystackWebhook)

	// ── Protected routes ──────────────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSvc))

		// User management
		r.Route("/user", func(r chi.Router) {
			r.Post("/update", userH.UpdateProfile)
			r.Post("/reset_password", userH.ResetPassword)
			r.Post("/change_password", userH.ChangePassword)

			// Roles
			r.Get("/roles", userH.GetRoles)

			// Favourites
			r.Route("/favourite", func(r chi.Router) {
				r.Post("/", userH.AddFavourite)
				r.Get("/favs", userH.GetFavourites)
				r.Delete("/delete/{message_id}", userH.DeleteFavourite)
			})

			// Devices
			r.Route("/devices", func(r chi.Router) {
				r.Get("/all", userH.GetDevices)
				r.Post("/add", userH.RegisterDevice)
				r.Delete("/delete/{identifier}", userH.DeleteDevice)
				r.Put("/update/{identifier}/{status}", userH.UpdateDeviceStatus)
			})

			// App version
			r.Route("/app_version", func(r chi.Router) {
				r.Get("/version/{version_id}", userH.GetAppVersion)
			})
		})

		// Audio messages
		r.Route("/audio_message", func(r chi.Router) {
			r.Get("/", contentH.ListMessages)
			r.Post("/", contentH.CreateMessage)
			r.Get("/id/message/{message_id}", contentH.GetMessage)
			r.Put("/update/{message_id}", contentH.UpdateMessage)
			r.Delete("/delete/{message_id}", contentH.DeleteMessage)

			r.Post("/meditation", contentH.CreateMeditation)
			r.Get("/meditations", contentH.ListMeditations)
			r.Get("/meditation/{meditation_id}", contentH.GetMeditation)
			r.Put("/meditation/{meditation_id}", contentH.UpdateMeditation)
			r.Delete("/meditation/delete/{meditation_id}", contentH.DeleteMeditation)
		})

		// Audio series
		r.Route("/audio_series", func(r chi.Router) {
			r.Get("/", contentH.ListSeries)
			r.Post("/", contentH.CreateSeries)
			r.Get("/id/series/{series_id}", contentH.GetSeries)
			r.Put("/update/{series_id}", contentH.UpdateSeries)
			r.Delete("/delete/{series_id}", contentH.DeleteSeries)
			r.Get("/home", contentH.GetHomepage)
		})

		// Subscriptions
		r.Route("/subscription", func(r chi.Router) {
			r.Get("/", subH.ListSubscriptions)
			r.Post("/verify", subH.VerifySubscription)
			r.Delete("/disable/{code}", subH.DisableSubscription)
			r.Post("/transaction", subH.InitializeTransaction)

			r.Route("/plan", func(r chi.Router) {
				r.Get("/", subH.ListPlans)
				r.Post("/", subH.CreatePlan)
				r.Get("/{id}", subH.GetPlan)
				r.Delete("/{id}", subH.DeletePlan)
				r.Get("/offering", subH.ListPlanOfferings)
				r.Post("/offering", subH.CreatePlanOffering)
			})

			r.Route("/offering", func(r chi.Router) {
				r.Get("/", subH.ListOfferings)
				r.Post("/", subH.CreateOffering)
				r.Delete("/delete/{offering_id}", subH.DeleteOffering)
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSvc))
		r.Use(middleware.RequireAdmin)

		// Admin
		r.Route("/admin", func(r chi.Router) {

			// User management
			r.Post("/user/roles", userH.AssignRoles)
			r.Put("/user/app_version/update", userH.UpdateAppVersion)

			// File upload
			r.Route("/upload", func(r chi.Router) {
				r.Post("/", uploadH.UploadFile)                   // Server-side upload
				r.Get("/presigned", uploadH.GeneratePresignedURL) // Direct browser-to-S3 upload
			})

			// Global parameters
			r.Get("/global", adminH.GetGlobalParameters)
			r.Put("/global", adminH.UpdateGlobalParameters)
		})
	})

	return r
}
