package paystack

import (
	"bitbucket.org/hofng/hofApp/pkg/events"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type EventType string

type SubscriptionEventResponse struct {
	Subscription PaystackSubscription `json:"subscription"`
	PaystackCustomerSubscription
	SubscriptionCreatedEvent
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

		e.logger.Info("SubsscriptionCreateEvent", zap.Any("sub response", a.Data.Subscription))
		e.logger.Info("SubsscriptionCreateEvent", zap.Any("all response", *a))
		return nil
	})

	e.ChargeSuccessEvent.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("ChargeSuccess", zap.Any("sub response", a.Data.Subscription))
		e.logger.Info("ChargeSuccess", zap.Any("all response", *a))
		return nil
	})

	return e
}

type SubscriptionCreatedEvent struct {
	Event string `json:"event"`
	Data  struct {
		Domain           string      `json:"domain"`
		Status           string      `json:"status"`
		SubscriptionCode string      `json:"subscription_code"`
		Amount           int         `json:"amount"`
		CronExpression   string      `json:"cron_expression"`
		NextPaymentDate  time.Time   `json:"next_payment_date"`
		OpenInvoice      interface{} `json:"open_invoice"`
		CreatedAt        time.Time   `json:"createdAt"`
		Plan             struct {
			Name         string      `json:"name"`
			PlanCode     string      `json:"plan_code"`
			Description  interface{} `json:"description"`
			Amount       int         `json:"amount"`
			Interval     string      `json:"interval"`
			SendInvoices bool        `json:"send_invoices"`
			SendSms      bool        `json:"send_sms"`
			Currency     string      `json:"currency"`
		} `json:"plan"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
			CountryCode       string `json:"country_code"`
			Brand             string `json:"brand"`
			AccountName       string `json:"account_name"`
		} `json:"authorization"`
		Customer struct {
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
			Phone        string `json:"phone"`
			Metadata     struct {
			} `json:"metadata"`
			RiskAction string `json:"risk_action"`
		} `json:"customer"`
		CreatedAt1 time.Time `json:"created_at"`
	} `json:"data"`
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
