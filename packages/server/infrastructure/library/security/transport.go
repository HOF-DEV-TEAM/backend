package security

import (
	"context"
	"net/http"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"github.com/golang-jwt/jwt"
)

//TODO: change this into a callback that returns a middlware
func (config *SecurityConfig) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString,  err := config.FromContext(ctx)
		if err != nil {
			http_helper.EncodeJSONError(ctx, http_helper.ErrUnauthorized, w)
			return
		}

		claims := &JWTClaim{}
		token, err := jwt.ParseWithClaims((tokenString), claims, func(t *jwt.Token) (interface{}, error) {			
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {							
				return nil, http_helper.ErrUnauthorized
			}
			
			return []byte(config.JWTSecret), nil
		})

		if err != nil {
			http_helper.EncodeJSONError(ctx, err, w)	
			return	
		}

		if !token.Valid {			
			http_helper.EncodeJSONError(ctx, http_helper.ErrUnauthorized, w)
			return
		}

		newCtx := context.WithValue(ctx, config.JWTClaimsContextKey, claims)
		
		// Token is authenticated, pass it through
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func (v *JWTClaim) PutUserIDAndSign(config *SecurityConfig, userId string) (string, error){
	v.JWTClaimsMain.LoggedInUserId = userId
	return v.Sign(config)
}


