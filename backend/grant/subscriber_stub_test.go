package grant

import "github.com/syncloud/redirect/model"

type subscriberStub struct {
	user  *model.User
	id    string
	kind  int
	plan  string
	calls int
	err   error
}

func (s *subscriberStub) Subscribe(user *model.User, subscriptionId string, subscriptionType int, plan string) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	s.user = user
	s.id = subscriptionId
	s.kind = subscriptionType
	s.plan = plan
	user.Subscribe(subscriptionId, subscriptionType, plan)
	return nil
}
