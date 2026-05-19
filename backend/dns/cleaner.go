package dns

import (
	"fmt"
	"github.com/syncloud/redirect/metrics"
	"github.com/syncloud/redirect/model"
	"time"
)

type Database interface {
	GetDomainTokenUpdatedBefore(before time.Time) (string, error)
	GetDomainByToken(token string) (*model.Domain, error)
	UpdateDomain(domain *model.Domain) error
	GetUser(id int64) (*model.User, error)
}

type Remover interface {
	DeleteDomain(userId int64, domainName string) error
}

type Mail interface {
	SendDnsCleanNotification(to string, userDomain string) error
}

type Cleaner struct {
	database Database
	remover  Remover
	mail     Mail
	metrics  *metrics.Metrics
}

func NewCleaner(database Database, dns Remover, mail Mail, m *metrics.Metrics) *Cleaner {
	return &Cleaner{
		database: database,
		remover:  dns,
		mail:     mail,
		metrics:  m,
	}
}

func (c *Cleaner) Start() error {
	go func() {
		for {
			err := c.Clean(time.Now())
			if err != nil {
				fmt.Printf("error: %v", err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

func (c *Cleaner) Clean(now time.Time) error {
	monthOld := now.AddDate(0, -1, 0)
	token, err := c.database.GetDomainTokenUpdatedBefore(monthOld)
	if err != nil {
		c.metrics.Cleaner("error")
		return err
	}
	if token == "" {
		//fmt.Printf("not found\n")
		return nil
	}
	domain, err := c.database.GetDomainByToken(token)
	if err != nil {
		c.metrics.Cleaner("error")
		return err
	}
	if domain == nil {
		fmt.Printf("token not found: %s\n", token)
		return nil
	}
	lastUpdate := domain.LastUpdate
	format := "nil"
	if lastUpdate != nil {
		if !lastUpdate.Before(monthOld) {
			return nil
		}
		format = lastUpdate.Format(time.RFC3339)
	}
	user, err := c.database.GetUser(domain.UserId)
	if err != nil {
		c.metrics.Cleaner("error")
		return err
	}
	fmt.Printf("id: %d, domain: %s, last update: %s, user subscribed: %v\n", domain.Id, domain.Name, format, user.IsSubscribed())
	if !user.IsSubscribed() {
		c.metrics.Cleaner("delete")
		err = c.remover.DeleteDomain(user.Id, domain.Name)
		if err != nil {
			c.metrics.Cleaner("error")
			return err
		}
		err = c.mail.SendDnsCleanNotification(user.Email, domain.Name)
		if err != nil {
			c.metrics.Cleaner("error")
			fmt.Printf("cannot send dns clean email: %s\n", err)
		}
	} else {
		domain.LastUpdate = &now
		err = c.database.UpdateDomain(domain)
		if err != nil {
			c.metrics.Cleaner("error")
			return err
		}
	}
	return nil
}
