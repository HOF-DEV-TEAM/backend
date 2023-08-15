package user

import (
	"bitbucket.org/hofng/hofApp/infrastructure/config"
	"bitbucket.org/hofng/hofApp/infrastructure/library"
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/mailer"
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"math"
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
	ForgotPassword(request ForgotPasswordPayload) error
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
	SendEmailVerificationLink(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context) error
}

type UserService struct {
	repo        Repository
	log         *zap.Logger
	config      *config.ServerConfig
	idGenerator library.IDGenerator
}

func NewService(repo Repository, log *zap.Logger, config *config.ServerConfig) *UserService {
	return &UserService{log: log, repo: repo, config: config, idGenerator: library.NewIDGenerator()}
}

func (s *UserService) validateStruct(v interface{}) error {
	validate := validator.New()

	return validate.Struct(v)
}

func (s *UserService) validateSignUpStruct(user *SignUpUser) error {
	validate := validator.New()

	return validate.Struct(user)
}

func (s *UserService) SignUp(ctx context.Context, user *SignUpUser, devices []Device) (*User, error) {
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

func (s *UserService) CreateUser(ctx context.Context, user *User) (*User, error) {

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

func (s *UserService) ForgotPassword(request ForgotPasswordPayload) error {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		return err
	}

	otpResponse, err := s.repo.ForgotPassword(request)
	if err != nil {
		return err
	}

	messageID, err := s.idGenerator.IDGenerate()
	if err != nil {
		return err
	}

	expirationTime := time.Unix(otpResponse.ExpireTimeInSeconds, 0)
	expiresIn := expirationTime.Sub(time.Now()).Minutes()

	bucketPath := "https://hof-s3.s3.eu-west-2.amazonaws.com/email_template_images"
	message := mailer.Message{
		ID:     messageID,
		Title:  "Password Reset OTP",
		Target: otpResponse.Target,
		DataMap: map[string]string{
			"User":              otpResponse.User,
			"OTP":               otpResponse.OTP,
			"ExpiresIn":         fmt.Sprintf("%v", math.Ceil(expiresIn/5)*5),
			"HofRoundLogo":      fmt.Sprintf("%s/HoF_Logo_White.png", bucketPath),
			"ThisIsHome1":       fmt.Sprintf("%s/home1.jpg", bucketPath),
			"ThisIsHome2":       fmt.Sprintf("%s/home2.jpg", bucketPath),
			"ThisIsHome3":       fmt.Sprintf("%s/home3.jpg", bucketPath),
			"Instagram":         fmt.Sprintf("%s/instagram2x.png", bucketPath),
			"Facebook":          fmt.Sprintf("%s/facebook2x.png", bucketPath),
			"Twitter":           fmt.Sprintf("%s/twitter2x.png", bucketPath),
			"YouTube":           fmt.Sprintf("%s/youtube2x.png", bucketPath),
			"HOFHorizontalLogo": fmt.Sprintf("%s/hof_horizontal_logo.png", bucketPath),
		},
	}
	err = mailer.SendMail(message, s.config.Mailer.PasswordResetMailPath, s.log, &s.config.Mailer)
	if err != nil {
		return err
	}

	// Temporary return statement pending the mail
	return nil
}

func (s *UserService) VerifyPasswordResetOTP(ctx context.Context, request *VerifyOTP) (*UserAndToken, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {

	}
	user, err := s.repo.VerifyPasswordResetOTP(request)
	if err != nil {
		return nil, err
	}

	// recover claims from JWT
	c, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
	if !ok {
		s.log.Info("msg",
			zap.String("JWTError", "broken"),
			zap.String(s.config.Security.JWTContextKey, ""),
		)
	}

	updatedJWTToken, err := c.PutUserIDAndSign(&s.config.Security, user.ID)

	if err != nil {
		return nil, err
	}

	return &UserAndToken{Token: updatedJWTToken}, nil
}

func (s *UserService) ResetPassword(ctx context.Context, request ResetPasswordPayload) (uuid.UUID, error) {
	validate := validator.New()
	err := validate.Struct(request)
	if err != nil {
		return uuid.Nil, err
	}

	if request.Password != request.PasswordConfirm {
		return uuid.Nil, errors.New("compare user password error")
	}

	request.Password = fmt.Sprintf("%x", md5.Sum([]byte(strings.TrimSpace(request.Password))))

	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
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

func (s *UserService) validateFavouriteStruct(audioSeries *Favourites) error {
	validate := validator.New()

	return validate.Struct(audioSeries)
}

func (s *UserService) CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error) {
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
	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
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

func (s *UserService) GetFavourites(ctx context.Context) (GetFavouritesResponse, error) {
	result := GetFavouritesResponse{Favourites: []*FavMessageJSON{}}

	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
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

func (s *UserService) DeleteFavourite(ctx context.Context, messageId string) (uuid.UUID, error) {
	messageID, err := uuid.FromString(messageId)
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
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

func (s *UserService) ChangePassword(ctx context.Context, request ChangePasswordPayload) (uuid.UUID, error) {
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

	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
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

func (s *UserService) UpdateUserProfile(ctx context.Context, user *User) (uuid.UUID, error) {
	claims, ok := ctx.Value(security.JWTClaimsContextKey).(*security.JWTClaim[any])
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

func (s *UserService) BuildDevice(ctx context.Context, input *DeviceManager, email string) (*DeviceManager, error) {
	deviceManager, err := s.repo.BuildDevice(ctx, input, email)
	if err != nil {
		return nil, err
	}

	return deviceManager, nil
}

func (s *UserService) GetDevices(ctx context.Context) (*DeviceManager, error) {
	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
	if !ok {
		return nil, http_helper.ErrInvalidAccount
	}

	devices, err := s.repo.GetDevices(ctx, claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (s *UserService) DeleteDevice(ctx context.Context, identifier string) (string, error) {
	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
	if !ok {
		return "", http_helper.ErrInvalidAccount
	}

	deletedDevice, err := s.repo.DeleteDevice(ctx, identifier, claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return "", err
	}

	return deletedDevice, nil
}

func (s *UserService) UpdateDevice(ctx context.Context, status, identifier string) (*DeviceManager, error) {
	claims, ok := ctx.Value(s.config.Security.JWTClaimsContextKey).(*security.JWTClaim[any])
	if !ok {
		return nil, http_helper.ErrInvalidAccount
	}

	updatedDevice, err := s.repo.UpdateDevice(ctx, claims.JWTClaimsMain.LoggedInUserId, status, identifier)
	if err != nil {
		return nil, err
	}

	return updatedDevice, nil
}

func (s *UserService) UpdateAppVersion(ctx context.Context, version VersionManager) (uuid.UUID, error) {
	version.LastUpdated = sql.NullString{Valid: true, String: time.Now().Format(time.RFC3339)}
	result, err := s.repo.UpdateAppVersion(ctx, version)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}

func (s *UserService) GetAppVersion(ctx context.Context, versionID string) (*VersionManager, error) {
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

func (s *UserService) SendEmailVerificationLink(ctx context.Context, email string) error {
	user, err := s.repo.GetByEmail(ctx, email)

	if err != nil {
		return err
	}

	claims := &security.JWTClaim[security.EmailVerificationClaim]{}

	expiresAt := jwt.NewNumericDate(time.Now().Add(time.Hour * 2))

	verificationCliam := security.EmailVerificationClaim{Type: "email_verification", Email: user.Email, ExpiresAt: expiresAt}
	claims.JWTClaimsMain.Claims = verificationCliam

	tokenString, err := claims.Sign(&s.config.Security, expiresAt)

	if err != nil {
		return err
	}

	messageID, err := s.idGenerator.IDGenerate()
	if err != nil {
		return err
	}

	expiresIn := verificationCliam.ExpiresAt.Sub(time.Now()).Minutes()
	message := mailer.Message{
		ID:     messageID,
		Title:  "Verify Email",
		Target: user.Email,
		DataMap: map[string]string{
			"User":             strings.Title(fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
			"ExpiresIn":        fmt.Sprintf("%v", math.Ceil(expiresIn/5)*5),
			"VerificationLink": fmt.Sprintf("%s/verify_email/%s", s.config.ServerUrl, tokenString),
			"HofRoundLogo":     fmt.Sprintf("%s%sHoF_Logo_White.png", s.config.AwsConfiguration.BaseURL, s.config.AwsConfiguration.BucketPath),
			"ThisIsHome1":      fmt.Sprintf("%s%sthisIsHome.jpeg", s.config.AwsConfiguration.BaseURL, s.config.AwsConfiguration.BucketPath),
		},
	}
	err = mailer.SendMail(message, "verify_email.page.tmpl", s.log, &s.config.Mailer)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) VerifyEmail(ctx context.Context) error {
	//user claim is valid at this point
	claims, ok := ctx.Value(security.JWTClaimsContextKey).(*security.JWTClaim[security.EmailVerificationClaim])

	if !ok {
		s.log.Info("msg",
			zap.String("JWTError", "broken"),
			zap.String(security.JWTClaimsContextKey, ""),
		)
		return errors.New("invalid link")
	}

	claim := claims.JWTClaimsMain.Claims

	if claim.Type != "email_verification" {
		return errors.New("invalid link")
	}

	user, err := s.repo.GetByEmail(ctx, fmt.Sprintf("%s", claim.Email))

	if err != nil {
		return err
	}

	user.IsVerified = EmailVerified
	err = s.repo.UpdateUserIsVerified(ctx, user.ID, EmailVerified)

	if err != nil {
		return err
	}
	return nil
}
