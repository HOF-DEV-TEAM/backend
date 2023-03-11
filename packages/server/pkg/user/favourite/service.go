package favourite

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
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
	CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error)
	GetFavourites(ctx context.Context) (GetFavouritesResponse, error)
	DeleteFavourite(ctx context.Context, favId string) (uuid.UUID, error)
}
type favouritesService struct {
	repo   Repository
	log    *zap.Logger
	config *security.SecurityConfig
}

func NewService(repo Repository, log *zap.Logger, config *security.SecurityConfig) Service {
	return &favouritesService{log: log, repo: repo, config: config}
}

func (s *favouritesService) validateStruct(fav *Favourites) error {
	validate := validator.New()

	return validate.Struct(fav)
}

func (s *favouritesService) validateFavouriteStruct(audioSeries *Favourites) error {
	validate := validator.New()

	return validate.Struct(audioSeries)
}

func (s *favouritesService) CreateFavourite(ctx context.Context, favourite *Favourites) (*Favourites, error) {
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
			default:
				s.log.Info("untyped validation error", zap.String("field", e.StructField()))
			}
		}
		return nil, err
	}
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return nil, http_helper.ErrInvalidAccount
	}
	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return nil, err
	}

	favourite.UserID = userId
	result, err := s.repo.CreateFavourite(ctx, favourite)
	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		s.log.Error("msg",
			zap.String("method", "CreateFavourite"),
			zap.String("error", err.Error()),
		)
		return nil, err
	}
	return result, nil
}

func (s *favouritesService) GetFavourites(ctx context.Context) (GetFavouritesResponse, error) {
	result := GetFavouritesResponse{}

	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return result, http_helper.ErrInvalidAccount
	}

	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return result, err
	}

	fav, count, err := s.repo.GetFavourites(ctx, userId)
	if err == sql.ErrNoRows {
		return result, err
	}

	result.Favourites = []*FavMessageJSON{}

	for _, as := range fav {
		result.Favourites = append(result.Favourites, NewJSONFavMessage(as))
	}

	result.Pagination = PageResponse{
		TotalResults: int32(count),
	}

	return result, nil
}

func (s *favouritesService) DeleteFavourite(ctx context.Context, messageId string) (uuid.UUID, error) {
	messageID, err := uuid.FromString(messageId)
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := ctx.Value(s.config.JWTClaimsContextKey).(*security.JWTClaim)
	if !ok {
		return uuid.Nil, http_helper.ErrInvalidAccount
	}

	userId, err := uuid.FromString(claims.JWTClaimsMain.LoggedInUserId)
	if err != nil {
		return uuid.Nil, err
	}
	result, err := s.repo.DeleteFavourite(ctx, messageID, userId)
	if err != nil {
		return uuid.Nil, err
	}

	return result, nil
}
