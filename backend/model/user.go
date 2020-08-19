package model

import "time"

type User struct {
	Id           uint64
	Email        string
	PasswordHash string
	Active       bool
	UpdateToken  string
	Unsubscribed bool
	Timestamp    time.Time
}
