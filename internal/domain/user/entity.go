package user

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// VerificationStatus tracks the verification state of a user's contact info.
type VerificationStatus uint8

const (
	Unverified            VerificationStatus = 0
	PhoneVerified         VerificationStatus = 1
	EmailVerified         VerificationStatus = 2
	EmailAndPhoneVerified VerificationStatus = 3
)

// PasswordVersion identifies the hashing algorithm used for a stored password.
type PasswordVersion string

const (
	PasswordVersionBcrypt PasswordVersion = "bcrypt"
	PasswordVersionMD5    PasswordVersion = "md5" // legacy only
)

// RoleName enumerates the roles a user may hold within the church system.
type RoleName string

const (
	RoleSteward      RoleName = "steward"
	RoleMember       RoleName = "member"
	RoleChurchFriend RoleName = "church_friend"
	RoleTeamLead     RoleName = "team_lead"
	RoleChurchAdmin  RoleName = "church_admin"
)

// AllRoles lists every valid role name.
var AllRoles = []RoleName{
	RoleSteward,
	RoleMember,
	RoleChurchFriend,
	RoleTeamLead,
	RoleChurchAdmin,
}

// Role is a named permission group that can be assigned to many users.
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        RoleName  `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string    `gorm:"type:varchar(200)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Role) TableName() string { return "roles" }

// User is the root aggregate for all user-related business logic.
type User struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FirstName            string             `gorm:"type:varchar(200);not null"`
	LastName             string             `gorm:"type:varchar(200);not null"`
	UserName             string             `gorm:"type:varchar(200)"`
	Email                string             `gorm:"type:varchar(200);uniqueIndex;not null"`
	Password             string             `gorm:"type:varchar(255);not null"`
	PasswordVersion      PasswordVersion    `gorm:"type:varchar(10);default:'bcrypt'"`
	Mobile               *string            `gorm:"type:varchar(15)"`
	Address              *string            `gorm:"type:varchar(100)"`
	Gender               *string            `gorm:"type:varchar(10)"`
	IsVerified           VerificationStatus `gorm:"type:smallint;default:0"`
	PaystackCustomerCode *string            `gorm:"type:varchar(255)"`
	PaystackCustomerID   *string            `gorm:"type:varchar(255)"`
	Roles                []Role             `gorm:"many2many:user_roles;"`
	CreatedAt            time.Time          `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt            time.Time          `gorm:"column:last_updated;autoUpdateTime"`
}

func (User) TableName() string { return "users" }

// HasRole reports whether the user holds the named role.
func (u *User) HasRole(name RoleName) bool {
	for _, r := range u.Roles {
		if r.Name == name {
			return true
		}
	}
	return false
}

// FullName returns the user's formatted full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// PasswordToken stores a one-time token used for password reset flows.
type PasswordToken struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email              string    `gorm:"type:varchar(200);not null;index"`
	PasswordResetToken string    `gorm:"type:varchar(255)"`
	PasswordResetAt    int64     `gorm:"type:bigint"`
	Validated          bool      `gorm:"default:false"`
}

func (PasswordToken) TableName() string { return "user_password_token" }

// DeviceStatus represents the active/inactive state of a registered device.
type DeviceStatus string

const (
	DeviceStatusActive   DeviceStatus = "ACTIVE"
	DeviceStatusInactive DeviceStatus = "INACTIVE"
)

// Device is a single registered device belonging to a user.
type Device struct {
	ID          string       `json:"id"`
	Who         string       `json:"who"`
	Identifier  string       `json:"identifier,omitempty"`
	Os          string       `json:"os,omitempty"`
	Brand       string       `json:"brand,omitempty"`
	Version     string       `json:"version"`
	Status      DeviceStatus `json:"status,omitempty"`
	DateAdded   string       `json:"date_added,omitempty"`
	LastUpdated string       `json:"last_updated,omitempty"`
}

// DeviceList is a JSON-serialisable slice of devices for storage in a JSONB column.
type DeviceList []Device

func (d DeviceList) Value() (driver.Value, error) {
	if d == nil {
		return "[]", nil
	}
	b, err := json.Marshal(d)
	return string(b), err
}

func (d *DeviceList) Scan(src interface{}) error {
	var source []byte
	switch v := src.(type) {
	case []byte:
		source = v
	case string:
		source = []byte(v)
	default:
		return errors.New("unsupported DeviceList scan type")
	}
	return json.Unmarshal(source, d)
}

// DeviceRecord is the database row that holds a user's device list as JSONB.
type DeviceRecord struct {
	ID      uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"`
	Devices DeviceList `gorm:"type:jsonb;serializer:json"`
}

func (DeviceRecord) TableName() string { return "devices" }

// AppVersion holds the metadata for a specific application build.
type AppVersion struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Version   string    `gorm:"type:varchar(200);not null"`
	Force     bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:last_updated;autoUpdateTime"`
}

func (AppVersion) TableName() string { return "app_version" }

// FavouriteItem is a single bookmarked message stored inside the JSONB array.
type FavouriteItem struct {
	MessageID string `json:"message_id"`
	SeriesID  string `json:"series_id,omitempty"`
	Fav       bool   `json:"fav"`
	DateAdded string `json:"date_added,omitempty"`
}

// FavouriteList is a JSON-serialisable slice of favourite items.
type FavouriteList []FavouriteItem

func (f FavouriteList) Value() (driver.Value, error) {
	if f == nil {
		return "[]", nil
	}
	b, err := json.Marshal(f)
	return string(b), err
}

func (f *FavouriteList) Scan(src interface{}) error {
	var source []byte
	switch v := src.(type) {
	case []byte:
		source = v
	case string:
		source = []byte(v)
	default:
		return errors.New("unsupported FavouriteList scan type")
	}
	return json.Unmarshal(source, f)
}

// FavouriteRecord is the database row that stores a user's favourites as JSONB.
type FavouriteRecord struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex"`
	Fav       FavouriteList `gorm:"type:jsonb;serializer:json"`
	CreatedAt time.Time     `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt time.Time     `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt *time.Time    `gorm:"column:deleted_at"`
}

func (FavouriteRecord) TableName() string { return "favourites" }

// FavouriteMessage is the enriched view returned when a user fetches their favourites.
type FavouriteMessage struct {
	FavouriteID uuid.UUID `json:"favourite_id"`
	UserID      uuid.UUID `json:"user_id"`
	MessageID   uuid.UUID `json:"message_id"`
	Fav         bool      `json:"fav"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	ImageURL    string    `json:"image_url"`
	AudioURL    string    `json:"audio_url"`
	Description string    `json:"description"`
	IsFree      bool      `json:"is_free"`
}
