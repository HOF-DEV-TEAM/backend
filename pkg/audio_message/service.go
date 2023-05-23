package audio_message

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"github.com/go-playground/validator"
	"go.uber.org/zap"
)

var (
	ErrFieldRequired = errors.New("field is required")
)

type Service interface {
	GetAudioMessages(ctx context.Context, search *Filter) (GetAudiosMessagesResponse, error)
	CreateAudioMessage(ctx context.Context, audioMessage *AudioMessage) (*AudioMessage, error)
	CreateAudioSeries(ctx context.Context, audioSeries *AudioSeries) (*AudioSeries, error)
	GetAudioSeries(ctx context.Context) (GetAudiosSeriesResponse, error)
	GetAudioMessageByID(ctx context.Context, messageId string) (*AudioMessageJSON, error)
	GetAudioSeriesByID(ctx context.Context, seriesId string) (*AudioSeriesJSON, error)
	UpdateAudioMessagesByID(ctx context.Context, message AudioMessage, messageId string) (uuid.UUID, error)
	UpdateAudioSeriesByID(ctx context.Context, series AudioSeries, seriesId string) (uuid.UUID, error)
	DeleteAudioMessagesByID(ctx context.Context, messageId string) (uuid.UUID, error)
	DeleteAudioSeriesByID(ctx context.Context, seriesId string) (uuid.UUID, error)
	HomePageDirectory(ctx context.Context) (*Homepage, error)
	CreateMeditation(ctx context.Context, meditation *Meditation) (string, error)
	CreateMeditations(ctx context.Context, meditation []*Meditation) (*MeditationResponse, error)
	UpdateMeditationByID(ctx context.Context, status string, meditationID string) (*string, error)
	GetMeditations(ctx context.Context) ([]Meditation, error)
}

type FilterType string

const (
	SeriesID       = FilterType("series_id")
	AudioMessageID = FilterType("id")
)

var FilterList = []FilterType{
	SeriesID,
	AudioMessageID,
}

type Filter struct {
	SeriesID string `filter:"series_id"`
}
type audioMessageService struct {
	repo   Repository
	log    *zap.Logger
	config *security.SecurityConfig
}

func NewService(repo Repository, log *zap.Logger, config *security.SecurityConfig) Service {
	return &audioMessageService{log: log, repo: repo, config: config}
}

func (svc *audioMessageService) validateStruct(audioMessage *AudioMessage) error {
	validate := validator.New()

	return validate.Struct(audioMessage)
}

func (svc *audioMessageService) validateAudioSeriesStruct(audioSeries *AudioSeries) error {
	validate := validator.New()

	return validate.Struct(audioSeries)
}

func (svc *audioMessageService) CreateAudioMessage(ctx context.Context, audioMessage *AudioMessage) (*AudioMessage, error) {

	err := svc.validateStruct(audioMessage)

	if err != nil {
		tErr, ok := err.(validator.ValidationErrors)

		if !ok {
			return nil, fmt.Errorf("unknown validation error")
		}

		for _, e := range tErr {
			switch e.StructField() {
			case "Title":
				return nil, ErrFieldRequired
			case "Author":
				return nil, ErrFieldRequired
			case "AudioUrl":
				return nil, ErrFieldRequired
			default:
				svc.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	result, err := svc.repo.CreateAudioMessage(ctx, audioMessage)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "CreateAudioMessage"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	return result, nil
}

func (svc *audioMessageService) CreateAudioSeries(ctx context.Context, audioSeries *AudioSeries) (*AudioSeries, error) {

	err := svc.validateAudioSeriesStruct(audioSeries)

	if err != nil {
		tErr, ok := err.(validator.ValidationErrors)

		if !ok {
			return nil, fmt.Errorf("unknown validation error")
		}

		for _, e := range tErr {
			switch e.StructField() {
			case "Title":
				return nil, ErrFieldRequired
			case "ImageUrl":
				return nil, ErrFieldRequired
			default:
				svc.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	result, err := svc.repo.CreateAudioSeries(ctx, audioSeries)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "CreateAudioSeries"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	return result, nil
}

func (svc *audioMessageService) GetAudioSeries(ctx context.Context) (GetAudiosSeriesResponse, error) {
	result := GetAudiosSeriesResponse{}
	audioSeries, count, err := svc.repo.GetAudioSeries(ctx)

	if err == sql.ErrNoRows {
		return result, err
	}

	result.AudioSeries = []*AudioSeriesJSON{}

	for _, as := range audioSeries {
		result.AudioSeries = append(result.AudioSeries, NewJSONAudioSeries(as))
	}

	result.Pagination = PageResponse{
		TotalResults: int32(count),
	}

	return result, nil
}

func (svc *audioMessageService) GetAudioMessages(ctx context.Context, search *Filter) (GetAudiosMessagesResponse, error) {
	result := GetAudiosMessagesResponse{}

	audioMessages, count, err := svc.repo.GetAudioMessages(ctx, search)

	if err == sql.ErrNoRows {
		return result, err
	}

	result.AudioMessages = []*AudioMessageJSON{}

	for _, as := range audioMessages {
		result.AudioMessages = append(result.AudioMessages, NewJSONAudioMessage(as))
	}

	result.Pagination = PageResponse{
		TotalResults: int32(count),
	}

	return result, nil
}

func (svc *audioMessageService) GetAudioMessageByID(ctx context.Context, messageId string) (*AudioMessageJSON, error) {
	id, err := uuid.FromString(messageId)
	if err != nil {
		return nil, err
	}
	audioMessage, err := svc.repo.GetAudioMessageByID(ctx, id)
	if err != nil {
		return nil, err
	}
	audioMessageJSON := NewJSONAudioMessage(audioMessage)
	return audioMessageJSON, nil
}

func (svc *audioMessageService) GetAudioSeriesByID(ctx context.Context, seriesId string) (*AudioSeriesJSON, error) {
	id, err := uuid.FromString(seriesId)
	if err != nil {
		return nil, err
	}

	audioSeries, err := svc.repo.GetAudioSeriesByID(ctx, id)
	if err != nil {
		return nil, err
	}

	audioSeriesJSON := NewJSONAudioSeries(audioSeries)
	return audioSeriesJSON, nil
}

func (svc *audioMessageService) UpdateAudioMessagesByID(ctx context.Context, message AudioMessage, messageId string) (uuid.UUID, error) {
	id, err := uuid.FromString(messageId)
	if err != nil {
		return uuid.Nil, err
	}

	message.LastUpdated = sql.NullString{
		String: time.Now().Format(time.RFC3339),
	}
	result, err := svc.repo.UpdateAudioMessagesByID(ctx, message, id)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}
func (svc *audioMessageService) UpdateAudioSeriesByID(ctx context.Context, series AudioSeries, seriesId string) (uuid.UUID, error) {
	id, err := uuid.FromString(seriesId)
	if err != nil {
		return uuid.Nil, err
	}
	series.LastUpdated = sql.NullString{
		String: time.Now().Format(time.RFC3339),
	}
	result, err := svc.repo.UpdateAudioSeriesByID(ctx, series, id)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}

func (svc *audioMessageService) DeleteAudioMessagesByID(ctx context.Context, messageId string) (uuid.UUID, error) {
	id, err := uuid.FromString(messageId)
	if err != nil {
		return uuid.Nil, err
	}

	deletedAt := sql.NullString{
		String: time.Now().Format(time.RFC3339),
		Valid:  true,
	}
	result, err := svc.repo.DeleteAudioMessagesByID(ctx, id, deletedAt)
	if err != nil {
		return uuid.Nil, err
	}

	return result, err
}

func (svc *audioMessageService) DeleteAudioSeriesByID(ctx context.Context, seriesId string) (uuid.UUID, error) {
	id, err := uuid.FromString(seriesId)
	if err != nil {
		return uuid.Nil, err
	}
	deletedAt := sql.NullString{
		String: time.Now().Format(time.RFC3339),
		Valid:  true,
	}

	result, err := svc.repo.DeleteAudioSeriesByID(ctx, id, deletedAt)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}

func (svc *audioMessageService) HomePageDirectory(ctx context.Context) (*Homepage, error) {
	homePage, err := svc.repo.HomePageDirectory(ctx)
	if err == sql.ErrNoRows {
		return nil, err
	}

	return homePage, nil

}

func (svc *audioMessageService) UpdateMeditationByID(ctx context.Context, status string, meditationID string) (*string, error) {
	deletedAt := sql.NullString{
		String: time.Now().Format(time.RFC3339),
	}
	result, err := svc.repo.UpdateMeditationByID(ctx, status, meditationID, deletedAt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (svc *audioMessageService) CreateMeditations(ctx context.Context, meditation []*Meditation) (*MeditationResponse, error) {
	var med []*Meditation
	for _, m := range meditation {
		m.DateAdded = sql.NullString{Valid: true, String: time.Now().Format(time.RFC3339)}
		med = append(med, m)
	}

	result, err := svc.repo.CreateMeditations(ctx, med)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "CreateMeditation"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}

	return result, nil
}

func (svc *audioMessageService) CreateMeditation(ctx context.Context, meditation *Meditation) (string, error) {
	meditation.DateAdded = sql.NullString{Valid: true, String: time.Now().Format(time.RFC3339)}

	result, err := svc.repo.CreateMeditation(ctx, meditation)

	if err == sql.ErrNoRows {
		return "", err
	}

	if err != nil {
		svc.log.Error("msg",
			zap.String("method", "CreateMeditation"),
			zap.String("error", err.Error()),
		)
		return "", err
	}

	return result, nil
}

func (svc *audioMessageService) GetMeditations(ctx context.Context) ([]Meditation, error) {
	meditation, err := svc.repo.GetMeditations(ctx)
	if err == sql.ErrNoRows {
		return nil, err
	}

	return meditation, nil

}
