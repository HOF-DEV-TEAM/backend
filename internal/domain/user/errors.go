package user

import "errors"

var (
	// ErrInvalidCredentials is returned when sign-in credentials do not match.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrEmailAlreadyInUse is returned when an email address is already registered.
	ErrEmailAlreadyInUse = errors.New("an account with this email already exists")
	// ErrDeviceNotFound is returned when a device record cannot be found.
	ErrDeviceNotFound = errors.New("device not found for this account")
	// ErrExpiredOTP is returned when a verification code has expired.
	ErrExpiredOTP = errors.New("the verification code has expired")
	// ErrInvalidOTP is returned when a verification code is invalid.
	ErrInvalidOTP = errors.New("the verification code is invalid")
	// ErrAlreadyValidatedOTP is returned when a verification code was already used.
	ErrAlreadyValidatedOTP = errors.New("this verification code has already been used")
	// ErrPasswordMismatch is returned when the current password does not match.
	ErrPasswordMismatch = errors.New("old password does not match")
	// ErrPasswordConfirm is returned when the password confirmation does not match.
	ErrPasswordConfirm = errors.New("password and confirmation do not match")
	// ErrInvalidRole is returned when one or more roles are invalid.
	ErrInvalidRole = errors.New("one or more roles are invalid")
)
