package service

import (
	"fmt"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/utils"
	"github.com/syncloud/redirect/validator"
	"log"
	"time"
)

type Domains struct {
	amazonDns dns.Dns
	db        DomainsDb
	users     DomainsUsers
	domain    string
}

type DomainsDb interface {
	GetDomainByToken(token string) (*model.Domain, error)
	GetUserDomains(userId int64) ([]*model.Domain, error)
	GetUser(id int64) (*model.User, error)
	DeleteAllDomains(userId int64) error
	GetDomainByName(name string) (*model.Domain, error)
	InsertDomain(domain *model.Domain) error
	UpdateDomain(domain *model.Domain) error
	DeleteDomain(domainId uint64) error
}

type DomainsUsers interface {
	Authenticate(email *string, password *string) (*model.User, error)
}

func NewDomains(dnsImpl dns.Dns, db DomainsDb, users DomainsUsers, domain string) *Domains {
	return &Domains{amazonDns: dnsImpl, db: db, users: users, domain: domain}
}

func (d *Domains) GetDomain(token string) (*model.Domain, error) {
	fieldValidator := validator.New()
	fieldValidator.Token(&token)
	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
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
	domain.BackwardCompatibleDomain(d.domain)
	return domain, nil
}

func (d *Domains) GetDomains(user *model.User) ([]*model.Domain, error) {
	domains, err := d.db.GetUserDomains(user.Id)
	if err != nil {
		return nil, err
	}
	for _, domain := range domains {
		domain.BackwardCompatibleDomain(d.domain)
	}
	return domains, nil
}

func (d *Domains) DeleteAllDomains(userId int64) error {
	domains, err := d.db.GetUserDomains(userId)
	if err != nil {
		return err
	}

	for _, domain := range domains {
		err = d.amazonDns.DeleteDomain(domain)
		if err != nil {
			return err
		}
	}
	err = d.db.DeleteAllDomains(userId)
	if err != nil {
		return err
	}
	return nil
}

func (d *Domains) DeleteDomain(userId int64, domainName string) error {
	domain, err := d.db.GetDomainByName(domainName)
	if err != nil {
		return err
	}
	if domain == nil || domain.UserId != userId {
		return fmt.Errorf("not found")
	}
	err = d.amazonDns.DeleteDomain(domain)
	if err != nil {
		return err
	}
	return d.db.DeleteDomain(domain.Id)
}

func (d *Domains) Availability(request model.DomainAvailabilityRequest) (*model.Domain, error) {
	user, err := d.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	fieldValidator := validator.New()
	domainField := "domain"
	domainName := fieldValidator.Domain(request.Domain, domainField, d.domain)
	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}

	domain, err := d.find(domainName, user, domainField)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (d *Domains) find(domain *string, user *model.User, field string) (*model.Domain, error) {
	foundDomain, err := d.db.GetDomainByName(*domain)
	if err != nil {
		return nil, err
	}
	log.Printf("domain: %v, found: %v, user: %v\n", *domain, foundDomain, user)
	if foundDomain != nil && foundDomain.UserId != user.Id {
		return nil, &model.ParameterError{ParameterErrors: &[]model.ParameterMessages{{
			Parameter: field, Messages: []string{"User domain name is already in use"},
		}}}
	}
	return foundDomain, err
}

func (d *Domains) DomainAcquire(request model.DomainAcquireRequest, domainField string) (*model.Domain, error) {

	user, err := d.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	fieldValidator := validator.New()

	domainName := fieldValidator.Domain(request.Domain, domainField, d.domain)

	deviceMacAddress := fieldValidator.DeviceMacAddress(request.DeviceMacAddress)
	fieldValidator.DeviceName(request.DeviceName)
	fieldValidator.DeviceTitle(request.DeviceTitle)

	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}

	domain, err := d.find(domainName, user, domainField)
	if err != nil {
		return nil, err
	}
	updateToken := utils.Uuid()
	if domain == nil {
		domain = &model.Domain{
			Name:             *domainName,
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
	domain.BackwardCompatibleDomain(d.domain)
	log.Println("domain acquired")
	return domain, nil
}

func (d *Domains) Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, error) {
	fieldValidator := validator.New()
	fieldValidator.Token(request.Token)
	ip := fieldValidator.Ip(request.Ip, requestIp)
	ipv6 := request.Ipv6
	dkimKey := request.DkimKey
	fieldValidator.LocalIp(request.LocalIp)
	mapLocalAddress := request.MapLocalAddress
	platformVersion := request.PlatformVersion
	webProtocol := fieldValidator.WebProtocol(request.WebProtocol)
	webLocalPort := fieldValidator.WebLocalPort(request.WebLocalPort)
	webPort := request.WebPort

	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
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
		err := d.amazonDns.UpdateDomain(domain)
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
	domain.BackwardCompatibleDomain(d.domain)

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
