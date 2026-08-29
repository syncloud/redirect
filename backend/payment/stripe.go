package payment

import (
	"strings"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/syncloud/redirect/product"
	"go.uber.org/zap"
)

type Stripe struct {
	secretKey  string
	successUrl string
	cancelUrl  string
	logger     *zap.Logger
}

func NewStripe(secretKey, successUrl, cancelUrl string, logger *zap.Logger) *Stripe {
	return &Stripe{secretKey: secretKey, successUrl: successUrl, cancelUrl: cancelUrl, logger: logger}
}

func (s *Stripe) Start(order *product.Order, description string) (string, error) {
	stripe.Key = s.secretKey
	created, err := session.New(&stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(s.successUrl),
		CancelURL:         stripe.String(s.cancelUrl),
		CustomerEmail:     stripe.String(order.Email),
		ClientReferenceID: stripe.String(order.Reference),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			Quantity: stripe.Int64(1),
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:   stripe.String(strings.ToLower(product.Currency)),
				UnitAmount: stripe.Int64(int64(order.Total)),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(description),
				},
			},
		}},
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

func (s *Stripe) Paid(providerReference string) (bool, int, string, error) {
	stripe.Key = s.secretKey
	found, err := session.Get(providerReference, nil)
	if err != nil {
		return false, 0, "", err
	}
	if found.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return false, 0, "", nil
	}
	return true, int(found.AmountTotal), strings.ToUpper(string(found.Currency)), nil
}
