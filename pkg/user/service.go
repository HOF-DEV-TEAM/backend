package user

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library"
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"strings"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var ErrFieldRequired = errors.New("field is required")

type Service interface {
	SignUp(ctx context.Context, user *SignUpUser, devices []Device) (*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
	ForgotPassword(request ForgotPasswordPayload) (*OTPResponse, error)
	VerifyPasswordResetOTP(ctx context.Context, request *VerifyOTP) (*UserAndToken, error)
	ResetPassword(ctx context.Context, request ResetPasswordPayload) (uuid.UUID, error)
	ChangePassword(ctx context.Context, request ChangePasswordPayload) (uuid.UUID, error)
	CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error)
	GetFavourites(ctx context.Context) (GetFavouritesResponse, error)
	DeleteFavourite(ctx context.Context, favId string) (uuid.UUID, error)
	UpdateUserProfile(ctx context.Context, user *User) (uuid.UUID, error)
	BuildDevice(ctx context.Context, input *DeviceManager, email string) (*DeviceManager, error)
	GetDevices(ctx context.Context) (*DeviceManager, error)
	DeleteDevice(ctx context.Context, identifier string) (string, error)
	UpdateDevice(ctx context.Context, status, identifier string) (*DeviceManager, error)
	UpdateAppVersion(ctx context.Context, version VersionManager) (uuid.UUID, error)
	GetAppVersion(ctx context.Context, versionID string) (*VersionManager, error)
}

type userService struct {
	repo        Repository
	log         *zap.Logger
	config      *security.SecurityConfig
	idGenerator library.IDGenerator
}

func NewService(repo Repository, log *zap.Logger, config *security.SecurityConfig) Service {
	return &userService{log: log, repo: repo, config: config, idGenerator: library.NewIDGenerator()}
}

func (s *userService) validateStruct(v interface{}) error {
	validate := validator.New()

	return validate.Struct(v)
}

func (s *userService) validateSignUpStruct(user *SignUpUser) error {
	validate := validator.New()

	return validate.Struct(user)
}

func (s *userService) SignUp(ctx context.Context, user *SignUpUser, devices []Device) (*User, error) {
	err := s.validateSignUpStruct(user)

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
				s.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	_, err = s.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		// user exists
		return nil, http_helper.ErrUserExists
	}

	// leading and trailing whitespaces
	password := fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(user.Password))))

	u := &User{
		Email:     user.Email,
		Password:  password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	deviceManager := &DeviceManager{Devices: devices}

	result, err := s.repo.SignUpUser(ctx, u, deviceManager)

	if err != nil {
		s.log.Error("msg",
			zap.String("method", "Create"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	return result, nil
}

func (s *userService) CreateUser(ctx context.Context, user *User) (*User, error) {

	err := s.validateStruct(user)

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
				s.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	_, err = s.repo.GetByEmail(ctx, user.Email)
	if err == nil {
		// user exists
		return nil, http_helper.ErrUserExists
	}

	// leading and trailing whitespaces
	user.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(user.Password))))

	result, err := s.repo.CreateUser(ctx, user)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		s.log.Error("msg",
			zap.String("method", "Create"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	return result, nil
}

func (s *userService) ForgotPassword(request ForgotPasswordPayload) (*OTPResponse, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	otpResponse, err := s.repo.ForgotPassword(request)
	if err != nil {
		return nil, err
	}

	// TODO insert mailer function

	// Temporary return statement pending the mail
	return otpResponse, nil
}

func (s *userService) VerifyPasswordResetOTP(ctx context.Context, request *VerifyOTP) (*UserAndToken, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	user, err := s.repo.VerifyPasswordResetOTP(request)
	if err != nil {
		return nil, err
	}

	// recover claims from JWT
	c, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		s.log.Info("msg",
			zap.String("JWTError", "broken"),
			zap.String(s.config.JWTContextKey, ""),
		)
	}

	updatedJWTToken, err := c.PutUserIDAndSign(s.config, user.ID)

	if err != nil {
		return nil, err
	}

	return &UserAndToken{Token: updatedJWTToken}, nil
}

func (s *userService) ResetPassword(ctx context.Context, request ResetPasswordPayload) (uuid.UUID, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		return uuid.Nil, err
	}

	if request.Password != request.PasswordConfirm {
		return uuid.Nil, errors.New("compare user password error")
	}

	request.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(request.Password))))

	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}
	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return uuid.Nil, err
	}

	resp, err := s.repo.ResetPassword(ctx, userId, ResetPasswordPayload{
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
	result := GetFavouritesResponse{Favourites: []*FavMessageJSON{}}

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
		return result, nil
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

func (s *userService) UpdateUserProfile(ctx context.Context, user *User) (uuid.UUID, error) {
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}

	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return uuid.Nil, err
	}

	user.LastUpdated = sql.NullString{
		Valid:  true,
		String: time.Now().Format(time.RFC3339),
	}
	updateUser := UpdateUser{
		UserName:    user.UserName,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Mobile:      user.Mobile.String,
		Address:     user.Address,
		Gender:      user.Gender,
		LastUpdated: user.LastUpdated.String,
	}
	result, err := s.repo.UpdateUserProfile(ctx, userId, &updateUser)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}

func (s *userService) BuildDevice(ctx context.Context, input *DeviceManager, email string) (*DeviceManager, error) {
	deviceManager, err := s.repo.BuildDevice(ctx, input, email)
	if err != nil {
		return nil, err
	}

	return deviceManager, nil
}

func (s *userService) GetDevices(ctx context.Context) (*DeviceManager, error) {
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return nil, http_helper.ErrInvalidAccount
	}

	devices, err := s.repo.GetDevices(ctx, claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (s *userService) DeleteDevice(ctx context.Context, identifier string) (string, error) {
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return "", http_helper.ErrInvalidAccount
	}

	deletedDevice, err := s.repo.DeleteDevice(ctx, identifier, claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return "", err
	}

	return deletedDevice, nil
}

func (s *userService) UpdateDevice(ctx context.Context, status, identifier string) (*DeviceManager, error) {
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return nil, http_helper.ErrInvalidAccount
	}

	updatedDevice, err := s.repo.UpdateDevice(ctx, claims.JWTClaimsMain.LoggedInUserId, status, identifier)
	if err != nil {
		return nil, err
	}

	return updatedDevice, nil
}

func (s *userService) UpdateAppVersion(ctx context.Context, version VersionManager) (uuid.UUID, error) {
	version.LastUpdated = sql.NullString{Valid: true, String: time.Now().Format(time.RFC3339)}
	result, err := s.repo.UpdateAppVersion(ctx, version)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}

func (s *userService) GetAppVersion(ctx context.Context, versionID string) (*VersionManager, error) {
	id, err := uuid.FromString(versionID)
	if err != nil {
		return nil, err
	}

	appVersion, err := s.repo.GetAppVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	return appVersion, nil

}
