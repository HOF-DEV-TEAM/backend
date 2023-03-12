package user

import (
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
	Create(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetById(ctx context.Context, id string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	ForgotPassword(request ForgotPasswordPayload, passwordResetToken string) (*OTPResponse, error)
	VerifyPasswordToken(request ResetPasswordPayload, passwordTokenParam string) (string, error)
	ResetPassword(request ResetPasswordPayload) (uuid.UUID, error)
	UpdatePaystack(ctx context.Context, user *User) (uuid.UUID, error)
	Close() error
}

type userRepository struct {
	db           *sql.DB
	log          *zap.Logger
	getEmailStmt *sql.Stmt
	getIdStmt    *sql.Stmt
	otpGenerator OtpGenerator
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &userRepository{db: db, log: logger}
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

func (r userRepository) Create(ctx context.Context, user *User) (*User, error) {
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
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	defer tx.Rollback()

	tmpSmt, err := tx.PrepareContext(ctx, SQL)

	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	var createdUserId string

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
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	user.ID = createdUserId
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

func (r userRepository) GetById(ctx context.Context, id string) (*User, error) {
	return r.getUser(ctx, "id", id)
}

func (r userRepository) Login(ctx context.Context, email, password string) (*User, error) {
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

func (r userRepository) ForgotPassword(request ForgotPasswordPayload, passwordResetToken string) (*OTPResponse, error) {
	ctx := context.Background()
	otpRequest := OTPRequest{
		Target: request.Email,
	}
	otpResponse, err := r.requestOTP(ctx, otpRequest)
	if err != nil {
		return nil, err
	}

	err = r.saveOTP(ctx, *otpResponse)
	if err != nil {
		return nil, err
	}
	return otpResponse, nil
}

func (r *userRepository) VerifyPasswordToken(request ResetPasswordPayload, passwordTokenParam string) (string, error) {
	currentTime := time.Now().Unix()
	sqlQuery := `SELECT password_reset_token FROM user_password_token WHERE email = $1 AND password_reset_token=$2 AND password_reset_at >= $3`
	stmt, err := r.db.Prepare(sqlQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return "", err
	}
	var userPasswordToken UserPasswordToken
	row := stmt.QueryRow(request.Email, passwordTokenParam, currentTime)
	if err := row.Scan(&userPasswordToken.PasswordResetToken); err != nil {
		return "", err
	}

	return userPasswordToken.PasswordResetToken, nil
}

func (r *userRepository) ResetPassword(request ResetPasswordPayload) (uuid.UUID, error) {
	sqlQuery := `UPDATE users SET password_hash=$2 WHERE email = $1 RETURNING id`
	stmt, err := r.db.Prepare(sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return uuid.Nil, err
	}
	var userID uuid.UUID
	row := stmt.QueryRow(request.Email, request.Password)
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

func (r userRepository) requestOTP(ctx context.Context, request OTPRequest) (*OTPResponse, error) {
	_, err := r.GetByEmail(ctx, request.Target)
	if err != nil {
		return nil, err
	}
	expirationDuration := time.Duration(90) * time.Second

	otpResponse := OTPResponse{
		Target:              request.Target,
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
	err = tmpSmt.QueryRowContext(ctx, target).Scan(&passwordToken.ID, &passwordToken.Email, &passwordToken.PasswordResetToken, &passwordToken.PasswordResetAt)

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
		sqlQuery := `UPDATE user_password_token SET password_reset_token=$2, password_reset_at=$3 WHERE email = $1 RETURNING id`

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
	}
	return nil
}

func (r userRepository) verifyOTP(ctx context.Context, request OTPResponse) error {
	var userPasswordTokenID uuid.UUID
	passwordToken, err := r.getOTP(ctx, request.Target)
	if err != nil {
		return err
	}

	if passwordToken.Validated {
		return errors.New("already validated verification code")
	}

	if time.Unix(passwordToken.PasswordResetAt, 0).Before(time.Now()) {
		return errors.New("expired verification code")
	}

	sqlQuery := `INSERT INTO user_password_token(validated) VALUES ($1) RETURNING id`
	tmpSmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return err
	}

	err = tmpSmt.QueryRowContext(ctx, "true").Scan(&userPasswordTokenID)
	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return err
	}

	return nil
}
