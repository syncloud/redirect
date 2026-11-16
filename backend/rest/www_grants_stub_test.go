package rest

import (
	"time"

	"github.com/syncloud/redirect/grant"
	"github.com/syncloud/redirect/model"
)

type WwwGrantsStub struct {
	redeemed   string
	forUser    int64
	redemption *grant.Redemption
	err        error
}

func (s *WwwGrantsStub) Redeem(user *model.User, code string, _ time.Time) (*grant.Redemption, error) {
	s.redeemed = code
	s.forUser = user.Id
	if s.err != nil {
		return nil, s.err
	}
	return s.redemption, nil
}
