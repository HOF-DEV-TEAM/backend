package favourite

import (
	"context"
	"database/sql"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type Repository interface {
	CreateFavourite(ctx context.Context, favourite *Favourite) (*Favourite, error)
	GetFavourites(ctx context.Context) ([]*FavMessage, int, error)
	DeleteFavourite(ctx context.Context, favId uuid.UUID) (uuid.UUID, error)
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
	const sqlQuery = `INSERT INTO favourites (user_id, message_id, series_id, fav, date_added, deleted_at) SELECT $1, $2, $3, $4, $5, $6 WHERE NOT EXISTS (SELECT user_id FROM favourites WHERE user_id = $1 AND message_id = $2) RETURNING id`

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

func (r favoriteRepository) GetFavourites(ctx context.Context) ([]*FavMessage, int, error) {
	sqlQuery := `SELECT favourites.id, favourites.user_id, favourites.fav,audio_messages.id, audio_messages.series_id, audio_messages.title, audio_messages.author, audio_messages.image_url, audio_messages.audio_url, audio_messages.description FROM favourites INNER JOIN audio_messages ON favourites.message_id=audio_messages.id`
	//const SQL = "SELECT * FROM favourites"

	var favs []*FavMessage
	getFavsStmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", sqlQuery),
		)
		return favs, 0, err
	}

	rows, err := getFavsStmt.QueryContext(ctx)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return favs, 0, err
	}

	for rows.Next() {
		var as FavMessage

		if err := rows.Scan(
			&as.ID,
			&as.UserID,
			&as.Fav,
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
			return favs, 0, err
		}

		favs = append(favs, &as)
	}

	return favs, 0, nil
}

func (r favoriteRepository) DeleteFavourite(ctx context.Context, favId uuid.UUID) (uuid.UUID, error) {
	sqlQuery := `DELETE FROM favourites where id=$1`

	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("DeleteFavourite", zap.String("error preparing statement", err.Error()), zap.String("sqlQuery : ", sqlQuery))

		return uuid.Nil, err
	}

	row := stmt.QueryRowContext(ctx, favId)
	if err := row.Scan(); err != nil {
		if err == sql.ErrNoRows {
			return favId, nil
		}
	}
	return uuid.Nil, err
}
