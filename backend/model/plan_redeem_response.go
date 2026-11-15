package model

import "time"

type PlanRedeemResponse struct {
	Plan   string    `json:"plan"`
	EndsAt time.Time `json:"ends_at"`
}
