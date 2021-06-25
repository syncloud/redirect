package model

import "time"

// UserResponse TODO: Migrate mobile apps to a separate user and domains calls
type UserResponse struct {
	Email           string    `json:"email,omitempty"`
	Active          bool      `json:"active,omitempty"`
	UpdateToken     string    `json:"update_token,omitempty"`
	Unsubscribed    bool      `json:"unsubscribed"`
	PremiumStatusId int       `json:"premium_status_id,omitempty"`
	Timestamp       time.Time `json:"timestamp,omitempty"`
	Domains         []*Domain `json:"domains"`
}
