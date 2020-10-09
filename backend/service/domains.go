package service

import (
	"fmt"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/utils"
	"log"
	"time"
)

type Domains struct {
	amazonDns    dns.Dns
	db           *db.MySql
	domain       string
	users        *Users
	hostedZoneId string
}

func NewDomains(dnsImpl dns.Dns, db *db.MySql, domain string, users *Users, hostedZoneId string) *Domains {
	return &Domains{amazonDns: dnsImpl, db: db, domain: domain, users: users, hostedZoneId: hostedZoneId}
}

func (d *Domains) GetDomain(token string) (*model.Domain, error) {
	validator := NewValidator()
	validator.Token(&token)
	if validator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: validator.ToParametersMessages()}
	}
	domain, err := d.db.GetDomainByToken(token)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("unknown domain update token")}
	}
	user, err := d.db.GetUser(domain.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Active {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("unknown domain update token")}
	}

	return domain, nil
}

func (d *Domains) DomainAcquire(request model.DomainAcquireRequest) (*model.Domain, error) {

	user, err := d.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	validator := NewValidator()
	userDomain := validator.newUserDomain(request.UserDomain)
	deviceMacAddress := validator.deviceMacAddress(request.DeviceMacAddress)
	validator.deviceName(request.DeviceName)
	validator.deviceTitle(request.DeviceTitle)

	if validator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: validator.ToParametersMessages()}
	}

	domain, err := d.db.GetDomainByUserDomain(*userDomain)
	if err != nil {
		return nil, err
	}
	log.Printf("domain: %v, user: %v\n", domain, user)
	if domain != nil && domain.UserId != user.Id {
		return nil, &model.ParameterError{ParameterErrors: &[]model.ParameterMessages{{
			Parameter: "user_domain", Messages: []string{"User domain name is already in use"},
		}}}
	}
	updateToken := utils.Uuid()
	log.Println("uuid", updateToken)
	if domain == nil {
		domain = &model.Domain{
			UserDomain:       *userDomain,
			DeviceMacAddress: deviceMacAddress,
			DeviceName:       request.DeviceName,
			DeviceTitle:      request.DeviceTitle,
			UpdateToken:      &updateToken,
			UserId:           user.Id,
		}
		err := d.db.InsertDomain(domain)
		if err != nil {
			return nil, err
		}

	} else {
		domain.UpdateToken = &updateToken
		domain.DeviceMacAddress = deviceMacAddress
		domain.DeviceName = request.DeviceName
		domain.DeviceTitle = request.DeviceTitle

		err := d.db.UpdateDomain(domain)
		if err != nil {
			return nil, err
		}

	}
	log.Println("domain acquired")
	return domain, nil
}

func (d *Domains) CustomDomainAcquire(request model.CustomDomainAcquireRequest) (*model.Domain, error) {

	user, err := d.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	account, err := d.db.GetPremiumAccount(user.Id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, &model.ServiceError{
			InternalError: fmt.Errorf(
				"your account does not have a premium service activated, please contact support",
			),
		}
	}
	validator := NewValidator()
	newDomain := validator.newDomain(request.Password, d.domain)
	if validator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: validator.ToParametersMessages()}
	}

	domain, err := d.db.GetCustomDomainByDomain(*newDomain)
	if err != nil {
		return nil, err
	}
	log.Printf("domain: %v, user: %v\n", domain, user)
	if domain != nil && domain.UserId != user.Id {
		return nil, &model.ParameterError{ParameterErrors: &[]model.ParameterMessages{{
			Parameter: "domain", Messages: []string{"Domain name is already in use"},
		}}}
	}
	if domain.HostedZoneId == nil {
		activeDomains, err := d.db.GetActiveCustomDomains(user.Id)
		if err != nil {
			return nil, err
		}
		if len(activeDomains) > 0 {
			return nil, &model.ParameterError{ParameterErrors: &[]model.ParameterMessages{{
				Parameter: "domain", Messages: []string{"You already have an existing domain"},
			}}}
		}
		hostedZoneId, err := d.amazonDns.CreateHostedZone(domain.Domain)
		domain.HostedZoneId = hostedZoneId
	}

	err = d.amazonDns.UpdateDomain(domain.Domain, domain.Ip, domain.Ipv6, domain.DkimKey, *domain.HostedZoneId)
	if err != nil {
		return nil, err
	}

	updateToken := utils.Uuid()
	log.Println("uuid", updateToken)
	if domain == nil {
		domain = &model.CustomDomain{
			Domain:      *newDomain,
			UpdateToken: &updateToken,
			UserId:      user.Id,
		}
		err := d.db.InsertCustomDomain(domain)
		if err != nil {
			return nil, err
		}
	} else {
		domain.UpdateToken = &updateToken
		err := d.db.UpdateCustomDomain(domain)
		if err != nil {
			return nil, err
		}
	}

	log.Println("custom domain acquired")
	return nil, nil
}

func (d *Domains) Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, error) {
	validator := NewValidator()
	validator.Token(request.Token)
	ip := validator.Ip(request.Ip, requestIp)
	ipv6 := request.Ipv6
	dkimKey := request.DkimKey
	validator.localIp(request.LocalIp)
	mapLocalAddress := request.MapLocalAddress
	platformVersion := request.PlatformVersion
	webProtocol := validator.webProtocol(request.WebProtocol)
	webLocalPort := validator.webLocalPort(request.WebLocalPort)
	webPort := request.WebPort

	if validator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: validator.ToParametersMessages()}
	}

	domain, err := d.db.GetDomainByToken(*request.Token)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("unknown domain update token")}
	}

	user, err := d.db.GetUser(domain.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Active {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("unknown domain update token")}
	}

	changed := Changed(
		domain.MapLocalAddress, domain.Ip, domain.Ipv6, domain.DkimKey, domain.LocalIp,
		mapLocalAddress, ip, ipv6, dkimKey, request.LocalIp)

	domain.Ip = ip
	domain.LocalIp = request.LocalIp
	domain.Ipv6 = ipv6
	domain.DkimKey = dkimKey
	domain.MapLocalAddress = mapLocalAddress
	domain.PlatformVersion = platformVersion
	domain.WebProtocol = webProtocol
	domain.WebLocalPort = webLocalPort
	domain.WebPort = webPort

	if changed {
		fullDomain := domain.DnsName(d.domain)
		ipv4 := domain.DnsIpv4()
		ipv6 := domain.DnsIpv6()
		dkim := domain.DkimKey
		err := d.amazonDns.UpdateDomain(fullDomain, ipv4, ipv6, dkim, d.hostedZoneId)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	domain.LastUpdate = &now
	err = d.db.UpdateDomain(domain)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

func Changed(
	existingMapLocalAddress bool, existingIp *string, existingIpv6 *string, existingDkimKey *string, existingLocalIp *string,
	newMapLocalAddress bool, newIp *string, newIpv6 *string, newDkimKey *string, newLocalIp *string) bool {

	changed := (existingMapLocalAddress != newMapLocalAddress) ||
		!Equals(existingIp, newIp) ||
		!Equals(existingLocalIp, newLocalIp) ||
		!Equals(existingIpv6, newIpv6) ||
		!Equals(existingDkimKey, newDkimKey)

	return changed
}

func Equals(left *string, right *string) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return *left == *right
}
