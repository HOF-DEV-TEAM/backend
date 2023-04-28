package subscription

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bitbucket.org/hofng/hofApp/infrastructure/library/urlqueryhelper"
	"go.uber.org/zap"
)

type Repository interface {
	CreateSubscriptionOffering(ctx context.Context, offering *SubscriptionOfferingRequest) (string, error)
	CreateSubscriptionPlan(ctx context.Context, plan *SubscriptionPlan) (*SubscriptionPlan, error)
	GetPlan(ctx context.Context, planCode string) (*SubscriptionPlan, error)
	GetSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetSubscriptionByUserAndPlanId(ctx context.Context, userId, planId string) (*Subscription, error)
	CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetSubscriptionPlanOfferings(ctx context.Context) ([]*SubscriptionPlanOffering, int, error)
	CreateSubscriptionPlanOffering(ctx context.Context, sub *SubscriptionPlanOffering) (string, error)
	Close() error
}

type subscriptionRepo struct {
	db           *sql.DB
	getPlanStmt  *sql.Stmt
	log          *zap.Logger
	queryHandler urlqueryhelper.QueryHelper
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &subscriptionRepo{db: db, log: logger, queryHandler: urlqueryhelper.NewQueryHelper()}
}

func (r subscriptionRepo) Close() error {
	if r.getPlanStmt != nil {
		if err := r.getPlanStmt.Close(); err != nil {
			return err
		}
	}
	return nil
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
		"fee," +
		"type," +
		"currency," +
		"code," +
		"plan_id," +
		"date_added," +
		"last_updated" +
		") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) " +
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
		plan.Fee,
		plan.Type,
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

func (r *subscriptionRepo) CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error) {
	const SQL = "INSERT INTO subscriptions (" +
		"status," +
		"user_id," +
		"subscription_plan_id," +
		"date_added," +
		"last_updated" +
		") VALUES ($1, $2, $3, $4, $5) " +
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

	var subsriptionId string

	err = tmpSmt.QueryRowContext(ctx,
		sub.Status,
		sub.UserID,
		sub.SubscriptionPlanID,
		sub.DateAdded,
		sub.LastUpdated,
	).Scan(&subsriptionId)

	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return nil, err
	}

	err = tx.Commit()

	if err != nil {
		return nil, err
	}

	sub.ID = subsriptionId
	r.log.Info("msg", zap.String("subscription created successfully", ""), zap.String("sub", fmt.Sprintf("%+v", sub)))
	return sub, nil
}

func (r subscriptionRepo) GetPlan(ctx context.Context, planCode string) (*SubscriptionPlan, error) {
	query := "SELECT " +
		"id," +
		"status," +
		"code " +
		"FROM subscription_plans WHERE code = $1"

	var err error
	// first call, prepare statement for reuse
	if r.getPlanStmt == nil {
		r.getPlanStmt, err = r.db.PrepareContext(ctx, query)

		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", query))
			return nil, err
		}
	}

	row := r.getPlanStmt.QueryRowContext(ctx, planCode)

	var plan SubscriptionPlan

	err = row.Scan(
		&plan.ID,
		&plan.Status,
		&plan.Code,
	)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", query),
		)
		return nil, err
	}

	return &plan, nil
}

func (r subscriptionRepo) GetSubscriptionByUserAndPlanId(ctx context.Context, userId, planId string) (*Subscription, error) {
	sub := &Subscription{SubscriptionPlanID: planId, UserID: userId}
	return r.GetSubscription(ctx, sub)
}

func (r subscriptionRepo) GetSubscription(ctx context.Context, sub *Subscription) (*Subscription, error) {
	whereQuery := r.queryHandler.WhereQueryHelper(*sub)
	query := "SELECT " +
		"s.id, " +
		"s.status, " +
		"s.user_id, " +
		"s.subscription_plan_id, " +
		// "s.next_payment_date, " +
		"sp.type, " +
		"sp.freq, " +
		"sp.fee, " +
		"sp.currency, " +
		"sp.code " +
		"FROM subscriptions s " +
		"LEFT JOIN subscription_plans sp " +
		"ON sp.id = s.subscription_plan_id " +
		"WHERE s.status = 1" + whereQuery +
		" LIMIT 1;"

	getSubStmt, err := r.db.PrepareContext(ctx, query)

	r.log.Info("msg",  zap.String("query", query))
	
	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", query))
		return nil, err
	}

	row := getSubStmt.QueryRowContext(ctx)

	err = row.Scan(
		&sub.ID,
		&sub.Status,
		&sub.UserID,
		&sub.SubscriptionPlanID,
		// &sub.NextPaymentDate,
		&sub.Type,
		&sub.Freq,
		&sub.Fee,
		&sub.Currency,
		&sub.PlanCode,
	)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		r.log.Info("msg",
			zap.String("error querying", ""),
			zap.String("error", err.Error()),
			zap.String("query", query),
		)
		return nil, err
	}

	return sub, nil
}

func (r subscriptionRepo) GetSubscriptionPlanOfferings(ctx context.Context) ([]*SubscriptionPlanOffering, int, error) {
	query := "SELECT sp.id, sp.code, sp.fee, sp.currency, sp.freq, COALESCE(sp.type, 0), so.name FROM subscription_plan_offerings s " +
		"LEFT JOIN subscription_plans sp " +
		"ON sp.id = s.subscription_plan_id " +
		"LEFT JOIN subscription_offerings so " +
		"ON so.id = s.subscription_offering_id " +
		"GROUP BY sp.id, sp.fee, sp.freq, sp.type, so.name;"

	getSubOfferingStmt, err := r.db.PrepareContext(ctx, query)

	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", query))
		return nil, 0, err
	}

	subs := []*SubscriptionPlanOffering{}

	rows, err := getSubOfferingStmt.QueryContext(ctx)

	defer rows.Close()

	if err == sql.ErrNoRows {
		return subs, 0, err
	}

	for rows.Next() {
		var sub SubscriptionPlanOffering

		if err := rows.Scan(
			&sub.SubscriptionPlanID,
			&sub.PlanCode,
			&sub.Fee,
			&sub.Currency,
			&sub.Freq,
			&sub.Type,
			&sub.Name,
		); err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", query),
			)
			return subs, 0, err
		}

		subs = append(subs, &sub)
	}

	return subs, 0, nil

}

func (r *subscriptionRepo) CreateSubscriptionPlanOffering(ctx context.Context, offering *SubscriptionPlanOffering) (string, error) {
	const SQL = "INSERT INTO subscription_plan_offerings (" +
		"subscription_plan_id," +
		"subscription_offering_id," +
		"date_added," +
		"last_updated" +
		") VALUES ($1, $2, $3, $4) " +
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

	var subsriptionPlanOfferingId string

	err = tmpSmt.QueryRowContext(ctx,
		offering.SubscriptionPlanID,
		offering.SubscriptionOfferingID,
		offering.DateAdded,
		offering.LastUpdated,
	).Scan(&subsriptionPlanOfferingId)

	if err != nil {
		r.log.Info("error", zap.String("error", err.Error()), zap.String("query", SQL))
		return "", err
	}

	err = tx.Commit()

	if err != nil {
		return "", err
	}

	return subsriptionPlanOfferingId, nil
}
