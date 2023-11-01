package model

type PlanResponse struct {
	PlanMonthlyId string `json:"plan_monthly_id,omitempty"`
	PlanAnnualId  string `json:"plan_annual_id,omitempty"`
	ClientId      string `json:"client_id,omitempty"`
}
