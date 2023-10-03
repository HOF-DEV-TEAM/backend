package paystack

import (
	"bitbucket.org/hofng/hofApp/pkg/events"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/user"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type EventType string

type SubscriptionEventResponse struct {
	Subscription PaystackSubscription `json:"subscription"`
	PaystackCustomerSubscription
}

type EventResponse struct {
	Event EventType                 `json:"event"`
	Data  SubscriptionEventResponse `json:"data"`
}

const (
	ChargeSuccessEvent      = EventType("charge.success")
	InvoiceUpdateEvent      = EventType("invoice.update")
	NotRenewEvent           = EventType("subscription.not_renew")
	SubscriptionCreateEvent = EventType("subscription.create")
)

// concrete implementation of events interface
type PaystackEvents struct {
	svc                      *PaystackService
	userRepo                 user.Repository
	subRepo                  subscription.Repository
	logger                   *zap.Logger
	InvoiceUpdateEvent       *events.Observable[*EventResponse, EventType]
	ChargeSuccessEvent       *events.Observable[*EventResponse, EventType]
	NotRenewEvent            *events.Observable[*EventResponse, EventType]
	SubsscriptionCreateEvent *events.Observable[*EventResponse, EventType]
}

func New(svc *PaystackService, logger *zap.Logger) *PaystackEvents {
	return &PaystackEvents{svc: svc, logger: logger}
}

func (e *PaystackEvents) Listen() *PaystackEvents {
	//listen for events

	e.InvoiceUpdateEvent = events.NewObservable[*EventResponse, EventType](InvoiceUpdateEvent)
	e.ChargeSuccessEvent = events.NewObservable[*EventResponse, EventType](ChargeSuccessEvent)
	e.NotRenewEvent = events.NewObservable[*EventResponse, EventType](NotRenewEvent)
	e.SubsscriptionCreateEvent = events.NewObservable[*EventResponse, EventType](SubscriptionCreateEvent)

	e.InvoiceUpdateEvent.Watch(e.svc.HandleInvoiceUpdate)
	e.NotRenewEvent.Watch(e.svc.HandleCancelSubscription)
	//e.SubsscriptionCreateEvent.Watch(e.svc.HandleSubscriptionCreate)

	e.SubsscriptionCreateEvent.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("NewSubscriptionCreateEvent", zap.Any("all response", a.Data.PaystackCustomerSubscription))

		//paystackUser, err := e.userRepo.GetByCustomerCode(ctx, a.Data.PaystackCustomerSubscription.Customer.CustomerCode)
		//if err != nil || paystackUser == nil {
		//	return err
		//}
		//
		//subPlan, err := e.subRepo.GetPlan(ctx, a.Data.PaystackCustomerSubscription.Plan.PlanCode)
		////subplan exists at this point
		//if err != nil {
		//	return err
		//}
		////check if subscription exists locally
		//sub := &subscription.Subscription{UserID: paystackUser.ID, SubCode: a.Data.PaystackCustomerSubscription.SubscriptionCode}
		//
		//subResult, err := e.subRepo.GetSubscription(ctx, sub)
		//if err != nil && err != sql.ErrNoRows {
		//	return err
		//}
		//now := sql.NullString{
		//	String: time.Now().Format(time.RFC3339),
		//	Valid:  true,
		//}
		//
		//newSub := &subscription.Subscription{
		//	SubCode:         a.Data.PaystackCustomerSubscription.SubscriptionCode,
		//	NextPaymentDate: parseDateTime(a.Data.PaystackCustomerSubscription.NextPaymentDate),
		//	LastUpdated:     now,
		//	Status:          1,
		//}
		//
		//if subResult != nil {
		//	//subscription already exists; update next payment date and subscription
		//	_, err := e.subRepo.UpdateSubscription(ctx, paystackUser.ID, newSub)
		//	if err != nil {
		//		return err
		//	}
		//	return nil
		//}
		////create new sub
		//newSub.DateAdded = now
		//newSub.UserID = paystackUser.ID
		//newSub.SubscriptionPlanID = subPlan.ID
		//newSub.NextPaymentDate = sql.NullString{
		//	String: a.Data.PaystackCustomerSubscription.NextPaymentDate,
		//	Valid:  true,
		//}
		//_, err = e.subRepo.CreateSubscription(ctx, newSub)
		//if err != nil {
		//	return err
		//}
		return nil
	})

	e.ChargeSuccessEvent.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("ChargeSuccess", zap.Any("sub response", a.Data.Subscription))
		e.logger.Info("ChargeSuccess", zap.Any("all response", *a))
		return nil
	})

	return e
}

func parseEvent(r *http.Request) (*EventResponse, error) {
	bytes, errRead := io.ReadAll(r.Body)

	if errRead != nil {
		return nil, errRead
	}

	var event EventResponse

	json.Unmarshal(bytes, &event)

	return &event, nil
}

func (g *PaystackEvents) HandleEventRequest(req *http.Request) error {
	event, err := parseEvent(req)
	if err != nil {
		return err
	}

	g.logger.Info("eventPayload", zap.Any("Paystack Payload", event))
	ctx := req.Context()

	switch event.Event {
	case InvoiceUpdateEvent:
		return g.InvoiceUpdateEvent.Set(ctx, event)
	case ChargeSuccessEvent:
		return g.ChargeSuccessEvent.Set(ctx, event)
	case NotRenewEvent:
		return g.NotRenewEvent.Set(ctx, event)
	case SubscriptionCreateEvent:
		return g.SubsscriptionCreateEvent.Set(ctx, event)
	}

	return nil
}

func HandleSubscriptionEvents(eventHandler *PaystackEvents) http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) {
		err := eventHandler.HandleEventRequest(r)

		if err != nil {
			eventHandler.logger.Error("msg", zap.String("error", err.Error()))
			return
		}
	}
}
