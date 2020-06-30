package db

import "time"

type User struct {
	id           uint64
	email        string
	passwordHash string
	active       bool
	updateToken  string
	unsubscribed bool
	timestamp    time.Time
	isPremium    bool
}
