package model

import "time"

type Action struct {
	Id           uint64
	ActionTypeId uint64
	UserId       int64
	Token        string
	Timestamp    time.Time
}
