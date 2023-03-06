package favourite

import (
	"context"
	"database/sql"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type Repository interface {
	CreateFavourite(ctx context.Context, favourite *Favourite) (*Favourite, error)
	GetFavourites(ctx context.Context) ([]*Favourite, int, error)
	DeleteFavourite(ctx context.Context, favId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error)
	Close() error
}

type favoriteRepository struct {
	db           *sql.DB
	log          *zap.Logger
	getEmailStmt *sql.Stmt
	getIdStmt    *sql.Stmt
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &favoriteRepository{db: db, log: logger}
}

func (r favoriteRepository) Close() error {
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

func (r favoriteRepository) CreateFavourite(ctx context.Context, favourite *Favourite) (*Favourite, error) {

	const sqlQuery = `INSERT INTO favourites (user_id, message_id, series_id, fav, date_added, deleted_at) VALUES  ($1, $2, $3, $4, $5, $6) RETURNING id`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.log.Info("msg", zap.String("error starting transaction", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}
	defer tx.Rollback()

	tmpSmt, err := tx.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}
	var (
		favouriteID uuid.UUID
	)

	err = tmpSmt.QueryRowContext(ctx,
		favourite.UserID,
		favourite.MessageID,
		favourite.SeriesID,
		favourite.Fav,
		favourite.DateAdded,
		favourite.DeletedAt,
	).Scan(&favouriteID)
	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	favourite.ID = favouriteID

	return favourite, nil
}

func (r favoriteRepository) GetFavourites(ctx context.Context) ([]*Favourite, int, error) {
	const SQL = "SELECT * FROM favourites"

	var favs []*Favourite
	getFavsStmt, err := r.db.PrepareContext(ctx, SQL)

	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", SQL),
		)
		return favs, 0, err
	}

	rows, err := getFavsStmt.QueryContext(ctx)

	defer rows.Close()

	if err == sql.ErrNoRows {
		return favs, 0, err
	}

	for rows.Next() {
		var as Favourite

		if err := rows.Scan(
			&as.ID,
			&as.UserID,
			&as.MessageID,
			&as.SeriesID,
			&as.Fav,
			&as.DateAdded,
			&as.DeletedAt,
		); err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", SQL),
			)
			return favs, 0, err
		}

		favs = append(favs, &as)
	}

	return favs, 0, nil
}

func (r favoriteRepository) DeleteFavourite(ctx context.Context, favId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error) {
	return uuid.Nil, nil
}
