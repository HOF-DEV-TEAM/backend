package favourite

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

var (
	ErrFieldRequired = errors.New("field is required")
)

type Service interface {
	CreateFavourite(ctx context.Context, favourite *Favourite) (*Favourite, error)
	GetFavourites(ctx context.Context) ([]*Favourite, int, error)
	DeleteFavourite(ctx context.Context, favId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error)
}
type favouritesService struct {
	repo   Repository
	log    *zap.Logger
	config *security.SecurityConfig
}

func NewService(repo Repository, log *zap.Logger, config *security.SecurityConfig) Service {
	return &favouritesService{log: log, repo: repo, config: config}
}

func (s *favouritesService) validateStruct(audioMessage *Favourite) error {
	validate := validator.New()

	return validate.Struct(audioMessage)
}

func (s *favouritesService) validateAudioSeriesStruct(audioSeries *Favourite) error {
	validate := validator.New()

	return validate.Struct(audioSeries)
}

func (s *favouritesService) CreateFavourite(ctx context.Context, favourite *Favourite) (*Favourite, error) {
	err := s.validateStruct(favourite)
	if err != nil {
		tErr, ok := err.(validator.ValidationErrors)

		if !ok {
			return nil, fmt.Errorf("unknown validation error")
		}

		for _, e := range tErr {
			switch e.StructField() {
			case "UserID":
				return nil, ErrFieldRequired
			case "MessageID":
				return nil, ErrFieldRequired
			default:
				s.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}

	result, err := s.repo.CreateFavourite(ctx, favourite)
	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		s.log.Error("msg",
			zap.String("method", "CreateAudioMessage"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}
	return result, nil
}

func (s *favouritesService) GetFavourites(ctx context.Context) ([]*Favourite, int, error) {
	return nil, 0, nil
}

func (s *favouritesService) DeleteFavourite(ctx context.Context, favId uuid.UUID, deletedAt sql.NullString) (uuid.UUID, error) {
	return uuid.Nil, nil
}
