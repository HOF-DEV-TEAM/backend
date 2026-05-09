package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	domainSub "bitbucket.org/hofng/hofApp/internal/domain/subscription"
	domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"
	"bitbucket.org/hofng/hofApp/internal/infrastructure/security"
)

// Service handles authentication operations (login, token refresh).
type Service interface {
	Login(ctx context.Context, req LoginRequest) (*SessionResponse, error)
	AdminLogin(ctx context.Context, req AdminLoginRequest) (*SessionResponse, error)
	Authenticate(ctx context.Context, req AuthenticateRequest) (*SessionResponse, error)
}

type authService struct {
	userRepo domainUser.Repository
	subRepo  domainSub.Repository
	jwtSvc   *security.JWTService
	log      *zap.Logger
}

// NewService constructs the authentication application service.
func NewService(
	userRepo domainUser.Repository,
	subRepo domainSub.Repository,
	jwtSvc *security.JWTService,
	log *zap.Logger,
) Service {
	return &authService{
		userRepo: userRepo,
		subRepo:  subRepo,
		jwtSvc:   jwtSvc,
		log:      log,
	}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*SessionResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, shared.ErrUnauthorized{Message: "invalid email or password"}
		}
		return nil, fmt.Errorf("login: %w", err)
	}

	if err := s.checkPassword(ctx, u, req.Password); err != nil {
		return nil, err
	}

	// Upsert device on every login so the list stays current.
	// Non-fatal: a device error must not block the session.
	if req.Device != nil && req.Device.Identifier != "" {
		s.upsertDevice(ctx, u, req.Device)
	}

	return s.buildSession(ctx, u)
}

func (s *authService) AdminLogin(ctx context.Context, req AdminLoginRequest) (*SessionResponse, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, shared.ErrUnauthorized{Message: "invalid email or password"}
		}
		return nil, fmt.Errorf("admin login: %w", err)
	}

	if err := s.checkPassword(ctx, u, req.Password); err != nil {
		return nil, err
	}

	return s.buildSession(ctx, u)
}

func (s *authService) Authenticate(ctx context.Context, req AuthenticateRequest) (*SessionResponse, error) {
	// The refresh token must be fully valid (not expired).
	refreshClaims, err := s.jwtSvc.Parse(req.RefreshToken)
	if err != nil {
		return nil, shared.ErrUnauthorized{Message: "invalid refresh token"}
	}

	// Parse the access token (ignore expiry; only check user ID consistency).
	accessClaims, _ := s.jwtSvc.Parse(req.Token)
	if accessClaims != nil && accessClaims.UserID != refreshClaims.UserID {
		return nil, shared.ErrUnauthorized{Message: "token user mismatch"}
	}

	uid, err := uuid.Parse(refreshClaims.UserID)
	if err != nil {
		return nil, shared.ErrUnauthorized{Message: "malformed user id in token"}
	}

	u, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, shared.ErrUnauthorized{Message: "user not found"}
		}
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return s.buildSession(ctx, u)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (s *authService) checkPassword(ctx context.Context, u *domainUser.User, plaintext string) error {
	switch u.PasswordVersion {
	case domainUser.PasswordVersionBcrypt, "":
		if err := security.CheckPasswordBcrypt(u.Password, plaintext); err != nil {
			return shared.ErrUnauthorized{Message: "invalid email or password"}
		}
	case domainUser.PasswordVersionMD5:
		// Legacy MD5 path — upgrade the hash on successful login.
		if security.MD5Hash(plaintext) != u.Password {
			return shared.ErrUnauthorized{Message: "invalid email or password"}
		}
		hashed, err := security.HashPassword(plaintext)
		if err != nil {
			s.log.Error("upgrading legacy password hash", zap.Error(err))
		} else {
			if upgradeErr := s.userRepo.UpdatePassword(ctx, u.ID, hashed, domainUser.PasswordVersionBcrypt); upgradeErr != nil {
				s.log.Warn("failed to upgrade legacy password hash", zap.Error(upgradeErr))
			}
		}
	}
	return nil
}

func (s *authService) buildSession(ctx context.Context, u *domainUser.User) (*SessionResponse, error) {
	// Extract user roles for JWT claims
	roleNames := make([]string, len(u.Roles))
	for i := range u.Roles {
		roleNames[i] = string(u.Roles[i].Name)
	}

	accessToken, err := s.jwtSvc.IssueAccessTokenWithRoles(u.ID.String(), roleNames)
	if err != nil {
		return nil, fmt.Errorf("issuing access token: %w", err)
	}

	refreshToken, err := s.jwtSvc.IssueRefreshToken(u.ID.String())
	if err != nil {
		return nil, fmt.Errorf("issuing refresh token: %w", err)
	}

	subDTO := s.resolveSubscription(ctx, u)
	globalParams := s.resolveGlobalParameters(ctx)

	return &SessionResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: UserDTO{
			ID:         u.ID.String(),
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			Email:      u.Email,
			IsVerified: uint8(u.IsVerified),
			Roles:      roleNames,
		},
		Subscription:     subDTO,
		GlobalParameters: globalParams,
	}, nil
}

func (s *authService) resolveGlobalParameters(ctx context.Context) GlobalParamsDTO {
	params, err := s.subRepo.GetGlobalParameters(ctx)
	if err != nil {
		return GlobalParamsDTO{ActivateSubscription: true}
	}
	return GlobalParamsDTO{ActivateSubscription: params.ActivateSubscription}
}

// upsertDevice refreshes an existing device entry or appends a new one.
// Called after every successful login — errors are logged but never surface to the caller.
func (s *authService) upsertDevice(ctx context.Context, u *domainUser.User, input *DeviceInput) {
	rec, err := s.userRepo.GetDeviceRecord(ctx, u.ID)
	if shared.IsNotFound(err) {
		rec = &domainUser.DeviceRecord{UserID: u.ID}
	} else if err != nil {
		s.log.Warn("login: getting device record", zap.String("user_id", u.ID.String()), zap.Error(err))
		return
	}

	now := time.Now().Format(time.RFC3339)

	for i, d := range rec.Devices {
		if d.Identifier == input.Identifier {
			// Known device — refresh metadata and ensure it is active.
			rec.Devices[i].Os = input.Os
			rec.Devices[i].Brand = input.Brand
			rec.Devices[i].Version = input.Version
			rec.Devices[i].Status = domainUser.DeviceStatusActive
			rec.Devices[i].LastUpdated = now
			if err := s.userRepo.UpsertDeviceRecord(ctx, rec); err != nil {
				s.log.Warn("login: updating device record", zap.Error(err))
			}
			return
		}
	}

	// New device detected — append it.
	s.log.Info("new device detected on login",
		zap.String("user_id", u.ID.String()),
		zap.String("identifier", input.Identifier))

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
	if err := s.userRepo.UpsertDeviceRecord(ctx, rec); err != nil {
		s.log.Warn("login: inserting new device record", zap.Error(err))
	}
}

func (s *authService) resolveSubscription(ctx context.Context, u *domainUser.User) SubscriptionDTO {
	sub, err := s.subRepo.GetSubscriptionByUserID(ctx, u.ID)
	if err != nil {
		return SubscriptionDTO{Status: int(domainSub.StatusInactive)}
	}

	status := sub.Status
	if sub.IsExpired() {
		status = domainSub.StatusCanceled
	}

	dto := SubscriptionDTO{Status: int(status)}
	if sub.NextPaymentDate != nil {
		dto.NextPaymentDate = sub.NextPaymentDate.Format(time.RFC3339)
	}
	if sub.Plan != nil {
		dto.PlanName = sub.Plan.Name
	}
	return dto
}
