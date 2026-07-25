package subscription

import (
	"context"
	"github.com/plutov/paypal/v4"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
	"os"
)

type PayPal struct {
	client           *paypal.Client
	clientId         string
	planMonthlyId    string
	planAnnualId     string
	planMaxMonthlyId string
	planMaxAnnualId  string
	logger           *zap.Logger
}

func New(clientID, secretID, url, planMonthlyId, planAnnualId, planMaxMonthlyId, planMaxAnnualId string, logger *zap.Logger) (*PayPal, error) {
	c, err := paypal.NewClient(clientID, secretID, url)
	if err != nil {
		return nil, err
	}
	c.SetLog(os.Stdout)

	return &PayPal{
		client:           c,
		clientId:         clientID,
		planMonthlyId:    planMonthlyId,
		planAnnualId:     planAnnualId,
		planMaxMonthlyId: planMaxMonthlyId,
		planMaxAnnualId:  planMaxAnnualId,
		logger:           logger,
	}, nil
}

func (p *PayPal) MaxEnabled() bool {
	return p.planMaxMonthlyId != "" && p.planMaxAnnualId != ""
}

func (p *PayPal) Tier(planId string) string {
	if planId != "" && (planId == p.planMaxMonthlyId || planId == p.planMaxAnnualId) {
		return model.PlanMax
	}
	return model.PlanPro
}

func (p *PayPal) Plans() model.PlanResponse {
	return model.PlanResponse{
		PlanMonthlyId:    p.planMonthlyId,
		PlanAnnualId:     p.planAnnualId,
		PlanMaxMonthlyId: p.planMaxMonthlyId,
		PlanMaxAnnualId:  p.planMaxAnnualId,
		ClientId:         p.clientId,
		PayPalMaxEnabled: p.MaxEnabled(),
	}
}

func (p *PayPal) Unsubscribe(id string) error {
	_, err := p.client.GetAccessToken(context.Background())
	if err != nil {
		return err
	}
	return p.client.CancelSubscription(context.Background(), id, "user action")
}

func (p *PayPal) GetSubscriptionDetails(id string) (*paypal.SubscriptionDetailResp, error) {
	_, err := p.client.GetAccessToken(context.Background())
	if err != nil {
		return nil, err
	}
	return p.client.GetSubscriptionDetails(context.Background(), id)
}

func (p *PayPal) PlanId(id string) (string, error) {
	details, err := p.GetSubscriptionDetails(id)
	if err != nil {
		return "", err
	}
	return details.PlanID, nil
}

func (p *PayPal) IsActive(id string) (bool, error) {
	details, err := p.GetSubscriptionDetails(id)
	if err != nil {
		return false, err
	}
	return details.SubscriptionStatus == paypal.SubscriptionStatusActive, nil
}
