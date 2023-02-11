package user

import (
	"context"
	"database/sql"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"go.uber.org/zap"
)

type Repository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	ForgotPassword(request ForgotPasswordPayload, passwordResetToken string) (*User, error)
	VerifyPasswordToken(request ResetPasswordPayload, passwordTokenParam string) (string, error)
	ResetPassword(request ResetPasswordPayload) (int, error)
	Close() error
}

type userRepository struct {
	db           *sql.DB
	log          *zap.Logger
	getEmailStmt *sql.Stmt
	getIdStmt    *sql.Stmt
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

	var createdUserId int

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

func (r userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
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
		"is_verified " +
		"FROM users WHERE email = $1"

	var err error
	// first call, prepare statement for reuse
	if r.getEmailStmt == nil {
		r.getEmailStmt, err = r.db.PrepareContext(ctx, SQL)

		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", SQL))
			return nil, err
		}
	}

	var user User

	err = r.getEmailStmt.QueryRowContext(ctx, email).Scan(
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
	)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", SQL),
			zap.String("email", email),
		)
		return nil, err
	}

	return &user, nil
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

func (r userRepository) ForgotPassword(request ForgotPasswordPayload, passwordResetToken string) (*User, error) {
	ctx := context.Background()
	user, err := r.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	err = r.GetUserPasswordToken(user, request.Email, passwordResetToken)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r userRepository) GetUserPasswordToken(user *User, email, passwordResetToken string) error {
	userPasswordToken := UserPasswordToken{
		Email:              user.Email,
		PasswordResetToken: passwordResetToken,
		PasswordResetAt:    time.Now().Add(time.Minute * 15).Unix(),
	}
	getQuery := `SELECT id FROM user_password_token WHERE email = $1`
	tmpSmt, err := r.db.Prepare(getQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", getQuery))
		return err
	}

	var userPasswordTokenID int
	row := tmpSmt.QueryRow(email).Scan(&userPasswordTokenID)
	switch {
	case row == sql.ErrNoRows:
		sqlQuery := `INSERT INTO user_password_token(email, password_reset_token, password_reset_at) VALUES ($1, $2, $3) RETURNING id`
		tmpSmt, err := r.db.Prepare(sqlQuery)
		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}

		err = tmpSmt.QueryRow(userPasswordToken.Email, userPasswordToken.PasswordResetToken, userPasswordToken.PasswordResetAt).Scan(&userPasswordTokenID)
		if err != nil {
			r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}
	case row != sql.ErrNoRows:
		sqlQuery := `UPDATE user_password_token SET password_reset_token=$2, password_reset_at=$3 WHERE email = $1 RETURNING id`

		tmpSmt, err := r.db.Prepare(sqlQuery)
		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}

		err = tmpSmt.QueryRow(userPasswordToken.Email, userPasswordToken.PasswordResetToken, userPasswordToken.PasswordResetAt).Scan(&userPasswordTokenID)
		if err != nil {
			r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return err
		}
	}
	return nil
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

func (r *userRepository) ResetPassword(request ResetPasswordPayload) (int, error) {
	sqlQuery := `UPDATE users SET password_hash=$2 WHERE email = $1 RETURNING id`
	stmt, err := r.db.Prepare(sqlQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return 0, err
	}
	var userID int
	row := stmt.QueryRow(request.Email, request.Password)
	if err := row.Scan(&userID); err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return 0, err
	}
	return userID, nil
}
