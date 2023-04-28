package security

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
	
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNoTokenFound = errors.New("no token found")
)

type SecurityConfig struct {
	JWTSecret 		string `env:"JWT_SECRET"`
	JWTKeyString 	string `env:"JWT_SIGNING_KEY"`
	JWTContextKey	string
	JWTClaimsContextKey string
	JWTExpiration 	time.Duration	
}

type JWTClaim struct {
	JWTClaimsMain jwtClaims `json:"v1"`
	jwt.RegisteredClaims
}

type jwtClaims struct {
	LoggedInUserId 	string 	`json:"userId"`
	Claims 			string 	`json:"claims"`
}

func (v *JWTClaim) Sign(config *SecurityConfig) (string, error) {	
	v.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt: jwt.NewNumericDate(jwt.TimeFunc()),
	}
	
	token := jwt.NewWithClaims((jwt.SigningMethodHS256), v)		
	return token.SignedString([]byte(config.JWTSecret))
}

//TODO: validate approach for this longer lived token - ideally this should come from DB
func (v *JWTClaim) CreateRefreshToken(config *SecurityConfig) (string, error) {	
	v.RegisteredClaims = jwt.RegisteredClaims{		
		IssuedAt: jwt.NewNumericDate(jwt.TimeFunc()),
	}
	
	token := jwt.NewWithClaims((jwt.SigningMethodHS256), v)		
	return token.SignedString([]byte(config.JWTSecret))
}


func TokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// TokenFromHeader tries to retreive the token string from the
// "Authorization" reqeust header: "Authorization: BEARER T".
func TokenFromHeader(r *http.Request) string {
	// Get token from authorization header.
	bearer := r.Header.Get("Authorization")	
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}


func (config *SecurityConfig) FromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value(config.JWTContextKey).(string)

	if !ok {
		return "", ErrUnauthorized
	}
	return token, nil
}

func (config *SecurityConfig) Verifier() func(http.Handler) http.Handler {
	return config.Verify(TokenFromHeader, TokenFromCookie)
}

func (config *SecurityConfig) Verify(findTokenFns ...func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			token, err := VerifyRequest(r, findTokenFns...)
			ctx = config.NewContext(ctx, token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func (config *SecurityConfig) NewContext(ctx context.Context, t string, err error) context.Context {
	ctx = context.WithValue(ctx, config.JWTContextKey, t)	
	return ctx
}

func (config *SecurityConfig) AddClaimToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {		
		ctx := context.WithValue(r.Context(), config.JWTClaimsContextKey, &JWTClaim{})
		next.ServeHTTP(w, r.WithContext(ctx))	
	})
}

func VerifyRequest(r *http.Request, findTokenFns ...func(r *http.Request) string) (string, error) {
	var tokenString string
	for _, fn := range findTokenFns {
		tokenString = fn(r)
		if tokenString != "" {
			break
		}
	}
	if tokenString == "" {
		return "", ErrNoTokenFound
	}
	return tokenString, nil
}
