package user

import (
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
	"time"
)

type Database interface {
	GetNextUserId(after int64) (int64, error)
	GetUser(id int64) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(userId int64) error
	DeleteActions(userId int64) error
}

type State interface {
	Get() (int64, error)
	Set(userId int64) error
}

type Mail interface {
	SendTrial(to string) error
	SendAccountLockSoon(to string) error
	SendAccountLocked(to string) error
	SendAccountRemoved(to string) error
}

type Remover interface {
	DeleteAllDomains(userId int64) error
}

type Cleaner struct {
	database Database
	state    State
	mail     Mail
	remover  Remover
	enabled  bool
	logger   *zap.Logger
}

func NewCleaner(database Database, state State, mail Mail, remover Remover, enabled bool, logger *zap.Logger) *Cleaner {
	return &Cleaner{
		database: database,
		state:    state,
		mail:     mail,
		remover:  remover,
		enabled:  enabled,
		logger:   logger,
	}
}

func (c *Cleaner) Start() error {
	if !c.enabled {
		c.logger.Warn("user cleaner is disabled")
		return nil
	}

	go func() {
		for {
			err := c.Clean(time.Now())
			if err != nil {
				c.logger.Error("unable to clean users", zap.Error(err))
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

func (c *Cleaner) Clean(now time.Time) error {
	userId, err := c.state.Get()
	if err != nil {
		return err
	}

	id, err := c.database.GetNextUserId(userId)
	if err != nil {
		return err
	}
	if id == 0 {
		return c.state.Set(id)
	}
	user, err := c.database.GetUser(id)
	if err != nil {
		return err
	}
	if user.IsSubscribed() {
		return c.state.Set(id)
	}
	if user.IsStatusCreated() {
		user.TrialEmailSent(now)
		err = c.database.UpdateUser(user)
		if err != nil {
			return err
		}
		err = c.mail.SendTrial(user.Email)
		if err != nil {
			return err
		}
	}
	if user.IsReadyForLockEmail(now) {
		user.LockEmailSent(now)
		err = c.database.UpdateUser(user)
		if err != nil {
			return err
		}
		err = c.mail.SendAccountLockSoon(user.Email)
		if err != nil {
			return err
		}
	}
	if user.IsReadyForAccountLock(now) {
		user.Lock(now)
		err = c.remover.DeleteAllDomains(id)
		if err != nil {
			return err
		}
		err = c.database.UpdateUser(user)
		if err != nil {
			return err
		}
		err = c.mail.SendAccountLocked(user.Email)
		if err != nil {
			return err
		}
	}
	if user.IsReadyForAccountRemove(now) {
		err = c.remover.DeleteAllDomains(id)
		if err != nil {
			return err
		}
		err = c.database.DeleteActions(user.Id)
		if err != nil {
			return err
		}
		err = c.database.DeleteUser(user.Id)
		if err != nil {
			return err
		}
		err = c.mail.SendAccountRemoved(user.Email)
		if err != nil {
			return err
		}
	}
	//TODO: lock accounts without subscription after 1 day
	return c.state.Set(id)
}
