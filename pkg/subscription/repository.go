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
	GetSubscriptionPlans(ctx context.Context) ([]*SubscriptionPlan, int, error)
	DeleteSubscriptionPlanById(ctx context.Context, id string) (string, error)
	GetSubscriptions(ctx context.Context) ([]*Subscription, int, error)
	GetSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetSubscriptionPlanById(ctx context.Context, subPlanId string) (*SubscriptionPlan, error)
	GetSubscriptionByUserAndPlanId(ctx context.Context, userId, planId string) (*Subscription, error)
	GetSubscriptionByCode(ctx context.Context, subCode string) (*Subscription, error)
	CreateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error)
	UpdateSubscription(ctx context.Context, userId string, sub *Subscription) (string, error)
	GetSubscriptionPlanOfferings(ctx context.Context) ([]*SubscriptionPlanOffering, int, error)
	CreateSubscriptionPlanOffering(ctx context.Context, sub *SubscriptionPlanOffering) (string, error)
	Close() error
}

type subscriptionRepo struct {
	db           *sql.DB
	getPlanStmt  *sql.Stmt
	log          *zap.Logger
	queryHandler urlqueryhelper.QueryHelper
	queryBuilder *urlqueryhelper.QueryBuilder
}

func NewRepository(db *sql.DB, logger *zap.Logger) Repository {
	return &subscriptionRepo{db: db, log: logger, queryHandler: urlqueryhelper.NewQueryHelper(), queryBuilder: urlqueryhelper.NewQueryBuilder()}
}

func (r subscriptionRepo) Close() error {
	if r.getPlanStmt != nil {
		if err := r.getPlanStmt.Close(); err != nil {
			return err
		}
	}
	return nil
}

var getSubscriptionQuery = "SELECT " +
	"s.id, " +
	"s.status, " +
	"s.user_id, " +
	"s.subscription_plan_id, " +
	"s.next_payment_date, " +
	"s.sub_code, " +
	"sp.type, " +
	"sp.freq, " +
	"sp.fee, " +
	"sp.currency, " +
	"sp.code " +
	"FROM subscriptions s " +
	"LEFT JOIN subscription_plans sp " +
	"ON sp.id = s.subscription_plan_id " +
	"WHERE s.status != 0" + "%s" +
	" ORDER BY s.date_added" +
	" LIMIT 1;"

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
		"next_payment_date," +
		"sub_code," +
		"date_added," +
		"last_updated" +
		") VALUES ($1, $2, $3, $4, $5, $6, $7) " +
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
		sub.NextPaymentDate,
		sub.SubCode,
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

func (r subscriptionRepo) UpdateSubscription(ctx context.Context, userId string, sub *Subscription) (string, error) {
	subWhere := Subscription{
		UserID: userId,
	}

	var subId string
	whereQuery := r.queryBuilder.Where(subWhere)
	setQuery := r.queryBuilder.Set(*sub)

	sqlQuery := `UPDATE subscriptions SET ` + setQuery + " WHERE status != 0 AND " + whereQuery + " RETURNING id"

	err := r.db.QueryRowContext(ctx, sqlQuery).Scan(&subId)
	if err != nil {
		r.log.Error("UpdateSubscription", zap.String("error scanning row", err.Error()))
		return "", err
	}
	return subId, nil
}

func (r subscriptionRepo) getPlan(ctx context.Context, subPlan *SubscriptionPlan) (*SubscriptionPlan, error) {
	whereQuery := r.queryBuilder.Where(*subPlan)
	query := `SELECT 
    	id, 
    	name,
    	fee,
    	type,
    	freq,
    	currency,
    	status, 
    	code,
    	date_added,
    	last_updated,
    	plan_id,
    	subscription_provider_id
		FROM subscription_plans
		WHERE deleted_at IS NULL AND %s;`

	var err error
	// first call, prepare statement for reuse
	if r.getPlanStmt == nil {
		r.getPlanStmt, err = r.db.PrepareContext(ctx, fmt.Sprintf(query, whereQuery))

		if err != nil {
			r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", query))
			return nil, err
		}
	}

	row := r.getPlanStmt.QueryRowContext(ctx)

	var plan SubscriptionPlan

	err = row.Scan(
		&plan.ID,
		&plan.Name,
		&plan.Fee,
		&plan.Type,
		&plan.Freq,
		&plan.Currency,
		&plan.Status,
		&plan.Code,
		&plan.DateAdded,
		&plan.LastUpdated,
		&plan.PlanId,
		&plan.SubscritpionProviderID,
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

func (r subscriptionRepo) GetPlan(ctx context.Context, planCode string) (*SubscriptionPlan, error) {
	return r.getPlan(ctx, &SubscriptionPlan{Code: planCode})
}

func (r subscriptionRepo) GetSubscriptionPlanById(ctx context.Context, planId string) (*SubscriptionPlan, error) {
	return r.getPlan(ctx, &SubscriptionPlan{ID: planId})
}

func (r subscriptionRepo) GetSubscriptionPlans(ctx context.Context) ([]*SubscriptionPlan, int, error) {
	query := `
	SELECT 
    	id, 
    	name,
    	fee,
    	type,
    	freq,
    	currency,
    	status, 
    	code FROM subscription_plans
		WHERE deleted_at IS NULL;
`

	getPlansSmt, err := r.db.PrepareContext(ctx, query)
	rows, err := getPlansSmt.QueryContext(ctx)
	defer rows.Close()

	plans := []*SubscriptionPlan{}

	if err == sql.ErrNoRows || err != nil {
		return plans, 0, err
	}

	for rows.Next() {
		var plan SubscriptionPlan

		err = rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.Fee,
			&plan.Type,
			&plan.Freq,
			&plan.Currency,
			&plan.Status,
			&plan.Code,
		)

		if err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", query),
			)
			return plans, 0, err
		}

		plans = append(plans, &plan)
	}
	return plans, 0, nil
}

func (r subscriptionRepo) DeleteSubscriptionPlanById(ctx context.Context, subPlanId string) (string, error) {
	sqlQuery := `UPDATE subscription_plans SET deleted_at=$1 WHERE id=$2 RETURNING id`
	stmt, err := r.db.PrepareContext(ctx, sqlQuery)
	if err != nil {
		r.log.Error("DeleteAudioMessagesByID", zap.String("error preparing statement", err.Error()), zap.String("sqlQuery : ", sqlQuery))

		return "", err
	}

	deletedAt := sql.NullString{
		String: time.Now().Format(time.RFC3339),
		Valid:  true,
	}
	row := stmt.QueryRowContext(ctx, deletedAt, subPlanId)
	if err := row.Scan(&subPlanId); err != nil {
		r.log.Error("DeleteSubscriptionPlanById", zap.String("error scanning row", err.Error()))
		return "", err
	}
	return subPlanId, nil
}

func (r subscriptionRepo) GetSubscriptionByUserAndPlanId(ctx context.Context, userId, planId string) (*Subscription, error) {
	sub := &Subscription{SubscriptionPlanID: planId, UserID: userId}
	return r.GetSubscription(ctx, sub)
}

func (r subscriptionRepo) GetSubscriptionByCode(ctx context.Context, subCode string) (*Subscription, error) {
	sub := &Subscription{SubCode: subCode}
	return r.GetSubscription(ctx, sub)
}

func (r subscriptionRepo) GetSubscription(ctx context.Context, sub *Subscription) (*Subscription, error) {
	whereQuery := r.queryHandler.WhereQueryHelper(*sub)
	query := fmt.Sprintf(getSubscriptionQuery, whereQuery)

	getSubStmt, err := r.db.PrepareContext(ctx, query)

	r.log.Info("msg", zap.String("query", query))

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
		&sub.NextPaymentDate,
		&sub.SubCode,
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

func (r subscriptionRepo) GetSubscriptions(ctx context.Context) ([]*Subscription, int, error) {
	query := fmt.Sprintf(getSubscriptionQuery, "")
	getSubStmt, err := r.db.PrepareContext(ctx, query)

	r.log.Info("msg", zap.String("query", query))

	if err != nil {
		r.log.Info("msg", zap.String("error preparing statement", ""), zap.String("error", err.Error()), zap.String("query", getSubscriptionQuery))
		return nil, 0, err
	}

	rows, err := getSubStmt.QueryContext(ctx)

	defer rows.Close()

	subs := []*Subscription{}

	if err == sql.ErrNoRows {
		return subs, 0, err
	}

	for rows.Next() {
		var sub Subscription
		err = rows.Scan(
			&sub.ID,
			&sub.Status,
			&sub.UserID,
			&sub.SubscriptionPlanID,
			&sub.NextPaymentDate,
			&sub.SubCode,
			&sub.Type,
			&sub.Freq,
			&sub.Fee,
			&sub.Currency,
			&sub.PlanCode,
		)

		if err != nil {
			r.log.Info("msg",
				zap.String("error querying", ""),
				zap.String("error", err.Error()),
				zap.String("query", getSubscriptionQuery),
			)
			return subs, 0, err
		}

		subs = append(subs, &sub)
	}
	return subs, 0, nil
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
