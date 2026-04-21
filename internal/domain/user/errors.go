package user

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailAlreadyInUse   = errors.New("an account with this email already exists")
	ErrDeviceNotFound      = errors.New("device not found for this account")
	ErrExpiredOTP          = errors.New("the verification code has expired")
	ErrInvalidOTP          = errors.New("the verification code is invalid")
	ErrAlreadyValidatedOTP = errors.New("this verification code has already been used")
	ErrPasswordMismatch    = errors.New("old password does not match")
	ErrPasswordConfirm     = errors.New("password and confirmation do not match")
	ErrInvalidRole         = errors.New("one or more roles are invalid")
)
