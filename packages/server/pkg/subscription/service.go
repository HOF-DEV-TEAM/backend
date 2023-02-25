package subscription

type SubscriptionService interface {	
	CreateSubscription()
	CancelSubscription()
	ChangeSubscription()
	MakePayment()
} 


type subscriptionSvc struct {
	subProvider SubscriptionService //implements subsription service
}

func (ss *subscriptionSvc) CreateSubscription() {
	ss.subProvider.CreateSubscription()
}

func (ss *subscriptionSvc) CancelSubscription() {
	ss.subProvider.ChangeSubscription()
}

func (ss *subscriptionSvc) ChangeSubscription() {
	ss.subProvider.CancelSubscription()
}

func (ss *subscriptionSvc) MakePayment() {
	ss.subProvider.MakePayment()
}