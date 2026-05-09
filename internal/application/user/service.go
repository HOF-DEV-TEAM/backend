package user

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/mailer"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
)

var validate = validator.New()

// Service exposes all user-management use cases.
type Service interface {
	SignUp(ctx context.Context, req SignUpRequest) (*domainUser.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*domainUser.User, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error

	AssignRoles(ctx context.Context, req AssignRolesRequest) error
	RemoveRoles(ctx context.Context, userID uuid.UUID, roles []string) error
	GetRoles(ctx context.Context, userID uuid.UUID) ([]domainUser.Role, error)
	AdminSignup(ctx context.Context, req AdminSignupRequest) (*domainUser.User, error)
	DeleteAdmin(ctx context.Context, callerID, targetID uuid.UUID) error

	AddFavourite(ctx context.Context, userID uuid.UUID, req AddFavouriteRequest) error
	GetFavourites(ctx context.Context, userID uuid.UUID) ([]domainUser.FavouriteMessage, error)
	DeleteFavourite(ctx context.Context, userID, messageID uuid.UUID) error

	RegisterDevice(ctx context.Context, userID uuid.UUID, device DeviceInput) (*domainUser.DeviceRecord, error)
	GetDevices(ctx context.Context, userID uuid.UUID) (*domainUser.DeviceRecord, error)
	DeleteDevice(ctx context.Context, userID uuid.UUID, identifier string) error
	UpdateDeviceStatus(ctx context.Context, userID uuid.UUID, identifier string, status domainUser.DeviceStatus) error

	GetAppVersion(ctx context.Context, versionID uuid.UUID) (*domainUser.AppVersion, error)
	UpdateAppVersion(ctx context.Context, req AppVersionUpdateRequest) error

	SendEmailVerification(ctx context.Context, email string, serverURL string) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
}

type userService struct {
	repo   domainUser.Repository
	mailer mailer.EmailSender
	jwtSvc *security.JWTService
	log    *zap.Logger
}

// NewService creates the user application service.
func NewService(
	repo domainUser.Repository,
	mailer mailer.EmailSender,
	jwtSvc *security.JWTService,
	log *zap.Logger,
) Service {
	return &userService{
		repo:   repo,
		mailer: mailer,
		jwtSvc: jwtSvc,
		log:    log,
	}
}

// ── Account ───────────────────────────────────────────────────────────────────

func (s *userService) SignUp(ctx context.Context, req SignUpRequest) (*domainUser.User, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("sign up: %w", err)
	}

	u := &domainUser.User{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Password:        hashed,
		PasswordVersion: domainUser.PasswordVersionBcrypt,
		IsVerified:      domainUser.Unverified,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	// Assign default member role.
	if err := s.repo.AssignRoles(ctx, u.ID, []domainUser.RoleName{domainUser.RoleMember}); err != nil {
		s.log.Warn("could not assign default member role", zap.Error(err))
	}

	// Register devices if provided.
	if len(req.Devices) > 0 {
		rec := buildDeviceRecords(u.ID, req.Devices)
		if err := s.repo.UpsertDeviceRecord(ctx, rec); err != nil {
			s.log.Warn("could not register devices during sign up", zap.Error(err))
		}
	}

	return s.repo.GetByID(ctx, u.ID)
}

// AdminSignup creates a new admin user with automatic church_admin role assignment.
func (s *userService) AdminSignup(ctx context.Context, req AdminSignupRequest) (*domainUser.User, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	// Check if admin user already exists
	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		return nil, shared.ErrConflict{Message: "admin user already exists"}
	} else if !shared.IsNotFound(err) {
		s.log.Error("admin signup: lookup failed", zap.Error(err))
		return nil, fmt.Errorf("admin signup: %w", err)
	}

	// Hash password before creating admin user
	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		s.log.Error("admin signup: hashing password", zap.String("email", req.Email), zap.Error(err))
		return nil, fmt.Errorf("admin signup: hashing password: %w", err)
	}

	// Create admin user with church_admin role
	u := &domainUser.User{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Email:           req.Email,
		Password:        hashed,
		PasswordVersion: domainUser.PasswordVersionBcrypt,
		IsVerified:      1, // Auto-verify admin users
	}

	if err := s.repo.Create(ctx, u); err != nil {
		s.log.Error("admin signup: creating user", zap.String("email", req.Email), zap.Error(err))
		return nil, err
	}

	// Set as admin
	if err := s.repo.AssignRoles(ctx, u.ID, []domainUser.RoleName{domainUser.RoleChurchAdmin}); err != nil {
		s.log.Warn("could not assign default member role",
			zap.String("email", req.Email),
			zap.String("user_id", u.ID.String()),
			zap.Error(err))
	}

	s.log.Info("admin user created",
		zap.String("email", req.Email),
		zap.String("name", fmt.Sprintf("%s %s", req.FirstName, req.LastName)),
	)

	return s.repo.GetByID(ctx, u.ID)
}

func (s *userService) DeleteAdmin(ctx context.Context, callerID, targetID uuid.UUID) error {
	if callerID == targetID {
		return shared.ErrForbidden{Message: "you cannot delete your own admin account"}
	}

	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	if !target.HasRole(domainUser.RoleChurchAdmin) {
		return shared.ErrNotFound{Resource: "admin", ID: targetID.String()}
	}

	return s.repo.DeleteUser(ctx, targetID)
}

func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error {
	if err := validate.Struct(req); err != nil {
		return shared.ErrInvalidInput{Message: err.Error()}
	}

	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	u.UserName = req.UserName
	u.FirstName = req.FirstName
	u.LastName = req.LastName
	if req.Mobile != "" {
		u.Mobile = &req.Mobile
	}
	if req.Address != "" {
		u.Address = &req.Address
	}
	if req.Gender != "" {
		u.Gender = &req.Gender
	}

	return s.repo.Update(ctx, u)
}

// ── Password management ───────────────────────────────────────────────────────

func (s *userService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	u, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if shared.IsNotFound(err) {
			// Return no error to avoid email enumeration.
			return nil
		}
		return fmt.Errorf("forgot password: %w", err)
	}

	otp := generateOTP()
	expiry := time.Now().Add(5 * time.Minute).Unix()

	token := &domainUser.PasswordToken{
		Email:              req.Email,
		PasswordResetToken: otp,
		PasswordResetAt:    expiry,
	}
	if err := s.repo.UpsertPasswordToken(ctx, token); err != nil {
		return fmt.Errorf("storing OTP: %w", err)
	}

	if err := s.mailer.SendPasswordReset(req.Email, u.FullName(), otp); err != nil {
		s.log.Error("enqueuing password reset email", zap.Error(err))
	}

	return nil
}

func (s *userService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*domainUser.User, error) {
	token, err := s.repo.GetPasswordToken(ctx, req.Email)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, domainUser.ErrInvalidOTP
		}
		return nil, err
	}

	if token.Validated {
		return nil, domainUser.ErrAlreadyValidatedOTP
	}
	if time.Unix(token.PasswordResetAt, 0).Before(time.Now()) {
		return nil, domainUser.ErrExpiredOTP
	}
	if token.PasswordResetToken != req.OTP {
		return nil, domainUser.ErrInvalidOTP
	}

	if err := s.repo.MarkPasswordTokenValidated(ctx, req.Email); err != nil {
		return nil, fmt.Errorf("marking OTP validated: %w", err)
	}

	return s.repo.GetByEmail(ctx, req.Email)
}

func (s *userService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if req.Password != req.PasswordConfirm {
		return domainUser.ErrPasswordConfirm
	}

	// Require the OTP to have been validated before allowing the password change.
	token, err := s.repo.GetPasswordToken(ctx, req.Email)
	if err != nil || !token.Validated {
		return shared.ErrForbidden{Message: "OTP not verified"}
	}

	u, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, u.ID, hashed, domainUser.PasswordVersionBcrypt)
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return domainUser.ErrPasswordConfirm
	}

	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password.
	if err := security.CheckPasswordBcrypt(u.Password, req.OldPassword); err != nil {
		if security.MD5Hash(req.OldPassword) != u.Password {
			return domainUser.ErrPasswordMismatch
		}
	}

	hashed, err := security.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, userID, hashed, domainUser.PasswordVersionBcrypt)
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (s *userService) AssignRoles(ctx context.Context, req AssignRolesRequest) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return shared.ErrInvalidInput{Field: "user_id", Message: "invalid UUID"}
	}

	names := make([]domainUser.RoleName, len(req.Roles))
	for i, r := range req.Roles {
		names[i] = domainUser.RoleName(r)
	}
	return s.repo.AssignRoles(ctx, userID, names)
}

func (s *userService) RemoveRoles(ctx context.Context, userID uuid.UUID, roles []string) error {
	names := make([]domainUser.RoleName, len(roles))
	for i, r := range roles {
		names[i] = domainUser.RoleName(r)
	}
	return s.repo.RemoveRoles(ctx, userID, names)
}

func (s *userService) GetRoles(ctx context.Context, userID uuid.UUID) ([]domainUser.Role, error) {
	return s.repo.GetRoles(ctx, userID)
}

// ── Favourites ────────────────────────────────────────────────────────────────

func (s *userService) AddFavourite(ctx context.Context, userID uuid.UUID, req AddFavouriteRequest) error {
	rec, err := s.repo.GetFavouriteRecord(ctx, userID)
	if shared.IsNotFound(err) {
		rec = &domainUser.FavouriteRecord{UserID: userID}
	} else if err != nil {
		return err
	}

	item := domainUser.FavouriteItem{
		MessageID: req.MessageID,
		SeriesID:  req.SeriesID,
		Fav:       true,
		DateAdded: time.Now().Format(time.RFC3339),
	}
	rec.Fav = append(rec.Fav, item)
	return s.repo.UpsertFavourite(ctx, rec)
}

func (s *userService) GetFavourites(ctx context.Context, userID uuid.UUID) ([]domainUser.FavouriteMessage, error) {
	return s.repo.GetFavouriteMessages(ctx, userID)
}

func (s *userService) DeleteFavourite(ctx context.Context, userID, messageID uuid.UUID) error {
	return s.repo.DeleteFavouriteItem(ctx, userID, messageID)
}

// ── Devices ───────────────────────────────────────────────────────────────────

// RegisterDevice upserts a device by identifier: if the identifier already exists the
// entry is refreshed (os, version, status=ACTIVE); otherwise a new entry is appended.
// This is safe to call on every login — it will never create duplicate device entries.
func (s *userService) RegisterDevice(ctx context.Context, userID uuid.UUID, input DeviceInput) (*domainUser.DeviceRecord, error) {
	rec, err := s.repo.GetDeviceRecord(ctx, userID)
	if shared.IsNotFound(err) {
		rec = &domainUser.DeviceRecord{UserID: userID}
	} else if err != nil {
		return nil, err
	}

	now := time.Now().Format(time.RFC3339)

	for i, d := range rec.Devices {
		if d.Identifier == input.Identifier {
			// Known device — refresh metadata and mark active.
			rec.Devices[i].Os = input.Os
			rec.Devices[i].Brand = input.Brand
			rec.Devices[i].Version = input.Version
			rec.Devices[i].Status = domainUser.DeviceStatusActive
			rec.Devices[i].LastUpdated = now
			if err := s.repo.UpsertDeviceRecord(ctx, rec); err != nil {
				return nil, err
			}
			return rec, nil
		}
	}

	// New device — append it.
	rec.Devices = append(rec.Devices, domainUser.Device{
		ID:        uuid.NewString(),
		Who:       input.Who,
		Identifier: input.Identifier,
		Os:        input.Os,
		Brand:     input.Brand,
		Version:   input.Version,
		Status:    domainUser.DeviceStatusActive,
		DateAdded: now,
	})

	if err := s.repo.UpsertDeviceRecord(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *userService) GetDevices(ctx context.Context, userID uuid.UUID) (*domainUser.DeviceRecord, error) {
	return s.repo.GetDeviceRecord(ctx, userID)
}

func (s *userService) DeleteDevice(ctx context.Context, userID uuid.UUID, identifier string) error {
	return s.repo.DeleteDevice(ctx, userID, identifier)
}

func (s *userService) UpdateDeviceStatus(ctx context.Context, userID uuid.UUID, identifier string, status domainUser.DeviceStatus) error {
	return s.repo.UpdateDeviceStatus(ctx, userID, identifier, status)
}

// ── App version ───────────────────────────────────────────────────────────────

func (s *userService) GetAppVersion(ctx context.Context, versionID uuid.UUID) (*domainUser.AppVersion, error) {
	return s.repo.GetAppVersion(ctx, versionID)
}

func (s *userService) UpdateAppVersion(ctx context.Context, req AppVersionUpdateRequest) error {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return shared.ErrInvalidInput{Field: "id", Message: "invalid UUID"}
	}

	v, err := s.repo.GetAppVersion(ctx, id)
	if err != nil {
		return err
	}

	v.Version = req.Version
	v.Force = req.Force
	return s.repo.UpdateAppVersion(ctx, v)
}

// ── Email verification ────────────────────────────────────────────────────────

func (s *userService) SendEmailVerification(ctx context.Context, email string, serverURL string) error {
	if err := validate.Var(email, "required,email"); err != nil {
		return shared.ErrInvalidInput{Field: "email", Message: "invalid email"}
	}
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil
		}
		return err
	}

	if s.mailer != nil {
		token, err := s.jwtSvc.IssueEmailVerificationToken(u.ID.String())
		if err != nil {
			return fmt.Errorf("issuing email verification token: %w", err)
		}

		link := fmt.Sprintf("%s/verify_email/%s", serverURL, token)

		if err := s.mailer.SendEmailVerification(u.Email, u.FullName(), link); err != nil {
			s.log.Error("enqueuing email verification", zap.Error(err))
		}
	} else {
		s.log.Warn("mailer not configured, skipping email verification", zap.String("email", u.Email))
	}

	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	return s.repo.UpdateVerificationStatus(ctx, userID, domainUser.EmailVerified)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func generateOTP() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.Itoa(100000 + rng.Intn(900000))
}

func buildDeviceRecords(userID uuid.UUID, inputs []DeviceInput) *domainUser.DeviceRecord {
	seen := make(map[string]bool, len(inputs))
	devices := make(domainUser.DeviceList, 0, len(inputs))
	now := time.Now().Format(time.RFC3339)
	for _, input := range inputs {
		if seen[input.Identifier] {
			continue // deduplicate within the signup payload
		}
		seen[input.Identifier] = true
		devices = append(devices, domainUser.Device{
			ID:        uuid.NewString(),
			Who:       input.Who,
			Identifier: input.Identifier,
			Os:        input.Os,
			Brand:     input.Brand,
			Version:   input.Version,
			Status:    domainUser.DeviceStatusActive,
			DateAdded: now,
		})
	}
	return &domainUser.DeviceRecord{UserID: userID, Devices: devices}
}
