package model

type PlanResponse struct {
	PlanMonthlyId    string `json:"plan_monthly_id,omitempty"`
	PlanAnnualId     string `json:"plan_annual_id,omitempty"`
	PlanMaxMonthlyId string `json:"plan_max_monthly_id,omitempty"`
	PlanMaxAnnualId  string `json:"plan_max_annual_id,omitempty"`
	ClientId         string `json:"client_id,omitempty"`
	SdkUrl           string `json:"sdk_url,omitempty"`
	StripeMaxEnabled bool   `json:"stripe_max_enabled"`
	PayPalMaxEnabled bool   `json:"paypal_max_enabled"`
}
