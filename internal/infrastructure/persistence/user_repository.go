package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainUser "bitbucket.org/hofng/hofApp/internal/domain/user"
	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewUserRepository returns a GORM-backed implementation of user.Repository.
func NewUserRepository(db *gorm.DB, log *zap.Logger) domainUser.Repository {
	return &userRepository{db: db, log: log}
}

// ── User CRUD ─────────────────────────────────────────────────────────────────

func (r *userRepository) Create(ctx context.Context, u *domainUser.User) error {
	if result := r.db.WithContext(ctx).Create(u); result.Error != nil {
		if isUniqueViolation(result.Error) {
			return shared.ErrAlreadyExists{Resource: "user", Field: "email", Value: u.Email}
		}
		return fmt.Errorf("creating user: %w", result.Error)
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainUser.User, error) {
	var u domainUser.User
	result := r.db.WithContext(ctx).Preload("Roles").First(&u, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "user", ID: id.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting user by id: %w", result.Error)
	}
	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	var u domainUser.User
	result := r.db.WithContext(ctx).Preload("Roles").First(&u, "email = ?", email)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "user", ID: email}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting user by email: %w", result.Error)
	}
	return &u, nil
}

func (r *userRepository) GetByCustomerCode(ctx context.Context, code string) (*domainUser.User, error) {
	var u domainUser.User
	result := r.db.WithContext(ctx).Preload("Roles").First(&u, "paystack_customer_code = ?", code)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "user", ID: code}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting user by customer code: %w", result.Error)
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, u *domainUser.User) error {
	result := r.db.WithContext(ctx).Save(u)
	if result.Error != nil {
		return fmt.Errorf("updating user: %w", result.Error)
	}
	return nil
}

func (r *userRepository) UpdateVerificationStatus(ctx context.Context, id uuid.UUID, status domainUser.VerificationStatus) error {
	result := r.db.WithContext(ctx).Model(&domainUser.User{}).
		Where("id = ?", id).
		Update("is_verified", status)
	if result.Error != nil {
		return fmt.Errorf("updating verification status: %w", result.Error)
	}
	return nil
}

// ── Password ──────────────────────────────────────────────────────────────────

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashed string, version domainUser.PasswordVersion) error {
	result := r.db.WithContext(ctx).Model(&domainUser.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password":         hashed,
			"password_version": version,
		})
	if result.Error != nil {
		return fmt.Errorf("updating password: %w", result.Error)
	}
	return nil
}

// ── Paystack ──────────────────────────────────────────────────────────────────

func (r *userRepository) UpdatePaystackInfo(ctx context.Context, id uuid.UUID, code, customerID string) error {
	result := r.db.WithContext(ctx).Model(&domainUser.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"paystack_customer_code": code,
			"paystack_customer_id":   customerID,
		})
	if result.Error != nil {
		return fmt.Errorf("updating paystack info: %w", result.Error)
	}
	return nil
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (r *userRepository) AssignRoles(ctx context.Context, userID uuid.UUID, names []domainUser.RoleName) error {
	var roles []domainUser.Role
	if result := r.db.WithContext(ctx).Where("name IN ?", names).Find(&roles); result.Error != nil {
		return fmt.Errorf("finding roles: %w", result.Error)
	}

	u := &domainUser.User{}
	u.ID = userID
	if err := r.db.WithContext(ctx).Model(u).Association("Roles").Append(roles); err != nil {
		return fmt.Errorf("assigning roles: %w", err)
	}
	return nil
}

func (r *userRepository) RemoveRoles(ctx context.Context, userID uuid.UUID, names []domainUser.RoleName) error {
	var roles []domainUser.Role
	if result := r.db.WithContext(ctx).Where("name IN ?", names).Find(&roles); result.Error != nil {
		return fmt.Errorf("finding roles: %w", result.Error)
	}

	u := &domainUser.User{}
	u.ID = userID
	if err := r.db.WithContext(ctx).Model(u).Association("Roles").Delete(roles); err != nil {
		return fmt.Errorf("removing roles: %w", err)
	}
	return nil
}

func (r *userRepository) GetRoles(ctx context.Context, userID uuid.UUID) ([]domainUser.Role, error) {
	var u domainUser.User
	u.ID = userID
	var roles []domainUser.Role
	if err := r.db.WithContext(ctx).Model(&u).Association("Roles").Find(&roles); err != nil {
		return nil, fmt.Errorf("getting roles: %w", err)
	}
	return roles, nil
}

// ── Password tokens ───────────────────────────────────────────────────────────

func (r *userRepository) UpsertPasswordToken(ctx context.Context, token *domainUser.PasswordToken) error {
	result := r.db.WithContext(ctx).
		Where(domainUser.PasswordToken{Email: token.Email}).
		Assign(domainUser.PasswordToken{
			PasswordResetToken: token.PasswordResetToken,
			PasswordResetAt:    token.PasswordResetAt,
			Validated:          false,
		}).
		FirstOrCreate(token)
	if result.Error != nil {
		return fmt.Errorf("upserting password token: %w", result.Error)
	}

	// If the record already existed, update its fields.
	if result.RowsAffected == 0 {
		result = r.db.WithContext(ctx).Model(token).Updates(map[string]any{
			"password_reset_token": token.PasswordResetToken,
			"password_reset_at":    token.PasswordResetAt,
			"validated":            false,
		})
		if result.Error != nil {
			return fmt.Errorf("updating password token: %w", result.Error)
		}
	}
	return nil
}

func (r *userRepository) GetPasswordToken(ctx context.Context, email string) (*domainUser.PasswordToken, error) {
	var token domainUser.PasswordToken
	result := r.db.WithContext(ctx).First(&token, "email = ?", email)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "password token", ID: email}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting password token: %w", result.Error)
	}
	return &token, nil
}

func (r *userRepository) MarkPasswordTokenValidated(ctx context.Context, email string) error {
	result := r.db.WithContext(ctx).Model(&domainUser.PasswordToken{}).
		Where("email = ?", email).
		Update("validated", true)
	if result.Error != nil {
		return fmt.Errorf("marking password token validated: %w", result.Error)
	}
	return nil
}

// ── Favourites ────────────────────────────────────────────────────────────────

func (r *userRepository) GetFavouriteRecord(ctx context.Context, userID uuid.UUID) (*domainUser.FavouriteRecord, error) {
	var fav domainUser.FavouriteRecord
	result := r.db.WithContext(ctx).First(&fav, "user_id = ?", userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "favourite", ID: userID.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting favourite record: %w", result.Error)
	}
	return &fav, nil
}

func (r *userRepository) UpsertFavourite(ctx context.Context, record *domainUser.FavouriteRecord) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", record.UserID).
		Assign(domainUser.FavouriteRecord{Fav: record.Fav}).
		FirstOrCreate(record)
	if result.Error != nil {
		return fmt.Errorf("upserting favourite: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		result = r.db.WithContext(ctx).Model(record).Update("fav", record.Fav)
		if result.Error != nil {
			return fmt.Errorf("updating favourite: %w", result.Error)
		}
	}
	return nil
}

func (r *userRepository) GetFavouriteMessages(ctx context.Context, userID uuid.UUID) ([]domainUser.FavouriteMessage, error) {
	type favRow struct {
		FavouriteID uuid.UUID
		UserID      uuid.UUID
		MessageID   uuid.UUID
		Fav         bool
		Title       string
		Author      string
		ImageURL    string
		AudioURL    string
		Description string
		IsFree      bool
	}

	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT
			f.id AS favourite_id,
			f.user_id,
			(item->>'message_id')::uuid AS message_id,
			(item->>'fav')::boolean AS fav,
			m.title,
			m.author,
			m.image_url,
			m.audio_url,
			m.description,
			m.is_free
		FROM favourites f
		CROSS JOIN LATERAL jsonb_array_elements(f.fav) AS arr(item)
		JOIN audio_messages m ON m.id = (item->>'message_id')::uuid
		WHERE f.user_id = ? AND m.deleted_at IS NULL
	`, userID).Rows()
	if err != nil {
		return nil, fmt.Errorf("querying favourite messages: %w", err)
	}
	defer rows.Close()

	var results []domainUser.FavouriteMessage
	for rows.Next() {
		var row domainUser.FavouriteMessage
		if err := r.db.ScanRows(rows, &row); err != nil {
			return nil, fmt.Errorf("scanning favourite message: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

func (r *userRepository) DeleteFavouriteItem(ctx context.Context, userID, messageID uuid.UUID) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE favourites
		SET fav = fav - (
			SELECT (position - 1)::int
			FROM favourites, jsonb_array_elements(fav) WITH ORDINALITY arr(item, position)
			WHERE user_id = ? AND item->>'message_id' = ?
		)
		WHERE user_id = ?
	`, userID, messageID.String(), userID)
	if result.Error != nil {
		return fmt.Errorf("deleting favourite item: %w", result.Error)
	}
	return nil
}

// ── Devices ───────────────────────────────────────────────────────────────────

func (r *userRepository) GetDeviceRecord(ctx context.Context, userID uuid.UUID) (*domainUser.DeviceRecord, error) {
	var rec domainUser.DeviceRecord
	result := r.db.WithContext(ctx).First(&rec, "user_id = ?", userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "device record", ID: userID.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting device record: %w", result.Error)
	}
	return &rec, nil
}

func (r *userRepository) UpsertDeviceRecord(ctx context.Context, rec *domainUser.DeviceRecord) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", rec.UserID).
		Assign(domainUser.DeviceRecord{Devices: rec.Devices}).
		FirstOrCreate(rec)
	if result.Error != nil {
		return fmt.Errorf("upserting device record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		result = r.db.WithContext(ctx).Model(rec).Update("devices", rec.Devices)
		if result.Error != nil {
			return fmt.Errorf("updating device record: %w", result.Error)
		}
	}
	return nil
}

func (r *userRepository) DeleteDevice(ctx context.Context, userID uuid.UUID, identifier string) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE devices
		SET devices = devices - (
			SELECT (position - 1)::int
			FROM devices, jsonb_array_elements(devices) WITH ORDINALITY arr(item, position)
			WHERE user_id = ? AND item->>'identifier' = ?
		)
		WHERE user_id = ?
	`, userID, identifier, userID)
	if result.Error != nil {
		return fmt.Errorf("deleting device: %w", result.Error)
	}
	return nil
}

func (r *userRepository) UpdateDeviceStatus(ctx context.Context, userID uuid.UUID, identifier string, status domainUser.DeviceStatus) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE devices
		SET devices = (
			SELECT jsonb_agg(
				CASE WHEN elem->>'identifier' = ?
					THEN jsonb_set(elem, '{status}', to_jsonb(?::text))
					ELSE elem
				END
			)
			FROM devices, LATERAL jsonb_array_elements(devices) AS arr(elem)
			WHERE user_id = ?
		)
		WHERE user_id = ?
	`, identifier, string(status), userID, userID)
	if result.Error != nil {
		return fmt.Errorf("updating device status: %w", result.Error)
	}
	return nil
}

// ── App version ───────────────────────────────────────────────────────────────

func (r *userRepository) GetAppVersion(ctx context.Context, id uuid.UUID) (*domainUser.AppVersion, error) {
	var v domainUser.AppVersion
	result := r.db.WithContext(ctx).First(&v, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "app version", ID: id.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting app version: %w", result.Error)
	}
	return &v, nil
}

func (r *userRepository) UpdateAppVersion(ctx context.Context, v *domainUser.AppVersion) error {
	v.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).Save(v)
	if result.Error != nil {
		return fmt.Errorf("updating app version: %w", result.Error)
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func isUniqueViolation(err error) bool {
	return err != nil && (containsString(err.Error(), "unique") || containsString(err.Error(), "duplicate"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
