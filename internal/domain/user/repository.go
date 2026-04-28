package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines every persistence operation the domain needs on users.
// Implementations live in the infrastructure layer.
type Repository interface {
	// User CRUD
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByCustomerCode(ctx context.Context, code string) (*User, error)
	Update(ctx context.Context, u *User) error
	UpdateVerificationStatus(ctx context.Context, id uuid.UUID, status VerificationStatus) error

	// Password
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string, version PasswordVersion) error

	// Paystack
	UpdatePaystackInfo(ctx context.Context, id uuid.UUID, customerCode, customerID string) error

	// Roles
	AssignRoles(ctx context.Context, userID uuid.UUID, roles []RoleName) error
	RemoveRoles(ctx context.Context, userID uuid.UUID, roles []RoleName) error
	GetRoles(ctx context.Context, userID uuid.UUID) ([]Role, error)

	// Password reset tokens
	UpsertPasswordToken(ctx context.Context, token *PasswordToken) error
	GetPasswordToken(ctx context.Context, email string) (*PasswordToken, error)
	MarkPasswordTokenValidated(ctx context.Context, email string) error

	// Favorites
	GetFavouriteRecord(ctx context.Context, userID uuid.UUID) (*FavouriteRecord, error)
	UpsertFavourite(ctx context.Context, record *FavouriteRecord) error
	GetFavouriteMessages(ctx context.Context, userID uuid.UUID) ([]FavouriteMessage, error)
	DeleteFavouriteItem(ctx context.Context, userID, messageID uuid.UUID) error

	// Devices
	GetDeviceRecord(ctx context.Context, userID uuid.UUID) (*DeviceRecord, error)
	UpsertDeviceRecord(ctx context.Context, record *DeviceRecord) error
	DeleteDevice(ctx context.Context, userID uuid.UUID, identifier string) error
	UpdateDeviceStatus(ctx context.Context, userID uuid.UUID, identifier string, status DeviceStatus) error

	// App version
	GetAppVersion(ctx context.Context, id uuid.UUID) (*AppVersion, error)
	UpdateAppVersion(ctx context.Context, v *AppVersion) error
}
