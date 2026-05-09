// @title           HOF Backend API
// @version         2.0
// @description     Heritage of Faith Church — audio content and subscription platform.
// @termsOfService  https://hofng.org/terms
// @contact.name    HOF Dev Team
// @contact.email   dev@hofng.org
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// Host and scheme are set dynamically at startup from SERVER_URL — see NewRouter.
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Type "Bearer" followed by a space and the JWT access token.
package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	appAuth "bitbucket.org/hofng/hofApp/internal/application/auth"
	appContent "bitbucket.org/hofng/hofApp/internal/application/content"
	appSub "bitbucket.org/hofng/hofApp/internal/application/subscription"
	appUser "bitbucket.org/hofng/hofApp/internal/application/user"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/database"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/logger"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/mailer"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/payment/paystack"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/persistence"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/storage"
	httpServer "bitbucket.org/hofng/hofApp/internal/interfaces/http"
	"go.uber.org/zap"
)

func main() {
	// ── Load .env (silently ignored in production where env vars are injected) ─
	_ = godotenv.Overload() // .env always wins in dev; no-op if file absent (production)

	// ── Logger ────────────────────────────────────────────────────────────────
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}
	zapLog, err := logger.New(appEnv)
	if err != nil {
		log.Fatalf("initializing logger: %v", err)
	}
	defer func() {
		_ = zapLog.Sync()
	}()

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load(zapLog)
	if err != nil {
		zapLog.Fatal("loading config", zap.Error(err))
	}

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := database.Connect(&cfg.Database, zapLog)
	if err != nil {
		zapLog.Fatal("connecting to database", zap.Error(err))
	}

	if err := database.RunMigrations(db, "./migrations", zapLog); err != nil {
		zapLog.Fatal("running migrations", zap.Error(err))
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := persistence.NewUserRepository(db, zapLog)
	contentRepo := persistence.NewContentRepository(db, zapLog)
	subRepo := persistence.NewSubscriptionRepository(db, zapLog)

	// ── Infrastructure services ───────────────────────────────────────────────
	jwtSvc := security.NewJWTService(cfg.Security.JWTSigningKey)
	mailSvc := mailer.New(&cfg.Mailer, zapLog)
	emailQueue := mailer.NewEmailQueue(mailSvc, db, zapLog)
	emailQueue.Start(context.Background())

	var fileStorage storage.Storage
	if storageSvc, err := storage.NewStorage(cfg, zapLog); err != nil {
		zapLog.Warn("Storage unavailable — file uploads disabled", zap.Error(err))
	} else {
		fileStorage = storageSvc
	}

	paystackSvc := paystack.NewService(paystack.NewClient(cfg.Paystack, zapLog), zapLog)

	// ── Application services ──────────────────────────────────────────────────
	authSvc := appAuth.NewService(userRepo, subRepo, jwtSvc, zapLog)
	userSvc := appUser.NewService(userRepo, emailQueue, jwtSvc, zapLog)
	contentSvc := appContent.NewService(contentRepo, zapLog)
	subSvc := appSub.NewService(subRepo, paystackSvc, userRepo, zapLog)

	// ── HTTP server ───────────────────────────────────────────────────────────
	srv := httpServer.NewServer(
		cfg.HTTPPort,
		jwtSvc,
		cfg.ServerURL,
		cfg.Mailer.TemplatePath,
		cfg.Paystack.Secret,
		authSvc,
		userSvc,
		contentSvc,
		subSvc,
		fileStorage,
		zapLog,
	)

	if err := srv.Run(); err != nil {
		zapLog.Fatal("server error", zap.Error(err))
	}
}
