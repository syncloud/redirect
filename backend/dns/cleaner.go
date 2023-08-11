package dns

import (
	"fmt"
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
	DeleteDomainRecords(domain *model.Domain) error
}

type Mail interface {
	SendDnsCleanNotification(to string, userDomain string) error
}

type Cleaner struct {
	database Database
	dns      Remover
	mail     Mail
}

func NewCleaner(database Database, dns Remover, mail Mail) *Cleaner {
	return &Cleaner{
		database: database,
		dns:      dns,
		mail:     mail,
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
		return err
	}
	if token == "" {
		//fmt.Printf("not found\n")
		return nil
	}
	domain, err := c.database.GetDomainByToken(token)
	if err != nil {
		return err
	}
	if domain == nil {
		fmt.Printf("token not found: %s\n", token)
		return nil
	}
	if !domain.LastUpdate.Before(monthOld) {
		return nil
	}
	user, err := c.database.GetUser(domain.UserId)
	if err != nil {
		return err
	}
	fmt.Printf("id: %d, domain: %s, last update: %s, user subscribed: %v\n", domain.Id, domain.Name, domain.LastUpdate.Format(time.RFC3339), user.SubscriptionId != nil)
	if user.SubscriptionId == nil {
		err = c.dns.DeleteDomainRecords(domain)
		if err != nil {
			return err
		}
		domain.Ip = nil
		err = c.mail.SendDnsCleanNotification(user.Email, domain.Name)
		if err != nil {
			fmt.Printf("cannot send dns clean email: %s\n", err)
		}
	}
	domain.LastUpdate = &now
	err = c.database.UpdateDomain(domain)
	if err != nil {
		return err
	}
	return nil
}
