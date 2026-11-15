package rest

import (
	"time"

	"github.com/syncloud/redirect/grant"
	"github.com/syncloud/redirect/model"
)

type WwwGrants interface {
	Redeem(user *model.User, code string, now time.Time) (*grant.Redemption, error)
}
