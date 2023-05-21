package paystack

import (
	"bitbucket.org/hofng/hofApp/pkg/events"
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
	e.SubsscriptionCreateEvent.Watch(e.svc.HandleSubscriptionCreate)

	e.ChargeSuccessEvent.Watch(func(ctx context.Context, a *EventResponse) error {
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
