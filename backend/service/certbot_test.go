package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
)

type CertbotDnsStub struct {
}

func (c CertbotDnsStub) CreateCertbotRecord(hostedZoneId string, name string, value string) error {
	return nil
}

func (c CertbotDnsStub) DeleteCertbotRecord(hostedZoneId string, name string, value string) error {
	return nil
}

type CertbotDbStub struct {
	Name string
}

func (db *CertbotDbStub) GetDomainByToken(token string) (*model.Domain, error) {
	return &model.Domain{Name: db.Name, UserId: 1, HostedZoneId: ""}, nil
}

func TestPresentMyDomain(t *testing.T) {

	db := &CertbotDbStub{Name: "test.syncloud.it"}
	dnsStub := &CertbotDnsStub{}
	certbot := NewCertbot(db, dnsStub)

	err := certbot.Present("token", "acme-123.test.syncloud.it", "123value")

	assert.Nil(t, err)
}

func TestPresentNotMyDomain(t *testing.T) {

	db := &CertbotDbStub{Name: "test1.syncloud.it"}
	dnsStub := &CertbotDnsStub{}
	certbot := NewCertbot(db, dnsStub)

	err := certbot.Present("token", "acme-123.test.syncloud.it", "123value")

	assert.NotNil(t, err)
}

func TestPresentNotMyDomainContains(t *testing.T) {

	db := &CertbotDbStub{Name: "1test.syncloud.it"}
	dnsStub := &CertbotDnsStub{}
	certbot := NewCertbot(db, dnsStub)

	err := certbot.Present("token", "acme-123.11test.syncloud.it", "123value")

	assert.NotNil(t, err)
}

func TestCleanUpMyDomain(t *testing.T) {

	db := &CertbotDbStub{Name: "test.syncloud.it"}
	dnsStub := &CertbotDnsStub{}
	certbot := NewCertbot(db, dnsStub)

	err := certbot.CleanUp("token", "acme-123.test.syncloud.it", "123value")

	assert.Nil(t, err)
}

func TestCleanUpNotMyDomain(t *testing.T) {

	db := &CertbotDbStub{Name: "test1.syncloud.it"}
	dnsStub := &CertbotDnsStub{}
	certbot := NewCertbot(db, dnsStub)

	err := certbot.CleanUp("token", "acme-123.test.syncloud.it", "123value")

	assert.NotNil(t, err)
}

func TestCleanUpNotMyDomainContains(t *testing.T) {

	db := &CertbotDbStub{Name: "1test.syncloud.it"}
	dnsStub := &CertbotDnsStub{}
	certbot := NewCertbot(db, dnsStub)

	err := certbot.CleanUp("token", "acme-123.11test.syncloud.it", "123value")

	assert.NotNil(t, err)
}
