package user

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
	"strings"

	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Service interface {
	SignUp(ctx context.Context, user *SignUpUser) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	Login(ctx context.Context, email, password string) (*UserAndToken, error)
	ForgotPassword(request ForgotPasswordPayload) (*OTPResponse, error)
	VerifyPasswordResetOTP(request *VerifyOTP) (*UserAndToken, error)
	ResetPassword(ctx context.Context, request ResetPasswordPayload) (uuid.UUID, error)
}

type userService struct {
	repo   Repository
	log    *zap.Logger
	config *security.SecurityConfig
}

func NewService(repo Repository, log *zap.Logger, config *security.SecurityConfig) Service {
	return &userService{log: log, repo: repo, config: config}
}

func (s *userService) validateStruct(user *User) error {
	validate := validator.New()

	return validate.Struct(user)
}

func (s *userService) validateSignUpStruct(user *SignUpUser) error {
	validate := validator.New()

	return validate.Struct(user)
}

func (svc *userService) SignUp(ctx context.Context, user *SignUpUser) (*User, error) {

	err := svc.validateSignUpStruct(user)

	if err != nil {
		tErr, ok := err.(validator.ValidationErrors)

		if !ok {
			return nil, fmt.Errorf("unknown validation error")
		}

		for _, e := range tErr {
			switch e.StructField() {
			case "Email":
				return nil, http_helper.ErrEmailRequired
			default:
				svc.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	_, err = svc.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		// user exists
		return nil, http_helper.ErrUserExists
	}

	// leading and trailing whitespaces
	password := fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(user.Password))))

	result, err := svc.repo.Create(
		ctx,
		&User{
			Email:     user.Email,
			Password:  password,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "Create"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	return result, nil
}

func (svc *userService) CreateUser(ctx context.Context, user *User) (*User, error) {

	err := svc.validateStruct(user)

	if err != nil {
		tErr, ok := err.(validator.ValidationErrors)

		if !ok {
			return nil, fmt.Errorf("unknown validation error")
		}

		for _, e := range tErr {
			switch e.StructField() {
			case "Email":
				return nil, http_helper.ErrEmailRequired
			default:
				svc.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	_, err = svc.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		// user exists
		return nil, http_helper.ErrUserExists
	}

	// leading and trailing whitespaces
	user.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(user.Password))))

	result, err := svc.repo.Create(ctx, user)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "Create"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	return result, nil
}

func (svc *userService) Login(ctx context.Context, email, password string) (*UserAndToken, error) {
	err := validator.New().Struct(LoginUser{
		Email:    email,
		Password: password,
	})

	// If either Email or Password field is empty
	if err != nil {
		return nil, http_helper.ErrEmptyLoginCredentials
	}

	// md5 hash prior to sending it to repository
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	result, err := svc.repo.Login(ctx, email, hashedPassword)

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

	// recover claims from JWT
	claims, ok := ctx.Value(svc.config.JWTClaimsContextKey).(*security.JWTClaim)

	if !ok {
		svc.log.Info("msg",
			zap.String("JWTError", "broken"),
			zap.String(svc.config.JWTContextKey, ""),
		)
	}

	updatedJWTToken, err := claims.PutUserIDAndSign(svc.config, result.ID)

	if err != nil {
		return nil, err
	}

	return &UserAndToken{User: result, Token: updatedJWTToken}, nil
}

func (svc *userService) ForgotPassword(request ForgotPasswordPayload) (*OTPResponse, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	otpResponse, err := svc.repo.ForgotPassword(request)
	if err != nil {
		return nil, err
	}

	// TODO insert mailer function

	// Temporary return statement pending the mail
	return otpResponse, nil
}

func (svc *userService) VerifyPasswordResetOTP(request *VerifyOTP) (*UserAndToken, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	user, err := svc.repo.VerifyPasswordResetOTP(request)
	if err != nil {
		return nil, err
	}

	var c *security.JWTClaim
	updatedJWTToken, err := c.PutUserIDAndSign(svc.config, user.ID)

	if err != nil {
		return nil, err
	}

	return &UserAndToken{Token: updatedJWTToken}, nil
}

func (svc *userService) ResetPassword(ctx context.Context, request ResetPasswordPayload) (uuid.UUID, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		return uuid.Nil, err
	}

	if request.Password != request.PasswordConfirm {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}

	request.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(request.Password))))

	claims, ok := ctx.Value(svc.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}

	userId, err := svc.repo.ResetPassword(ResetPasswordPayload{
		ID:              claims.JWTClaimsMain.LoggedInUserId,
		Password:        request.Password,
		PasswordConfirm: request.PasswordConfirm,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}
