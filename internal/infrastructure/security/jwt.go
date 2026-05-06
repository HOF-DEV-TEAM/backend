// Package security provides JWT signing/verification and password hashing utilities.
package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
)

const (
	accessTokenTTL      = 48 * time.Hour
	refreshTokenTTL     = 30 * 24 * time.Hour
	emailVerifyTokenTTL = 24 * time.Hour

	contextKeyToken  = contextKey("jwt_token")
	contextKeyClaims = contextKey("jwt_claims")
)

type contextKey string

// Claims is the standard JWT payload for this application.
type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles,omitempty"` // User roles for authorization
	Type   string   `json:"typ,omitempty"`   // "access", "refresh", "email_verify"
	jwt.RegisteredClaims
}

// JWTService signs and verifies JWTs for the application.
type JWTService struct {
	signingKey []byte
}

// NewJWTService creates a JWTService with the provided HMAC signing key.
func NewJWTService(signingKey string) *JWTService {
	return &JWTService{signingKey: []byte(signingKey)}
}

// IssueAccessToken signs a short-lived access token for userID.
func (s *JWTService) IssueAccessToken(userID string) (string, error) {
	return s.sign(userID, nil, accessTokenTTL, "access")
}

// IssueAccessTokenWithRoles signs a short-lived access token for userID with roles.
func (s *JWTService) IssueAccessTokenWithRoles(userID string, roles []string) (string, error) {
	return s.sign(userID, roles, accessTokenTTL, "access")
}

// IssueRefreshToken signs a long-lived refresh token for userID.
func (s *JWTService) IssueRefreshToken(userID string) (string, error) {
	return s.sign(userID, nil, refreshTokenTTL, "refresh")
}

// IssueEmailVerificationToken signs a 24-hour token for email verification links.
// The token includes a "typ: email_verify" claim that middleware.Authenticate rejects
// for API access, limiting blast radius if the verification link is intercepted.
func (s *JWTService) IssueEmailVerificationToken(userID string) (string, error) {
	return s.sign(userID, nil, emailVerifyTokenTTL, "email_verify")
}

func (s *JWTService) sign(userID string, roles []string, ttl time.Duration, typ string) (string, error) {
	claims := Claims{
		UserID: userID,
		Roles:  roles,
		Type:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return signed, nil
}

// Parse validates a token string and returns the embedded claims.
func (s *JWTService) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, shared.ErrUnauthorized{Message: fmt.Sprintf("unexpected signing method: %v", t.Header["alg"])}
		}
		return s.signingKey, nil
	})
	if err != nil {
		return nil, shared.ErrUnauthorized{Message: "invalid or expired token"}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, shared.ErrUnauthorized{Message: "invalid token"}
	}
	return claims, nil
}

// ClaimsFromContext retrieves JWT claims stored in ctx by the auth middleware.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(contextKeyClaims).(*Claims)
	return c, ok
}

// WithClaims returns a new context containing the given claims.
func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, contextKeyClaims, c)
}

// HasRole checks if the claims contain the specified role.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the claims contain any of the specified roles.
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has admin privileges.
func (c *Claims) IsAdmin() bool {
	return c.HasRole("church_admin")
}

// Middleware attaches JWT claims to the context when a valid Bearer token is present.
// It is non-blocking: requests without a token pass through untouched.
// Use middleware.Authenticate on protected route groups to enforce presence.
func (s *JWTService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr != "" {
			if claims, err := s.Parse(tokenStr); err == nil {
				r = r.WithContext(WithClaims(r.Context(), claims))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// PathTokenMiddleware extracts the token from a URL path parameter named "token".
func (s *JWTService) PathTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path-based tokens are embedded in the URL, e.g. /verify_email/{token}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) == 0 {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[len(parts)-1]

		claims, err := s.Parse(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := WithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
	}

	// Fallback: check cookie named "jwt"
	if cookie, err := r.Cookie("jwt"); err == nil {
		return cookie.Value
	}
	return ""
}
