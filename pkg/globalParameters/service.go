package globalParameters

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/security"
	"context"
	"database/sql"
	"go.uber.org/zap"
	"time"
)

type Service interface {
	UpdateGlobalVariables(ctx context.Context, globalParameters *UpdateGlobalParameters) (*GlobalParameters, error)
	GetGlobalVariables(ctx context.Context) (*GlobalParameters, error)
}

type globalService struct {
	repo   Repository
	log    *zap.Logger
	config *security.SecurityConfig
}

func NewService(repo Repository, log *zap.Logger, config *security.SecurityConfig) Service {
	return &globalService{log: log, repo: repo, config: config}
}

func (g globalService) UpdateGlobalVariables(ctx context.Context, globalParameters *UpdateGlobalParameters) (*GlobalParameters, error) {
	_, ok := ctx.Value(g.config.JWTClaimsContextKey).(*security.JWTClaim[any])
	if !ok {
		return nil, http_helper.ErrUnauthorized
	}

	globalParameters.LastUpdated = sql.NullString{
		String: time.Now().Format(time.RFC3339),
		Valid:  true,
	}
	result, err := g.repo.UpdateGlobalVariables(ctx, globalParameters)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (g globalService) GetGlobalVariables(ctx context.Context) (*GlobalParameters, error) {
	globalVariables, err := g.repo.GetGlobalVariables(ctx)
	if err != nil {
		return nil, err
	}
	return globalVariables, nil

}
