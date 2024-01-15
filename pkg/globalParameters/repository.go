package globalParameters

import (
	"bitbucket.org/hofng/hofApp/infrastructure/library/http_helper"
	"bitbucket.org/hofng/hofApp/infrastructure/library/urlqueryhelper"
	"context"
	"database/sql"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type Repository interface {
	UpdateGlobalVariables(ctx context.Context, globalParameters *UpdateGlobalParameters) (*GlobalParameters, error)
	GetGlobalVariables(ctx context.Context) (*GlobalParameters, error)
	Close() error
}

type globalRepository struct {
	db           *sql.DB
	log          *zap.Logger
	getEmailStmt *sql.Stmt
	getIdStmt    *sql.Stmt
	queryHandler urlqueryhelper.QueryHelper
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &globalRepository{db: db, log: logger, queryHandler: urlqueryhelper.NewQueryHelper()}
}

func (r globalRepository) Close() error {
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

func (r globalRepository) UpdateGlobalVariables(ctx context.Context, globalParameters *UpdateGlobalParameters) (*GlobalParameters, error) {

	id, err := uuid.FromString(globalParameters.ID)
	if err != nil {
		return nil, err
	}

	globalID := struct {
		Id uuid.UUID `sql:"id"`
	}{
		Id: id,
	}

	var globalVariables GlobalParameters
	whereQuery := r.queryHandler.WhereQueryHelper(globalID)
	setQuery := r.queryHandler.SetQueryHelper(*globalParameters)
	sqlQuery := `UPDATE global_variables SET ` + setQuery + " WHERE " + whereQuery + " RETURNING *"
	err = r.db.QueryRowContext(ctx, sqlQuery).Scan(&globalVariables.ID, &globalVariables.ActivateSubscription, &globalVariables.LastUpdated, &globalVariables.DateCreated)
	if err != nil {
		r.log.Error("UpdateGlobalVariables", zap.String("error scanning row", err.Error()))
		return nil, err
	}
	return &globalVariables, nil

}

func (r globalRepository) GetGlobalVariables(ctx context.Context) (*GlobalParameters, error) {
	sqlQuery := `SELECT * FROM global_variables`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, err
	}
	var globalParameter GlobalParameters
	err = stmt.QueryRowContext(ctx).Scan(
		&globalParameter.ID,
		&globalParameter.ActivateSubscription,
		&globalParameter.LastUpdated,
		&globalParameter.DateCreated,
	)
	if err != nil {
		r.log.Error("msg", zap.String("error retrieving data", ""), zap.String("error", err.Error()), zap.String("query", sqlQuery))
		return nil, http_helper.ErrNotFound

	}
	return &globalParameter, nil
}
