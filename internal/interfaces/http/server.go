// Package http provides the HTTP server lifecycle helpers.
package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appAuth "bitbucket.org/hofng/hofApp/internal/application/auth"
	appContent "bitbucket.org/hofng/hofApp/internal/application/content"
	appSub "bitbucket.org/hofng/hofApp/internal/application/subscription"
	appUser "bitbucket.org/hofng/hofApp/internal/application/user"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/storage"
	"go.uber.org/zap"
)

// Server wraps an http.Server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
	log        *zap.Logger
}

// NewServer builds an HTTP server with all routes wired.
func NewServer(
	port int,
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
) *Server {
	router := NewRouter(jwtSvc, serverURL, templatePath, paystackSecret, authSvc, userSvc, contentSvc, subSvc, fileStorage, log)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
	}
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
func (s *Server) Run() error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("server starting", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		s.log.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	s.log.Info("server stopped gracefully")
	return nil
}
