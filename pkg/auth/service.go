package auth

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/user"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type Service interface {
	Login(ctx context.Context, email, password, deviceIdentifier string) (*UserSession, error)
	Authenticate(ctx context.Context, token, refreshToken string) (*UserSession, error)
}

type authService struct {
	userRepo   user.Repository
	subService subscription.Service
	log        *zap.Logger
	config     *security.SecurityConfig
}

func NewService(userRepo user.Repository, subService subscription.Service, log *zap.Logger, config *security.SecurityConfig) Service {
	return &authService{log: log, userRepo: userRepo, subService: subService, config: config}
}

func (svc *authService) createSession(ctx context.Context, user *user.User) (*UserSession, error) {
	sub, err := svc.subService.GetSubscription(ctx, user.ID)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// recover claims from JWT
	claims, ok := ctx.Value(svc.config.JWTClaimsContextKey).(*security.JWTClaim)

	if !ok {
		svc.log.Info("msg",
			zap.String("JWTError", "broken"),
			zap.String(svc.config.JWTContextKey, ""),
		)
	}

	updatedJWTToken, err := claims.PutUserIDAndSign(svc.config, user.ID)

	if err != nil {
		return nil, err
	}

	refreshToken, err := claims.CreateRefreshToken(svc.config)

	if err != nil {
		return nil, err
	}

	var subJSON *subscription.SubscriptionJSON

	if sub != nil {
		subJSON = sub.ToJSON()
	}

	return &UserSession{
		User:         user.ToJSON(),
		Subscription: subJSON,
		Token:        updatedJWTToken,
		RefreshToken: refreshToken,
	}, nil
}

func (svc *authService) Authenticate(ctx context.Context, authToken, refeshToken string) (*UserSession, error) {
	if authToken == "" {
		return nil, http_helper.ErrNoTokenFound
	}

	//validate refresh token
	_, authCliams, err := svc.config.ValidateJWT(authToken)

	if e, ok := err.(*jwt.ValidationError); ok {
		if e.Errors != jwt.ValidationErrorClaimsInvalid {
			return nil, err
		}
	}

	//validate refresh token
	_, claims, err := svc.config.ValidateJWT(refeshToken)
	if err != nil {
		return nil, err
	}

	//compare claims - for now only the userId
	userId := claims.JWTClaimsMain.LoggedInUserId

	if userId != authCliams.JWTClaimsMain.LoggedInUserId {
		return nil, http_helper.ErrInvalidAccount
	}

	user, err := svc.userRepo.GetById(ctx, userId)

	if err != nil {
		return nil, err
	}
	return svc.createSession(ctx, user)
}

func (svc *authService) Login(ctx context.Context, email, password, deviceIdentifier string) (*UserSession, error) {
	err := validator.New().Struct(LoginUser{
		Email:            email,
		Password:         password,
		DeviceIdentifier: deviceIdentifier,
	})

	// If either Email or Password field is empty
	if err != nil {
		return nil, http_helper.ErrEmptyLoginCredentials
	}

	// md5 hash prior to sending it to repository
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	result, err := svc.userRepo.Login(ctx, email, hashedPassword, deviceIdentifier)

	if err == http_helper.ErrUserPwd {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "Login"),
			zap.String("error", err.Error()),
		)
		return nil, http_helper.ErrQueryRepository
	}

	return svc.createSession(ctx, result)
}
