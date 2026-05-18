package user

import (
	"github.com/plutov/paypal/v4"
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
	SendPlanUnSubscribed(to string) error
}

type Remover interface {
	DeleteAllDomains(userId int64) error
}

type PayPalSubscriptionChecker interface {
	GetSubscriptionDetails(id string) (*paypal.SubscriptionDetailResp, error)
}

type Cleaner struct {
	database Database
	state    State
	mail     Mail
	remover  Remover
	checker  PayPalSubscriptionChecker
	enabled  bool
	logger   *zap.Logger
}

func NewCleaner(
	database Database,
	state State,
	mail Mail,
	remover Remover,
	checker PayPalSubscriptionChecker,
	enabled bool,
	logger *zap.Logger) *Cleaner {
	return &Cleaner{
		database: database,
		state:    state,
		mail:     mail,
		remover:  remover,
		checker:  checker,
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
			time.Sleep(60 * time.Second)
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
	c.logger.Info("cleaner", zap.Int64("user id", id))

	if id == 0 {
		return c.state.Set(id)
	}
	user, err := c.database.GetUser(id)
	if err != nil {
		return err
	}
	if user.IsSubscribed() {
		c.logger.Info("cleaner user subscribed")

		if !user.IsPayPal() {
			c.logger.Info("cleaner not paypal user")

			return c.state.Set(id)
		}

		details, err := c.checker.GetSubscriptionDetails(*user.SubscriptionId)
		if err != nil {
			return err
		}
		if details.SubscriptionStatus == paypal.SubscriptionStatusActive {
			return c.state.Set(id)
		}

		c.logger.Info("paypal subscription is not active", zap.String("status", string(details.SubscriptionStatus)))
		user.UnSubscribe(now)
		err = c.database.UpdateUser(user)
		if err != nil {
			return err
		}
		err = c.remover.DeleteAllDomains(id)
		if err != nil {
			return err
		}
		return c.mail.SendPlanUnSubscribed(user.Email)

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
	return c.state.Set(id)
}
