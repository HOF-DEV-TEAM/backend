//go:build integration

package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	appAuth "bitbucket.org/hofng/hofApp/internal/application/auth"
	appContent "bitbucket.org/hofng/hofApp/internal/application/content"
	appSub "bitbucket.org/hofng/hofApp/internal/application/subscription"
	appUser "bitbucket.org/hofng/hofApp/internal/application/user"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/config"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/database"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/persistence"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
	httpServer "bitbucket.org/hofng/hofApp/internal/interfaces/http"
	"go.uber.org/zap"
)

// testServer spins up a real HTTP server backed by the test PostgreSQL database.
// It runs all migrations before each test suite.
func testServer(t *testing.T) *httptest.Server {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping integration tests")
	}

	log, _ := zap.NewDevelopment()

	dbCfg := &config.DatabaseConfig{URL: dsn}
	db, err := database.Connect(dbCfg, log)
	if err != nil {
		t.Fatalf("connect to test DB: %v", err)
	}
	if err := database.RunMigrations(db, "./migrations", log); err != nil {
		// Migrations path relative to repo root; try from project root
		if err2 := database.RunMigrations(db, "../../../migrations", log); err2 != nil {
			t.Fatalf("run migrations: %v (also tried relative path: %v)", err, err2)
		}
	}

	jwtKey := os.Getenv("JWT_SIGNING_KEY")
	if jwtKey == "" {
		jwtKey = "integration-test-key-not-for-production"
	}

	jwtSvc := security.NewJWTService(jwtKey)
	userRepo := persistence.NewUserRepository(db, log)
	contentRepo := persistence.NewContentRepository(db, log)
	subRepo := persistence.NewSubscriptionRepository(db, log)

	authSvc := appAuth.NewService(userRepo, subRepo, jwtSvc, log)
	userSvc := appUser.NewService(userRepo, nil, jwtSvc, log)
	contentSvc := appContent.NewService(contentRepo, log)
	subSvc := appSub.NewService(subRepo, nil, userRepo, log)

	router := httpServer.NewRouter(jwtSvc, "http://localhost", "", authSvc, userSvc, contentSvc, subSvc, nil, log)
	return httptest.NewServer(router)
}

// ── Health check ──────────────────────────────────────────────────────────────

func TestIntegration_HealthCheck(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// ── Sign up + Sign in flow ────────────────────────────────────────────────────

func TestIntegration_SignUp_SignIn(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	email := fmt.Sprintf("integration_%d@example.com", time.Now().UnixNano())

	// Sign up
	body := map[string]interface{}{
		"first_name": "Integration",
		"last_name":  "Test",
		"email":      email,
		"password":   "pass123456",
	}
	respSignUp := postJSON(t, srv.URL+"/session/sign_up", body)
	if respSignUp.StatusCode != http.StatusOK && respSignUp.StatusCode != http.StatusCreated {
		t.Errorf("sign_up status = %d, want 200/201", respSignUp.StatusCode)
	}

	var signUpResult struct {
		Success bool `json:"success"`
		Data    struct {
			Email string `json:"Email"`
		} `json:"data"`
	}
	decodeJSON(t, respSignUp, &signUpResult)
	if !signUpResult.Success {
		t.Fatal("sign_up returned success:false")
	}

	// Sign in
	loginBody := map[string]string{"email": email, "password": "pass123456"}
	respLogin := postJSON(t, srv.URL+"/session/sign_in", loginBody)
	if respLogin.StatusCode != http.StatusOK {
		t.Errorf("sign_in status = %d, want 200", respLogin.StatusCode)
	}

	var loginResult struct {
		Success bool `json:"success"`
		Data    struct {
			Token            string `json:"token"`
			GlobalParameters struct {
				ActivateSubscription bool `json:"activate_subscription"`
			} `json:"global_parameters"`
		} `json:"data"`
	}
	decodeJSON(t, respLogin, &loginResult)
	if !loginResult.Success {
		t.Fatal("sign_in returned success:false")
	}
	if loginResult.Data.Token == "" {
		t.Error("expected non-empty token")
	}

	// GlobalParameters must be present in session
	// (it defaults to true on first read via seeding)
	t.Logf("activate_subscription = %v", loginResult.Data.GlobalParameters.ActivateSubscription)
}

func TestIntegration_SignIn_InvalidCredentials(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	body := map[string]string{"email": "nobody@example.com", "password": "wrongpassword"}
	resp := postJSON(t, srv.URL+"/session/sign_in", body)

	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("sign_in with bad creds: status = %d, want 401 or 400", resp.StatusCode)
	}
}

// ── Protected route requires auth ─────────────────────────────────────────────

func TestIntegration_ProtectedRoute_NoAuth(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/user/roles")
	if err != nil {
		t.Fatalf("GET /user/roles: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

// ── Webhook always returns 200 ────────────────────────────────────────────────

func TestIntegration_Webhook_AlwaysOK(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", `{}`},
		{"charge.success", `{"event":"charge.success","data":{}}`},
		{"unknown event", `{"event":"unknown.event","data":{}}`},
		{"bad JSON", `not-json`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Post(
				srv.URL+"/subscription/webhook",
				"application/json",
				bytes.NewBufferString(tt.body),
			)
			if err != nil {
				t.Fatalf("POST /subscription/webhook: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("webhook status = %d, want 200", resp.StatusCode)
			}
		})
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func postJSON(t *testing.T, url string, body interface{}) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
