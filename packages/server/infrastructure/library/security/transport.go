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
			http_helper.EncodeJSONError(ctx, ErrUnauthorized, w)
			return
		}

		claims := &JWTClaim{}
		token, err := jwt.ParseWithClaims((tokenString), claims, func(t *jwt.Token) (interface{}, error) {			
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {							
				return nil, ErrUnauthorized
			}
			
			return []byte(config.JWTSecret), nil
		})

		if err != nil {
			if e, ok := err.(*jwt.ValidationError); ok {
				switch {
				case e.Errors&jwt.ValidationErrorMalformed !=0:
					//Token is malformed
					http_helper.EncodeJSONError(ctx, ErrUnauthorized, w)
					return
				case e.Errors&jwt.ValidationErrorExpired !=0:
					//Token is expired
					http_helper.EncodeJSONError(ctx, ErrUnauthorized, w)
					return
				case e.Errors&jwt.ValidationErrorNotValidYet !=0:					
					//Token is not active yet
					http_helper.EncodeJSONError(ctx, ErrUnauthorized, w)
					return
				case e.Inner != nil:
					http_helper.EncodeJSONError(ctx, e.Inner, w)
					return
				}
			}
		}
		if !token.Valid {			
			http_helper.EncodeJSONError(ctx, ErrUnauthorized, w)
			return
		}

		newCtx := context.WithValue(ctx, config.JWTClaimsContextKey, claims)
		
		// Token is authenticated, pass it through
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func (v *JWTClaim) PutUserIDAndSign(config *SecurityConfig, userId int) (string, error){
	v.JWTClaimsMain.LoggedInUserId = userId
	return v.Sign(config)
}
