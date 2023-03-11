package user

import (
	"context"
	"database/sql"
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
	ForgotPassword(request ForgotPasswordPayload, passwordResetToken string) (*User, error)
	VerifyPasswordToken(request ResetPasswordPayload, passwordTokenParam string) (string, error)
	ResetPassword(request ResetPasswordPayload) (uuid.UUID, error)
	UpdatePaystack(ctx context.Context, user *User) (uuid.UUID, error)
	CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error)
	GetFavourites(ctx context.Context, userId uuid.UUID) ([]*FavMessage, int, error)
	DeleteFavourite(ctx context.Context, messageId, userId uuid.UUID) (uuid.UUID, error)
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

	var userPasswordTokenID uuid.UUID
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

func (r userRepository) getFavourites(ctx context.Context, userId uuid.UUID) (*Favourites, error) {
	getQuery := `SELECT id, fav FROM favourites WHERE user_id = $1`

	tmpSmt, err := r.db.PrepareContext(ctx, getQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", getQuery))
		return nil, err
	}
	var (
		favouriteID uuid.UUID
		savedFav    Favourite
	)

	err = tmpSmt.QueryRowContext(ctx, userId).Scan(&favouriteID, &savedFav)
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

	sqlQuery := `SELECT id, series_id, title, author, image_url, audio_url, description FROM audio_messages WHERE id = ANY($1)`

	var favss []*FavMessage
	getFavsStmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Info("msg",
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
			&as.SeriesID,
			&as.Title,
			&as.Author,
			&as.ImageUrl,
			&as.AudioUrl,
			&as.Description,
		); err != nil {
			r.log.Info("msg",
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
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
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
		favs, err := Value(favourite.Fav)
		if err != nil {
			r.log.Info("error", zap.String("marshal field", err.Error()))
			return nil, err
		}

		err = tmpSmt.QueryRowContext(ctx, favourite.UserID, favs).Scan(&favouriteID)
		if err != nil {
			r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}
		favourite.ID = favouriteID
	case err != sql.ErrNoRows:
		const sqlQuery = `UPDATE favourites SET fav = COALESCE(fav, '[]'::jsonb) || $2 ::jsonb WHERE user_id=$1;`

		favs, err := Value(favourite.Fav)
		if err != nil {
			r.log.Info("error", zap.String("marshal field", err.Error()))
			return nil, err
		}
		_, err = r.db.ExecContext(ctx, sqlQuery, favourite.UserID, favs)
		if err != nil {
			r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
			return nil, err
		}
		favourite.ID = allFavs.ID

	}

	return favourite, nil
}
