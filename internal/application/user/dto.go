package user

import domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"

// UserResponse is the safe public representation of a user — no password fields.
type UserResponse struct {
	ID         string   `json:"id"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Email      string   `json:"email"`
	IsVerified uint8    `json:"is_verified"`
	Roles      []string `json:"roles"`
}

// ToUserResponse converts a domain User to the safe HTTP response DTO.
func ToUserResponse(u *domainUser.User) UserResponse {
	roles := make([]string, len(u.Roles))
	for i := range u.Roles {
		roles[i] = string(u.Roles[i].Name)
	}
	return UserResponse{
		ID:         u.ID.String(),
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Email:      u.Email,
		IsVerified: uint8(u.IsVerified),
		Roles:      roles,
	}
}

// SignUpRequest is the payload for creating a new user account.
type SignUpRequest struct {
	FirstName string        `json:"first_name" validate:"required"`
	LastName  string        `json:"last_name"  validate:"required"`
	Email     string        `json:"email"      validate:"required,email"`
	Password  string        `json:"password"   validate:"required,min=6"`
	Devices   []DeviceInput `json:"devices"`
}

// UpdateProfileRequest carries fields the user may change on their profile.
type UpdateProfileRequest struct {
	UserName  string `json:"username"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"  validate:"required"`
	Mobile    string `json:"mobile"`
	Address   string `json:"address"`
	Gender    string `json:"gender"`
}

// ForgotPasswordRequest initiates the password reset flow.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// VerifyOTPRequest confirms a one-time password sent during reset.
type VerifyOTPRequest struct {
	Email string `json:"target" validate:"required,email"`
	OTP   string `json:"otp"    validate:"required"`
}

// ResetPasswordRequest completes the password reset after OTP verification.
type ResetPasswordRequest struct {
	Email           string `json:"email"            validate:"required,email"`
	Password        string `json:"password"         validate:"required,min=6"`
	PasswordConfirm string `json:"password_confirm" validate:"required,min=6"`
}

// ChangePasswordRequest allows an authenticated user to change their password.
type ChangePasswordRequest struct {
	Email           string `json:"email"              validate:"required,email"`
	OldPassword     string `json:"old_password"       validate:"required"`
	NewPassword     string `json:"new_password"       validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_new_password" validate:"required,min=6"`
}

// AssignRolesRequest sets a user's roles (replaces, does not append).
type AssignRolesRequest struct {
	UserID string   `json:"user_id" validate:"required,uuid"`
	Roles  []string `json:"roles"   validate:"required,min=1"`
}

// AdminSignupRequest is the admin signup payload with auto role assignment.
type AdminSignupRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"  validate:"required"`
	Email     string `json:"email"      validate:"required,email"`
	Password  string `json:"password"   validate:"required,min=6"`
}

// AddFavouriteRequest bookmarks a message for a user.
type AddFavouriteRequest struct {
	MessageID string `json:"message_id" validate:"required,uuid"`
	SeriesID  string `json:"series_id"`
}

// DeviceInput is the device descriptor sent during signup or device registration.
type DeviceInput struct {
	Who        string `json:"who"`
	Identifier string `json:"identifier" validate:"required"`
	Os         string `json:"os"`
	Brand      string `json:"brand"`
	Version    string `json:"version"`
}

// UpdateDeviceStatusRequest changes the active/inactive state of a device.
type UpdateDeviceStatusRequest struct {
	Identifier string                  `json:"identifier" validate:"required"`
	Status     domainUser.DeviceStatus `json:"status"     validate:"required"`
}

// AppVersionUpdateRequest updates the promoted app version record.
type AppVersionUpdateRequest struct {
	ID      string `json:"id"      validate:"required,uuid"`
	Version string `json:"version" validate:"required"`
	Force   bool   `json:"force"`
}

// SendEmailVerificationRequest triggers a verification email.
type SendEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}
