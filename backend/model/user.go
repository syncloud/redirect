package model

import "time"

type User struct {
	Id                  int64     `json:"-"`
	Email               string    `json:"email,omitempty"`
	PasswordHash        string    `json:"-"`
	Active              bool      `json:"active,omitempty"`
	UpdateToken         string    `json:"update_token,omitempty"`
	NotificationEnabled bool      `json:"notification_enabled"`
	Timestamp           time.Time `json:"timestamp,omitempty"`
	SubscriptionId      *string   `json:"subscription_id,omitempty"`
}
