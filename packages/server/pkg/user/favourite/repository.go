package favourite

import (
	"context"
	"database/sql"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type Repository interface {
	CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error)
	GetFavourites(ctx context.Context, userId uuid.UUID) ([]*FavMessage, int, error)
	DeleteFavourite(ctx context.Context, messageId, userId uuid.UUID) (uuid.UUID, error)
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

func (r favoriteRepository) getFavourites(ctx context.Context, userId uuid.UUID) (*Favourites, error) {
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

	err = tmpSmt.QueryRowContext(ctx, "61e346e0-dbb8-40ef-8c79-65612d0a69a1").Scan(&favouriteID, &savedFav)
	favourite := &Favourites{
		ID:     favouriteID,
		UserID: userId,
		Fav:    savedFav,
	}

	return favourite, err
}

func (r favoriteRepository) GetFavourites(ctx context.Context, userId uuid.UUID) ([]*FavMessage, int, error) {
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

func (r favoriteRepository) DeleteFavourite(ctx context.Context, messageId, userID uuid.UUID) (uuid.UUID, error) {
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

func (r favoriteRepository) CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error) {
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
