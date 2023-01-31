package user

import (
	"database/sql"
	"errors"
	"time"
)

type IsVerifiedEnum uint16

const (
	Unverfified           = IsVerifiedEnum(0)
	PhoneVerified         = IsVerifiedEnum(1)
	EmailVerified         = IsVerifiedEnum(2)
	EmailAndPhoneVerified = IsVerifiedEnum(3)
)

func (e IsVerifiedEnum) String() string {
	switch e {
	case Unverfified:
		return "0"
	case PhoneVerified:
		return "1"
	case EmailVerified:
		return "2"
	case EmailAndPhoneVerified:
		return "3"
	default:
		return "invalid"
	}
}

// MarshalText interface implementation IsVerifiedEnum into text.
func (e IsVerifiedEnum) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

// UnMarshalText interface implementation IsVerifiedEnum into text.
func (e *IsVerifiedEnum) UnMarshalText(from []byte) error {
	switch string(from) {
	case "o":
		*e = Unverfified
	case "1":
		*e = PhoneVerified
	case "2":
		*e = EmailVerified
	case "3":
		*e = EmailAndPhoneVerified
	default:
		return errors.New("invalid IsVerifiedEnum")
	}
	return nil
}

type User struct {
	// The ULID of a user
	ID           int            `sql:"id"`
	UserName     string         `sql:"username"`
	Password     string         `sql:"password" validate:"min=6"`
	FirstName    string         `sql:"first_name" validate:"required"`
	LastName     string         `sql:"last_name" validate:"required"`
	Email        string         `sql:"email" validate:"required,email"`
	Mobile       sql.NullString `sql:"mobile"`
	Address      string         `sql:"address"`
	Gender       string         `sql:"gender"`
	PasswordHash sql.NullString `sql:"password_hash"`
	IsVerified   IsVerifiedEnum `sql:"is_verified"`
} // @name User

type LoginUser struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type SignUpUser struct {
	FirstName    	string         	`validate:"required"`
	LastName     	string         	`validate:"required"`
	Email   	 	string 			`validate:"required,email"`
	Password 		string 			`validate:"required"`
}


type UserPasswordToken struct {
	ID                 int       `sql:"id"`
	Email              string    `json:"email" validate:"required,email"`
	PasswordResetToken string    `sql:"password_reset_token"`
	PasswordResetAt    time.Time `sql:"password_reset_at"`
}
type ForgotPasswordPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordPayload struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"min=6" binding:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"min=6" binding:"required"`
}
