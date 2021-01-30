package model

import "time"

type User struct {
	Id              int64    `json:"-"`
	Email           string    `json:"email,omitempty"`
	PasswordHash    string    `json:"-"`
	Active          bool      `json:"active,omitempty"`
	UpdateToken     string    `json:"update_token,omitempty"`
	Unsubscribed    bool      `json:"unsubscribed,omitempty"`
	PremiumStatusId int       `json:"premium_status_id,omitempty"`
	Timestamp       time.Time `json:"timestamp,omitempty"`
}

