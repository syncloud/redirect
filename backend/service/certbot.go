package service

import (
	"fmt"
	"github.com/syncloud/redirect/model"
	"strings"
)

type Certbot struct {
	db        CertbotDb
	amazonDns CertbotDns
}

func NewCertbot(db CertbotDb, amazonDns CertbotDns) *Certbot {
	return &Certbot{
		db:        db,
		amazonDns: amazonDns,
	}
}

type CertbotDb interface {
	GetDomainByToken(token string) (*model.Domain, error)
}

type CertbotDns interface {
	CreateCertbotRecord(hostedZoneId string, name string, value string) error
	DeleteCertbotRecord(hostedZoneId string, name string, value string) error
}

func (c Certbot) Present(token string, fqdn string, value string) error {
	domain, err := c.db.GetDomainByToken(token)
	if err != nil {
		return err
	}
	if !strings.Contains(fqdn, fmt.Sprintf(".%s", domain.Name)) {
		return fmt.Errorf("only same domain is allowed")
	}
	err = c.amazonDns.CreateCertbotRecord(domain.HostedZoneId, fqdn, value)
	return err
}

func (c Certbot) CleanUp(token string, fqdn string, value string) error {
	domain, err := c.db.GetDomainByToken(token)
	if err != nil {
		return err
	}
	if !strings.Contains(fqdn, fmt.Sprintf(".%s", domain.Name)) {
		return fmt.Errorf("only same domain is allowed")
	}
	err = c.amazonDns.DeleteCertbotRecord(domain.HostedZoneId, fqdn, value)
	return err
}
