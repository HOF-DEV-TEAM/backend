package audio_message

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/urlqueryhelper"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type Repository interface {
	CreateAudioMessage(ctx context.Context, audioMessage *AudioMessage) (*AudioMessage, error)
	CreateAudioSeries(ctx context.Context, audioSeries *AudioSeries) (*AudioSeries, error)
	GetAudioMessages(ctx context.Context, search *Filter) ([]*AudioMessage, int, error)
	GetAudioSeries(ctx context.Context) ([]*AudioSeries, int, error)
	GetAudioMessageByID(ctx context.Context, messageId uuid.UUID) (*AudioMessage, error)
	GetAudioSeriesByID(ctx context.Context, seriesId uuid.UUID) (*AudioSeries, error)
	UpdateAudioMessagesByID(ctx context.Context, message AudioMessage, messageId uuid.UUID) (uuid.UUID, error)
	UpdateAudioSeriesByID(ctx context.Context, series AudioSeries, seriesId uuid.UUID) (uuid.UUID, error)
	DeleteAudioMessagesByID(ctx context.Context, messageId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error)
	DeleteAudioSeriesByID(ctx context.Context, seriesId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error)
	HomePageDirectory(ctx context.Context) (*Homepage, error)
	CreateMeditation(ctx context.Context, meditation *Meditation) (string, error)
	CreateMeditations(ctx context.Context, meditation []*Meditation) (*MeditationResponse, error)
	UpdateMeditationByID(ctx context.Context, status string, meditationID string) (*string, error)
	GetMeditations(ctx context.Context) ([]Meditation, error)
	GetMeditation(ctx context.Context, meditationId string) (*Meditation, error)
	DeleteMeditationByID(ctx context.Context, meditationId string) (*DefaultResponse, error)
	Close() error
}

type audioMessageRepository struct {
	db           *sql.DB
	log          *zap.Logger
	getEmailStmt *sql.Stmt
	getIdStmt    *sql.Stmt
	queryHandler urlqueryhelper.QueryHelper
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &audioMessageRepository{db: db, log: logger, queryHandler: urlqueryhelper.NewQueryHelper()}
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
	const SQL = "INSERT INTO audio_messages (" +
		"title," +
		"author," +
		"image_url," +
		"audio_url," +
		"description," +
		"date_added," +
		"last_updated," +
		"series_id," +
		"date_released" +
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) " +
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
		audioMessage.DateReleased,
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
		"last_updated," +
		"date_released," +
		"of_the_month" +
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

	var createdAudioSeriesId string

	err = tmpSmt.QueryRowContext(ctx,
		audioSeries.Title,
		audioSeries.Author,
		audioSeries.ImageUrl,
		audioSeries.Description,
		audioSeries.DateAdded,
		audioSeries.LastUpdated,
		audioSeries.DateReleased,
		audioSeries.OfTheMonth,
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
	const SQL = "SELECT * FROM audio_series WHERE deleted_at IS NULL"

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
			&as.DeletedAt,
			&as.DateReleased,
			&as.OfTheMonth,
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
		r.log.Error("msg",
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
			&as.DeletedAt,
			&as.DateReleased,
		); err != nil {
			r.log.Error("msg",
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
		break
	case "?":
		sqlSmt += " AND series_id IS NULL"
	default:
		sqlSmt += " AND series_id=$1"
		queryParams = append(queryParams, filter.SeriesID)
	}
	return sqlSmt, queryParams, nil
}

// TODO: implement pagination
func (r audioMessageRepository) GetAudioMessages(ctx context.Context, search *Filter) ([]*AudioMessage, int, error) {
	var sqlStmt string
	sqlStmt = "SELECT * FROM audio_messages WHERE deleted_at IS NULL"

	query, queryParams, err := buildQuery(sqlStmt, search)

	if err != nil {
		return []*AudioMessage{}, 0, err
	}

	return r.getAudioMessages(ctx, query, queryParams)
}

func (r audioMessageRepository) GetAudioMessageByID(ctx context.Context, messageId uuid.UUID) (*AudioMessage, error) {
	sqlQuery := `SELECT * FROM audio_messages WHERE id=$1 AND deleted_at IS NULL`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
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
		&audioMessage.DeletedAt,
		&audioMessage.DateReleased,
	)
	if err != nil {
		r.log.Error("msg", zap.String("error retrieving data", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, http_helper.ErrNotFound

	}
	return &audioMessage, nil
}

func (r audioMessageRepository) GetAudioSeriesByID(ctx context.Context, seriesId uuid.UUID) (*AudioSeries, error) {
	sqlQuery := `SELECT * FROM audio_series WHERE id=$1 AND deleted_at IS NULL`

	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
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
		&audioSeries.DeletedAt,
		&audioSeries.DateReleased,
		&audioSeries.OfTheMonth,
	)
	if err != nil {
		r.log.Error("msg", zap.String("error retrieving data", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, http_helper.ErrNotFound

	}
	return &audioSeries, nil
}

func (r audioMessageRepository) UpdateAudioMessagesByID(ctx context.Context, message AudioMessage, messageId uuid.UUID) (uuid.UUID, error) {
	id := struct {
		Id uuid.UUID `sql:"id"`
	}{
		Id: messageId,
	}
	whereQuery := r.queryHandler.WhereQueryHelper(id)
	setQuery := r.queryHandler.SetQueryHelper(message)
	sqlQuery := `UPDATE audio_messages SET ` + setQuery + " WHERE " + whereQuery + " RETURNING id"
	err := r.db.QueryRowContext(ctx, sqlQuery).Scan(&messageId)
	if err != nil {
		r.log.Error("UpdateAudioMessagesByID", zap.String("error scanning row", err.Error()))
		return uuid.Nil, err
	}
	return messageId, nil
}

func (r audioMessageRepository) UpdateAudioSeriesByID(ctx context.Context, series AudioSeries, seriesId uuid.UUID) (uuid.UUID, error) {
	id := struct {
		Id uuid.UUID `sql:"id"`
	}{
		Id: seriesId,
	}
	whereQuery := r.queryHandler.WhereQueryHelper(id)
	setQuery := r.queryHandler.SetQueryHelper(series)

	sqlQuery := `UPDATE audio_series SET ` + setQuery + " WHERE " + whereQuery + " RETURNING id"
	err := r.db.QueryRowContext(ctx, sqlQuery).Scan(&seriesId)
	if err != nil {
		r.log.Error("UpdateAudioSeriesByID", zap.String("error scanning row", err.Error()))
		return uuid.Nil, err
	}
	return seriesId, nil
}

func (r audioMessageRepository) DeleteAudioMessagesByID(ctx context.Context, messageId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error) {
	sqlQuery := `UPDATE audio_messages SET deleted_at=$2 WHERE id=$1 RETURNING id`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("DeleteAudioMessagesByID", zap.String("error preparing statement", err.Error()), zap.String("sqlQuery : ", sqlQuery))

		return uuid.Nil, err
	}
	row := stmt.QueryRowContext(ctx, messageId, deletedAt)
	if err := row.Scan(&messageId); err != nil {
		r.log.Error("DeleteAudioMessagesByID", zap.String("error scanning row", err.Error()))
		return uuid.Nil, err
	}
	return messageId, nil
}

func (r audioMessageRepository) DeleteAudioSeriesByID(ctx context.Context, seriesId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error) {
	sqlQuery := `UPDATE audio_series SET deleted_at=$2 WHERE id=$1 RETURNING id`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("DeleteAudioSeriesByID", zap.String("error preparing statement", err.Error()), zap.String("sqlQuery : ", sqlQuery))

		return uuid.Nil, err
	}
	row := stmt.QueryRowContext(ctx, seriesId, deletedAt)
	if err := row.Scan(&seriesId); err != nil {
		r.log.Error("DeleteAudioSeriesByID", zap.String("error scanning row", err.Error()))
		return uuid.Nil, err
	}
	return seriesId, nil
}

func (r audioMessageRepository) HomePageDirectory(ctx context.Context) (*Homepage, error) {

	const (
		meditationSQL = `SELECT * FROM meditation WHERE deleted_at IS NULL AND LOWER(status)='active'`
		seriesSQL     = `SELECT * FROM audio_series WHERE deleted_at IS NULL AND of_the_month=true`
	)

	var audioSeries []*AudioSeries
	getAudioSeriesStmt, err := r.db.PrepareContext(ctx, seriesSQL)
	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", seriesSQL),
		)
		return nil, err
	}

	rows, err := getAudioSeriesStmt.QueryContext(ctx)

	defer rows.Close()

	if err == sql.ErrNoRows {
		return nil, err
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
			&as.DeletedAt,
			&as.DateReleased,
			&as.OfTheMonth,
		); err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", seriesSQL),
			)
			return nil, err
		}

		audioSeries = append(audioSeries, &as)
	}

	var meditation []*Meditation
	getMeditationStmt, err := r.db.PrepareContext(ctx, meditationSQL)
	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", meditationSQL),
		)
		return nil, err
	}

	rows, err = getMeditationStmt.QueryContext(ctx)

	defer rows.Close()

	if err == sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var as Meditation

		if err := rows.Scan(
			&as.ID,
			&as.Name,
			&as.Image,
			&as.Status,
			&as.DateAdded,
			&as.DeletedAt,
		); err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", meditationSQL),
			)
			return nil, err
		}

		meditation = append(meditation, &as)
	}

	home := Homepage{
		AudioSeries: audioSeries,
		Meditation:  meditation,
	}
	return &home, nil
}

func (r audioMessageRepository) CreateMeditation(ctx context.Context, meditation *Meditation) (string, error) {
	sqlStr := "INSERT INTO meditation (name, image_url, status, date_added) VALUES ($1, $2, $3, $4) RETURNING id;"

	stmt, err := r.db.PrepareContext(ctx, sqlStr)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlStr))
		return "", err
	}

	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		r.log.Error("error", zap.String("error", err.Error()), zap.String("query", sqlStr))
		return "", err
	}

	defer tx.Rollback()

	var meditationId string

	err = stmt.QueryRowContext(ctx,
		meditation.Name,
		meditation.Image,
		meditation.Status,
		meditation.DateAdded,
	).Scan(&meditationId)

	if err != nil {
		r.log.Error("error", zap.String("error", err.Error()), zap.String("query", sqlStr))
		return "", err
	}

	err = tx.Commit()

	if err != nil {
		return "", err
	}

	return meditationId, err
}

func (r audioMessageRepository) CreateMeditations(ctx context.Context, meditation []*Meditation) (*MeditationResponse, error) {
	sqlStr := "INSERT INTO meditation (name, image_url, status, date_added) VALUES "
	var vals []interface{}

	if len(meditation) > 10 {
		return nil, errors.New("insertion cannot be greater than 10")
	}
	for i, row := range meditation {
		p1 := i * 4

		sqlStr += fmt.Sprintf("($%d,$%d,$%d,$%d),", p1+1, p1+2, p1+3, p1+4)
		vals = append(vals, row.Name, row.Image, row.Status, row.DateAdded)
	}
	//trim the last ,
	sqlStr = sqlStr[0 : len(sqlStr)-1]

	stmt, err := r.db.PrepareContext(ctx, sqlStr)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlStr))
		return nil, err
	}

	result, err := stmt.ExecContext(ctx, vals...)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &MeditationResponse{AffectedRows: affected}, nil
}

func (r audioMessageRepository) UpdateMeditationByID(ctx context.Context, status string, meditationID string) (*string, error) {
	var medID string
	sqlQuery := `UPDATE meditation SET status=$2 WHERE id=$1 RETURNING id`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("UpdateMeditationByID", zap.String("error preparing statement", err.Error()), zap.String("sqlQuery : ", sqlQuery))

		return nil, err
	}
	row := stmt.QueryRowContext(ctx, meditationID, status)
	if err := row.Scan(&medID); err != nil {
		r.log.Error("UpdateMeditationByID", zap.String("error scanning row", err.Error()))
		return nil, err
	}
	return &medID, nil
}

func (r audioMessageRepository) GetMeditation(ctx context.Context, meditationId string) (*Meditation, error) {
	meditationSQL := `SELECT * FROM meditation WHERE id=$1 AND status=$2 LIMIT 1;`

	getMeditationStmt, err := r.db.PrepareContext(ctx, meditationSQL)
	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", meditationSQL),
		)
		return nil, err
	}

	row := getMeditationStmt.QueryRowContext(ctx, meditationId, "active")
	var meditation Meditation

	if err := row.Scan(
		&meditation.ID,
		&meditation.Name,
		&meditation.Image,
		&meditation.Status,
		&meditation.DateAdded,
		&meditation.DeletedAt,
	); err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", meditationSQL),
		)
		return nil, err
	}

	return &meditation, nil
}

func (r audioMessageRepository) GetMeditations(ctx context.Context) ([]Meditation, error) {
	meditationSQL := `SELECT * FROM meditation WHERE deleted_at IS NULL`

	var meditation []Meditation
	getMeditationStmt, err := r.db.PrepareContext(ctx, meditationSQL)
	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", meditationSQL),
		)
		return nil, err
	}

	rows, err := getMeditationStmt.QueryContext(ctx)

	defer rows.Close()

	if err == sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var as Meditation

		if err := rows.Scan(
			&as.ID,
			&as.Name,
			&as.Image,
			&as.Status,
			&as.DateAdded,
			&as.DeletedAt,
		); err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", meditationSQL),
			)
			return nil, err
		}

		meditation = append(meditation, as)
	}

	return meditation, nil
}

func (r audioMessageRepository) DeleteMeditationByID(ctx context.Context, meditationId string) (*DefaultResponse, error) {
	sqlQuery := `DELETE FROM meditation WHERE id=$1`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("DeleteMeditationByID", zap.String("error preparing statement", err.Error()), zap.String("sqlQuery : ", sqlQuery))

		return nil, err
	}
	row := stmt.QueryRowContext(ctx, meditationId)
	if err := row.Scan(&meditationId); err != nil {
		if err == sql.ErrNoRows {
			r.log.Error("DeleteMeditationByID", zap.String("error scanning row", err.Error()))
		} else {
			r.log.Error("DeleteMeditationByID", zap.String("error scanning row", err.Error()))
			return nil, err
		}
	}
	return &DefaultResponse{
		Success: true,
		Message: "meditation deleted successfully",
	}, nil
}
