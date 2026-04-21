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
