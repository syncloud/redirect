package service

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/syncloud/redirect/change"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/utils"
	"github.com/syncloud/redirect/validation"
	"log"
	"time"
)

type Domains struct {
	amazonDns        DomainsDns
	db               DomainsDb
	users            DomainsUsers
	domain           string
	freeHostedZoneId string
	detector         change.Detector
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

type DomainsDns interface {
	CreateHostedZone(domain string) (*string, error)
	DeleteHostedZone(hostedZoneId string) error
	UpdateDomainRecords(domain *model.Domain) error
	DeleteDomainRecords(domain *model.Domain) error
	DeleteCertbotRecord(hostedZoneId string, name string) error
	GetHostedZoneNameServers(id string) ([]*string, error)
}

func NewDomains(
	dnsImpl DomainsDns,
	db DomainsDb,
	users DomainsUsers,
	domain string,
	freeHostedZoneId string,
	detector change.Detector,
) *Domains {
	return &Domains{
		amazonDns:        dnsImpl,
		db:               db,
		users:            users,
		domain:           domain,
		freeHostedZoneId: freeHostedZoneId,
		detector:         detector}
}

func (d *Domains) GetDomain(token string) (*model.Domain, error) {
	fieldValidator := validation.New()
	fieldValidator.Token(&token)
	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}
	domain, err := d.db.GetDomainByToken(token)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return nil, model.NewServiceError("unknown domain update token")
	}
	user, err := d.db.GetUser(domain.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Active {
		return nil, model.NewServiceError("unknown domain update token")
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
		if d.freeHostedZoneId != domain.HostedZoneId {
			nameServers, err := d.amazonDns.GetHostedZoneNameServers(domain.HostedZoneId)
			if err != nil {
				return nil, err
			}
			domain.NameServers = nameServers
		}
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
		err = d.deleteDomain(domain)
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
	err = d.deleteDomain(domain)
	if err != nil {
		return err
	}
	return d.db.DeleteDomain(domain.Id)
}

func (d *Domains) deleteDomain(domain *model.Domain) error {
	err := d.amazonDns.DeleteDomainRecords(domain)
	if err != nil {
		return err
	}

	if d.freeHostedZoneId != domain.HostedZoneId {
		err = d.amazonDns.DeleteCertbotRecord(domain.HostedZoneId, fmt.Sprintf("_acme-challenge.%s", domain.FQDN()))
		if err != nil {
			return err
		}
		err = d.amazonDns.DeleteHostedZone(domain.HostedZoneId)
		var aErr awserr.Error
		if errors.As(err, &aErr) {
			switch aErr.Code() {
			case route53.ErrCodeNoSuchHostedZone:
				log.Printf("no such hosted zone: %s, ignoring", domain.HostedZoneId)
				return nil
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Domains) Availability(request model.DomainAvailabilityRequest) (*model.Domain, error) {
	user, err := d.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	fieldValidator := validation.New()
	domainField := "domain"
	fieldValidator.Domain(request.Domain, domainField, d.domain)
	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}

	domain, err := d.findAndCheck(request.Domain, request.IsFree(d.domain), user, domainField)
	if err != nil {
		return nil, err
	}
	return domain, nil
}

func (d *Domains) findAndCheck(domain *string, isFree bool, user *model.User, field string) (*model.Domain, error) {
	foundDomain, err := d.db.GetDomainByName(*domain)
	if err != nil {
		return nil, err
	}
	if foundDomain != nil {
		log.Printf("domain: %s, found: %s, owner: %d, requester: %d\n", *domain, foundDomain.Name, foundDomain.UserId, user.Id)
		if foundDomain.UserId != user.Id {
			return nil, &model.ParameterError{ParameterErrors: &[]model.ParameterMessages{{
				Parameter: field, Messages: []string{"User domain name is already in use"},
			}}}
		}
	} else {
		if !isFree && user.SubscriptionId == nil {
			return nil, fmt.Errorf("non free domain name requires a premium subscription")
		}
	}

	return foundDomain, err
}

func (d *Domains) DomainAcquire(request model.DomainAcquireRequest, domainField string) (*model.Domain, error) {

	user, err := d.users.Authenticate(request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	validator := validation.New()

	validator.Domain(request.Domain, domainField, d.domain)

	deviceMacAddress := validator.DeviceMacAddress(request.DeviceMacAddress)
	validator.DeviceName(request.DeviceName)
	validator.DeviceTitle(request.DeviceTitle)

	if validator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: validator.ToParametersMessages()}
	}

	isFree := request.IsFree(d.domain)
	domain, err := d.findAndCheck(request.Domain, isFree, user, domainField)
	if err != nil {
		return nil, err
	}
	updateToken := utils.Uuid()
	now := time.Now()
	if domain == nil {
		domain = &model.Domain{
			Name:             *request.Domain,
			DeviceMacAddress: deviceMacAddress,
			DeviceName:       request.DeviceName,
			DeviceTitle:      request.DeviceTitle,
			UpdateToken:      &updateToken,
			UserId:           user.Id,
			HostedZoneId:     d.freeHostedZoneId,
			LastUpdate:       &now,
		}
		if !isFree {
			id, err := d.amazonDns.CreateHostedZone(domain.Name)
			if err != nil {
				return nil, err
			}
			domain.HostedZoneId = *id
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
		domain.LastUpdate = &now

		err := d.db.UpdateDomain(domain)
		if err != nil {
			return nil, err
		}

	}
	domain.BackwardCompatibleDomain(d.domain)
	log.Printf("domain acquired %s, new token: %s\n", domain.Name, *domain.UpdateToken)
	return domain, nil
}

func (d *Domains) Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, error) {
	fieldValidator := validation.New()
	fieldValidator.Token(request.Token)
	ip := fieldValidator.Ip(request.Ip, requestIp)
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
		return nil, model.NewServiceError("unknown domain update token")
	}

	user, err := d.db.GetUser(domain.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Active {
		return nil, model.NewServiceError("unknown domain update token")
	}

	var ipv4 *string
	var localIpv4 *string
	if request.Ipv4Enabled {
		ipv4 = ip
		localIpv4 = request.LocalIp
	}

	var ipv6 *string
	if request.Ipv6Enabled {
		ipv6 = request.Ipv6
	}

	changed := d.detector.Changed(
		domain.MapLocalAddress, domain.Ip, domain.Ipv6, domain.DkimKey, domain.LocalIp,
		mapLocalAddress, ipv4, ipv6, dkimKey, localIpv4)

	domain.Ip = ipv4
	domain.LocalIp = localIpv4
	domain.Ipv6 = ipv6
	domain.DkimKey = dkimKey
	domain.MapLocalAddress = mapLocalAddress
	domain.PlatformVersion = platformVersion
	domain.WebProtocol = webProtocol
	domain.WebLocalPort = webLocalPort
	domain.WebPort = webPort

	if changed {
		err := d.amazonDns.UpdateDomainRecords(domain)
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
