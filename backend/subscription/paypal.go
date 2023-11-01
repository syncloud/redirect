package subscription

import (
	"context"
	"github.com/plutov/paypal/v4"
	"go.uber.org/zap"
	"os"
)

type PayPal struct {
	client *paypal.Client
	logger *zap.Logger
}

func New(url string, logger *zap.Logger) (*PayPal, error) {
	c, err := paypal.NewClient("clientID", "secretID", url)
	if err != nil {
		return nil, err
	}
	c.SetLog(os.Stdout)

	return &PayPal{
		client: c,
		logger: logger,
	}, nil
}

func (p *PayPal) Unsubscribe(id string) error {
	_, err := p.client.GetAccessToken(context.Background())
	if err != nil {
		return err
	}
	return p.client.DeactivateSubscriptionPlans(context.Background(), id)
}
