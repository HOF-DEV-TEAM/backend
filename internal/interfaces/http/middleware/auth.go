// Package middleware provides JWT authentication helpers for HTTP routes.
package middleware

import (
	"context"
	"net/http"

	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
	"bitbucket.org/hofng/hofApp/internal/interfaces/http/response"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "authenticated_user_id"

// Authenticate is an HTTP middleware that enforces JWT authentication.
// On success it stores the authenticated user's UUID in the request context.
func Authenticate(jwtSvc *security.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := security.ClaimsFromContext(r.Context())
			if !ok {
				response.Unauthorized(w)
				return
			}

			// Reject non-access tokens (e.g., email_verify, refresh) for API access
			if claims.Type != "" && claims.Type != "access" {
				response.Unauthorized(w)
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				response.Unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext retrieves the authenticated user ID placed by the Authenticate middleware.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

// RequireAdmin is middleware that ensures the authenticated user has admin role.
// Must be used after Authenticate middleware.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := security.ClaimsFromContext(r.Context())
		if !ok {
			response.Unauthorized(w)
			return
		}

		if !claims.IsAdmin() {
			response.Forbidden(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole is middleware that ensures the authenticated user has the specified role.
// Must be used after Authenticate middleware.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := security.ClaimsFromContext(r.Context())
			if !ok {
				response.Unauthorized(w)
				return
			}

			if !claims.HasRole(role) {
				response.Forbidden(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole is middleware that ensures the authenticated user has any of the specified roles.
// Must be used after Authenticate middleware.
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := security.ClaimsFromContext(r.Context())
			if !ok {
				response.Unauthorized(w)
				return
			}

			if !claims.HasAnyRole(roles...) {
				response.Forbidden(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
