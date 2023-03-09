package user

import (
	"database/sql"
	"database/sql/driver"
	"errors"
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
	case "0":
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

// sql/driver.Valuer interface implementation for IsVerifiedEnum
func (e IsVerifiedEnum) Value() (driver.Value, error) {
	switch e {
	case Unverfified:
		return 0, nil
	case PhoneVerified:
		return 1, nil
	case EmailVerified:
		return 2, nil
	case EmailAndPhoneVerified:		
		return 3, nil
	default:
		return 0, nil
	}
}

type User struct {
	ID                   string         `sql:"id"`
	UserName             string         `sql:"username"`
	Password             string         `sql:"password" validate:"min=6"`
	FirstName            string         `sql:"first_name" validate:"required"`
	LastName             string         `sql:"last_name" validate:"required"`
	Email                string         `sql:"email" validate:"required,email"`
	Mobile               sql.NullString `sql:"mobile"`
	Address              string         `sql:"address"`
	Gender               string         `sql:"gender"`
	PasswordHash         sql.NullString `sql:"password_hash"`
	IsVerified           IsVerifiedEnum `sql:"is_verified"`
	PaystackCustomerCode sql.NullString `sql:"paystack_customer_code"`
	PaystackCustomerId   sql.NullString `sql:"paystack_customer_id"`
}

type LoginUser struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type SignUpUser struct {
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
	Email     string `validate:"required,email"`
	Password  string `validate:"required"`
}

type UserPasswordToken struct {
	ID                 string `sql:"id"`
	Email              string `sql:"email" validate:"required,email"`
	PasswordResetToken string `sql:"password_reset_token"`
	PasswordResetAt    int64  `sql:"password_reset_at"`
}

type ForgotPasswordPayload struct {
	// The ULID of ForgotPasswordPayload
	Email string `json:"email" validate:"required,email"`
} //	@name	ForgotPasswordPayload

// ForgotPasswordResponse temporary response pending an email client
type ForgotPasswordResponse struct {
	URL string `json:"url"`
} //	@name	ForgotPasswordResponse

type ResetPasswordPayload struct {
	// The ULID of ResetPasswordPayload
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"min=6" binding:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"min=6" binding:"required"`
} //	@name	ResetPasswordPayload
