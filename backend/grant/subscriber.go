package grant

import "github.com/syncloud/redirect/model"

type Subscriber interface {
	Subscribe(user *model.User, subscriptionId string, subscriptionType int, plan string) error
}
