package user

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"strings"

	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var ErrFieldRequired = errors.New("field is required")

type Service interface {
	SignUp(ctx context.Context, user *SignUpUser) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	Login(ctx context.Context, email, password string) (*UserAndToken, error)
	ForgotPassword(request ForgotPasswordPayload) (*OTPResponse, error)
	VerifyPasswordResetOTP(ctx context.Context, request *VerifyOTP) (*UserAndToken, error)
	ResetPassword(ctx context.Context, request ResetPasswordPayload) (uuid.UUID, error)
	ChangePassword(ctx context.Context, request ChangePasswordPayload) (uuid.UUID, error)
	CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error)
	GetFavourites(ctx context.Context) (GetFavouritesResponse, error)
	DeleteFavourite(ctx context.Context, favId string) (uuid.UUID, error)
}

type userService struct {
	repo   Repository
	log    *zap.Logger
	config *security.SecurityConfig
}

func NewService(repo Repository, log *zap.Logger, config *security.SecurityConfig) Service {
	return &userService{log: log, repo: repo, config: config}
}

func (s *userService) validateStruct(v interface{}) error {
	validate := validator.New()

	return validate.Struct(v)
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

func (svc *userService) VerifyPasswordResetOTP(ctx context.Context, request *VerifyOTP) (*UserAndToken, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	user, err := svc.repo.VerifyPasswordResetOTP(request)
	if err != nil {
		return nil, err
	}

	// recover claims from JWT
	c, ok := ctx.Value(svc.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		svc.log.Info("msg",
			zap.String("JWTError", "broken"),
			zap.String(svc.config.JWTContextKey, ""),
		)
	}

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
		return uuid.Nil, errors.New("compare user password error")
	}

	request.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(request.Password))))

	claims, ok := ctx.Value(svc.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}
	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return uuid.Nil, err
	}

	resp, err := svc.repo.ResetPassword(ctx, userId, ResetPasswordPayload{
		Email:           request.Email,
		Password:        request.Password,
		PasswordConfirm: request.PasswordConfirm,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return resp, nil
}

func (s *userService) validateFavouriteStruct(audioSeries *Favourites) error {
	validate := validator.New()

	return validate.Struct(audioSeries)
}

func (s *userService) CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error) {
	err := s.validateStruct(favourite)
	if err != nil {
		tErr, ok := err.(validator.ValidationErrors)

		if !ok {
			return nil, fmt.Errorf("unknown validation error")
		}

		for _, e := range tErr {
			switch e.StructField() {
			case "UserID":
				return nil, ErrFieldRequired
			default:
				s.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return nil, http_helper.ErrInvalidAccount
	}
	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return nil, err
	}

	favourite.UserID = userId
	result, err := s.repo.CreateFavourite(ctx, favourite)
	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		s.log.Error("msg",
			zap.String("method", "CreateFavourite"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}
	return result, nil
}

func (s *userService) GetFavourites(ctx context.Context) (GetFavouritesResponse, error) {
	result := GetFavouritesResponse{}

	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return result, http_helper.ErrInvalidAccount
	}

	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return result, err
	}

	fav, count, err := s.repo.GetFavourites(ctx, userId)
	if err == sql.ErrNoRows {
		return result, err
	}

	result.Favourites = []*FavMessageJSON{}

	for _, as := range fav {
		result.Favourites = append(result.Favourites, NewJSONFavMessage(as))
	}

	result.Pagination = PageResponse{
		TotalResults: int32(count),
	}

	return result, nil
}

func (s *userService) DeleteFavourite(ctx context.Context, messageId string) (uuid.UUID, error) {
	messageID, err := uuid.FromString(messageId)
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}

	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return uuid.Nil, err
	}
	result, err := s.repo.DeleteFavourite(ctx, messageID, userId)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}

func (s *userService) ChangePassword(ctx context.Context, request ChangePasswordPayload) (uuid.UUID, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		return uuid.Nil, err
	}

	if request.NewPassword != request.ConfirmNewPassword {
		return uuid.Nil, errors.New("compare user password error")
	}

	request.OldPassword = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(request.OldPassword))))
	request.NewPassword = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(request.NewPassword))))

	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}
	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return uuid.Nil, err
	}

	resp, err := s.repo.ChangePassword(ctx, userId, ChangePasswordPayload{
		Email:              request.Email,
		OldPassword:        request.OldPassword,
		NewPassword:        request.NewPassword,
		ConfirmNewPassword: request.ConfirmNewPassword,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return resp, nil
}
