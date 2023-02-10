package user

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/errorHandler"
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/security"	
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

//TODO: Better error handling

var (
	ErrQueryRepository       = errors.New("there was an error executing the query")
	ErrUserPwd               = errors.New("invalid login credentials")
	ErrEmptyLoginCredentials = errors.New("invalid login credentials")
	ErrEmailRequired         = errors.New("email is required")
	ErrNotFound              = errors.New("not found")
	ErrUnauthoriedRequest    = errors.New("unauthorized request. please check your credentials")
	ErrUserExists            = errors.New("user with the same email address already exists")
)

type Service interface {
	SignUp(ctx context.Context, user *SignUpUser) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	Login(ctx context.Context, email, password string) (*UserAndToken, error)
	ForgotPassword(request ForgotPasswordPayload) (interface{}, error)
	VerifyPasswordToken(request ResetPasswordPayload, passwordTokenParam string) (string, error)
	ResetPassword(request ResetPasswordPayload) (int, error)
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
				return nil, ErrEmailRequired
			default:
				svc.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	_, err = svc.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		// user exists
		return nil, ErrUserExists
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
				return nil, ErrEmailRequired
			default:
				svc.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	_, err = svc.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		// user exists
		return nil, ErrUserExists
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
		return nil, ErrEmptyLoginCredentials
	}

	// md5 hash prior to sending it to repository
	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	result, err := svc.repo.Login(ctx, email, hashedPassword)

	if err == ErrUserPwd {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "Login"),
			zap.String("error", err.Error()),
		)
		return nil, ErrQueryRepository
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

func randStr(n int, charset []byte, seededRand *rand.Rand) string {
	b := make([]byte, n)
	for i := range b {
		// randomly select 1 character from given charset
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (svc *userService) ForgotPassword(request ForgotPasswordPayload) (interface{}, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	passwordResetToken := EncodeString(randStr(10, charset, seededRand))

	user, err := svc.repo.ForgotPassword(request, passwordResetToken)
	if err != nil {
		return nil, err
	}
	// TODO insert mailer function
	// Temporary return statement pending the mail
	fmt.Println(user)
	return struct {
		URL string `json:"url"`
	}{fmt.Sprint("https://hof-backend.herokuapp.com/user/resetPassword/", passwordResetToken)}, nil
}

func (svc *userService) VerifyPasswordToken(request ResetPasswordPayload, passwordTokenParam string) (string, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	userPasswordToken, err := svc.repo.VerifyPasswordToken(request, passwordTokenParam)
	if err != nil {
		return "", err
	}
	return userPasswordToken, nil
}

func (svc *userService) ResetPassword(request ResetPasswordPayload) (int, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		return 0, err
	}

	if request.Password != request.PasswordConfirm {
		return 0, errorHandler.Format(errorHandler.InvalidRequest, errors.New(errorHandler.Message(errorHandler.InvalidRequest)))
	}

	request.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(request.Password))))

	userId, err := svc.repo.ResetPassword(ResetPasswordPayload{
		Email:           request.Email,
		Password:        request.Password,
		PasswordConfirm: request.PasswordConfirm,
	})
	if err != nil {
		return 0, err
	}
	return userId, nil
}
