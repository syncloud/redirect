package model

import "time"

type Action struct {
	Id           uint64
	ActionTypeId uint64
	UserId       int64
	Token        string
	Timestamp    time.Time
	SentAt       *time.Time
	Attempts     int
}

type PendingActivation struct {
	ActionId uint64
	Token    string
	Email    string
	Attempts int
}
