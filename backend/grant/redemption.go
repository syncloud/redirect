package grant

import "time"

type Redemption struct {
	Id      int64
	GrantId int64
	UserId  int64
	Plan    string
	EndsAt  time.Time
}
