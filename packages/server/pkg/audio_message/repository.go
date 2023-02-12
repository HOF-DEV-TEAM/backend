package audio_message

import (
	"context"
	"database/sql"

	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"go.uber.org/zap"
)

type Repository interface {
	CreateAudioMessage(ctx context.Context, audioMessage *AudioMessage) (*AudioMessage, error)
	CreateAudioSeries(ctx context.Context, audioSeries *AudioSeries) (*AudioSeries, error)
	GetAudioMessages(ctx context.Context, search *Filter) ([]*AudioMessage, int, error)
	GetAudioSeries(ctx context.Context) ([]*AudioSeries, int, error)
	GetAudioMessageByID(ctx context.Context, messageId string) (*AudioMessage, error)
	GetAudioSeriesByID(ctx context.Context, seriesId string) (*AudioSeries, error)
	Close() error
}

type audioMessageRepository struct {
	db           *sql.DB
	log          *zap.Logger
	getEmailStmt *sql.Stmt
	getIdStmt    *sql.Stmt
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &audioMessageRepository{db: db, log: logger}
}

func (r audioMessageRepository) Close() error {
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

func (r audioMessageRepository) CreateAudioMessage(ctx context.Context, audioMessage *AudioMessage) (*AudioMessage, error) {
	// sql insert query, primary key provided by autoincrement
	const SQL = "INSERT INTO audio_messages (" +
		"title," +
		"author," +
		"image_url," +
		"audio_url," +
		"description," +
		"date_added," +
		"last_updated," +
		"series_id" +
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8) " +
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

	var createdAudioMessageId string

	err = tmpSmt.QueryRowContext(ctx,
		audioMessage.Title,
		audioMessage.Author,
		audioMessage.ImageUrl,
		audioMessage.AudioUrl,
		audioMessage.Description,
		audioMessage.DateAdded,
		audioMessage.LastUpdated,
		audioMessage.SeriesID,
	).Scan(&createdAudioMessageId)

	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	audioMessage.ID = createdAudioMessageId
	return audioMessage, nil
}

func (r audioMessageRepository) CreateAudioSeries(ctx context.Context, audioSeries *AudioSeries) (*AudioSeries, error) {
	// sql insert query, primary key provided by autoincrement
	const SQL = "INSERT INTO audio_series (" +
		"title," +
		"author," +
		"image_url," +
		"description," +
		"date_added," +
		"last_updated" +
		") VALUES ($1, $2, $3, $4, $5, $6) " +
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

	var createdAudioSeriesId string

	err = tmpSmt.QueryRowContext(ctx,
		audioSeries.Title,
		audioSeries.Author,
		audioSeries.ImageUrl,
		audioSeries.Description,
		audioSeries.DateAdded,
		audioSeries.LastUpdated,
	).Scan(&createdAudioSeriesId)

	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	audioSeries.ID = createdAudioSeriesId
	return audioSeries, nil
}

func (r audioMessageRepository) GetAudioSeries(ctx context.Context) ([]*AudioSeries, int, error) {
	const SQL = "SELECT * FROM audio_series"

	var audioSeries []*AudioSeries
	getAudioSeriesStmt, err := r.db.PrepareContext(ctx, SQL)

	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", SQL),
		)
		return audioSeries, 0, err
	}

	rows, err := getAudioSeriesStmt.QueryContext(ctx)

	defer rows.Close()

	if err == sql.ErrNoRows {
		return audioSeries, 0, err
	}

	for rows.Next() {
		var as AudioSeries

		if err := rows.Scan(
			&as.ID,
			&as.Title,
			&as.Author,
			&as.Description,
			&as.ImageUrl,
			&as.DateAdded,
			&as.LastUpdated,
		); err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", SQL),
			)
			return audioSeries, 0, err
		}

		audioSeries = append(audioSeries, &as)
	}

	return audioSeries, 0, nil
}

func (r audioMessageRepository) getAudioMessages(ctx context.Context, query string, queryParams []interface{}) ([]*AudioMessage, int, error) {
	var audioMessages []*AudioMessage
	getAudioMessagesStmt, err := r.db.PrepareContext(ctx, query)

	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", query),
		)
		return audioMessages, 0, err
	}

	rows, err := getAudioMessagesStmt.QueryContext(ctx, queryParams...)

	if err == sql.ErrNoRows {
		return audioMessages, 0, err
	}

	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", query),
		)
		return audioMessages, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var as AudioMessage

		if err := rows.Scan(
			&as.ID,
			&as.Title,
			&as.Author,
			&as.ImageUrl,
			&as.AudioUrl,
			&as.Description,
			&as.DateAdded,
			&as.LastUpdated,
			&as.SeriesID,
		); err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", query),
			)
			return audioMessages, 0, err
		}

		audioMessages = append(audioMessages, &as)
	}

	return audioMessages, 0, nil
}

func buildQuery(query string, filter *Filter) (string, []interface{}, error) {
	queryParams := []interface{}{}

	sqlSmt := query	
	switch filter.SeriesID {
	case "", "*":
		break;
	case "?":
		sqlSmt += " WHERE series_id IS NULL"		
	default:		
		sqlSmt += " WHERE series_id=$1"
		queryParams = append(queryParams, filter.SeriesID)
	}
	return sqlSmt, queryParams, nil
}

// TODO: implement pagination
func (r audioMessageRepository) GetAudioMessages(ctx context.Context, search *Filter) ([]*AudioMessage, int, error) {
	var sqlStmt string
	sqlStmt = "SELECT * FROM audio_messages"

	query, queryParams, err := buildQuery(sqlStmt, search)

	if err != nil {
		return []*AudioMessage{}, 0, err
	}
	
	return r.getAudioMessages(ctx, query, queryParams)
}

func (r audioMessageRepository) GetAudioMessageByID(ctx context.Context, messageId string) (*AudioMessage, error) {
	sqlQuery := `SELECT * FROM audio_messages WHERE id=$1`

	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}
	var audioMessage AudioMessage
	err = stmt.QueryRowContext(ctx, messageId).Scan(
		&audioMessage.ID,
		&audioMessage.Title,
		&audioMessage.Author,
		&audioMessage.ImageUrl,
		&audioMessage.AudioUrl,
		&audioMessage.Description,
		&audioMessage.DateAdded,
		&audioMessage.LastUpdated,
		&audioMessage.SeriesID,
	)
	if err != nil {
		r.log.Info("msg", zap.String("error retrieving data", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, http_helper.ErrNotFound

	}
	return &audioMessage, nil
}

func (r audioMessageRepository) GetAudioSeriesByID(ctx context.Context, seriesId string) (*AudioSeries, error) {
	sqlQuery := `SELECT * FROM audio_series WHERE id=$1`

	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}
	var audioSeries AudioSeries
	err = stmt.QueryRowContext(ctx, seriesId).Scan(
		&audioSeries.ID,
		&audioSeries.Title,
		&audioSeries.Author,
		&audioSeries.Description,
		&audioSeries.ImageUrl,
		&audioSeries.DateAdded,
		&audioSeries.LastUpdated,
	)
	if err != nil {
		r.log.Info("msg", zap.String("error retrieving data", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, http_helper.ErrNotFound

	}
	return &audioSeries, nil
}
