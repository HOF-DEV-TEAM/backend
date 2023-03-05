package subscription

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

type Repository interface {
	CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error)
	CreateSubscriptionPlan(ctx context.Context, plan *SubscriptionPlan) (*SubscriptionPlan, error)
}

type subscriptionRepo struct {
	db  *sql.DB
	log *zap.Logger
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &subscriptionRepo{db: db, log: logger}
}

func (r *subscriptionRepo) CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error) {
	const SQL = "INSERT INTO subscription_offerings (" +
		"name," +
		"date_added" +
		") VALUES ($1, $2) " +
		"RETURNING id"

	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", SQL))
		return "", err
	}

	defer tx.Rollback()

	tmpSmt, err := tx.PrepareContext(ctx, SQL)

	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", SQL))
		return "", err
	}

	var subsriptionOfferingId string

	err = tmpSmt.QueryRowContext(ctx,
		offering.Name,
		sql.NullString{
			String: time.Now().Format(time.RFC3339),
			Valid:  true,
		},
	).Scan(&subsriptionOfferingId)

	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return "", err
	}

	err = tx.Commit()

	if err != nil {
		return "", err
	}
	return subsriptionOfferingId, nil
}

func (r *subscriptionRepo) CreateSubscriptionPlan(ctx context.Context, plan *SubscriptionPlan) (*SubscriptionPlan, error) {
	const SQL = "INSERT INTO subscription_plans (" +
		"name," +
		"status," +
		"freq," +
		"currency," +
		"code," +
		"plan_id," +
		"date_added," +
		"last_updated" +
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

	var subsriptionPlanId string

	err = tmpSmt.QueryRowContext(ctx,
		plan.Name,
		plan.Status,
		plan.Freq,
		plan.Currency,
		plan.Code,
		plan.PlanId,
		plan.DateAdded,
		plan.LastUpdated,	
	).Scan(&subsriptionPlanId)

	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	plan.ID = subsriptionPlanId
	return plan, nil
}