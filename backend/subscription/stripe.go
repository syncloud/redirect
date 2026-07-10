package subscription

import (
	"fmt"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	stripesub "github.com/stripe/stripe-go/v81/subscription"
	"go.uber.org/zap"
)

const (
	StripePlanMonthly = "monthly"
	StripePlanAnnual  = "annual"
)

type Stripe struct {
	secretKey      string
	priceMonthlyId string
	priceAnnualId  string
	successUrl     string
	cancelUrl      string
	logger         *zap.Logger
}

func NewStripe(secretKey, priceMonthlyId, priceAnnualId, successUrl, cancelUrl string, logger *zap.Logger) *Stripe {
	return &Stripe{
		secretKey:      secretKey,
		priceMonthlyId: priceMonthlyId,
		priceAnnualId:  priceAnnualId,
		successUrl:     successUrl,
		cancelUrl:      cancelUrl,
		logger:         logger,
	}
}

func (s *Stripe) Enabled() bool {
	return s.secretKey != "" && s.priceMonthlyId != "" && s.priceAnnualId != ""
}

func (s *Stripe) priceId(plan string) (string, error) {
	switch plan {
	case StripePlanAnnual:
		return s.priceAnnualId, nil
	case StripePlanMonthly:
		return s.priceMonthlyId, nil
	default:
		return "", fmt.Errorf("unknown stripe plan: %s", plan)
	}
}

func (s *Stripe) CreateCheckout(plan string) (string, error) {
	if !s.Enabled() {
		return "", fmt.Errorf("stripe is not configured")
	}
	priceId, err := s.priceId(plan)
	if err != nil {
		return "", err
	}
	stripe.Key = s.secretKey
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceId),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.successUrl),
		CancelURL:  stripe.String(s.cancelUrl),
	}
	checkoutSession, err := session.New(params)
	if err != nil {
		s.logger.Error("unable to create stripe checkout session", zap.Error(err))
		return "", err
	}
	return checkoutSession.URL, nil
}

func (s *Stripe) GetCheckoutSubscription(sessionId string) (string, error) {
	if !s.Enabled() {
		return "", fmt.Errorf("stripe is not configured")
	}
	stripe.Key = s.secretKey
	checkoutSession, err := session.Get(sessionId, nil)
	if err != nil {
		s.logger.Error("unable to get stripe checkout session", zap.Error(err))
		return "", err
	}
	if checkoutSession.PaymentStatus == stripe.CheckoutSessionPaymentStatusUnpaid {
		return "", fmt.Errorf("stripe checkout is not paid")
	}
	if checkoutSession.Subscription == nil {
		return "", fmt.Errorf("stripe checkout has no subscription")
	}
	return checkoutSession.Subscription.ID, nil
}

func (s *Stripe) Unsubscribe(id string) error {
	stripe.Key = s.secretKey
	_, err := stripesub.Cancel(id, nil)
	return err
}

func (s *Stripe) IsActive(id string) (bool, error) {
	stripe.Key = s.secretKey
	sub, err := stripesub.Get(id, nil)
	if err != nil {
		return false, err
	}
	return sub.Status == stripe.SubscriptionStatusActive || sub.Status == stripe.SubscriptionStatusTrialing, nil
}
