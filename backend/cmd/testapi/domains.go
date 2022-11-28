package main

import "github.com/syncloud/redirect/model"

type TestDomains struct {
}

func (d *TestDomains) DomainAcquire(request model.DomainAcquireRequest, domainField string) (*model.Domain, error) {
	return &model.Domain{Name: *request.Domain}, nil
}

func (d *TestDomains) Availability(request model.DomainAvailabilityRequest) (*model.Domain, error) {
	return &model.Domain{}, nil
}

func (d *TestDomains) Update(request model.DomainUpdateRequest, requestIp *string) (*model.Domain, error) {
	return &model.Domain{}, nil
}

func (d *TestDomains) GetDomain(token string) (*model.Domain, error) {
	return &model.Domain{}, nil
}

func (d *TestDomains) GetDomains(user *model.User) ([]*model.Domain, error) {
	return []*model.Domain{}, nil
}
