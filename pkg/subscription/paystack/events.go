package paystack

import (
	"bitbucket.org/hofng/hofApp/pkg/events"
	"bitbucket.org/hofng/hofApp/pkg/subscription"
	"bitbucket.org/hofng/hofApp/pkg/user"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	InvoiceUpdatedEvent
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
	InvoicePaymentFailed    = EventType("invoice.payment_failed")
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
	InvoicePaymentFailed     *events.Observable[*EventResponse, EventType]
}

func New(svc *PaystackService, userRepo user.Repository, subRepo subscription.Repository, logger *zap.Logger) *PaystackEvents {
	return &PaystackEvents{svc: svc, userRepo: userRepo, subRepo: subRepo, logger: logger}
}

func (e *PaystackEvents) Listen() *PaystackEvents {
	//listen for events

	e.InvoiceUpdateEvent = events.NewObservable[*EventResponse, EventType](InvoiceUpdateEvent)
	e.ChargeSuccessEvent = events.NewObservable[*EventResponse, EventType](ChargeSuccessEvent)
	e.NotRenewEvent = events.NewObservable[*EventResponse, EventType](NotRenewEvent)
	e.InvoicePaymentFailed = events.NewObservable[*EventResponse, EventType](InvoicePaymentFailed)
	e.SubsscriptionCreateEvent = events.NewObservable[*EventResponse, EventType](SubscriptionCreateEvent)

	e.InvoiceUpdateEvent.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("InvoiceUpdateEvent", zap.Any("all response", a.Data.PaystackCustomerSubscription))
		storeUser, err := e.userRepo.GetByCustomerCode(ctx, a.Data.PaystackCustomerSubscription.Customer.CustomerCode)

		if err != nil || storeUser == nil {
			return err
		}

		//get additional information from paystack
		existingSub, err := e.svc.payStackClient.GetSubscription(ctx, a.Data.PaystackCustomerSubscription.SubscriptionCode)

		if err != nil || existingSub == nil {
			return err
		}

		//check if subscription exists locally
		sub := &subscription.Subscription{UserID: storeUser.ID}

		subResult, err := e.subRepo.GetSubscription(ctx, sub)

		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if subResult != nil {
			now := sql.NullString{
				String: time.Now().Format(time.RFC3339),
				Valid:  true,
			}

			newSub := &subscription.Subscription{
				SubCode:         a.Data.PaystackCustomerSubscription.SubscriptionCode,
				NextPaymentDate: parseDateTime(a.Data.PaystackCustomerSubscription.NextPaymentDate),
				LastUpdated:     now,
				Status:          1,
			}

			//subscription already exists; update next payment date and subscription
			_, err := e.subRepo.UpdateSubscription(ctx, storeUser.ID, newSub)
			if err != nil {
				return err
			}
		}

		return nil
	})

	e.NotRenewEvent.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("NotRenewEvent", zap.Any("all response", a.Data.PaystackCustomerSubscription))

		user, err := e.userRepo.GetByCustomerCode(ctx, a.Data.Customer.CustomerCode)

		if err != nil || user == nil {
			return err
		}

		sub := &subscription.Subscription{UserID: user.ID, SubCode: a.Data.SubscriptionCode}

		subResult, err := e.subRepo.GetSubscription(ctx, sub)

		if err != nil && err != sql.ErrNoRows {
			return err
		}

		newSub := &subscription.Subscription{
			//NextPaymentDate: sql.NullString{Valid: true},
			LastUpdated: sql.NullString{
				String: time.Now().Format(time.RFC3339),
				Valid:  true,
			},
			Status: 3,
		}

		if subResult != nil {
			//subscription already exists; update next payment date and subscription
			_, err := e.subRepo.UpdateSubscription(ctx, user.ID, newSub)
			if err != nil {
				return err
			}
			return nil
		}

		return nil
	})

	e.SubsscriptionCreateEvent.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("NewSubscriptionCreateEvent", zap.Any("all response", a.Data.PaystackCustomerSubscription))

		storeUser, err := e.userRepo.GetByEmail(ctx, a.Data.PaystackCustomerSubscription.Customer.Email)
		if err != nil || storeUser == nil {
			e.logger.Error("GetByCustomerCode", zap.Any("all response", a.Data.PaystackCustomerSubscription), zap.Error(err))
			return err
		}

		subPlan, err := e.subRepo.GetPlan(ctx, a.Data.PaystackCustomerSubscription.Plan.PlanCode)
		//subplan exists at this point
		if err != nil {
			e.logger.Error("GetPlan", zap.Any("all response", a.Data.PaystackCustomerSubscription.Plan.PlanCode), zap.Error(err))
			return err
		}

		//check if subscription exists locally
		sub := &subscription.Subscription{UserID: storeUser.ID, SubCode: a.Data.PaystackCustomerSubscription.SubscriptionCode}

		subResult, err := e.subRepo.GetSubscription(ctx, sub)
		if err != nil && err != sql.ErrNoRows {
			e.logger.Error("GetSubscription", zap.Any("all response", sub), zap.Error(err))
			return err
		}
		now := sql.NullString{
			String: time.Now().Format(time.RFC3339),
			Valid:  true,
		}

		newSub := &subscription.Subscription{
			SubCode:         a.Data.PaystackCustomerSubscription.SubscriptionCode,
			NextPaymentDate: parseDateTime(a.Data.PaystackCustomerSubscription.NextPaymentDate),
			LastUpdated:     now,
			Status:          1,
		}

		if subResult != nil {
			e.logger.Error("UpdateSubscription1", zap.Any("subresult", subResult), zap.Error(err))

			//subscription already exists; update next payment date and subscription
			_, err := e.subRepo.UpdateSubscription(ctx, storeUser.ID, newSub)
			if err != nil {
				e.logger.Error("UpdateSubscription", zap.Any("all response", newSub), zap.Error(err))
				return err
			}
			return errors.New("subscription updated successfully")
		}

		//create new sub
		newSub.DateAdded = now
		newSub.UserID = storeUser.ID
		newSub.SubscriptionPlanID = subPlan.ID
		newSub.NextPaymentDate = sql.NullString{
			String: a.Data.PaystackCustomerSubscription.NextPaymentDate,
			Valid:  true,
		}

		e.logger.Error("UpdateSubscription2", zap.Any("subresult", newSub), zap.Error(err))
		_, err = e.subRepo.CreateSubscription(ctx, newSub)
		if err != nil {
			e.logger.Error("CreateSubscription", zap.Any("all response", newSub), zap.Error(err))
			return err
		}
		return nil
	})

	e.ChargeSuccessEvent.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("ChargeSuccessEvent", zap.Any("all response", a.Data.PaystackCustomerSubscription))
		storeUser, err := e.userRepo.GetByCustomerCode(ctx, a.Data.PaystackCustomerSubscription.Customer.CustomerCode)
		if err == sql.ErrNoRows {
			e.logger.Error("Charge.Success", zap.String("getCustomerByCode", "an error occurred"), zap.Error(err))
			return err
		}

		if storeUser.PaystackCustomerCode.String != "" {
			getSubscription, err := e.subRepo.GetSubscription(ctx, &subscription.Subscription{
				UserID: storeUser.ID,
			})
			if err == sql.ErrNoRows {
				e.logger.Error("Charge.Success", zap.String("GetSubscription", "an error occurred"), zap.Error(err))
				return err
			}

			if getSubscription != nil {

				//get additional information from paystack
				existingSub, err := e.svc.payStackClient.GetSubscription(ctx, getSubscription.SubCode)
				e.logger.Info("Charge.Success", zap.String("PaystackGetSubscription", "an error occurred"), zap.Error(err))
				e.logger.Info("Charge.Success", zap.String("PaystackGetSubscription", "subscription"), zap.Any("value", existingSub))

				switch {
				case err == nil && existingSub != nil:
					e.logger.Info("Charge.Success", zap.String("PaystackGetSubscription", "no error occurred"), zap.Error(err))

					now := sql.NullString{
						String: time.Now().Format(time.RFC3339),
						Valid:  true,
					}

					newSub := &subscription.Subscription{
						SubCode:         a.Data.PaystackCustomerSubscription.SubscriptionCode,
						NextPaymentDate: parseDateTime(existingSub.Data.NextPaymentDate),
						LastUpdated:     now,
						Status:          1,
					}
					e.logger.Error("ChargeSuccessEvent1", zap.Any("subresult", newSub), zap.Error(err))

					//subscription already exists; update next payment date and subscription
					_, err := e.subRepo.UpdateSubscription(ctx, storeUser.ID, newSub)
					if err != nil {
						return err
					}

				case err != nil:
					e.logger.Error("Charge.Success", zap.String("PaystackGetSubscription", "an error occurred"), zap.Error(err))
					return err

				}
			}

		}

		return nil
	})

	e.InvoicePaymentFailed.Watch(func(ctx context.Context, a *EventResponse) error {
		e.logger.Info("InvoicePaymentFailed", zap.Any("all response", a.Data.PaystackCustomerSubscription))

		user, err := e.userRepo.GetByCustomerCode(ctx, a.Data.Customer.CustomerCode)

		if err != nil || user == nil {
			return err
		}

		sub := &subscription.Subscription{UserID: user.ID, SubCode: a.Data.SubscriptionCode}

		subResult, err := e.subRepo.GetSubscription(ctx, sub)

		if err != nil && err != sql.ErrNoRows {
			return err
		}

		newSub := &subscription.Subscription{
			//NextPaymentDate: sql.NullString{Valid: true},
			LastUpdated: sql.NullString{
				String: time.Now().Format(time.RFC3339),
				Valid:  true,
			},
			Status: 2,
		}

		if subResult != nil {
			//subscription already exists; update next payment date and subscription
			_, err := e.subRepo.UpdateSubscription(ctx, user.ID, newSub)
			if err != nil {
				return err
			}
			return nil
		}

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
		NextPaymentDate  string      `json:"next_payment_date"`
		OpenInvoice      interface{} `json:"open_invoice"`
		CreatedAt        string      `json:"createdAt"`
		Plan             struct {
			Name         string      `json:"name"`
			PlanCode     string      `json:"plan_code"`
			Description  interface{} `json:"description"`
			Amount       int         `json:"amount"`
			Interval     string      `json:"interval"`
			SendInvoices int         `json:"send_invoices"`
			SendSms      int         `json:"send_sms"`
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
		CreatedAt1 string `json:"created_at"`
	} `json:"data"`
}

type InvoiceUpdatedEvent struct {
	Event string `json:"event"`
	Data  struct {
		Domain        string      `json:"domain"`
		InvoiceCode   string      `json:"invoice_code"`
		Amount        int         `json:"amount"`
		PeriodStart   time.Time   `json:"period_start"`
		PeriodEnd     time.Time   `json:"period_end"`
		Status        string      `json:"status"`
		Paid          bool        `json:"paid"`
		PaidAt        time.Time   `json:"paid_at"`
		Description   interface{} `json:"description"`
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
		Subscription struct {
			Status           string      `json:"status"`
			SubscriptionCode string      `json:"subscription_code"`
			Amount           int         `json:"amount"`
			CronExpression   string      `json:"cron_expression"`
			NextPaymentDate  time.Time   `json:"next_payment_date"`
			OpenInvoice      interface{} `json:"open_invoice"`
		} `json:"subscription"`
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
		Transaction struct {
			Reference string `json:"reference"`
			Status    string `json:"status"`
			Amount    int    `json:"amount"`
			Currency  string `json:"currency"`
		} `json:"transaction"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"data"`
}

func parseEvent(g *PaystackEvents, r *http.Request) (*EventResponse, error) {
	bytes, errRead := io.ReadAll(r.Body)

	if errRead != nil {
		return nil, errRead
	}

	var event EventResponse

	err := json.Unmarshal(bytes, &event)
	if err != nil {

		g.logger.Error("msg", zap.String("paystackSubEvent", err.Error()))
		return nil, err
	}

	return &event, nil
}

func (g *PaystackEvents) HandleEventRequest(req *http.Request) error {
	event, err := parseEvent(g, req)
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
	case InvoicePaymentFailed:
		return g.InvoicePaymentFailed.Set(ctx, event)
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
