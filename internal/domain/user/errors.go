package user

import "bitbucket.org/hofng/hofApp/internal/domain/shared"

var (
	// ErrInvalidCredentials is returned when sign-in credentials do not match.
	ErrInvalidCredentials = shared.ErrUnauthorized{Message: "invalid email or password"}
	// ErrEmailAlreadyInUse is returned when an email address is already registered.
	ErrEmailAlreadyInUse = shared.ErrAlreadyExists{Resource: "user", Field: "email", Value: ""}
	// ErrDeviceNotFound is returned when a device record cannot be found.
	ErrDeviceNotFound = shared.ErrNotFound{Resource: "device", ID: ""}
	// ErrExpiredOTP is returned when a verification code has expired.
	ErrExpiredOTP = shared.ErrInvalidInput{Field: "otp", Message: "the verification code has expired"}
	// ErrInvalidOTP is returned when a verification code is invalid.
	ErrInvalidOTP = shared.ErrInvalidInput{Field: "otp", Message: "the verification code is invalid"}
	// ErrAlreadyValidatedOTP is returned when a verification code was already used.
	ErrAlreadyValidatedOTP = shared.ErrInvalidInput{Field: "otp", Message: "this verification code has already been used"}
	// ErrPasswordMismatch is returned when the current password does not match.
	ErrPasswordMismatch = shared.ErrInvalidInput{Field: "old_password", Message: "old password does not match"}
	// ErrPasswordConfirm is returned when the password confirmation does not match.
	ErrPasswordConfirm = shared.ErrInvalidInput{Field: "confirm_password", Message: "password and confirmation do not match"}
	// ErrInvalidRole is returned when one or more roles are invalid.
	ErrInvalidRole = shared.ErrInvalidInput{Field: "roles", Message: "one or more roles are invalid"}
)
