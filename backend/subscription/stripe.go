package subscription

import (
	"fmt"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	stripesub "github.com/stripe/stripe-go/v81/subscription"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

const (
	StripePlanMonthly    = "monthly"
	StripePlanAnnual     = "annual"
	StripePlanMaxMonthly = "max_monthly"
	StripePlanMaxAnnual  = "max_annual"
)

type Stripe struct {
	secretKey         string
	priceMonthlyId    string
	priceAnnualId     string
	priceMaxMonthlyId string
	priceMaxAnnualId  string
	successUrl        string
	cancelUrl         string
	logger            *zap.Logger
}

func NewStripe(secretKey, priceMonthlyId, priceAnnualId, priceMaxMonthlyId, priceMaxAnnualId, successUrl, cancelUrl string, logger *zap.Logger) *Stripe {
	return &Stripe{
		secretKey:         secretKey,
		priceMonthlyId:    priceMonthlyId,
		priceAnnualId:     priceAnnualId,
		priceMaxMonthlyId: priceMaxMonthlyId,
		priceMaxAnnualId:  priceMaxAnnualId,
		successUrl:        successUrl,
		cancelUrl:         cancelUrl,
		logger:            logger,
	}
}

func (s *Stripe) Enabled() bool {
	return s.secretKey != "" && s.priceMonthlyId != "" && s.priceAnnualId != ""
}

func (s *Stripe) MaxEnabled() bool {
	return s.Enabled() && s.priceMaxMonthlyId != "" && s.priceMaxAnnualId != ""
}

func (s *Stripe) priceId(plan string) (string, error) {
	switch plan {
	case StripePlanAnnual:
		return s.priceAnnualId, nil
	case StripePlanMonthly:
		return s.priceMonthlyId, nil
	case StripePlanMaxAnnual:
		return s.priceMaxAnnualId, nil
	case StripePlanMaxMonthly:
		return s.priceMaxMonthlyId, nil
	default:
		return "", fmt.Errorf("unknown stripe plan: %s", plan)
	}
}

func (s *Stripe) tierForPrice(priceId string) string {
	if priceId != "" && (priceId == s.priceMaxMonthlyId || priceId == s.priceMaxAnnualId) {
		return model.PlanMax
	}
	return model.PlanPro
}

func (s *Stripe) CreateCheckout(plan string) (string, error) {
	if !s.Enabled() {
		return "", fmt.Errorf("stripe is not configured")
	}
	priceId, err := s.priceId(plan)
	if err != nil {
		return "", err
	}
	if priceId == "" {
		return "", fmt.Errorf("stripe plan not available: %s", plan)
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

func (s *Stripe) GetCheckoutSubscription(sessionId string) (string, string, error) {
	if !s.Enabled() {
		return "", "", fmt.Errorf("stripe is not configured")
	}
	stripe.Key = s.secretKey
	params := &stripe.CheckoutSessionParams{}
	params.AddExpand("line_items")
	checkoutSession, err := session.Get(sessionId, params)
	if err != nil {
		s.logger.Error("unable to get stripe checkout session", zap.Error(err))
		return "", "", err
	}
	if checkoutSession.PaymentStatus == stripe.CheckoutSessionPaymentStatusUnpaid {
		return "", "", fmt.Errorf("stripe checkout is not paid")
	}
	if checkoutSession.Subscription == nil {
		return "", "", fmt.Errorf("stripe checkout has no subscription")
	}
	plan := model.PlanPro
	if checkoutSession.LineItems != nil && len(checkoutSession.LineItems.Data) > 0 && checkoutSession.LineItems.Data[0].Price != nil {
		plan = s.tierForPrice(checkoutSession.LineItems.Data[0].Price.ID)
	}
	return checkoutSession.Subscription.ID, plan, nil
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
