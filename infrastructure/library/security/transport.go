package security

import (
	"context"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"github.com/golang-jwt/jwt"
)

func (config *SecurityConfig) ValidateJWT(tokenString string) (*jwt.Token, JWTClaim[any], error) {
	claims := &JWTClaim[any]{}
	token, err := jwt.ParseWithClaims((tokenString), claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http_helper.ErrUnauthorized
		}

		return []byte(config.JWTSecret), nil
	})
	return token, *claims, err
}

func (claims *JWTClaim[T]) Parse(ctx context.Context, config *SecurityConfig) error {
	tokenString, err := config.FromContext(ctx)
	if err != nil {
		return http_helper.ErrUnauthorized
	}

	token, err := jwt.ParseWithClaims((tokenString), claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http_helper.ErrUnauthorized
		}

		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		return http_helper.ErrTokenInvalid
	}

	if !token.Valid {
		return http_helper.ErrUnauthorized
	}
	return nil
}

func (config *SecurityConfig) AuthenticateVerifyEmail(next http.Handler) (fn http.Handler) {
	fn = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims := &JWTClaim[EmailVerificationClaim]{}

		err := claims.Parse(ctx, config)

		//Handle error in service
		if err != nil {
			next.ServeHTTP(w, r)
		}

		newCtx := context.WithValue(ctx, JWTClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
	return
}

func (config *SecurityConfig) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims := &JWTClaim[any]{}

		err := claims.Parse(ctx, config)
		if err != nil {
			http_helper.EncodeJSONError(ctx, err, w)
			return
		}

		newCtx := context.WithValue(ctx, JWTClaimsContextKey, claims)
		// Token is authenticated, pass it through
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}
