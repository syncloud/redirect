package service

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/model"
	"testing"
)

type DnsStub struct {
}

func (dns *DnsStub) DeleteHostedZone(hostedZoneId string) error {
	return nil
}

func (dns *DnsStub) CreateHostedZone(domain string) (*string, error) {
	id := "123"
	return &id, nil
}

func (dns *DnsStub) UpdateDomainRecords(domain *model.Domain) error {
	return nil
}

func (dns *DnsStub) DeleteDomainRecords(domain *model.Domain) error {
	return nil
}

var _ dns.Dns = (*DnsStub)(nil)

type DomainsDbStub struct {
	userId   int64
	found    bool
	updated  bool
	inserted bool
}

func (db *DomainsDbStub) GetDomainByToken(token string) (*model.Domain, error) {
	return nil, nil
}

func (db *DomainsDbStub) GetUserDomains(userId int64) ([]*model.Domain, error) {
	return nil, nil
}

func (db *DomainsDbStub) GetUser(id int64) (*model.User, error) {
	return nil, nil
}

func (db *DomainsDbStub) DeleteAllDomains(userId int64) error {
	return nil
}

func (db *DomainsDbStub) DeleteDomain(domainId uint64) error {
	return nil
}

func (db *DomainsDbStub) GetDomainByName(value string) (*model.Domain, error) {
	if db.found {
		return &model.Domain{Name: value, UserId: db.userId}, nil
	}
	return nil, nil
}
func (db *DomainsDbStub) InsertDomain(domain *model.Domain) error {
	db.inserted = true
	return nil

}
func (db *DomainsDbStub) UpdateDomain(domain *model.Domain) error {
	db.updated = true
	return nil

}

var _ DomainsDb = (*DomainsDbStub)(nil)

type DomainsUsersStub struct {
	userId        int64
	authenticated bool
	premiumStatus int
}

func (users *DomainsUsersStub) Authenticate(email *string, password *string) (*model.User, error) {
	if users.authenticated {
		return &model.User{Id: users.userId, Email: *email, Active: true, PremiumStatusId: users.premiumStatus}, nil
	}
	return nil, fmt.Errorf("authentication failed")
}

var _ DomainsUsers = (*DomainsUsersStub)(nil)

func TestNotChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.False(t, changed)
}

func TestIpChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "2"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestIpv6Changed(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "2"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestDkimChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "2"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestLocalIpChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "2"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		true, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestMapLocalAddressChanged(t *testing.T) {
	ip := "1"
	ipv6 := "1"
	dkim := "1"
	localIp := "1"

	newIp := "1"
	newIpv6 := "1"
	newDkim := "1"
	newLocalIp := "1"

	changed := Changed(
		true, &ip, &ipv6, &dkim, &localIp,
		false, &newIp, &newIpv6, &newDkim, &newLocalIp)

	assert.True(t, changed)
}

func TestEquals(t *testing.T) {
	ip := "127.0.0.1"
	ip1 := "127.0.0.1"

	assert.True(t, Equals(&ip, &ip1))
	assert.True(t, Equals(nil, nil))
	assert.False(t, Equals(&ip, nil))
	assert.False(t, Equals(nil, &ip))
}

func TestAcquireFreeDomain_ExistingMine(t *testing.T) {
	db := &DomainsDbStub{found: true, userId: 1}
	dnsStub := &DnsStub{}
	users := &DomainsUsersStub{authenticated: true, userId: 1}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	users := &DomainsUsersStub{authenticated: true, userId: 1, premiumStatus: PremiumStatusInactive}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	users := &DomainsUsersStub{authenticated: true, userId: 1, premiumStatus: PremiumStatusActive}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	users := &DomainsUsersStub{authenticated: true, userId: 1, premiumStatus: PremiumStatusInactive}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
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
	users := &DomainsUsersStub{authenticated: true, userId: 1, premiumStatus: PremiumStatusActive}
	domains := NewDomains(dnsStub, db, users, "syncloud.it", "")
	domain := "example.com"
	password := "password"
	email := "test@example.com"
	request := model.DomainAvailabilityRequest{Email: &email, Password: &password, Domain: &domain}
	result, err := domains.Availability(request)

	assert.Nil(t, err)
	assert.Nil(t, result)

}
