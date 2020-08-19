package service

import (
	"fmt"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/dns"
	"github.com/syncloud/redirect/model"
	"time"
)

type Service struct {
	amazonDns dns.Dns
	db        *db.MySql
	domain    string
}

func New(dnsImpl dns.Dns, db *db.MySql, domain string) *Service {
	return &Service{amazonDns: dnsImpl, db: db, domain: domain}
}

func (s *Service) GetDomain(token string) (*model.Domain, *model.Error) {
	validator := NewValidator()
	validator.Token(&token)
	if validator.HasErrors() {
		return nil, model.ParametersError(validator.ToParametersMessages())
	}
	domain, err := s.db.SelectDomainByToken(token)
	if err != nil {
		return nil, model.UnknownError(err)
	}
	if domain == nil {
		return nil, model.ServiceError(fmt.Errorf("unknown domain update token"))
	}
	user, err := s.db.GetUser(domain.UserId)
	if err != nil {
		return nil, model.UnknownError(err)
	}
	if !user.Active {
		return nil, model.ServiceError(fmt.Errorf("unknown domain update token"))
	}

	return domain, nil
}
func (s *Service) Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, *model.Error) {
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
		return nil, model.ParametersError(validator.ToParametersMessages())
	}

	domain, err := s.db.SelectDomainByToken(*request.Token)
	if err != nil {
		return nil, model.UnknownError(err)
	}
	if domain == nil {
		return nil, model.ServiceError(fmt.Errorf("unknown domain update token"))
	}

	user, err := s.db.GetUser(domain.UserId)
	if err != nil {
		return nil, model.UnknownError(err)
	}
	if !user.Active {
		return nil, model.ServiceError(fmt.Errorf("unknown domain update token"))
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
		err := s.amazonDns.UpdateDomain(s.domain, domain)
		if err != nil {
			return nil, model.UnknownError(err)
		}
	}

	now := time.Now()
	domain.LastUpdate = &now
	err = s.db.UpdateDomain(domain)
	if err != nil {
		return nil, model.UnknownError(err)
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
