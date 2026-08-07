package relay

import "github.com/syncloud/redirect/model"

type WarnUsers interface {
	GetUser(id int64) (*model.User, error)
}

type WarnMail interface {
	SendMailRelayLimitWarning(to string, used int64, limit int64) error
}

type LimitWarner struct {
	users WarnUsers
	mail  WarnMail
}

func NewLimitWarner(users WarnUsers, mail WarnMail) *LimitWarner {
	return &LimitWarner{users: users, mail: mail}
}

func (w *LimitWarner) Warn(userId int64, used int64, limit int64) error {
	user, err := w.users.GetUser(userId)
	if err != nil || user == nil {
		return err
	}
	return w.mail.SendMailRelayLimitWarning(user.Email, used, limit)
}
