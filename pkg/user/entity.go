package user

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/gofrs/uuid"
)

type IsVerifiedEnum uint16

const (
	Unverfified           = IsVerifiedEnum(0)
	PhoneVerified         = IsVerifiedEnum(1)
	EmailVerified         = IsVerifiedEnum(2)
	EmailAndPhoneVerified = IsVerifiedEnum(3)
	appVersionID          = "8383afb2-2a5c-4c10-aaf7-f10735b9d895"
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
	LatestAppVersion     VersionManager `sql:"latest_app_version"`
	DateAdded            sql.NullString `sql:"date_added"`
	LastUpdated          sql.NullString `sql:"last_updated"`
}

type SignUpUser struct {
	FirstName string   `validate:"required"`
	LastName  string   `validate:"required"`
	Email     string   `validate:"required,email"`
	Password  string   `validate:"required"`
	Devices   []Device `validate:"required"`
}

type UserPasswordToken struct {
	ID                 string `sql:"id"`
	Email              string `sql:"email" validate:"required,email"`
	PasswordResetToken string `sql:"password_reset_token"`
	PasswordResetAt    int64  `sql:"password_reset_at"`
	Validated          bool   `sql:"validated"`
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

type ChangePasswordPayload struct {
	// The ULID of ChangePasswordPayload
	Email              string `json:"email" validate:"required,email"`
	OldPassword        string `json:"old_password" binding:"required"`
	NewPassword        string `json:"new_password" validate:"min=6" binding:"required"`
	ConfirmNewPassword string `json:"confirm_new_password" validate:"min=6" binding:"required"`
} //	@name	ChangePasswordPayload

type Favourites struct {
	ID     uuid.UUID `sql:"id"`
	UserID uuid.UUID `sql:"user_id" validate:"required"`
	Fav    []FavBody `sql:"fav" json:"fav"`
}

type FavBody struct {
	MessageID uuid.UUID `sql:"message_id" json:"message_id"`
	SeriesID  string    `sql:"series_id" json:"series_id"`
	Fav       bool      `sql:"fav" json:"fav"`
	DateAdded string    `sql:"date_added" json:"date_added"`
	DeletedAt string    `sql:"deleted_at" json:"deleted_at"`
}

type FavMessage struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Fav         bool      `json:"fav"`
	MessageID   uuid.UUID `json:"message_id"`
	SeriesID    uuid.UUID `json:"series_id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	ImageUrl    string    `json:"image_url"`
	AudioUrl    string    `json:"audio_url"`
	Description string    `json:"description"`
}

type OTPRequest struct {
	Target string `json:"target"`
}

type OTPResponse struct {
	User                string `json:"user"`
	Target              string `json:"target"`
	OTP                 string `json:"otp"`
	ExpireTimeInSeconds int64  `json:"expireTimeInSeconds"`
}

type VerifyOTP struct {
	Target string `json:"target"`
	OTP    string `json:"otp"`
}

type UpdateUser struct {
	UserName    string `sql:"username"`
	FirstName   string `sql:"first_name" validate:"required"`
	LastName    string `sql:"last_name" validate:"required"`
	Mobile      string `sql:"mobile"`
	Address     string `sql:"address"`
	Gender      string `sql:"gender"`
	LastUpdated string `sql:"last_updated"`
}

type DeviceManager struct {
	ID      uuid.UUID `sql:"id" json:"id"`
	UserID  string    `sql:"user_id" json:"user_id"`
	Devices []Device  `sql:"devices" json:"devices"`
}

type Device struct {
	ID          uuid.UUID      `sql:"id" json:"id"`
	Who         string         `sql:"who" json:"who"`
	Identifier  string         `sql:"identifier" json:"identifier,omitempty"`
	Os          string         `sql:"os" json:"os,omitempty"`
	Brand       string         `sql:"brand" json:"brand,omitempty"`
	Version     string         `sql:"version" json:"version"`
	Status      DeviceStatus   `sql:"status" json:"status,omitempty"`
	DateAdded   sql.NullString `sql:"date_added" swaggertype:"string"`
	LastUpdated sql.NullString `sql:"last_updated" swaggertype:"string"`
}

// DeviceStatus enum type
type DeviceStatus string

// TODO use enum
const (
	ACTIVE_DEVICE_STATUS   DeviceStatus = "ACTIVE"
	INACTIVE_DEVICE_STATUS DeviceStatus = "INACTIVE"
)

type VersionManager struct {
	ID          string         `sql:"id" json:"id"`
	Version     string         `sql:"version" json:"version"`
	Force       *bool          `sql:"force" json:"force"`
	DateAdded   sql.NullString `sql:"date_added" json:"date_added" swaggertype:"string"`
	LastUpdated sql.NullString `sql:"last_updated" json:"last_updated" swaggertype:"string"`
}
