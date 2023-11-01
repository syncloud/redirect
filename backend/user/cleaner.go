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
}

type State interface {
	Get() (int64, error)
	Set(userId int64) error
}

type Mail interface {
	SendTrial(to string) error
	SendAccountLockSoon(to string) error
	SendAccountLocked(to string) error
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
	if user.IsSubscribed() || user.IsLocked() {
		return c.state.Set(id)
	}
	if user.IsStatusCreated() {
		user.TrialEmailSent()
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
		user.LockEmailSent()
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
		user.Lock()
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
	//TODO: remove account after 10 days from lock
	return c.state.Set(id)
}