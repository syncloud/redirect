package payment

import (
	"context"
	"strconv"
	"strings"

	"github.com/plutov/paypal/v4"
	"github.com/syncloud/redirect/product"
	"go.uber.org/zap"
)

type PayPal struct {
	client     *paypal.Client
	successUrl string
	cancelUrl  string
	logger     *zap.Logger
}

func NewPayPal(clientId, secretId, url, successUrl, cancelUrl string, logger *zap.Logger) (*PayPal, error) {
	client, err := paypal.NewClient(clientId, secretId, url)
	if err != nil {
		return nil, err
	}
	return &PayPal{client: client, successUrl: successUrl, cancelUrl: cancelUrl, logger: logger}, nil
}

func (p *PayPal) Start(order *product.Order, description string) (string, error) {
	ctx := context.Background()
	if _, err := p.client.GetAccessToken(ctx); err != nil {
		return "", err
	}
	created, err := p.client.CreateOrder(ctx, paypal.OrderIntentCapture,
		[]paypal.PurchaseUnitRequest{{
			Description: description,
			InvoiceID:   order.Reference,
			CustomID:    order.Reference,
			Amount: &paypal.PurchaseUnitAmount{
				Currency: product.Currency,
				Value:    product.Money(order.Total),
			},
		}}, nil,
		&paypal.ApplicationContext{ReturnURL: p.successUrl, CancelURL: p.cancelUrl})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

func (p *PayPal) Paid(providerReference string) (bool, int, string, error) {
	ctx := context.Background()
	if _, err := p.client.GetAccessToken(ctx); err != nil {
		return false, 0, "", err
	}
	captured, err := p.client.CaptureOrder(ctx, providerReference, paypal.CaptureOrderRequest{})
	if err != nil {
		return false, 0, "", err
	}
	if captured.Status != "COMPLETED" {
		return false, 0, "", nil
	}
	amount, currency := taken(captured)
	return true, amount, currency, nil
}

func taken(captured *paypal.CaptureOrderResponse) (int, string) {
	for _, unit := range captured.PurchaseUnits {
		for _, capture := range unit.Payments.Captures {
			return parse(capture.Amount.Value), capture.Amount.Currency
		}
	}
	return 0, ""
}

func parse(value string) int {
	parts := strings.SplitN(value, ".", 2)
	pounds, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	total := pounds * 100
	if len(parts) == 2 {
		padded := (parts[1] + "00")[:2]
		if fraction, err := strconv.Atoi(padded); err == nil {
			total += fraction
		}
	}
	return total
}
