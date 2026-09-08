package subscription

import (
	"context"
	"github.com/plutov/paypal/v4"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type PayPal struct {
	client           *paypal.Client
	clientId         string
	sdkUrl           string
	planMonthlyId    string
	planAnnualId     string
	planMaxMonthlyId string
	planMaxAnnualId  string
	returnUrl        string
	cancelUrl        string
	logger           *zap.Logger
}

func New(clientID, secretID, url, sdkUrl, planMonthlyId, planAnnualId, planMaxMonthlyId, planMaxAnnualId, returnUrl, cancelUrl string, logger *zap.Logger) (*PayPal, error) {
	c, err := paypal.NewClient(clientID, secretID, url)
	if err != nil {
		return nil, err
	}
	return &PayPal{
		client:           c,
		clientId:         clientID,
		sdkUrl:           sdkUrl,
		planMonthlyId:    planMonthlyId,
		planAnnualId:     planAnnualId,
		planMaxMonthlyId: planMaxMonthlyId,
		planMaxAnnualId:  planMaxAnnualId,
		returnUrl:        returnUrl,
		cancelUrl:        cancelUrl,
		logger:           logger,
	}, nil
}

func (p *PayPal) Period(planId string) string {
	if planId == p.planAnnualId || planId == p.planMaxAnnualId {
		return model.PeriodYear
	}
	return model.PeriodMonth
}

func (p *PayPal) AnnualPlanId(planId string) string {
	if planId == p.planMaxMonthlyId || planId == p.planMaxAnnualId {
		return p.planMaxAnnualId
	}
	return p.planAnnualId
}

func (p *PayPal) Revise(subscriptionId string, planId string) (string, error) {
	_, err := p.client.GetAccessToken(context.Background())
	if err != nil {
		return "", err
	}
	response, err := p.client.ReviseSubscription(context.Background(), subscriptionId, paypal.SubscriptionBase{
		PlanID: planId,
		ApplicationContext: &paypal.ApplicationContext{
			ReturnURL: p.returnUrl,
			CancelURL: p.cancelUrl,
		},
	})
	if err != nil {
		return "", err
	}
	for _, link := range response.Links {
		if link.Rel == "approve" {
			return link.Href, nil
		}
	}
	return "", nil
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
		SdkUrl:           p.sdkUrl,
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
