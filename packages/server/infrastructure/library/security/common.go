package security

import (
	"errors"
	"time"

	"github.com/go-chi/jwtauth"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type SecurityConfig struct {
	JWTKeyString 	string `env:"JWT_SIGNING_KEY"`
	JWTExpiration 	time.Duration
	TokenAuth		*jwtauth.JWTAuth
}


type JWTClaim struct {
	JWTClaims 	jwtClaims	`json:"v1"`	
}

type jwtClaims struct {
	Type 			string 	`json:"type"`
	LoggedInUserId 	int 	`json:"userId"`
}


func (s *SecurityConfig) PutUserIDAndSign(claims map[string]interface{}, userId int) (string, error) {
	claims["user_id"] = userId

	jwtauth.SetIssuedNow(claims)
	jwtauth.SetExpiry(claims, time.Now().Add(time.Hour * 2))
	_, tokenString, err := s.TokenAuth.Encode(claims)	


	
	if err != nil {
		return "", err
	}
	return tokenString, nil 	
}