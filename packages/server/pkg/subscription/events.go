package subscription

import (
	"context"	

	"bitbucket.org/hofng/hofApp/pkg/user"
	"go.uber.org/zap"
)

type Event interface {
	HandleEvent(ctx context.Context, event *SubscriptionEvent) 
}

type EventType string

const (
	ChargeSuccess      = EventType("charge.success")
	SubscriptionCreate = EventType("subscription.create")
)

type SubscriptionEvent  struct {
	Event EventType `json:"event"`
	Data SubscriptionResponseData `json:"data"`
}

type subEvent struct {
	userRepo user.Repository
	subRepo Repository	
	logger *zap.Logger
}

func NewSubEvent(userRepo user.Repository, subRepo Repository, logger *zap.Logger) Event {
	return &subEvent{userRepo: userRepo, subRepo: subRepo, logger: logger}
}

func(se *subEvent) HandleEvent(ctx context.Context, event *SubscriptionEvent) {
	switch event.Event {
	case SubscriptionCreate:
		sub := event.Data.ToSubscription()
		subPlanFromDB, err := se.subRepo.GetPlan(ctx, event.Data.Plan.PlanCode)

		if err != nil {
			se.logger.Info("msg", zap.String(string(event.Event), err.Error()))
			return 
		}

		userFromDB, err := se.userRepo.GetByEmail(ctx, event.Data.Customer.Email)

		if err != nil {
			se.logger.Info("msg", zap.String(string(event.Event), err.Error()))
		}

		sub.Status = 1
		sub.UserID = userFromDB.ID
		sub.SubscriptionPlanID = subPlanFromDB.ID
		se.subRepo.CreateSubscription(ctx, &sub)
	default:
	}
}