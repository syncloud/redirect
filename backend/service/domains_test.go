package service

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
)

type DnsStub struct {
	hostedZoneDeleted bool
	recordsDeleted    bool
	certbotDeleted    bool
	updated           bool
	error             error
}

func (dns *DnsStub) GetHostedZoneNameServers(_ string) ([]*string, error) {
	return []*string{aws.String("ns1.example.com")}, nil
}

func (dns *DnsStub) DeleteHostedZone(_ string) error {
	if dns.error != nil {
		return dns.error
	}
	dns.hostedZoneDeleted = true
	return nil
}

func (dns *DnsStub) CreateHostedZone(_ string) (*string, error) {
	if dns.error != nil {
		return nil, dns.error
	}
	id := "123"
	return &id, nil
}

func (dns *DnsStub) UpdateDomainRecords(_ *model.Domain) error {
	if dns.error != nil {
		return dns.error
	}
	dns.updated = true
	return nil
}

func (dns *DnsStub) DeleteDomainRecords(_ *model.Domain) error {
	if dns.error != nil {
		return dns.error
	}
	dns.recordsDeleted = true
	return nil
}

func (dns *DnsStub) DeleteCertbotRecord(_ string, _ string) error {
	if dns.error != nil {
		return dns.error
	}
	dns.certbotDeleted = true
	return nil
}

type DomainsDbStub struct {
	userId       int64
	found        bool
	updated      bool
	inserted     bool
	deleted      bool
	hostedZoneId string
	userStatus   int64
}

func (db *DomainsDbStub) GetDomainByToken(_ string) (*model.Domain, error) {
	if db.found {
		return &model.Domain{Name: "name", UserId: db.userId, HostedZoneId: db.hostedZoneId}, nil
	}
	return nil, nil
}

func (db *DomainsDbStub) GetUserDomains(_ int64) ([]*model.Domain, error) {
	if db.found {
		return []*model.Domain{{Name: "name", UserId: db.userId, HostedZoneId: db.hostedZoneId}}, nil
	}
	return nil, nil
}

func (db *DomainsDbStub) GetUser(_ int64) (*model.User, error) {
	if db.found {
		return &model.User{Id: db.userId, Active: true, Status: db.userStatus}, nil
	}
	return nil, nil
}

func (db *DomainsDbStub) DeleteAllDomains(_ int64) error {
	return nil
}

func (db *DomainsDbStub) DeleteDomain(_ uint64) error {
	db.deleted = true
	return nil
}

func (db *DomainsDbStub) GetDomainByName(value string) (*model.Domain, error) {
	if db.found {
		return &model.Domain{Name: value, UserId: db.userId, HostedZoneId: db.hostedZoneId}, nil
	}
	return nil, nil
}
func (db *DomainsDbStub) InsertDomain(_ *model.Domain) error {
	db.inserted = true
	return nil

}
func (db *DomainsDbStub) UpdateDomain(_ *model.Domain) error {
	db.updated = true
	return nil

}

type DomainsUsersStub struct {
	userId         int64
	authenticated  bool
	subscriptionId *string
}

func (users *DomainsUsersStub) Authenticate(email *string, _ *string) (*model.User, error) {
	if users.authenticated {
		return &model.User{Id: users.userId, Email: *email, Active: true, SubscriptionId: users.subscriptionId}, nil
	}
	return nil, fmt.Errorf("authentication failed")
}

type DetectorStub struct {
	changed bool
}

func (d *DetectorStub) Changed(
	_ bool,
	_ *string,
	_ *string,
	_ *string,
	_ *string,
	_ bool,
	_ *string,
	_ *string,
	_ *string,
	_ *string) bool {
	return d.changed
}

func TestAcquireFreeDomain_ExistingMine(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "test123.syncloud.it"
	password := "password"
	email := "test@example.com"
	mac := "11:22:33:44:55:66"
	deviceName := "name"
	deviceTitle := "title"
	request := model.DomainAcquireRequest{Email: &email, Password: &password, Domain: &domain, DeviceMacAddress: &mac, DeviceName: &deviceName, DeviceTitle: &deviceTitle}
	_, err := domains.DomainAcquire(request, "")

	assert.Nil(t, err)
	assert.True(t, db.updated)
	assert.False(t, db.inserted)
}

func TestAcquireFreeDomain_ExistingNotMine(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 2}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	userDomain := "test.syncloud.it"
	password := "password"
	email := "test@example.com"
	mac := "11:22:33:44:55:66"
	deviceName := "name"
	deviceTitle := "title"
	request := model.DomainAcquireRequest{Email: &email, Password: &password, DeprecatedUserDomain: &userDomain, DeviceMacAddress: &mac, DeviceName: &deviceName, DeviceTitle: &deviceTitle}
	_, err := domains.DomainAcquire(request, "")

	assert.NotNil(t, err)
	assert.False(t, db.updated)
	assert.False(t, db.inserted)

}

func TestAcquireFreeDomain_Available(t *testing.T) {
	db := &DomainsDbStub{found: false}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "test123.syncloud.it"
	password := "password"
	email := "test@example.com"
	mac := "11:22:33:44:55:66"
	deviceName := "name"
	deviceTitle := "title"
	request := model.DomainAcquireRequest{Email: &email, Password: &password, Domain: &domain, DeviceMacAddress: &mac, DeviceName: &deviceName, DeviceTitle: &deviceTitle}
	_, err := domains.DomainAcquire(request, "")

	assert.Nil(t, err)
	assert.False(t, db.updated)
	assert.True(t, db.inserted)

}

func TestAcquirePremiumDomain_FreeUser_NotAvailable(t *testing.T) {
	db := &DomainsDbStub{found: false}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "example.com"
	password := "password"
	email := "test@example.com"
	mac := "11:22:33:44:55:66"
	deviceName := "name"
	deviceTitle := "title"
	request := model.DomainAcquireRequest{Email: &email, Password: &password, Domain: &domain, DeviceMacAddress: &mac, DeviceName: &deviceName, DeviceTitle: &deviceTitle}
	_, err := domains.DomainAcquire(request, "")

	assert.NotNil(t, err)

}

func TestAcquirePremiumDomain_PremiumUser_Available(t *testing.T) {
	db := &DomainsDbStub{found: false}
	dnsStub := &DnsStub{}
	subscriptionId := "1"
	users := &DomainsUsersStub{authenticated: true, userId: 1, subscriptionId: &subscriptionId}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "example.com"
	password := "password"
	email := "test@example.com"
	mac := "11:22:33:44:55:66"
	deviceName := "name"
	deviceTitle := "title"
	request := model.DomainAcquireRequest{Email: &email, Password: &password, Domain: &domain, DeviceMacAddress: &mac, DeviceName: &deviceName, DeviceTitle: &deviceTitle}
	_, err := domains.DomainAcquire(request, "")

	assert.Nil(t, err)
	assert.False(t, db.updated)
	assert.True(t, db.inserted)

}

func TestFreeAvailability_SameUser(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "test123.syncloud.it"
	password := "password"
	email := "test@example.com"
	request := model.DomainAvailabilityRequest{Email: &email, Password: &password, Domain: &domain}
	result, err := domains.Availability(request)

	assert.Nil(t, err)
	assert.NotNil(t, result)

}

func TestFreeAvailability_OtherUser(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 2}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "test.syncloud.it"
	password := "password"
	email := "test@example.com"
	request := model.DomainAvailabilityRequest{Email: &email, Password: &password, Domain: &domain}
	result, err := domains.Availability(request)

	assert.NotNil(t, err)
	assert.Nil(t, result)

}

func TestFreeAvailability_Available(t *testing.T) {
	db := &DomainsDbStub{found: false}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "test123.syncloud.it"
	password := "password"
	email := "test@example.com"
	request := model.DomainAvailabilityRequest{Email: &email, Password: &password, Domain: &domain}
	result, err := domains.Availability(request)

	assert.Nil(t, err)
	assert.Nil(t, result)

}

func TestPremiumAvailability_FreeUser_NotAvailable(t *testing.T) {
	db := &DomainsDbStub{found: false}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "example.com"
	password := "password"
	email := "test@example.com"
	request := model.DomainAvailabilityRequest{Email: &email, Password: &password, Domain: &domain}
	result, err := domains.Availability(request)

	assert.NotNil(t, err)
	assert.Nil(t, result)

}

func TestPremiumAvailability_PremiumUser_NotAvailable(t *testing.T) {
	db := &DomainsDbStub{found: false}
	dnsStub := &DnsStub{}
	subscriptionId := "1"
	users := &DomainsUsersStub{authenticated: true, userId: 1, subscriptionId: &subscriptionId}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "", &DetectorStub{})
	domain := "example.com"
	password := "password"
	email := "test@example.com"
	request := model.DomainAvailabilityRequest{Email: &email, Password: &password, Domain: &domain}
	result, err := domains.Availability(request)

	assert.Nil(t, err)
	assert.Nil(t, result)

}

func TestDeleteDomain_Free_DeleteRecords(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1, hostedZoneId: "1"}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "1", &DetectorStub{})
	err := domains.DeleteDomain(1, "test.syncloud.it")

	assert.Nil(t, err)
	assert.False(t, dnsStub.hostedZoneDeleted)
	assert.True(t, dnsStub.recordsDeleted)
	assert.False(t, dnsStub.certbotDeleted)
	assert.True(t, db.deleted)

}

func TestDeleteDomain_Premium_DeleteRecordsAndHostedZone(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1, hostedZoneId: "1"}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "2", &DetectorStub{})
	err := domains.DeleteDomain(1, "test.com")

	assert.Nil(t, err)
	assert.True(t, dnsStub.hostedZoneDeleted)
	assert.True(t, dnsStub.recordsDeleted)
	assert.True(t, dnsStub.certbotDeleted)
	assert.True(t, db.deleted)

}

func TestDeleteDomain_Premium_DeleteRecordsAndHostedZone_IgnoreNoSuchHostedZoneError(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1, hostedZoneId: "1"}
	dnsStub := &DnsStub{error: awserr.New(route53.ErrCodeNoSuchHostedZone, "not found", nil)}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "2", &DetectorStub{})
	err := domains.DeleteDomain(1, "test.com")

	assert.Nil(t, err)
	assert.False(t, dnsStub.hostedZoneDeleted)
	assert.False(t, dnsStub.recordsDeleted)
	assert.False(t, dnsStub.certbotDeleted)
	assert.True(t, db.deleted)

}

func TestGetDomains_Free_NoNameServers(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1, hostedZoneId: "1"}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domainService := NewDomains(dnsStub, db, users, "syncloud.it", "1", &DetectorStub{})
	domains, err := domainService.GetDomains(&model.User{Id: 1})

	assert.Nil(t, err)
	assert.Empty(t, domains[0].NameServers)
}

func TestGetDomains_Premium_NameServers(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1, hostedZoneId: "1"}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domainService := NewDomains(dnsStub, db, users, "syncloud.it", "2", &DetectorStub{})
	domains, err := domainService.GetDomains(&model.User{Id: 1})

	assert.Nil(t, err)
	assert.NotEmpty(t, domains[0].NameServers)
}

func TestDomains_Update_Ipv6_Changed(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1, hostedZoneId: "1"}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	token := "123"
	ipv6 := "fe80::0000:0000:0000:0000"
	requestIp := "fe80::1111:1111:1111:1111"
	webLocalPort := 443
	webProtocol := "https"
	detector := &DetectorStub{changed: true}
	domainService := NewDomains(dnsStub, db, users, "syncloud.it", "2", detector)
	domain, err := domainService.Update(model.DomainUpdateRequest{
		MapLocalAddress: false,
		WebLocalPort:    &webLocalPort,
		WebProtocol:     &webProtocol,
		Token:           &token,
		Ipv6:            &ipv6,
		Ipv4Enabled:     false,
		Ipv6Enabled:     true,
	}, &requestIp)

	assert.Nil(t, err)
	assert.Equal(t, "fe80::0000:0000:0000:0000", *domain.Ipv6)
	assert.True(t, db.updated)
	assert.True(t, dnsStub.updated)
}

func TestDomains_Update_LockedUser_Error(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1, hostedZoneId: "1", userStatus: model.StatusLocked}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	token := "123"
	ipv6 := "fe80::0000:0000:0000:0000"
	requestIp := "fe80::1111:1111:1111:1111"
	webLocalPort := 443
	webProtocol := "https"
	detector := &DetectorStub{changed: true}
	domainService := NewDomains(dnsStub, db, users, "syncloud.it", "2", detector)
	_, err := domainService.Update(model.DomainUpdateRequest{
		MapLocalAddress: false,
		WebLocalPort:    &webLocalPort,
		WebProtocol:     &webProtocol,
		Token:           &token,
		Ipv6:            &ipv6,
		Ipv4Enabled:     false,
		Ipv6Enabled:     true,
	}, &requestIp)

	assert.Error(t, err)
}
