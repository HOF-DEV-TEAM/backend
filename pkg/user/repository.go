package user

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library"
	"bitbucket.org/hofng/hofApp/infrastructure/library/urlqueryhelper"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"go.uber.org/zap"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	SignUpUser(ctx context.Context, user *User, deviceManger *DeviceManager) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByCustomerCode(ctx context.Context, email string) (*User, error)
	GetById(ctx context.Context, id string) (*User, error)
	LoginWithEmailPasswordDevice(ctx context.Context, email, password, deviceIdentifier string) (*User, error)
	LoginWithEmailPassword(ctx context.Context, email, password string) (*User, error)
	ForgotPassword(request ForgotPasswordPayload) (*OTPResponse, error)
	VerifyPasswordResetOTP(request *VerifyOTP) (*User, error)
	ResetPassword(ctx context.Context, userId uuid.UUID, request ResetPasswordPayload) (uuid.UUID, error)
	ChangePassword(ctx context.Context, userId uuid.UUID, request ChangePasswordPayload) (uuid.UUID, error)
	UpdatePaystack(ctx context.Context, user *User) (uuid.UUID, error)
	CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error)
	GetFavourites(ctx context.Context, userId uuid.UUID) ([]*FavMessage, int, error)
	DeleteFavourite(ctx context.Context, messageId, userId uuid.UUID) (uuid.UUID, error)
	UpdateUserProfile(ctx context.Context, userId uuid.UUID, user *UpdateUser) (uuid.UUID, error)
	BuildDevice(ctx context.Context, input *DeviceManager, email string) (*DeviceManager, error)
	GetDevices(ctx context.Context, userId string) (*DeviceManager, error)
	DeleteDevice(ctx context.Context, identifier, userID string) (string, error)
	UpdateDevice(ctx context.Context, userId, status, identifier string) (*DeviceManager, error)
	UpdateAppVersion(ctx context.Context, version VersionManager) (uuid.UUID, error)
	GetAppVersion(ctx context.Context, versionID uuid.UUID) (*VersionManager, error)
	UpdateUserIsVerified(ctx context.Context, userId string, verifyField IsVerifiedEnum) error
	Close() error
}

type userRepository struct {
	db           *sql.DB
	log          *zap.Logger
	getEmailStmt *sql.Stmt
	getIdStmt    *sql.Stmt
	otpGenerator OtpGenerator
	queryHandler urlqueryhelper.QueryHelper
	idGenerator  library.IDGenerator
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &userRepository{db: db, log: logger, otpGenerator: NewOTPGenerator(), queryHandler: urlqueryhelper.NewQueryHelper(), idGenerator: library.NewIDGenerator()}
}

func (r userRepository) Close() error {
	if r.getEmailStmt != nil {
		if err := r.getEmailStmt.Close(); err != nil {
			return err
		}
	}

	if r.getIdStmt != nil {
		if err := r.getIdStmt.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (r userRepository) logSQLError(sqlStmt, errorMsg string, err error) {
	if errorMsg == "" {
		errorMsg = "error"
	}
	r.log.Info("msg", zap.String("error preparing statement", ""), zap.String(errorMsg, err.Error()), zap.String("query", sqlStmt))
}

func (r userRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
	// sql insert query, primary key provided by autoincrement
	const SQL = "INSERT INTO users (" +
		"username," +
		"password," +
		"first_name," +
		"last_name," +
		"email," +
		"mobile," +
		"address," +
		"gender," +
		"password_hash," +
		"is_verified" +
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) " +
		"RETURNING id"

	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		r.logSQLError(SQL, "", err)
		return nil, err
	}

	defer tx.Rollback()

	tmpSmt, err := tx.PrepareContext(ctx, SQL)
	if err != nil {
		r.logSQLError(SQL, "", err)
		return nil, err
	}

	var createdUserId string

	fmt.Printf("Printing user.... %+v\n", user)

	err = tmpSmt.QueryRowContext(ctx,
		user.UserName,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Mobile,
		user.Address,
		user.Gender,
		user.PasswordHash,
		user.IsVerified,
	).Scan(&createdUserId)

	if err != nil {
		r.log.Error("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	user.ID = createdUserId
	return user, err
}

func (r userRepository) SignUpUser(ctx context.Context, user *User, deviceManager *DeviceManager) (*User, error) {
	//var g errgroup.Group

	//g.Go(func() error {
	_, err := r.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	//return err
	//})

	//g.Go(func() error {
	_, err = r.BuildDevice(ctx, deviceManager, user.Email)
	//	return err
	//})

	if err != nil {
		return nil, err
	}

	//if err := g.Wait(); err != nil {
	//	if user.ID != "" {
	//		//rollback user
	//		r.rollBack(ctx, `DELETE FROM users WHERE id=$1 RETURNING id;`, user.ID)
	//	}
	//
	//	if deviceManager.ID != uuid.Nil {
	//		//rollback device
	//		r.rollBack(ctx, `DELETE FROM devices WHERE id=$1 RETURNING id;`, deviceManager.ID.String())
	//	}
	//	return nil, err
	//}
	//
	//appVersionId, err := r.idGenerator.IDGenerateFromString(appVersionID)
	//if err != nil {
	//	return nil, err
	//}
	//appVersion, err := r.GetAppVersion(ctx, appVersionId)
	//if err != nil {
	//	return nil, err
	//}
	//
	//user.LatestAppVersion = *appVersion

	return user, nil
}

func (r userRepository) getUser(ctx context.Context, field string, value string) (*User, error) {
	const SQL = "SELECT " +
		"id," +
		"username," +
		"password," +
		"first_name," +
		"last_name," +
		"email," +
		"mobile," +
		"address," +
		"gender," +
		"is_verified," +
		"paystack_customer_code " +
		"FROM users WHERE %s = $1"

	var err error
	// first call, prepare statement for reuse
	if r.getEmailStmt == nil {
		r.getEmailStmt, err = r.db.PrepareContext(ctx, fmt.Sprintf(SQL, field))

		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", SQL))
			return nil, err
		}
	}

	var user User

	err = r.getEmailStmt.QueryRowContext(ctx, value).Scan(
		&user.ID,
		&user.UserName,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Mobile,
		&user.Address,
		&user.Gender,
		&user.IsVerified,
		&user.PaystackCustomerCode,
	)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", SQL),
			zap.String(field, value),
		)
		return nil, err
	}

	return &user, nil
}

func (r userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return r.getUser(ctx, "email", email)
}

func (r userRepository) GetByCustomerCode(ctx context.Context, customerCode string) (*User, error) {
	return r.getUser(ctx, "paystack_customer_code", customerCode)
}

func (r userRepository) GetById(ctx context.Context, id string) (*User, error) {
	return r.getUser(ctx, "id", id)
}

func (r userRepository) LoginWithEmailPassword(ctx context.Context, email, password string) (*User, error) {
	existingUser, err := r.GetByEmail(ctx, email)
	if err == sql.ErrNoRows {
		return nil, http_helper.ErrUserPwd
	}

	if err != nil {
		return nil, err
	}

	if password != existingUser.Password {
		return nil, http_helper.ErrUserPwd
	}

	return existingUser, nil
}

func (r userRepository) LoginWithEmailPasswordDevice(ctx context.Context, email, password, deviceIdentifier string) (*User, error) {
	existingUser, err := r.LoginWithEmailPassword(ctx, email, password)
	if err != nil {
		return nil, err
	}

	_, err = r.GetCurrentDevice(ctx, existingUser.ID, deviceIdentifier)
	if err != nil {
		return nil, err
	}

	//appVersionId, err := r.idGenerator.IDGenerateFromString(appVersionID)
	//if err != nil {
	//	return nil, err
	//}

	//appVersion, err := r.GetAppVersion(ctx, appVersionId)
	//if err != nil {
	//	return nil, err
	//}
	//
	//existingUser.LatestAppVersion = *appVersion
	return existingUser, nil
}

func (r userRepository) ForgotPassword(request ForgotPasswordPayload) (*OTPResponse, error) {
	ctx := context.Background()
	otpResponse, err := r.requestOTP(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	err = r.saveOTP(ctx, *otpResponse)
	if err != nil {
		return nil, err
	}
	return otpResponse, nil
}

func (r userRepository) VerifyPasswordResetOTP(request *VerifyOTP) (*User, error) {
	ctx := context.Background()
	user, err := r.getUser(ctx, "email", request.Target)
	if err != nil {
		return nil, err
	}
	err = r.verifyOTP(ctx, request)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r userRepository) ResetPassword(ctx context.Context, userId uuid.UUID, request ResetPasswordPayload) (uuid.UUID, error) {
	user, err := r.GetById(context.Background(), userId.String())
	if err != nil {
		return uuid.Nil, err
	}
	return r.performPasswordChange(ctx, user, request.Email, request.Password)
}

func (r userRepository) ChangePassword(ctx context.Context, userId uuid.UUID, request ChangePasswordPayload) (uuid.UUID, error) {
	user, err := r.GetById(context.Background(), userId.String())
	if err != nil {
		return uuid.Nil, err
	}
	if user.Password != request.OldPassword {
		return uuid.Nil, errors.New("incorrect old password")
	}

	return r.performPasswordChange(ctx, user, request.Email, request.NewPassword)
}

func (r userRepository) performPasswordChange(ctx context.Context, user *User, email, password string) (uuid.UUID, error) {
	if user.Email != email {
		return uuid.Nil, errors.New("invalid user")
	}
	sqlQuery := `UPDATE users SET password=$2 WHERE id = $1 RETURNING id`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return uuid.Nil, err
	}
	var userID uuid.UUID
	row := stmt.QueryRowContext(ctx, user.ID, password)
	if err := row.Scan(&userID); err != nil {
		r.log.Error("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *userRepository) UpdatePaystack(ctx context.Context, user *User) (uuid.UUID, error) {
	sqlQuery := `UPDATE users SET paystack_customer_code=$1, paystack_customer_id=$2, is_verified=$3 WHERE id = $4 RETURNING id`
	stmt, err := r.db.Prepare(sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return uuid.Nil, err
	}
	var userID uuid.UUID
	row := stmt.QueryRowContext(
		ctx,
		user.PaystackCustomerCode,
		user.PaystackCustomerId,
		user.IsVerified,
		user.ID,
	)
	if err := row.Scan(&userID); err != nil {
		r.log.Error("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return uuid.Nil, err
	}
	return userID, nil
}

func (r userRepository) requestOTP(ctx context.Context, target string) (*OTPResponse, error) {
	user, err := r.GetByEmail(ctx, target)
	if err != nil {
		return nil, err
	}
	expirationDuration := time.Duration(5) * time.Minute

	otpResponse := OTPResponse{
		User:                fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Target:              target,
		OTP:                 r.otpGenerator.Generate(),
		ExpireTimeInSeconds: time.Now().Add(expirationDuration).Unix(),
	}

	return &otpResponse, nil
}

func (r userRepository) getOTP(ctx context.Context, target string) (*UserPasswordToken, error) {
	getQuery := `SELECT * FROM user_password_token WHERE email = $1`
	tmpSmt, err := r.db.PrepareContext(ctx, getQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", getQuery))
		return nil, err
	}

	var passwordToken UserPasswordToken
	err = tmpSmt.QueryRowContext(ctx, target).Scan(&passwordToken.ID, &passwordToken.Email, &passwordToken.PasswordResetToken, &passwordToken.PasswordResetAt, &passwordToken.Validated)

	return &passwordToken, err
}

func (r userRepository) saveOTP(ctx context.Context, request OTPResponse) error {
	_, row := r.getOTP(ctx, request.Target)
	var userPasswordTokenID uuid.UUID
	switch {
	case row == sql.ErrNoRows:
		sqlQuery := `INSERT INTO user_password_token(email, password_reset_token, password_reset_at) VALUES ($1, $2, $3) RETURNING id`
		tmpSmt, err := r.db.PrepareContext(ctx, sqlQuery)
		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}

		err = tmpSmt.QueryRowContext(ctx, request.Target, request.OTP, request.ExpireTimeInSeconds).Scan(&userPasswordTokenID)
		if err != nil {
			r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}
	case row != sql.ErrNoRows:
		sqlQuery := `UPDATE user_password_token SET password_reset_token=$2, password_reset_at=$3, validated=$4 WHERE email = $1 RETURNING id`

		tmpSmt, err := r.db.PrepareContext(ctx, sqlQuery)
		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}

		err = tmpSmt.QueryRowContext(ctx, request.Target, request.OTP, request.ExpireTimeInSeconds, false).Scan(&userPasswordTokenID)
		if err != nil {
			r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}
	}
	return nil
}

func (r userRepository) verifyOTP(ctx context.Context, request *VerifyOTP) error {
	var userPasswordTokenID uuid.UUID
	passwordToken, err := r.getOTP(ctx, request.Target)
	if err != nil {
		return err
	}

	if passwordToken.Validated {
		return errors.New("already validated verification code")
	}

	if time.Unix(passwordToken.PasswordResetAt, 0).Before(time.Now()) {
		err = r.updateOTPValidation(ctx, userPasswordTokenID)
		if err != nil {
			return err
		}
		return errors.New("expired verification code")
	}

	if passwordToken.PasswordResetToken != request.OTP {
		r.log.Error("msg", zap.String("error", "invalid verification code"))
		return errors.New("invalid verification code")
	}
	err = r.updateOTPValidation(ctx, userPasswordTokenID)
	if err != nil {
		return err
	}
	return nil
}

func (r userRepository) updateOTPValidation(ctx context.Context, tokenID uuid.UUID) error {
	sqlQuery := `UPDATE user_password_token SET validated=$1 RETURNING id`
	tmpSmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return err
	}

	err = tmpSmt.QueryRowContext(ctx, "true").Scan(&tokenID)
	if err != nil {
		r.log.Error("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return err
	}

	return nil
}

func (r userRepository) getFavourites(ctx context.Context, userId uuid.UUID) (*Favourites, error) {
	getQuery := `SELECT id, fav FROM favourites WHERE user_id = $1`

	tmpSmt, err := r.db.PrepareContext(ctx, getQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", getQuery))
		return nil, err
	}
	var (
		favouriteID uuid.UUID
		savedFav    ScanFavourite
	)

	err = tmpSmt.QueryRowContext(ctx, userId).Scan(&favouriteID, &savedFav)
	if err != nil {
		r.log.Error("QueryRowContext get favourites", zap.String("getFavourites", err.Error()), zap.String("query", getQuery))
		return nil, err
	}
	favourite := &Favourites{
		ID:     favouriteID,
		UserID: userId,
		Fav:    savedFav,
	}

	return favourite, err
}

func (r userRepository) GetFavourites(ctx context.Context, userId uuid.UUID) ([]*FavMessage, int, error) {
	var messageIds []uuid.UUID
	var as FavMessage

	favourites, err := r.getFavourites(ctx, userId)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			r.log.Info("error", zap.String("error", err.Error()))
			return nil, 0, err
		}
	}
	for _, fav := range favourites.Fav {
		messageIds = append(messageIds, fav.MessageID)
		as.Fav = fav.Fav
	}

	sqlQuery := `SELECT id, title, author, image_url, audio_url, description FROM audio_messages WHERE id = ANY($1)`

	var favss []*FavMessage
	getFavsStmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", sqlQuery),
		)
		return nil, 0, err
	}
	rows, err := getFavsStmt.QueryContext(ctx, messageIds)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return nil, 0, err
	}
	for rows.Next() {
		as.ID = favourites.ID
		as.UserID = favourites.UserID
		if err := rows.Scan(
			&as.MessageID,
			&as.Title,
			&as.Author,
			&as.ImageUrl,
			&as.AudioUrl,
			&as.Description,
		); err != nil {
			r.log.Error("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", sqlQuery),
			)
			return nil, 0, err
		}
		var ty = as
		favss = append(favss, &ty)
	}
	return favss, 0, nil
}

func (r userRepository) DeleteFavourite(ctx context.Context, messageId, userID uuid.UUID) (uuid.UUID, error) {
	const sqlQuery = `UPDATE favourites SET fav = fav - Cast((SELECT position - 1 FROM favourites, jsonb_array_elements(fav) with ordinality arr(item_object, position) WHERE user_id=$1 and item_object->>'message_id' = $2) as int) WHERE user_id=$1;`

	tmpSmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return uuid.Nil, err
	}

	err = tmpSmt.QueryRowContext(ctx, userID, messageId).Scan()
	if err == sql.ErrNoRows {
		return messageId, nil
	}

	return uuid.Nil, err
}

func (r userRepository) CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error) {
	var favouriteID uuid.UUID

	allFavs, err := r.getFavourites(ctx, favourite.UserID)
	switch {
	case err == sql.ErrNoRows:
		sqlQuery := `INSERT INTO favourites (user_id, fav) SELECT $1, $2 WHERE NOT EXISTS (SELECT user_id FROM favourites WHERE user_id = $1) RETURNING id`
		tmpSmt, err := r.db.Prepare(sqlQuery)
		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}
		favs, err := SaveFavouritesJSONBValue(favourite.Fav)
		if err != nil {
			r.log.Error("error", zap.String("marshal field", err.Error()))
			return nil, err
		}

		err = tmpSmt.QueryRowContext(ctx, favourite.UserID, favs).Scan(&favouriteID)
		if err != nil {
			r.log.Error("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}
		favourite.ID = favouriteID
	case err != sql.ErrNoRows:
		const sqlQuery = `UPDATE favourites SET fav = COALESCE(fav, '[]'::jsonb) || $2 ::jsonb WHERE user_id=$1;`

		favs, err := SaveFavouritesJSONBValue(favourite.Fav)
		if err != nil {
			r.log.Error("error", zap.String("marshal field", err.Error()))
			return nil, err
		}
		_, err = r.db.ExecContext(ctx, sqlQuery, favourite.UserID, favs)
		if err != nil {
			r.log.Error("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}
		favourite.ID = allFavs.ID

	}

	return favourite, nil
}

func (r userRepository) UpdateUserProfile(ctx context.Context, userId uuid.UUID, user *UpdateUser) (uuid.UUID, error) {
	id := struct {
		Id uuid.UUID `sql:"id"`
	}{
		Id: userId,
	}

	whereQuery := r.queryHandler.WhereQueryHelper(id)
	setQuery := r.queryHandler.SetQueryHelper(*user)
	sqlQuery := `UPDATE users SET ` + setQuery + " WHERE " + whereQuery + " RETURNING id"
	err := r.db.QueryRowContext(ctx, sqlQuery).Scan(&userId)
	if err != nil {
		r.log.Error("UpdateUserProfile", zap.String("error scanning row", err.Error()))
		return uuid.Nil, err
	}

	return userId, nil
}

func (r userRepository) GetDevices(ctx context.Context, userId string) (*DeviceManager, error) {
	getQuery := `SELECT id, user_id, devices FROM devices WHERE user_id = $1`

	tmpSmt, err := r.db.PrepareContext(ctx, getQuery)
	if err != nil {
		r.log.Error("PrepareContext get devices", zap.String("getDevices", err.Error()), zap.String("query", getQuery))
		return nil, err
	}

	var (
		devices      DeviceManager
		savedDevices ScanDevices
	)

	err = tmpSmt.QueryRowContext(ctx, userId).Scan(&devices.ID, &devices.UserID, &savedDevices)
	if err != nil {
		r.log.Error("QueryRowContext get devices", zap.String("getDevices", err.Error()), zap.String("query", getQuery))
	}
	devices.Devices = savedDevices
	return &devices, err
}

func (r userRepository) BuildDevice(ctx context.Context, input *DeviceManager, email string) (*DeviceManager, error) {
	user, err := r.getUser(ctx, "email", email)
	if err != nil {
		return nil, err
	}
	var devices []Device
	for _, device := range input.Devices {
		deviceID, err := r.idGenerator.IDGenerate()
		if err != nil {
			return nil, err
		}

		device.ID = deviceID
		device.DateAdded = sql.NullString{Valid: true, String: time.Now().Format(time.RFC3339)}
		devices = append(devices, device)
	}
	createdUserId := user.ID

	deviceManager := DeviceManager{
		UserID:  createdUserId,
		Devices: devices,
	}

	allDevices, err := r.GetDevices(ctx, createdUserId)
	switch {
	case err == sql.ErrNoRows:
		sqlQuery := `INSERT INTO devices (user_id, devices) SELECT $1, $2 WHERE NOT EXISTS (SELECT user_id FROM devices WHERE user_id = $1) RETURNING id`
		tmpSmt, err := r.db.Prepare(sqlQuery)
		if err != nil {
			r.log.Error("PrepareContext build devices", zap.String("buildDevice", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}
		devs, err := SaveDevicesJSONBValue(deviceManager.Devices)
		if err != nil {
			r.log.Info("error", zap.String("marshal field", err.Error()))
			return nil, err
		}

		err = tmpSmt.QueryRowContext(ctx, deviceManager.UserID, devs).Scan(&deviceManager.ID)
		if err != nil {
			r.log.Error("QueryRowContext build devices", zap.String("buildDevices", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}

	case err != sql.ErrNoRows:
		const sqlQuery = `UPDATE devices SET devices = COALESCE(devices, '[]'::jsonb) || $2 ::jsonb WHERE user_id=$1;`

		devs, err := SaveDevicesJSONBValue(deviceManager.Devices)
		if err != nil {
			r.log.Info("error", zap.String("marshal field", err.Error()))
			return nil, err
		}
		_, err = r.db.ExecContext(ctx, sqlQuery, deviceManager.UserID, devs)
		if err != nil {
			r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}
		deviceManager.ID = allDevices.ID

	}

	return &deviceManager, nil
}

func (r userRepository) DeleteDevice(ctx context.Context, identifier, userID string) (string, error) {
	const sqlQuery = `UPDATE devices SET devices = devices - Cast((SELECT position - 1 FROM devices, jsonb_array_elements(devices) with ordinality arr(item_object, position) WHERE user_id=$1 and item_object->>'identifier' = $2) as int) WHERE user_id=$1;`

	tmpSmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("PrepareContext Delete Device", zap.String("DeleteDevice", err.Error()), zap.String("query", sqlQuery))
		return "", err
	}
	err = tmpSmt.QueryRowContext(ctx, userID, identifier).Scan()
	if err == sql.ErrNoRows {
		return identifier, nil
	}

	return "", err
}

func (r userRepository) rollBack(ctx context.Context, rbQuery, id string) (string, error) {

	tmpSmt, err := r.db.PrepareContext(ctx, rbQuery)
	if err != nil {

		r.logSQLError(rbQuery, "Rollback..", err)
		return "", err
	}
	err = tmpSmt.QueryRowContext(ctx, id).Scan()
	if err == sql.ErrNoRows {
		return "", err
	}

	return id, err
}

func (r userRepository) GetCurrentDevice(ctx context.Context, userId, identifier string) (*DeviceManager, error) {
	getQuery := `SELECT id, user_id, arr.item_object FROM devices,jsonb_array_elements(devices) with ordinality arr(item_object, position) WHERE user_id=$1 and item_object->>'identifier'=$2`

	tmpSmt, err := r.db.PrepareContext(ctx, getQuery)
	if err != nil {
		r.log.Error("PrepareContext get devices", zap.String("getDevices", err.Error()), zap.String("query", getQuery))
		return nil, err
	}

	var (
		deviceID    string
		userID      string
		devices     DeviceManager
		savedDevice ScanDevice
	)

	err = tmpSmt.QueryRowContext(ctx, userId, identifier).Scan(&deviceID, &userID, &savedDevice)
	switch {
	case err == sql.ErrNoRows:
		r.log.Error("QueryRowContext get current device", zap.String("getCurrentDevice", err.Error()), zap.String("query", getQuery), zap.String("msg: ", "device does not exist"))
		return nil, errors.New("new device? device does not exist")
	}

	devices.Devices = append(devices.Devices, Device{
		ID:          savedDevice.ID,
		Who:         savedDevice.Who,
		Identifier:  savedDevice.Identifier,
		Os:          savedDevice.Os,
		Brand:       savedDevice.Brand,
		Version:     savedDevice.Version,
		Status:      savedDevice.Status,
		DateAdded:   savedDevice.DateAdded,
		LastUpdated: savedDevice.LastUpdated,
	})
	return &devices, err
}

func (r userRepository) UpdateDevice(ctx context.Context, userId, status, identifier string) (*DeviceManager, error) {
	sqlQuery := `UPDATE devices SET devices = 
    (
		SELECT jsonb_agg(
			CASE
			  WHEN elem->>'identifier' = $2 THEN
				jsonb_set(elem, '{status}', $3)
			  ELSE
				elem
			END
		)
		FROM devices, LATERAL jsonb_array_elements(devices) AS arr(elem)
		WHERE user_id = $1
	)
	WHERE user_id = $1;`

	tmpSmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("PrepareContext Update Device", zap.String("UpdateDevice", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}

	s := fmt.Sprintf("\"%s\"", status)
	err = tmpSmt.QueryRowContext(ctx, userId, identifier, s).Scan()
	if err == sql.ErrNoRows {
		return &DeviceManager{UserID: userId}, nil
	}

	return nil, nil
}

func (r userRepository) UpdateAppVersion(ctx context.Context, version VersionManager) (uuid.UUID, error) {
	id := struct {
		Id string `sql:"id"`
	}{
		Id: version.ID,
	}
	whereQuery := r.queryHandler.WhereQueryHelper(id)
	setQuery := r.queryHandler.SetQueryHelper(version)

	sqlQuery := `UPDATE app_version SET ` + setQuery + " WHERE " + whereQuery + " RETURNING id"
	var versionID uuid.UUID
	err := r.db.QueryRowContext(ctx, sqlQuery).Scan(&versionID)
	if err != nil {
		r.log.Error("QueryRowContext Update App Version", zap.String("UpdateAppVersion", err.Error()))
		return uuid.Nil, err
	}
	return versionID, nil

}

func (r userRepository) GetAppVersion(ctx context.Context, versionID uuid.UUID) (*VersionManager, error) {
	sqlQuery := `SELECT * FROM app_version WHERE id=$1`

	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("PrepareContext Get App Version", zap.String("GetAppVersion", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}
	var version VersionManager
	err = stmt.QueryRowContext(ctx, versionID).Scan(
		&version.ID,
		&version.Version,
		&version.Force,
		&version.DateAdded,
		&version.LastUpdated,
	)
	if err != nil {
		r.log.Error("QueryRowContext Get App Version", zap.String("GetAppVersion", err.Error()), zap.String("query", sqlQuery))
		return nil, http_helper.ErrNotFound

	}
	return &version, nil
}

func (r userRepository) UpdateUserIsVerified(ctx context.Context, userId string, verifyField IsVerifiedEnum) error {
	query := `UPDATE users SET is_verified=$1 WHERE id=$2 RETURNING id;`

	fmt.Println(verifyField, userId, "userId")
	err := r.db.QueryRowContext(ctx, query, verifyField, userId).Scan(&userId)

	if err != nil {
		r.log.Error("QueryRowContext Update users isVerified", zap.String("Update users is_verified field", err.Error()))
		return err
	}
	return nil
}
